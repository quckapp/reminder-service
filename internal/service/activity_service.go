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

type ActivityService struct {
	collection *mongo.Collection
}

func NewActivityService(db *mongo.Database) *ActivityService {
	return &ActivityService{collection: db.Collection("reminder_activity")}
}

func (s *ActivityService) Log(ctx context.Context, reminderID, userID, action string, details map[string]any) error {
	activity := &models.ReminderActivity{
		ReminderID: reminderID,
		UserID:     userID,
		Action:     action,
		Details:    details,
		CreatedAt:  time.Now(),
	}

	_, err := s.collection.InsertOne(ctx, activity)
	return err
}

func (s *ActivityService) GetByReminder(ctx context.Context, reminderID string, limit int64) ([]models.ReminderActivity, error) {
	if limit <= 0 {
		limit = 50
	}

	cursor, err := s.collection.Find(ctx, bson.M{"reminder_id": reminderID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.ReminderActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (s *ActivityService) GetByUser(ctx context.Context, userID string, limit int64) ([]models.ReminderActivity, error) {
	if limit <= 0 {
		limit = 50
	}

	cursor, err := s.collection.Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.ReminderActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (s *ActivityService) GetRecent(ctx context.Context, workspaceID string, limit int64) ([]models.ReminderActivity, error) {
	if limit <= 0 {
		limit = 50
	}

	// For workspace activity, we need a broader query approach
	cursor, err := s.collection.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.ReminderActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

var _ = primitive.ObjectID{}
