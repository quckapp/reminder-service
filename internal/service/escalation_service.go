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

type EscalationService struct {
	ruleCollection  *mongo.Collection
	eventCollection *mongo.Collection
}

func NewEscalationService(db *mongo.Database) *EscalationService {
	return &EscalationService{
		ruleCollection:  db.Collection("reminder_escalation_rules"),
		eventCollection: db.Collection("reminder_escalation_events"),
	}
}

func (s *EscalationService) CreateRule(ctx context.Context, userID, workspaceID string, req *models.CreateEscalationRequest) (*models.EscalationRule, error) {
	rule := &models.EscalationRule{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Priority:    req.Priority,
		Enabled:     req.Enabled,
		Actions:     req.Actions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := s.ruleCollection.InsertOne(ctx, rule)
	if err != nil {
		return nil, err
	}
	rule.ID = result.InsertedID.(primitive.ObjectID)
	return rule, nil
}

func (s *EscalationService) ListRules(ctx context.Context, userID, workspaceID string) ([]models.EscalationRule, error) {
	filter := bson.M{"user_id": userID}
	if workspaceID != "" {
		filter["workspace_id"] = workspaceID
	}

	cursor, err := s.ruleCollection.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []models.EscalationRule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *EscalationService) UpdateRule(ctx context.Context, ruleID string, req *models.UpdateEscalationRequest) (*models.EscalationRule, error) {
	objID, err := objectIDFromHex(ruleID)
	if err != nil {
		return nil, err
	}

	update := bson.M{"updated_at": time.Now()}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Priority != "" {
		update["priority"] = req.Priority
	}
	if req.Enabled != nil {
		update["enabled"] = *req.Enabled
	}
	if req.Actions != nil {
		update["actions"] = req.Actions
	}

	_, err = s.ruleCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	var rule models.EscalationRule
	if err := s.ruleCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *EscalationService) DeleteRule(ctx context.Context, ruleID string) error {
	objID, err := objectIDFromHex(ruleID)
	if err != nil {
		return err
	}
	_, err = s.ruleCollection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (s *EscalationService) GetHistory(ctx context.Context, reminderID string) ([]models.EscalationEvent, error) {
	cursor, err := s.eventCollection.Find(ctx, bson.M{"reminder_id": reminderID},
		options.Find().SetSort(bson.D{{Key: "executed_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.EscalationEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}
