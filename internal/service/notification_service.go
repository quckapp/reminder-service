package service

import (
	"context"
	"time"

	"reminder-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NotificationService struct {
	collection *mongo.Collection
}

func NewNotificationService(db *mongo.Database) *NotificationService {
	return &NotificationService{collection: db.Collection("notification_preferences")}
}

func (s *NotificationService) GetPreferences(ctx context.Context, userID, workspaceID string) (*models.NotificationPreference, error) {
	filter := bson.M{"user_id": userID}
	if workspaceID != "" {
		filter["workspace_id"] = workspaceID
	}

	var pref models.NotificationPreference
	err := s.collection.FindOne(ctx, filter).Decode(&pref)
	if err != nil {
		// Return default preferences
		return &models.NotificationPreference{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Channels:    []string{"in_app"},
			Enabled:     true,
			AdvanceLead: 0,
		}, nil
	}
	return &pref, nil
}

func (s *NotificationService) UpdatePreferences(ctx context.Context, userID, workspaceID string, req *models.UpdateNotificationPrefRequest) (*models.NotificationPreference, error) {
	filter := bson.M{"user_id": userID, "workspace_id": workspaceID}

	update := bson.M{"updated_at": time.Now()}
	if req.Channels != nil {
		update["channels"] = req.Channels
	}
	if req.QuietStart != "" {
		update["quiet_start"] = req.QuietStart
	}
	if req.QuietEnd != "" {
		update["quiet_end"] = req.QuietEnd
	}
	if req.Timezone != "" {
		update["timezone"] = req.Timezone
	}
	if req.AdvanceLead != nil {
		update["advance_lead"] = *req.AdvanceLead
	}
	if req.Enabled != nil {
		update["enabled"] = *req.Enabled
	}

	opts := primitive.ObjectID{}
	_ = opts

	// Upsert
	result := s.collection.FindOneAndUpdate(ctx, filter,
		bson.M{
			"$set": update,
			"$setOnInsert": bson.M{
				"user_id":      userID,
				"workspace_id": workspaceID,
				"created_at":   time.Now(),
			},
		},
	)

	if result.Err() != nil {
		// Insert new
		pref := &models.NotificationPreference{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Channels:    req.Channels,
			QuietStart:  req.QuietStart,
			QuietEnd:    req.QuietEnd,
			Timezone:    req.Timezone,
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if req.AdvanceLead != nil {
			pref.AdvanceLead = *req.AdvanceLead
		}
		if req.Enabled != nil {
			pref.Enabled = *req.Enabled
		}
		insertResult, err := s.collection.InsertOne(ctx, pref)
		if err != nil {
			return nil, err
		}
		pref.ID = insertResult.InsertedID.(primitive.ObjectID)
		return pref, nil
	}

	return s.GetPreferences(ctx, userID, workspaceID)
}

func (s *NotificationService) DeletePreferences(ctx context.Context, userID, workspaceID string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"user_id": userID, "workspace_id": workspaceID})
	return err
}
