package kafka

import (
	"context"
	"encoding/json"
	"log"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	service  *service.ReminderService
	topics   []string
	ready    chan bool
}

func NewConsumer(brokers []string, groupID string, svc *service.ReminderService) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		service:  svc,
		topics:   []string{"reminders.commands"},
		ready:    make(chan bool),
	}, nil
}

func (c *Consumer) Start() {
	ctx := context.Background()

	for {
		if err := c.consumer.Consume(ctx, c.topics, c); err != nil {
			log.Printf("Error from consumer: %v", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// ConsumerGroupHandler implementation
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		c.handleMessage(message)
		session.MarkMessage(message, "")
	}
	return nil
}

func (c *Consumer) handleMessage(message *sarama.ConsumerMessage) {
	var event map[string]any
	if err := json.Unmarshal(message.Value, &event); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	action, ok := event["action"].(string)
	if !ok {
		return
	}

	ctx := context.Background()

	switch action {
	case "create":
		c.handleCreate(ctx, event)
	case "cancel":
		c.handleCancel(ctx, event)
	case "snooze":
		c.handleSnooze(ctx, event)
	}
}

func (c *Consumer) handleCreate(ctx context.Context, event map[string]any) {
	// Parse and create reminder from event
	log.Printf("Received create reminder command: %v", event)
}

func (c *Consumer) handleCancel(ctx context.Context, event map[string]any) {
	reminderID, ok := event["reminder_id"].(string)
	if !ok {
		return
	}

	if err := c.service.Cancel(ctx, reminderID); err != nil {
		log.Printf("Error canceling reminder: %v", err)
	}
}

func (c *Consumer) handleSnooze(ctx context.Context, event map[string]any) {
	reminderID, ok := event["reminder_id"].(string)
	if !ok {
		return
	}

	var req models.SnoozeRequest
	if duration, ok := event["duration"].(string); ok {
		req.Duration = duration
	}

	log.Printf("Snooze reminder %s for %s", reminderID, req.Duration)
}
