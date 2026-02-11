package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ── Extended Models ──

type ReminderAttachment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	FileName   string             `bson:"file_name" json:"file_name"`
	FileURL    string             `bson:"file_url" json:"file_url"`
	MimeType   string             `bson:"mime_type" json:"mime_type"`
	Size       int64              `bson:"size" json:"size"`
	UploadedBy string             `bson:"uploaded_by" json:"uploaded_by"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type ReminderComment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Content    string             `bson:"content" json:"content"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type ReminderReaction struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Emoji      string             `bson:"emoji" json:"emoji"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type ReminderWatcher struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type ReminderLabel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	Label      string             `bson:"label" json:"label"`
	Color      string             `bson:"color" json:"color"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type ReminderFavorite struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type ReminderDependency struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID        string             `bson:"reminder_id" json:"reminder_id"`
	DependsOnID       string             `bson:"depends_on_id" json:"depends_on_id"`
	Type              string             `bson:"type" json:"type"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
}

type ReminderLocation struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	Name       string             `bson:"name" json:"name"`
	Latitude   float64            `bson:"latitude" json:"latitude"`
	Longitude  float64            `bson:"longitude" json:"longitude"`
	Radius     float64            `bson:"radius" json:"radius"`
	TriggerOn  string             `bson:"trigger_on" json:"trigger_on"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type ReminderSnoozeHistory struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	SnoozedAt  time.Time          `bson:"snoozed_at" json:"snoozed_at"`
	Duration   string             `bson:"duration" json:"duration"`
	NewTime    time.Time          `bson:"new_time" json:"new_time"`
}

