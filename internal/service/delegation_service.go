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

type DelegationService struct {
	collection *mongo.Collection
}

func NewDelegationService(db *mongo.Database) *DelegationService {
	return &DelegationService{collection: db.Collection("reminder_delegations")}
}

func (s *DelegationService) Delegate(ctx context.Context, reminderID, delegatedBy string, req *models.DelegateRequest) (*models.ReminderDelegation, error) {
	delegation := &models.ReminderDelegation{
		ReminderID:  reminderID,
		DelegatedBy: delegatedBy,
		DelegatedTo: req.DelegatedTo,
		Status:      models.DelegationPending,
		Message:     req.Message,
		CreatedAt:   time.Now(),
	}

	result, err := s.collection.InsertOne(ctx, delegation)
	if err != nil {
		return nil, err
	}
	delegation.ID = result.InsertedID.(primitive.ObjectID)
	return delegation, nil
}

func (s *DelegationService) Accept(ctx context.Context, delegationID string) (*models.ReminderDelegation, error) {
	return s.updateStatus(ctx, delegationID, models.DelegationAccepted)
}

func (s *DelegationService) Reject(ctx context.Context, delegationID string) (*models.ReminderDelegation, error) {
	return s.updateStatus(ctx, delegationID, models.DelegationRejected)
}

func (s *DelegationService) updateStatus(ctx context.Context, delegationID string, status models.DelegationStatus) (*models.ReminderDelegation, error) {
	objID, err := objectIDFromHex(delegationID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID},
		bson.M{"$set": bson.M{"status": status, "responded_at": now}})
	if err != nil {
		return nil, err
	}

	var delegation models.ReminderDelegation
	if err := s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&delegation); err != nil {
		return nil, err
	}
	return &delegation, nil
}

func (s *DelegationService) GetDelegatedReminders(ctx context.Context, userID string) ([]models.ReminderDelegation, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"delegated_to": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var delegations []models.ReminderDelegation
	if err := cursor.All(ctx, &delegations); err != nil {
		return nil, err
	}
	return delegations, nil
}
