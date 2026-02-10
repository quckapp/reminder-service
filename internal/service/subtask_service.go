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

type SubtaskService struct {
	collection *mongo.Collection
}

func NewSubtaskService(db *mongo.Database) *SubtaskService {
	return &SubtaskService{collection: db.Collection("reminder_subtasks")}
}

func (s *SubtaskService) Create(ctx context.Context, reminderID, userID string, req *models.CreateSubtaskRequest) (*models.ReminderSubtask, error) {
	subtask := &models.ReminderSubtask{
		ReminderID: reminderID,
		UserID:     userID,
		Title:      req.Title,
		Completed:  false,
		SortOrder:  req.SortOrder,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, subtask)
	if err != nil {
		return nil, err
	}
	subtask.ID = result.InsertedID.(primitive.ObjectID)
	return subtask, nil
}

func (s *SubtaskService) List(ctx context.Context, reminderID string) ([]models.ReminderSubtask, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"reminder_id": reminderID},
		options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subtasks []models.ReminderSubtask
	if err := cursor.All(ctx, &subtasks); err != nil {
		return nil, err
	}
	return subtasks, nil
}

func (s *SubtaskService) Update(ctx context.Context, subtaskID string, req *models.UpdateSubtaskRequest) (*models.ReminderSubtask, error) {
	objID, err := objectIDFromHex(subtaskID)
	if err != nil {
		return nil, err
	}

	update := bson.M{"updated_at": time.Now()}
	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.SortOrder != nil {
		update["sort_order"] = *req.SortOrder
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	var subtask models.ReminderSubtask
	if err := s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&subtask); err != nil {
		return nil, err
	}
	return &subtask, nil
}

func (s *SubtaskService) Delete(ctx context.Context, subtaskID string) error {
	objID, err := objectIDFromHex(subtaskID)
	if err != nil {
		return err
	}
	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (s *SubtaskService) ToggleComplete(ctx context.Context, subtaskID string) (*models.ReminderSubtask, error) {
	objID, err := objectIDFromHex(subtaskID)
	if err != nil {
		return nil, err
	}

	var subtask models.ReminderSubtask
	if err := s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&subtask); err != nil {
		return nil, err
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{"completed": !subtask.Completed, "updated_at": time.Now()},
	})
	if err != nil {
		return nil, err
	}

	subtask.Completed = !subtask.Completed
	return &subtask, nil
}