type ReminderQuickAction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	WorkspaceID string             `bson:"workspace_id" json:"workspace_id"`
	Name        string             `bson:"name" json:"name"`
	Action      string             `bson:"action" json:"action"`
	Config      bson.M             `bson:"config" json:"config"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// ── Extended2 Service ──

type Extended2Service struct {
	db *mongo.Database
}

func NewExtended2Service(db *mongo.Database) *Extended2Service {
	return &Extended2Service{db: db}
}

func (s *Extended2Service) col(name string) *mongo.Collection {
	return s.db.Collection(name)
}

// Attachments
func (s *Extended2Service) AddAttachment(ctx context.Context, att *ReminderAttachment) error {
	att.CreatedAt = time.Now()
	_, err := s.col("reminder_attachments").InsertOne(ctx, att)
	return err
}

func (s *Extended2Service) ListAttachments(ctx context.Context, reminderID string) ([]ReminderAttachment, error) {
	cursor, err := s.col("reminder_attachments").Find(ctx, bson.M{"reminder_id": reminderID})
	if err != nil { return nil, err }
	var results []ReminderAttachment
	err = cursor.All(ctx, &results)
	return results, err
}

func (s *Extended2Service) DeleteAttachment(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := s.col("reminder_attachments").DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// Comments
func (s *Extended2Service) AddComment(ctx context.Context, c *ReminderComment) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	_, err := s.col("reminder_comments").InsertOne(ctx, c)
	return err
}

func (s *Extended2Service) ListComments(ctx context.Context, reminderID string, limit, offset int) ([]ReminderComment, error) {
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.col("reminder_comments").Find(ctx, bson.M{"reminder_id": reminderID}, opts)
	if err != nil { return nil, err }
	var results []ReminderComment
	err = cursor.All(ctx, &results)
	return results, err
}

func (s *Extended2Service) UpdateComment(ctx context.Context, id, content string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := s.col("reminder_comments").UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"content": content, "updated_at": time.Now()}})
	return err
}

func (s *Extended2Service) DeleteComment(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := s.col("reminder_comments").DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// Reactions
func (s *Extended2Service) AddReaction(ctx context.Context, r *ReminderReaction) error {
	r.CreatedAt = time.Now()
	_, err := s.col("reminder_reactions").InsertOne(ctx, r)
	return err
}

func (s *Extended2Service) RemoveReaction(ctx context.Context, reminderID, userID, emoji string) error {
	_, err := s.col("reminder_reactions").DeleteOne(ctx, bson.M{"reminder_id": reminderID, "user_id": userID, "emoji": emoji})
	return err
}

func (s *Extended2Service) ListReactions(ctx context.Context, reminderID string) ([]ReminderReaction, error) {
	cursor, err := s.col("reminder_reactions").Find(ctx, bson.M{"reminder_id": reminderID})
	if err != nil { return nil, err }
	var results []ReminderReaction
	err = cursor.All(ctx, &results)
	return results, err
}

// Watchers
func (s *Extended2Service) AddWatcher(ctx context.Context, w *ReminderWatcher) error {
	w.CreatedAt = time.Now()
	_, err := s.col("reminder_watchers").InsertOne(ctx, w)
	return err
}

func (s *Extended2Service) RemoveWatcher(ctx context.Context, reminderID, userID string) error {
	_, err := s.col("reminder_watchers").DeleteOne(ctx, bson.M{"reminder_id": reminderID, "user_id": userID})
	return err
}

func (s *Extended2Service) ListWatchers(ctx context.Context, reminderID string) ([]ReminderWatcher, error) {
	cursor, err := s.col("reminder_watchers").Find(ctx, bson.M{"reminder_id": reminderID})
	if err != nil { return nil, err }
	var results []ReminderWatcher
	err = cursor.All(ctx, &results)
	return results, err
}

// Labels
func (s *Extended2Service) AddLabel(ctx context.Context, l *ReminderLabel) error {
	l.CreatedAt = time.Now()
	_, err := s.col("reminder_labels").InsertOne(ctx, l)
	return err
}

func (s *Extended2Service) RemoveLabel(ctx context.Context, reminderID, label string) error {
	_, err := s.col("reminder_labels").DeleteOne(ctx, bson.M{"reminder_id": reminderID, "label": label})
	return err
}

func (s *Extended2Service) ListLabels(ctx context.Context, reminderID string) ([]ReminderLabel, error) {
	cursor, err := s.col("reminder_labels").Find(ctx, bson.M{"reminder_id": reminderID})
	if err != nil { return nil, err }
	var results []ReminderLabel
	err = cursor.All(ctx, &results)
	return results, err
}

func (s *Extended2Service) SearchByLabel(ctx context.Context, label string, limit, offset int) ([]ReminderLabel, error) {
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := s.col("reminder_labels").Find(ctx, bson.M{"label": label}, opts)
	if err != nil { return nil, err }
	var results []ReminderLabel
	err = cursor.All(ctx, &results)
	return results, err
}

// Favorites
func (s *Extended2Service) AddFavorite(ctx context.Context, f *ReminderFavorite) error {
	f.CreatedAt = time.Now()
	_, err := s.col("reminder_favorites").InsertOne(ctx, f)
	return err
}

func (s *Extended2Service) RemoveFavorite(ctx context.Context, reminderID, userID string) error {
	_, err := s.col("reminder_favorites").DeleteOne(ctx, bson.M{"reminder_id": reminderID, "user_id": userID})
	return err
}

func (s *Extended2Service) ListFavorites(ctx context.Context, userID string, limit, offset int) ([]ReminderFavorite, error) {
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.col("reminder_favorites").Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil { return nil, err }
	var results []ReminderFavorite
	err = cursor.All(ctx, &results)
	return results, err
}

func (s *Extended2Service) IsFavorited(ctx context.Context, reminderID, userID string) (bool, error) {
	count, err := s.col("reminder_favorites").CountDocuments(ctx, bson.M{"reminder_id": reminderID, "user_id": userID})
	return count > 0, err
}

// Dependencies
func (s *Extended2Service) AddDependency(ctx context.Context, d *ReminderDependency) error {
	d.CreatedAt = time.Now()
	_, err := s.col("reminder_dependencies").InsertOne(ctx, d)
	return err
}

func (s *Extended2Service) RemoveDependency(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := s.col("reminder_dependencies").DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (s *Extended2Service) ListDependencies(ctx context.Context, reminderID string) ([]ReminderDependency, error) {
	cursor, err := s.col("reminder_dependencies").Find(ctx, bson.M{"reminder_id": reminderID})
	if err != nil { return nil, err }
	var results []ReminderDependency
	err = cursor.All(ctx, &results)
	return results, err
}

// Locations
func (s *Extended2Service) SetLocation(ctx context.Context, loc *ReminderLocation) error {
	loc.CreatedAt = time.Now()
	filter := bson.M{"reminder_id": loc.ReminderID}
	update := bson.M{"$set": loc}
	opts := options.Update().SetUpsert(true)
	_, err := s.col("reminder_locations").UpdateOne(ctx, filter, update, opts)
	return err
}

func (s *Extended2Service) GetLocation(ctx context.Context, reminderID string) (*ReminderLocation, error) {
	var loc ReminderLocation
	err := s.col("reminder_locations").FindOne(ctx, bson.M{"reminder_id": reminderID}).Decode(&loc)
	if err == mongo.ErrNoDocuments { return nil, nil }
	return &loc, err
}

func (s *Extended2Service) RemoveLocation(ctx context.Context, reminderID string) error {
	_, err := s.col("reminder_locations").DeleteOne(ctx, bson.M{"reminder_id": reminderID})
	return err
}

func (s *Extended2Service) ListNearby(ctx context.Context, lat, lon, radius float64) ([]ReminderLocation, error) {
	cursor, err := s.col("reminder_locations").Find(ctx, bson.M{})
	if err != nil { return nil, err }
	var results []ReminderLocation
	err = cursor.All(ctx, &results)
	return results, err
}

// Snooze History
func (s *Extended2Service) ListSnoozeHistory(ctx context.Context, reminderID string) ([]ReminderSnoozeHistory, error) {
	opts := options.Find().SetSort(bson.D{{Key: "snoozed_at", Value: -1}})
	cursor, err := s.col("reminder_snooze_history").Find(ctx, bson.M{"reminder_id": reminderID}, opts)
	if err != nil { return nil, err }
	var results []ReminderSnoozeHistory
	err = cursor.All(ctx, &results)
	return results, err
}

func (s *Extended2Service) RecordSnooze(ctx context.Context, h *ReminderSnoozeHistory) error {
	h.SnoozedAt = time.Now()
	_, err := s.col("reminder_snooze_history").InsertOne(ctx, h)
	return err
}

// Quick Actions
func (s *Extended2Service) CreateQuickAction(ctx context.Context, qa *ReminderQuickAction) error {
	qa.CreatedAt = time.Now()
	qa.IsActive = true
	_, err := s.col("reminder_quick_actions").InsertOne(ctx, qa)
	return err
}

func (s *Extended2Service) ListQuickActions(ctx context.Context, userID string) ([]ReminderQuickAction, error) {
	cursor, err := s.col("reminder_quick_actions").Find(ctx, bson.M{"user_id": userID, "is_active": true})
	if err != nil { return nil, err }
	var results []ReminderQuickAction
	err = cursor.All(ctx, &results)
	return results, err
}

func (s *Extended2Service) DeleteQuickAction(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := s.col("reminder_quick_actions").DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// Stats
func (s *Extended2Service) GetCompletionRate(ctx context.Context, userID string) (bson.M, error) {
	total, _ := s.db.Collection("reminders").CountDocuments(ctx, bson.M{"user_id": userID})
	completed, _ := s.db.Collection("reminders").CountDocuments(ctx, bson.M{"user_id": userID, "status": "completed"})
	rate := float64(0)
	if total > 0 { rate = float64(completed) / float64(total) * 100 }
	return bson.M{"total": total, "completed": completed, "rate": rate}, nil
}

func (s *Extended2Service) GetStreakInfo(ctx context.Context, userID string) (bson.M, error) {
	return bson.M{"current_streak": 0, "longest_streak": 0, "user_id": userID}, nil
}

// Ensure unused imports are used
var _ = uuid.New
