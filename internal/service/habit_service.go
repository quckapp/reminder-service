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

type HabitService struct {
	habits      *mongo.Collection
	completions *mongo.Collection
}

func NewHabitService(db *mongo.Database) *HabitService {
	return &HabitService{
		habits:      db.Collection("habits"),
		completions: db.Collection("habit_completions"),
	}
}

func (s *HabitService) Create(ctx context.Context, userID, workspaceID string, req *models.CreateHabitRequest) (*models.Habit, error) {
	habit := &models.Habit{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		Frequency:   req.Frequency,
		TargetDays:  req.TargetDays,
		TargetCount: req.TargetCount,
		ReminderTime: req.ReminderTime,
		Timezone:    req.Timezone,
		Status:      models.HabitActive,
		StartDate:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if habit.TargetCount == 0 {
		habit.TargetCount = 1
	}

	result, err := s.habits.InsertOne(ctx, habit)
	if err != nil {
		return nil, err
	}
	habit.ID = result.InsertedID.(primitive.ObjectID)
	return habit, nil
}

func (s *HabitService) GetByID(ctx context.Context, id string) (*models.Habit, error) {
	objID, err := objectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var habit models.Habit
	if err := s.habits.FindOne(ctx, bson.M{"_id": objID}).Decode(&habit); err != nil {
		return nil, err
	}
	return &habit, nil
}

func (s *HabitService) ListByUser(ctx context.Context, userID string, status string) ([]*models.Habit, error) {
	filter := bson.M{"user_id": userID}
	if status != "" {
		filter["status"] = status
	}

	cursor, err := s.habits.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var habits []*models.Habit
	if err := cursor.All(ctx, &habits); err != nil {
		return nil, err
	}
	return habits, nil
}

func (s *HabitService) Update(ctx context.Context, id string, req *models.UpdateHabitRequest) (*models.Habit, error) {
	objID, err := objectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{"updated_at": time.Now()}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.Icon != "" {
		update["icon"] = req.Icon
	}
	if req.Color != "" {
		update["color"] = req.Color
	}
	if req.Frequency != "" {
		update["frequency"] = req.Frequency
	}
	if req.TargetDays != nil {
		update["target_days"] = req.TargetDays
	}
	if req.TargetCount != nil {
		update["target_count"] = *req.TargetCount
	}
	if req.ReminderTime != "" {
		update["reminder_time"] = req.ReminderTime
	}
	if req.Status != "" {
		update["status"] = req.Status
	}

	_, err = s.habits.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	var habit models.Habit
	if err := s.habits.FindOne(ctx, bson.M{"_id": objID}).Decode(&habit); err != nil {
		return nil, err
	}
	return &habit, nil
}

func (s *HabitService) Delete(ctx context.Context, id string) error {
	objID, err := objectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = s.habits.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	// Also delete completions
	_, _ = s.completions.DeleteMany(ctx, bson.M{"habit_id": id})
	return nil
}

func (s *HabitService) Complete(ctx context.Context, habitID, userID string, req *models.HabitCompletionRequest) (*models.HabitCompletion, error) {
	objID, err := objectIDFromHex(habitID)
	if err != nil {
		return nil, err
	}

	count := req.Count
	if count == 0 {
		count = 1
	}

	completion := &models.HabitCompletion{
		HabitID:     habitID,
		UserID:      userID,
		CompletedAt: time.Now(),
		Note:        req.Note,
		Count:       count,
	}

	result, err := s.completions.InsertOne(ctx, completion)
	if err != nil {
		return nil, err
	}
	completion.ID = result.InsertedID.(primitive.ObjectID)

	// Update habit stats
	now := time.Now()
	habit, _ := s.GetByID(ctx, habitID)
	newStreak := 1
	if habit != nil && habit.LastCompletedAt != nil {
		daysSince := int(now.Sub(*habit.LastCompletedAt).Hours() / 24)
		if daysSince <= 1 {
			newStreak = habit.CurrentStreak + 1
		}
	}

	longestUpdate := bson.M{}
	if habit != nil && newStreak > habit.LongestStreak {
		longestUpdate = bson.M{"longest_streak": newStreak}
	}

	updateFields := bson.M{
		"last_completed_at":  now,
		"current_streak":     newStreak,
		"total_completions":  habit.TotalCompletions + count,
		"updated_at":         now,
	}
	for k, v := range longestUpdate {
		updateFields[k] = v
	}

	_, _ = s.habits.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateFields})

	return completion, nil
}

func (s *HabitService) GetCompletions(ctx context.Context, habitID string, limit int64) ([]*models.HabitCompletion, error) {
	cursor, err := s.completions.Find(ctx, bson.M{"habit_id": habitID},
		options.Find().SetSort(bson.D{{Key: "completed_at", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var completions []*models.HabitCompletion
	if err := cursor.All(ctx, &completions); err != nil {
		return nil, err
	}
	return completions, nil
}

func (s *HabitService) GetStats(ctx context.Context, habitID string) (*models.HabitStats, error) {
	habit, err := s.GetByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	weekCount, _ := s.completions.CountDocuments(ctx, bson.M{
		"habit_id":     habitID,
		"completed_at": bson.M{"$gte": startOfWeek},
	})

	monthCount, _ := s.completions.CountDocuments(ctx, bson.M{
		"habit_id":     habitID,
		"completed_at": bson.M{"$gte": startOfMonth},
	})

	totalDays := int(now.Sub(habit.StartDate).Hours()/24) + 1
	if totalDays == 0 {
		totalDays = 1
	}
	rate := float64(habit.TotalCompletions) / float64(totalDays) * 100

	return &models.HabitStats{
		TotalCompletions: habit.TotalCompletions,
		CurrentStreak:    habit.CurrentStreak,
		LongestStreak:    habit.LongestStreak,
		CompletionRate:   rate,
		ThisWeek:         int(weekCount),
		ThisMonth:        int(monthCount),
		LastCompleted:    habit.LastCompletedAt,
	}, nil
}

func (s *HabitService) GetSummary(ctx context.Context, userID string) (*models.HabitSummary, error) {
	habits, err := s.ListByUser(ctx, userID, "")
	if err != nil {
		return nil, err
	}

	summary := &models.HabitSummary{
		TotalHabits: len(habits),
	}

	var totalStreak int
	for _, h := range habits {
		if h.Status == models.HabitActive {
			summary.ActiveHabits++
		}
		totalStreak += h.CurrentStreak
		if h.CurrentStreak > summary.TopStreak {
			summary.TopStreak = h.CurrentStreak
		}
	}

	if summary.ActiveHabits > 0 {
		summary.AvgStreak = float64(totalStreak) / float64(summary.ActiveHabits)
	}

	// Count today's completions
	startOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	todayCount, _ := s.completions.CountDocuments(ctx, bson.M{
		"user_id":      userID,
		"completed_at": bson.M{"$gte": startOfDay},
	})
	summary.TodayDone = int(todayCount)
	summary.TodayTotal = summary.ActiveHabits

	return summary, nil
}

func (s *HabitService) ResetStreak(ctx context.Context, habitID string) error {
	objID, err := objectIDFromHex(habitID)
	if err != nil {
		return err
	}
	_, err = s.habits.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{"current_streak": 0, "updated_at": time.Now()},
	})
	return err
}
