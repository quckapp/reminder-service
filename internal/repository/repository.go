package repository

import (
	"context"
	"time"

	"reminder-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	Create(ctx context.Context, reminder *models.Reminder) error
	GetByID(ctx context.Context, id string) (*models.Reminder, error)
	GetByUserID(ctx context.Context, userID string, status *models.ReminderStatus) ([]*models.Reminder, error)
	GetByUserIDPaginated(ctx context.Context, userID string, status *models.ReminderStatus, skip, limit int64) ([]*models.Reminder, int64, error)
	GetByChannelID(ctx context.Context, channelID string, skip, limit int64) ([]*models.Reminder, int64, error)
	GetByWorkspaceID(ctx context.Context, workspaceID string, status *models.ReminderStatus, skip, limit int64) ([]*models.Reminder, int64, error)
	GetPendingReminders(ctx context.Context, before time.Time) ([]*models.Reminder, error)
	GetStats(ctx context.Context, userID string) (*models.ReminderStats, error)
	Update(ctx context.Context, id string, update *models.UpdateReminderRequest) error
	UpdateStatus(ctx context.Context, id string, status models.ReminderStatus) error
	BulkUpdateStatus(ctx context.Context, ids []string, status models.ReminderStatus) (int64, error)
	Delete(ctx context.Context, id string) error
	BulkDelete(ctx context.Context, ids []string) (int64, error)
	Count(ctx context.Context, filter bson.M) (int64, error)
	Close() error
}

type MongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(url, dbName string) (*MongoRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection("reminders")

	// Create indexes
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "remind_at", Value: 1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}}},
	}
	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &MongoRepository{
		client:     client,
		collection: collection,
	}, nil
}

func (r *MongoRepository) Create(ctx context.Context, reminder *models.Reminder) error {
	reminder.CreatedAt = time.Now()
	reminder.UpdatedAt = time.Now()
	reminder.Status = models.StatusPending

	result, err := r.collection.InsertOne(ctx, reminder)
	if err != nil {
		return err
	}

	reminder.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var reminder models.Reminder
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&reminder)
	if err != nil {
		return nil, err
	}

	return &reminder, nil
}

func (r *MongoRepository) GetByUserID(ctx context.Context, userID string, status *models.ReminderStatus) ([]*models.Reminder, error) {
	filter := bson.M{"user_id": userID}
	if status != nil {
		filter["status"] = *status
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "remind_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *MongoRepository) GetPendingReminders(ctx context.Context, before time.Time) ([]*models.Reminder, error) {
	filter := bson.M{
		"status":    models.StatusPending,
		"remind_at": bson.M{"$lte": before},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *MongoRepository) Update(ctx context.Context, id string, update *models.UpdateReminderRequest) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updateDoc := bson.M{"$set": bson.M{"updated_at": time.Now()}}
	setDoc := updateDoc["$set"].(bson.M)

	if update.Title != "" {
		setDoc["title"] = update.Title
	}
	if update.Description != "" {
		setDoc["description"] = update.Description
	}
	if update.RemindAt != nil {
		setDoc["remind_at"] = update.RemindAt
	}
	if update.Status != "" {
		setDoc["status"] = update.Status
	}
	if update.Recurrence != nil {
		setDoc["recurrence"] = update.Recurrence
	}
	if update.Metadata != nil {
		setDoc["metadata"] = update.Metadata
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, updateDoc)
	return err
}

func (r *MongoRepository) UpdateStatus(ctx context.Context, id string, status models.ReminderStatus) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == models.StatusTriggered {
		now := time.Now()
		update["$set"].(bson.M)["triggered_at"] = now
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *MongoRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *MongoRepository) GetByUserIDPaginated(ctx context.Context, userID string, status *models.ReminderStatus, skip, limit int64) ([]*models.Reminder, int64, error) {
	filter := bson.M{"user_id": userID}
	if status != nil {
		filter["status"] = *status
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "remind_at", Value: 1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, 0, err
	}

	return reminders, total, nil
}

func (r *MongoRepository) GetByChannelID(ctx context.Context, channelID string, skip, limit int64) ([]*models.Reminder, int64, error) {
	filter := bson.M{"channel_id": channelID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "remind_at", Value: 1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, 0, err
	}

	return reminders, total, nil
}

func (r *MongoRepository) GetByWorkspaceID(ctx context.Context, workspaceID string, status *models.ReminderStatus, skip, limit int64) ([]*models.Reminder, int64, error) {
	filter := bson.M{"workspace_id": workspaceID}
	if status != nil {
		filter["status"] = *status
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "remind_at", Value: 1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reminders []*models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, 0, err
	}

	return reminders, total, nil
}

func (r *MongoRepository) GetStats(ctx context.Context, userID string) (*models.ReminderStats, error) {
	stats := &models.ReminderStats{
		ByStatus: make(map[string]int64),
		ByType:   make(map[string]int64),
	}

	filter := bson.M{"user_id": userID}

	// Total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	// Count by status
	for _, status := range []models.ReminderStatus{models.StatusPending, models.StatusTriggered, models.StatusCompleted, models.StatusCancelled, models.StatusSnoozed} {
		f := bson.M{"user_id": userID, "status": status}
		count, _ := r.collection.CountDocuments(ctx, f)
		stats.ByStatus[string(status)] = count
	}

	// Count by type
	for _, rtype := range []models.ReminderType{models.ReminderTypeMessage, models.ReminderTypeTask, models.ReminderTypeCustom} {
		f := bson.M{"user_id": userID, "type": rtype}
		count, _ := r.collection.CountDocuments(ctx, f)
		stats.ByType[string(rtype)] = count
	}

	// Upcoming (pending, remind_at > now)
	upcoming, _ := r.collection.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"status":  models.StatusPending,
		"remind_at": bson.M{"$gt": time.Now()},
	})
	stats.Upcoming = upcoming

	// Overdue (pending, remind_at < now)
	overdue, _ := r.collection.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"status":  models.StatusPending,
		"remind_at": bson.M{"$lt": time.Now()},
	})
	stats.Overdue = overdue

	return stats, nil
}

func (r *MongoRepository) BulkUpdateStatus(ctx context.Context, ids []string, status models.ReminderStatus) (int64, error) {
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objID)
	}

	if len(objectIDs) == 0 {
		return 0, nil
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func (r *MongoRepository) BulkDelete(ctx context.Context, ids []string) (int64, error) {
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objID)
	}

	if len(objectIDs) == 0 {
		return 0, nil
	}

	result, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (r *MongoRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *MongoRepository) Database() *mongo.Database {
	return r.collection.Database()
}

func (r *MongoRepository) Close() error {
	return r.client.Disconnect(context.Background())
}
