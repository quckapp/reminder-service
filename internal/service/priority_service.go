package service

import (
	"context"
	"time"

	"reminder-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PriorityService struct {
	collection *mongo.Collection
}

func NewPriorityService(db *mongo.Database) *PriorityService {
	return &PriorityService{collection: db.Collection("reminders")}
}

func (s *PriorityService) SetPriority(ctx context.Context, reminderID string, priority models.ReminderPriority) error {
	objID, err := objectIDFromHex(reminderID)
	if err != nil {
		return err
	}
	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{"priority": priority, "updated_at": time.Now()},
	})
	return err
}

func (s *PriorityService) ListByPriority(ctx context.Context, userID string, priority models.ReminderPriority) ([]*models.Reminder, error) {
	filter := bson.M{"user_id": userID, "priority": priority}
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

func (s *PriorityService) GetDistribution(ctx context.Context, userID string) (*models.PriorityDistribution, error) {
	dist := &models.PriorityDistribution{}
	for _, p := range []struct {
		priority models.ReminderPriority
		field    *int64
	}{
		{models.PriorityLow, &dist.Low},
		{models.PriorityMedium, &dist.Medium},
		{models.PriorityHigh, &dist.High},
		{models.PriorityUrgent, &dist.Urgent},
	} {
		filter := bson.M{"user_id": userID, "priority": p.priority}
		count, _ := s.collection.CountDocuments(ctx, filter)
		*p.field = count
	}
	return dist, nil
}
