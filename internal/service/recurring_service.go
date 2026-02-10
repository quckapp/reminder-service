package service

import (
	"context"
	"time"

	"reminder-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RecurringService struct {
	patterns    *mongo.Collection
	occurrences *mongo.Collection
}

func NewRecurringService(db *mongo.Database) *RecurringService {
	return &RecurringService{
		patterns:    db.Collection("recurring_patterns"),
		occurrences: db.Collection("recurring_occurrences"),
	}
}

func (s *RecurringService) CreatePattern(ctx context.Context, userID, workspaceID string, req *models.CreateRecurringPatternRequest) (*models.RecurringPattern, error) {
	pattern := &models.RecurringPattern{
		UserID:         userID,
		WorkspaceID:    workspaceID,
		Name:           req.Name,
		Pattern:        req.Pattern,
		Interval:       req.Interval,
		DaysOfWeek:     req.DaysOfWeek,
		DaysOfMonth:    req.DaysOfMonth,
		MonthsOfYear:   req.MonthsOfYear,
		TimeOfDay:      req.TimeOfDay,
		Timezone:       req.Timezone,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		MaxOccurrences: req.MaxOccurrences,
		IsActive:       true,
		ReminderTitle:  req.ReminderTitle,
		ReminderType:   req.ReminderType,
		ReminderDesc:   req.ReminderDesc,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	next := s.calculateNextOccurrence(pattern)
	pattern.NextOccurrence = next

	result, err := s.patterns.InsertOne(ctx, pattern)
	if err != nil {
		return nil, err
	}
	pattern.ID = result.InsertedID.(primitive.ObjectID)
	return pattern, nil
}

func (s *RecurringService) GetPattern(ctx context.Context, id string) (*models.RecurringPattern, error) {
	objID, err := objectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var pattern models.RecurringPattern
	if err := s.patterns.FindOne(ctx, bson.M{"_id": objID}).Decode(&pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}

func (s *RecurringService) ListPatterns(ctx context.Context, userID string) ([]*models.RecurringPattern, error) {
	cursor, err := s.patterns.Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var patterns []*models.RecurringPattern
	if err := cursor.All(ctx, &patterns); err != nil {
		return nil, err
	}
	return patterns, nil
}

func (s *RecurringService) UpdatePattern(ctx context.Context, id string, req *models.UpdateRecurringPatternRequest) (*models.RecurringPattern, error) {
	objID, err := objectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{"updated_at": time.Now()}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Pattern != "" {
		update["pattern"] = req.Pattern
	}
	if req.Interval != nil {
		update["interval"] = *req.Interval
	}
	if req.DaysOfWeek != nil {
		update["days_of_week"] = req.DaysOfWeek
	}
	if req.DaysOfMonth != nil {
		update["days_of_month"] = req.DaysOfMonth
	}
	if req.MonthsOfYear != nil {
		update["months_of_year"] = req.MonthsOfYear
	}
	if req.TimeOfDay != "" {
		update["time_of_day"] = req.TimeOfDay
	}
	if req.Timezone != "" {
		update["timezone"] = req.Timezone
	}
	if req.EndDate != nil {
		update["end_date"] = req.EndDate
	}
	if req.MaxOccurrences != nil {
		update["max_occurrences"] = *req.MaxOccurrences
	}
	if req.IsActive != nil {
		update["is_active"] = *req.IsActive
	}
	if req.ReminderTitle != "" {
		update["reminder_title"] = req.ReminderTitle
	}
	if req.ReminderDesc != "" {
		update["reminder_desc"] = req.ReminderDesc
	}

	_, err = s.patterns.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	var pattern models.RecurringPattern
	if err := s.patterns.FindOne(ctx, bson.M{"_id": objID}).Decode(&pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}

func (s *RecurringService) DeletePattern(ctx context.Context, id string) error {
	objID, err := objectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = s.patterns.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (s *RecurringService) ToggleActive(ctx context.Context, id string) (*models.RecurringPattern, error) {
	pattern, err := s.GetPattern(ctx, id)
	if err != nil {
		return nil, err
	}

	objID, err := objectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	newActive := !pattern.IsActive
	_, err = s.patterns.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{"is_active": newActive, "updated_at": time.Now()},
	})
	if err != nil {
		return nil, err
	}

	pattern.IsActive = newActive
	return pattern, nil
}

func (s *RecurringService) ListOccurrences(ctx context.Context, patternID string, limit int64) ([]*models.RecurringOccurrence, error) {
	cursor, err := s.occurrences.Find(ctx, bson.M{"pattern_id": patternID},
		options.Find().SetSort(bson.D{{Key: "scheduled_at", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var occs []*models.RecurringOccurrence
	if err := cursor.All(ctx, &occs); err != nil {
		return nil, err
	}
	return occs, nil
}

func (s *RecurringService) RecordOccurrence(ctx context.Context, patternID, reminderID string, scheduledAt time.Time, occurrence int) error {
	occ := &models.RecurringOccurrence{
		PatternID:   patternID,
		ReminderID:  reminderID,
		Occurrence:  occurrence,
		ScheduledAt: scheduledAt,
		Status:      "created",
		CreatedAt:   time.Now(),
	}
	_, err := s.occurrences.InsertOne(ctx, occ)
	if err != nil {
		return err
	}

	objID, _ := objectIDFromHex(patternID)
	_, err = s.patterns.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{"last_triggered": time.Now(), "updated_at": time.Now()},
		"$inc": bson.M{"occurrence_count": 1},
	})
	return err
}

func (s *RecurringService) GetActivePatterns(ctx context.Context) ([]*models.RecurringPattern, error) {
	now := time.Now()
	filter := bson.M{
		"is_active": true,
		"$or": []bson.M{
			{"end_date": nil},
			{"end_date": bson.M{"$gt": now}},
		},
	}

	cursor, err := s.patterns.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var patterns []*models.RecurringPattern
	if err := cursor.All(ctx, &patterns); err != nil {
		return nil, err
	}
	return patterns, nil
}

func (s *RecurringService) calculateNextOccurrence(pattern *models.RecurringPattern) *time.Time {
	var base time.Time
	if pattern.LastTriggered != nil {
		base = *pattern.LastTriggered
	} else {
		base = pattern.StartDate
	}

	var next time.Time
	switch pattern.Pattern {
	case "daily":
		next = base.AddDate(0, 0, pattern.Interval)
	case "weekly":
		next = base.AddDate(0, 0, 7*pattern.Interval)
	case "monthly":
		next = base.AddDate(0, pattern.Interval, 0)
	case "yearly":
		next = base.AddDate(pattern.Interval, 0, 0)
	default:
		next = base.AddDate(0, 0, pattern.Interval)
	}

	if pattern.EndDate != nil && next.After(*pattern.EndDate) {
		return nil
	}
	if pattern.MaxOccurrences > 0 && pattern.OccurrenceCount >= pattern.MaxOccurrences {
		return nil
	}

	return &next
}
