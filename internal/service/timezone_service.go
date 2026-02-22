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

type TimezoneService struct {
	collection *mongo.Collection
}

func NewTimezoneService(db *mongo.Database) *TimezoneService {
	return &TimezoneService{collection: db.Collection("user_timezones")}
}

func (s *TimezoneService) GetUserTimezone(ctx context.Context, userID string) (*models.UserTimezone, error) {
	var tz models.UserTimezone
	err := s.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&tz)
	if err == mongo.ErrNoDocuments {
		return &models.UserTimezone{
			UserID:     userID,
			Timezone:   "UTC",
			UTCOffset:  0,
			AutoDetect: true,
			DSTEnabled: true,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &tz, nil
}

func (s *TimezoneService) SetUserTimezone(ctx context.Context, userID string, req *models.UpdateTimezoneRequest) (*models.UserTimezone, error) {
	loc, err := time.LoadLocation(req.Timezone)
	if err != nil {
		return nil, err
	}

	_, offset := time.Now().In(loc).Zone()
	offsetHours := offset / 3600

	tz := &models.UserTimezone{
		UserID:     userID,
		Timezone:   req.Timezone,
		UTCOffset:  offsetHours,
		AutoDetect: true,
		DSTEnabled: true,
		UpdatedAt:  time.Now(),
	}

	if req.AutoDetect != nil {
		tz.AutoDetect = *req.AutoDetect
	}
	if req.DSTEnabled != nil {
		tz.DSTEnabled = *req.DSTEnabled
	}

	opts := options.Update().SetUpsert(true)
	_, err = s.collection.UpdateOne(ctx, bson.M{"user_id": userID}, bson.M{
		"$set": bson.M{
			"timezone":    tz.Timezone,
			"utc_offset":  tz.UTCOffset,
			"auto_detect": tz.AutoDetect,
			"dst_enabled": tz.DSTEnabled,
			"updated_at":  time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	var result models.UserTimezone
	if err := s.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&result); err != nil {
		return tz, nil
	}
	return &result, nil
}

func (s *TimezoneService) ConvertTime(ctx context.Context, req *models.TimezoneConvertRequest) (*models.TimezoneConvertResponse, error) {
	fromLoc, err := time.LoadLocation(req.FromTZ)
	if err != nil {
		return nil, err
	}

	toLoc, err := time.LoadLocation(req.ToTZ)
	if err != nil {
		return nil, err
	}

	fromTime := req.Time.In(fromLoc)
	toTime := fromTime.In(toLoc)

	return &models.TimezoneConvertResponse{
		OriginalTime:  fromTime,
		ConvertedTime: toTime,
		FromTimezone:  req.FromTZ,
		ToTimezone:    req.ToTZ,
	}, nil
}

func (s *TimezoneService) GetWorldClock(ctx context.Context, timezones []string) ([]models.WorldClockEntry, error) {
	var entries []models.WorldClockEntry
	for _, tz := range timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			continue
		}
		now := time.Now().In(loc)
		_, offset := now.Zone()
		hours := offset / 3600
		mins := (offset % 3600) / 60

		var offsetStr string
		if mins != 0 {
			offsetStr = time.Now().In(loc).Format("-07:00")
		} else {
			if hours >= 0 {
				offsetStr = "+" + time.Duration(time.Duration(hours)*time.Hour).String()
			} else {
				offsetStr = time.Duration(time.Duration(hours)*time.Hour).String()
			}
		}

		entries = append(entries, models.WorldClockEntry{
			Timezone:    tz,
			CurrentTime: now,
			UTCOffset:   offsetStr,
		})
	}
	return entries, nil
}

func (s *TimezoneService) ListAvailableTimezones(ctx context.Context) []string {
	return []string{
		"UTC",
		"America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles",
		"America/Toronto", "America/Vancouver", "America/Sao_Paulo", "America/Mexico_City",
		"Europe/London", "Europe/Paris", "Europe/Berlin", "Europe/Moscow", "Europe/Istanbul",
		"Asia/Tokyo", "Asia/Shanghai", "Asia/Kolkata", "Asia/Dubai", "Asia/Singapore",
		"Asia/Seoul", "Asia/Hong_Kong", "Asia/Bangkok",
		"Australia/Sydney", "Australia/Melbourne", "Australia/Perth",
		"Pacific/Auckland", "Pacific/Honolulu",
		"Africa/Cairo", "Africa/Lagos", "Africa/Johannesburg",
	}
}

func (s *TimezoneService) GetBulkTimezones(ctx context.Context, userIDs []string) (map[string]*models.UserTimezone, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"user_id": bson.M{"$in": userIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]*models.UserTimezone)
	for cursor.Next(ctx) {
		var tz models.UserTimezone
		if err := cursor.Decode(&tz); err != nil {
			continue
		}
		result[tz.UserID] = &tz
	}

	// Fill defaults for missing users
	for _, uid := range userIDs {
		if _, ok := result[uid]; !ok {
			result[uid] = &models.UserTimezone{
				UserID:   uid,
				Timezone: "UTC",
			}
		}
	}

	return result, nil
}

// ── Helpers ──

func (s *TimezoneService) GetUserLocation(ctx context.Context, userID string) (*time.Location, error) {
	tz, err := s.GetUserTimezone(ctx, userID)
	if err != nil {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(tz.Timezone)
	if err != nil {
		return time.UTC, nil
	}
	return loc, nil
}

// Suppress unused import warning
var _ = primitive.ObjectID{}
