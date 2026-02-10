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

type TagService struct {
	collection *mongo.Collection
	tagMap     *mongo.Collection // reminder_tag_mappings
}

func NewTagService(db *mongo.Database) *TagService {
	return &TagService{
		collection: db.Collection("reminder_tags"),
		tagMap:     db.Collection("reminder_tag_mappings"),
	}
}

func (s *TagService) Create(ctx context.Context, userID, workspaceID string, req *models.CreateTagRequest) (*models.ReminderTag, error) {
	tag := &models.ReminderTag{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Color:       req.Color,
		CreatedAt:   time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, tag)
	if err != nil {
		return nil, err
	}
	tag.ID = result.InsertedID.(primitive.ObjectID)
	return tag, nil
}

func (s *TagService) GetByID(ctx context.Context, id string) (*models.ReminderTag, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var tag models.ReminderTag
	if err := s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *TagService) GetByUser(ctx context.Context, userID, workspaceID string) ([]models.ReminderTag, error) {
	filter := bson.M{"user_id": userID}
	if workspaceID != "" {
		filter["workspace_id"] = workspaceID
	}

	cursor, err := s.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tags []models.ReminderTag
	if err := cursor.All(ctx, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (s *TagService) Update(ctx context.Context, id string, req *models.UpdateTagRequest) (*models.ReminderTag, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Color != "" {
		update["color"] = req.Color
	}

	if len(update) == 0 {
		return s.GetByID(ctx, id)
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *TagService) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	// Clean up mappings
	_, _ = s.tagMap.DeleteMany(ctx, bson.M{"tag_id": id})
	return nil
}

func (s *TagService) TagReminder(ctx context.Context, reminderID string, tagIDs []string) error {
	for _, tagID := range tagIDs {
		doc := bson.M{
			"reminder_id": reminderID,
			"tag_id":      tagID,
			"created_at":  time.Now(),
		}
		_, _ = s.tagMap.InsertOne(ctx, doc)
	}
	return nil
}

func (s *TagService) UntagReminder(ctx context.Context, reminderID, tagID string) error {
	_, err := s.tagMap.DeleteOne(ctx, bson.M{"reminder_id": reminderID, "tag_id": tagID})
	return err
}

func (s *TagService) GetReminderTags(ctx context.Context, reminderID string) ([]models.ReminderTag, error) {
	cursor, err := s.tagMap.Find(ctx, bson.M{"reminder_id": reminderID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []struct {
		TagID string `bson:"tag_id"`
	}
	if err := cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}

	var tags []models.ReminderTag
	for _, m := range mappings {
		tag, err := s.GetByID(ctx, m.TagID)
		if err == nil {
			tags = append(tags, *tag)
		}
	}
	return tags, nil
}

func (s *TagService) GetRemindersByTag(ctx context.Context, tagID string) ([]string, error) {
	cursor, err := s.tagMap.Find(ctx, bson.M{"tag_id": tagID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []struct {
		ReminderID string `bson:"reminder_id"`
	}
	if err := cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}

	var ids []string
	for _, m := range mappings {
		ids = append(ids, m.ReminderID)
	}
	return ids, nil
}

func (s *TagService) BulkTag(ctx context.Context, reminderIDs, tagIDs []string) error {
	for _, rid := range reminderIDs {
		for _, tid := range tagIDs {
			doc := bson.M{
				"reminder_id": rid,
				"tag_id":      tid,
				"created_at":  time.Now(),
			}
			_, _ = s.tagMap.InsertOne(ctx, doc)
		}
	}
	return nil
}

// Ensure there's no duplicate import error
var _ = fmt.Errorf
