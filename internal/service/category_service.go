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

type CategoryService struct {
	collection *mongo.Collection
}

func NewCategoryService(db *mongo.Database) *CategoryService {
	return &CategoryService{collection: db.Collection("reminder_categories")}
}

func (s *CategoryService) Create(ctx context.Context, userID, workspaceID string, req *models.CreateCategoryRequest) (*models.ReminderCategory, error) {
	cat := &models.ReminderCategory{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, cat)
	if err != nil {
		return nil, err
	}
	cat.ID = result.InsertedID.(primitive.ObjectID)
	return cat, nil
}

func (s *CategoryService) List(ctx context.Context, userID, workspaceID string) ([]models.ReminderCategory, error) {
	filter := bson.M{"user_id": userID}
	if workspaceID != "" {
		filter["workspace_id"] = workspaceID
	}

	cursor, err := s.collection.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.ReminderCategory
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *CategoryService) Update(ctx context.Context, categoryID string, req *models.UpdateCategoryRequest) (*models.ReminderCategory, error) {
	objID, err := objectIDFromHex(categoryID)
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
	if req.Color != "" {
		update["color"] = req.Color
	}
	if req.Icon != "" {
		update["icon"] = req.Icon
	}
	if req.SortOrder != nil {
		update["sort_order"] = *req.SortOrder
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	var cat models.ReminderCategory
	if err := s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&cat); err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *CategoryService) Delete(ctx context.Context, categoryID string) error {
	objID, err := objectIDFromHex(categoryID)
	if err != nil {
		return err
	}
	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
