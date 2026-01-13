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
	GetPendingReminders(ctx context.Context, before time.Time) ([]*models.Reminder, error)
	Update(ctx context.Context, id string, update *models.UpdateReminderRequest) error
	UpdateStatus(ctx context.Context, id string, status models.ReminderStatus) error
	Delete(ctx context.Context, id string) error
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

func (r *MongoRepository) Close() error {
	return r.client.Disconnect(context.Background())
}
