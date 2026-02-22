package service

import (
	"context"
	"time"

	"reminder-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SearchService struct {
	collection *mongo.Collection
}

func NewSearchService(db *mongo.Database) *SearchService {
	return &SearchService{collection: db.Collection("reminders")}
}

func (s *SearchService) Search(ctx context.Context, params *models.ReminderSearchParams) (*models.PaginatedResponse, error) {
	params.Validate()

	filter := bson.M{}

	if params.UserID != "" {
		filter["user_id"] = params.UserID
	}
	if params.WorkspaceID != "" {
		filter["workspace_id"] = params.WorkspaceID
	}
	if params.Status != "" {
		filter["status"] = params.Status
	}
	if params.Type != "" {
		filter["type"] = params.Type
	}
	if params.Query != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": params.Query, "$options": "i"}},
			{"description": bson.M{"$regex": params.Query, "$options": "i"}},
		}
	}

	// Date range
	if params.DateFrom != "" || params.DateTo != "" {
		dateFilter := bson.M{}
		if params.DateFrom != "" {
			if t, err := time.Parse("2006-01-02", params.DateFrom); err == nil {
				dateFilter["$gte"] = t
			}
		}
		if params.DateTo != "" {
			if t, err := time.Parse("2006-01-02", params.DateTo); err == nil {
				dateFilter["$lte"] = t
			}
		}
		if len(dateFilter) > 0 {
			filter["remind_at"] = dateFilter
		}
	}

	// Count
	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	skip := int64((params.Page - 1) * params.PerPage)
	limit := int64(params.PerPage)

	opts := options.Find().
		SetSort(bson.D{{Key: "remind_at", Value: 1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + limit - 1) / limit)
	}

	return &models.PaginatedResponse{
		Data:       reminders,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}
