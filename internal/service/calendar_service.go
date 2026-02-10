package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"reminder-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CalendarService struct {
	syncCollection *mongo.Collection
	remCollection  *mongo.Collection
}

func NewCalendarService(db *mongo.Database) *CalendarService {
	return &CalendarService{
		syncCollection: db.Collection("reminder_calendar_syncs"),
		remCollection:  db.Collection("reminders"),
	}
}

func (s *CalendarService) ExportICal(ctx context.Context, userID string) (string, error) {
	cursor, err := s.remCollection.Find(ctx, bson.M{"user_id": userID, "status": "pending"},
		options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}}))
	if err != nil {
		return "", err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return "", err
	}

	ical := "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//QuckApp//Reminder//EN\n"
	for _, r := range reminders {
		ical += fmt.Sprintf("BEGIN:VEVENT\nUID:%s\nDTSTART:%s\nSUMMARY:%s\nDESCRIPTION:%s\nEND:VEVENT\n",
			r.ID.Hex(), r.RemindAt.Format("20060102T150405Z"), r.Title, r.Description)
	}
	ical += "END:VCALENDAR"
	return ical, nil
}

func (s *CalendarService) GetFeedURL(ctx context.Context, userID, workspaceID string, req *models.CalendarFeedRequest) (*models.CalendarSync, error) {
	// Check if feed already exists
	var existing models.CalendarSync
	err := s.syncCollection.FindOne(ctx, bson.M{"user_id": userID, "provider": req.Provider}).Decode(&existing)
	if err == nil {
		return &existing, nil
	}

	// Generate feed token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	sync := &models.CalendarSync{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Provider:    req.Provider,
		FeedToken:   token,
		FeedURL:     fmt.Sprintf("/api/v1/calendar/feed/%s", token),
		SyncEnabled: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := s.syncCollection.InsertOne(ctx, sync)
	if err != nil {
		return nil, err
	}
	sync.ID = result.InsertedID.(primitive.ObjectID)
	return sync, nil
}

func (s *CalendarService) SyncCalendar(ctx context.Context, userID string) error {
	now := time.Now()
	_, err := s.syncCollection.UpdateMany(ctx,
		bson.M{"user_id": userID, "sync_enabled": true},
		bson.M{"$set": bson.M{"last_synced_at": now, "updated_at": now}},
	)
	return err
}

func (s *CalendarService) GetCalendarView(ctx context.Context, userID string, start, end time.Time) ([]models.CalendarEvent, error) {
	filter := bson.M{
		"user_id": userID,
		"remind_at": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	cursor, err := s.remCollection.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}

	events := make([]models.CalendarEvent, 0, len(reminders))
	for _, r := range reminders {
		events = append(events, models.CalendarEvent{
			ID:          r.ID.Hex(),
			Title:       r.Title,
			Description: r.Description,
			StartTime:   r.RemindAt,
			EndTime:     r.RemindAt.Add(30 * time.Minute),
			ReminderID:  r.ID.Hex(),
			Status:      string(r.Status),
		})
	}
	return events, nil
}
