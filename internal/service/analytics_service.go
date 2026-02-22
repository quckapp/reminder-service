package service

import (
	"context"
	"time"

	"reminder-service/internal/models"
	"reminder-service/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AnalyticsService struct {
	repo       repository.Repository
	collection *mongo.Collection // reminders collection for aggregation
}

func NewAnalyticsService(repo repository.Repository, db *mongo.Database) *AnalyticsService {
	return &AnalyticsService{
		repo:       repo,
		collection: db.Collection("reminders"),
	}
}

func (s *AnalyticsService) GetWorkspaceAnalytics(ctx context.Context, workspaceID string) (*models.WorkspaceReminderAnalytics, error) {
	analytics := &models.WorkspaceReminderAnalytics{
		ByType:   make(map[string]int64),
		ByStatus: make(map[string]int64),
	}

	filter := bson.M{"workspace_id": workspaceID}

	// Total
	total, _ := s.collection.CountDocuments(ctx, filter)
	analytics.TotalReminders = total

	// Active (pending + snoozed)
	active, _ := s.collection.CountDocuments(ctx, bson.M{
		"workspace_id": workspaceID,
		"status":       bson.M{"$in": []string{"pending", "snoozed"}},
	})
	analytics.ActiveReminders = active

	// Completion rate
	completed, _ := s.collection.CountDocuments(ctx, bson.M{
		"workspace_id": workspaceID,
		"status":       "completed",
	})
	if total > 0 {
		analytics.CompletionRate = float64(completed) / float64(total) * 100
	}

	// By status
	for _, status := range []string{"pending", "triggered", "completed", "cancelled", "snoozed"} {
		count, _ := s.collection.CountDocuments(ctx, bson.M{"workspace_id": workspaceID, "status": status})
		analytics.ByStatus[status] = count
	}

	// By type
	for _, rtype := range []string{"message", "task", "custom"} {
		count, _ := s.collection.CountDocuments(ctx, bson.M{"workspace_id": workspaceID, "type": rtype})
		analytics.ByType[rtype] = count
	}

	// Top users (using aggregation)
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"workspace_id": workspaceID}}},
		{{Key: "$group", Value: bson.M{"_id": "$user_id", "count": bson.M{"$sum": 1}}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var result struct {
				UserID string `bson:"_id"`
				Count  int64  `bson:"count"`
			}
			if cursor.Decode(&result) == nil {
				analytics.TopUsers = append(analytics.TopUsers, models.UserActivity{
					UserID: result.UserID,
					Count:  result.Count,
				})
			}
		}
	}

	// Avg per user
	distinctUsers, _ := s.collection.Distinct(ctx, "user_id", filter)
	if len(distinctUsers) > 0 {
		analytics.AvgRemindersPerUser = float64(total) / float64(len(distinctUsers))
	}

	return analytics, nil
}

func (s *AnalyticsService) GetUpcoming(ctx context.Context, userID string, days, limit int) ([]*models.Reminder, error) {
	if days <= 0 {
		days = 7
	}
	if limit <= 0 {
		limit = 20
	}

	endDate := time.Now().Add(time.Duration(days) * 24 * time.Hour)

	filter := bson.M{
		"user_id": userID,
		"status":  "pending",
		"remind_at": bson.M{
			"$gte": time.Now(),
			"$lte": endDate,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}}).SetLimit(int64(limit))
	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}
	return reminders, nil
}

func (s *AnalyticsService) GetOverdue(ctx context.Context, userID string, limit int) ([]*models.Reminder, error) {
	if limit <= 0 {
		limit = 20
	}

	filter := bson.M{
		"user_id":   userID,
		"status":    "pending",
		"remind_at": bson.M{"$lt": time.Now()},
	}

	opts := options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}}).SetLimit(int64(limit))
	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}
	return reminders, nil
}

func (s *AnalyticsService) GetDueToday(ctx context.Context, userID string) ([]*models.Reminder, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"user_id": userID,
		"status":  "pending",
		"remind_at": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}

	cursor, err := s.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}
	return reminders, nil
}
