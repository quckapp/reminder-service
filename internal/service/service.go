package service

import (
	"context"
	"fmt"
	"time"

	"reminder-service/internal/models"
	"reminder-service/internal/repository"
)

// EventPublisher interface for publishing events
type EventPublisher interface {
	Publish(topic string, message interface{}) error
}

type ReminderService struct {
	repo     repository.Repository
	producer EventPublisher
}

func NewReminderService(repo repository.Repository, producer EventPublisher) *ReminderService {
	return &ReminderService{
		repo:     repo,
		producer: producer,
	}
}

func (s *ReminderService) Create(ctx context.Context, req *models.CreateReminderRequest) (*models.Reminder, error) {
	reminder := &models.Reminder{
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		ChannelID:   req.ChannelID,
		MessageID:   req.MessageID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		RemindAt:    req.RemindAt,
		Recurrence:  req.Recurrence,
		Metadata:    req.Metadata,
	}

	if err := s.repo.Create(ctx, reminder); err != nil {
		return nil, fmt.Errorf("failed to create reminder: %w", err)
	}

	// Publish event
	s.producer.Publish("reminders.created", map[string]any{
		"reminder_id":  reminder.ID.Hex(),
		"user_id":      reminder.UserID,
		"workspace_id": reminder.WorkspaceID,
		"remind_at":    reminder.RemindAt,
	})

	return reminder, nil
}

func (s *ReminderService) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ReminderService) GetByUserID(ctx context.Context, userID string, status *models.ReminderStatus) ([]*models.Reminder, error) {
	return s.repo.GetByUserID(ctx, userID, status)
}

func (s *ReminderService) Update(ctx context.Context, id string, req *models.UpdateReminderRequest) (*models.Reminder, error) {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, fmt.Errorf("failed to update reminder: %w", err)
	}

	reminder, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Publish event
	s.producer.Publish("reminders.updated", map[string]any{
		"reminder_id": id,
		"user_id":     reminder.UserID,
	})

	return reminder, nil
}

func (s *ReminderService) Snooze(ctx context.Context, id string, duration time.Duration) (*models.Reminder, error) {
	reminder, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newRemindAt := time.Now().Add(duration)
	update := &models.UpdateReminderRequest{
		RemindAt: &newRemindAt,
		Status:   models.StatusSnoozed,
	}

	if err := s.repo.Update(ctx, id, update); err != nil {
		return nil, fmt.Errorf("failed to snooze reminder: %w", err)
	}

	// Reset to pending after snooze update
	if err := s.repo.UpdateStatus(ctx, id, models.StatusPending); err != nil {
		return nil, err
	}

	reminder.RemindAt = newRemindAt
	reminder.Status = models.StatusPending

	// Publish event
	s.producer.Publish("reminders.snoozed", map[string]any{
		"reminder_id": id,
		"user_id":     reminder.UserID,
		"new_time":    newRemindAt,
	})

	return reminder, nil
}

func (s *ReminderService) Cancel(ctx context.Context, id string) error {
	reminder, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, models.StatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel reminder: %w", err)
	}

	// Publish event
	s.producer.Publish("reminders.cancelled", map[string]any{
		"reminder_id": id,
		"user_id":     reminder.UserID,
	})

	return nil
}

func (s *ReminderService) Delete(ctx context.Context, id string) error {
	reminder, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	// Publish event
	s.producer.Publish("reminders.deleted", map[string]any{
		"reminder_id": id,
		"user_id":     reminder.UserID,
	})

	return nil
}

func (s *ReminderService) TriggerReminder(ctx context.Context, reminder *models.Reminder) error {
	if err := s.repo.UpdateStatus(ctx, reminder.ID.Hex(), models.StatusTriggered); err != nil {
		return fmt.Errorf("failed to trigger reminder: %w", err)
	}

	// Publish notification event
	s.producer.Publish("notifications.send", map[string]any{
		"type":        "reminder",
		"user_id":     reminder.UserID,
		"title":       reminder.Title,
		"description": reminder.Description,
		"reminder_id": reminder.ID.Hex(),
		"channel_id":  reminder.ChannelID,
		"message_id":  reminder.MessageID,
		"metadata":    reminder.Metadata,
	})

	// Handle recurrence
	if reminder.Recurrence != nil {
		return s.scheduleNextRecurrence(ctx, reminder)
	}

	return nil
}

func (s *ReminderService) scheduleNextRecurrence(ctx context.Context, reminder *models.Reminder) error {
	nextTime := calculateNextOccurrence(reminder.RemindAt, reminder.Recurrence)

	if reminder.Recurrence.EndDate != nil && nextTime.After(*reminder.Recurrence.EndDate) {
		return nil // End of recurrence
	}

	newReminder := &models.Reminder{
		UserID:      reminder.UserID,
		WorkspaceID: reminder.WorkspaceID,
		ChannelID:   reminder.ChannelID,
		MessageID:   reminder.MessageID,
		Type:        reminder.Type,
		Title:       reminder.Title,
		Description: reminder.Description,
		RemindAt:    nextTime,
		Recurrence:  reminder.Recurrence,
		Metadata:    reminder.Metadata,
	}

	return s.repo.Create(ctx, newReminder)
}

