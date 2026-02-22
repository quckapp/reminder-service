package interfaces

import (
	"context"
)

// ReminderCanceller defines the interface for canceling reminders
// Used by kafka consumer to avoid import cycle
type ReminderCanceller interface {
	Cancel(ctx context.Context, id string) error
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Publish(topic string, message interface{}) error
}
