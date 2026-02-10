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

type SharingService struct {
	collection *mongo.Collection
}

func NewSharingService(db *mongo.Database) *SharingService {
	return &SharingService{collection: db.Collection("reminder_shares")}
}

func (s *SharingService) Share(ctx context.Context, reminderID, sharedBy string, req *models.ShareReminderRequest) (*models.ReminderShare, error) {
	share := &models.ReminderShare{
		ReminderID: reminderID,
		SharedBy:   sharedBy,
		SharedWith: req.SharedWith,
		Permission: req.Permission,
		CreatedAt:  time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, share)
	if err != nil {
		return nil, err
	}
	share.ID = result.InsertedID.(primitive.ObjectID)
	return share, nil
}

func (s *SharingService) GetSharesByReminder(ctx context.Context, reminderID string) ([]models.ReminderShare, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"reminder_id": reminderID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var shares []models.ReminderShare
	if err := cursor.All(ctx, &shares); err != nil {
		return nil, err
	}
	return shares, nil
}

func (s *SharingService) GetSharedWithUser(ctx context.Context, userID string) ([]models.ReminderShare, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"shared_with": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var shares []models.ReminderShare
	if err := cursor.All(ctx, &shares); err != nil {
		return nil, err
	}
	return shares, nil
}

func (s *SharingService) Unshare(ctx context.Context, reminderID, sharedWith string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"reminder_id": reminderID, "shared_with": sharedWith})
	return err
}

func (s *SharingService) HasAccess(ctx context.Context, reminderID, userID string) (bool, string, error) {
	var share models.ReminderShare
	err := s.collection.FindOne(ctx, bson.M{"reminder_id": reminderID, "shared_with": userID}).Decode(&share)
	if err != nil {
		return false, "", nil
	}
	return true, share.Permission, nil
}