func calculateNextOccurrence(current time.Time, recurrence *models.Recurrence) time.Time {
	switch recurrence.Pattern {
	case "daily":
		return current.AddDate(0, 0, recurrence.Interval)
	case "weekly":
		return current.AddDate(0, 0, 7*recurrence.Interval)
	case "monthly":
		return current.AddDate(0, recurrence.Interval, 0)
	case "yearly":
		return current.AddDate(recurrence.Interval, 0, 0)
	default:
		return current.AddDate(0, 0, 1)
	}
}

func (s *ReminderService) GetPendingReminders(ctx context.Context, before time.Time) ([]*models.Reminder, error) {
	return s.repo.GetPendingReminders(ctx, before)
}

// ── Paginated Queries ──

func (s *ReminderService) GetByUserIDPaginated(ctx context.Context, userID string, status *models.ReminderStatus, page *models.PaginationParams) (*models.PaginatedResponse, error) {
	page.Validate()
	reminders, total, err := s.repo.GetByUserIDPaginated(ctx, userID, status, page.Skip(), page.Limit())
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + page.Limit() - 1) / page.Limit())
	}

	return &models.PaginatedResponse{
		Data:       reminders,
		Total:      total,
		Page:       page.Page,
		PerPage:    page.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (s *ReminderService) GetByChannelID(ctx context.Context, channelID string, page *models.PaginationParams) (*models.PaginatedResponse, error) {
	page.Validate()
	reminders, total, err := s.repo.GetByChannelID(ctx, channelID, page.Skip(), page.Limit())
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + page.Limit() - 1) / page.Limit())
	}

	return &models.PaginatedResponse{
		Data:       reminders,
		Total:      total,
		Page:       page.Page,
		PerPage:    page.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (s *ReminderService) GetByWorkspaceID(ctx context.Context, workspaceID string, status *models.ReminderStatus, page *models.PaginationParams) (*models.PaginatedResponse, error) {
	page.Validate()
	reminders, total, err := s.repo.GetByWorkspaceID(ctx, workspaceID, status, page.Skip(), page.Limit())
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + page.Limit() - 1) / page.Limit())
	}

	return &models.PaginatedResponse{
		Data:       reminders,
		Total:      total,
		Page:       page.Page,
		PerPage:    page.PerPage,
		TotalPages: totalPages,
	}, nil
}

// ── Stats ──

func (s *ReminderService) GetStats(ctx context.Context, userID string) (*models.ReminderStats, error) {
	return s.repo.GetStats(ctx, userID)
}

// ── Complete ──

func (s *ReminderService) Complete(ctx context.Context, id string) error {
	reminder, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, models.StatusCompleted); err != nil {
		return fmt.Errorf("failed to complete reminder: %w", err)
	}

	s.producer.Publish("reminders.completed", map[string]any{
		"reminder_id": id,
		"user_id":     reminder.UserID,
	})

	return nil
}

// ── Bulk Operations ──

func (s *ReminderService) BulkCreate(ctx context.Context, reqs []models.CreateReminderRequest) *models.BulkActionResponse {
	resp := &models.BulkActionResponse{}
	for _, req := range reqs {
		reqCopy := req
		_, err := s.Create(ctx, &reqCopy)
		if err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s: %s", req.Title, err.Error()))
		} else {
			resp.Successful++
		}
	}
	return resp
}

func (s *ReminderService) BulkCancel(ctx context.Context, ids []string) *models.BulkActionResponse {
	resp := &models.BulkActionResponse{}
	count, err := s.repo.BulkUpdateStatus(ctx, ids, models.StatusCancelled)
	if err != nil {
		resp.Failed = len(ids)
		resp.Errors = append(resp.Errors, err.Error())
	} else {
		resp.Successful = int(count)
		resp.Failed = len(ids) - int(count)
	}

	s.producer.Publish("reminders.bulk_cancelled", map[string]any{
		"ids":       ids,
		"cancelled": count,
	})

	return resp
}

func (s *ReminderService) BulkDelete(ctx context.Context, ids []string) *models.BulkActionResponse {
	resp := &models.BulkActionResponse{}
	count, err := s.repo.BulkDelete(ctx, ids)
	if err != nil {
		resp.Failed = len(ids)
		resp.Errors = append(resp.Errors, err.Error())
	} else {
		resp.Successful = int(count)
		resp.Failed = len(ids) - int(count)
	}

	s.producer.Publish("reminders.bulk_deleted", map[string]any{
		"ids":     ids,
		"deleted": count,
	})

	return resp
}
