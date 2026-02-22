package service

import (
	"context"
	"fmt"
	"time"

	"reminder-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TemplateService struct {
	collection *mongo.Collection
}

func NewTemplateService(db *mongo.Database) *TemplateService {
	return &TemplateService{collection: db.Collection("reminder_templates")}
}

func (s *TemplateService) Create(ctx context.Context, userID, workspaceID string, req *models.CreateTemplateRequest) (*models.ReminderTemplate, error) {
	tmpl := &models.ReminderTemplate{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Recurrence:  req.Recurrence,
		Metadata:    req.Metadata,
		IsShared:    req.IsShared,
		UsageCount:  0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, tmpl)
	if err != nil {
		return nil, err
	}
	tmpl.ID = result.InsertedID.(primitive.ObjectID)
	return tmpl, nil
}

func (s *TemplateService) GetByID(ctx context.Context, id string) (*models.ReminderTemplate, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var tmpl models.ReminderTemplate
	if err := s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&tmpl); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (s *TemplateService) GetByUser(ctx context.Context, userID, workspaceID string) ([]models.ReminderTemplate, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID},
			{"workspace_id": workspaceID, "is_shared": true},
		},
	}

	cursor, err := s.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "usage_count", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []models.ReminderTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func (s *TemplateService) Update(ctx context.Context, id string, req *models.UpdateTemplateRequest) (*models.ReminderTemplate, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{"updated_at": time.Now()}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.Recurrence != nil {
		update["recurrence"] = req.Recurrence
	}
	if req.Metadata != nil {
		update["metadata"] = req.Metadata
	}
	update["is_shared"] = req.IsShared

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *TemplateService) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (s *TemplateService) IncrementUsage(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"usage_count": 1}})
	return err
}

func (s *TemplateService) GetPopular(ctx context.Context, workspaceID string, limit int64) ([]models.ReminderTemplate, error) {
	filter := bson.M{"workspace_id": workspaceID, "is_shared": true}
	opts := options.Find().SetSort(bson.D{{Key: "usage_count", Value: -1}}).SetLimit(limit)

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []models.ReminderTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

var _ = fmt.Errorf
