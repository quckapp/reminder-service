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

type NoteService struct {
	collection *mongo.Collection
}

func NewNoteService(db *mongo.Database) *NoteService {
	return &NoteService{collection: db.Collection("reminder_notes")}
}

func (s *NoteService) Create(ctx context.Context, reminderID, userID string, req *models.CreateNoteRequest) (*models.ReminderNote, error) {
	note := &models.ReminderNote{
		ReminderID: reminderID,
		UserID:     userID,
		Content:    req.Content,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, note)
	if err != nil {
		return nil, err
	}
	note.ID = result.InsertedID.(primitive.ObjectID)
	return note, nil
}

func (s *NoteService) GetByReminder(ctx context.Context, reminderID string) ([]models.ReminderNote, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"reminder_id": reminderID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notes []models.ReminderNote
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *NoteService) Update(ctx context.Context, noteID string, req *models.UpdateNoteRequest) (*models.ReminderNote, error) {
	objID, err := primitive.ObjectIDFromHex(noteID)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$set": bson.M{
			"content":    req.Content,
			"updated_at": time.Now(),
		},
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}

	var note models.ReminderNote
	if err := s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&note); err != nil {
		return nil, err
	}
	return &note, nil
}

func (s *NoteService) Delete(ctx context.Context, noteID string) error {
	objID, err := primitive.ObjectIDFromHex(noteID)
	if err != nil {
		return err
	}
	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (s *NoteService) CountByReminder(ctx context.Context, reminderID string) (int64, error) {
	return s.collection.CountDocuments(ctx, bson.M{"reminder_id": reminderID})
}
