package service

import (
	"context"
	"fmt"
	"time"

	"reminder-service/internal/models"
	"reminder-service/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExportService struct {
	repo       repository.Repository
	collection *mongo.Collection
}

func NewExportService(repo repository.Repository, db *mongo.Database) *ExportService {
	return &ExportService{
		repo:       repo,
		collection: db.Collection("reminders"),
	}
}

func (s *ExportService) Export(ctx context.Context, req *models.ExportRequest) (*models.ExportResponse, error) {
	filter := bson.M{"user_id": req.UserID}
	if req.Status != "" {
		filter["status"] = req.Status
	}

	cursor, err := s.collection.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}

	return &models.ExportResponse{
		Data:       reminders,
		Format:     req.Format,
		Count:      len(reminders),
		ExportedAt: time.Now(),
	}, nil
}

func (s *ExportService) Import(ctx context.Context, userID string, req *models.ImportRequest) *models.ImportResponse {
	resp := &models.ImportResponse{}

	for _, reminderReq := range req.Reminders {
		reminder := &models.Reminder{
			UserID:      reminderReq.UserID,
			WorkspaceID: reminderReq.WorkspaceID,
			ChannelID:   reminderReq.ChannelID,
			MessageID:   reminderReq.MessageID,
			Type:        reminderReq.Type,
			Title:       reminderReq.Title,
			Description: reminderReq.Description,
			RemindAt:    reminderReq.RemindAt,
			Recurrence:  reminderReq.Recurrence,
			Metadata:    reminderReq.Metadata,
		}

		if err := s.repo.Create(ctx, reminder); err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s: %s", reminderReq.Title, err.Error()))
		} else {
			resp.Imported++
		}
	}

	return resp
}
