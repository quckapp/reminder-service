package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReminderType string

const (
	ReminderTypeMessage ReminderType = "message"
	ReminderTypeTask    ReminderType = "task"
	ReminderTypeCustom  ReminderType = "custom"
)

type ReminderStatus string

const (
	StatusPending   ReminderStatus = "pending"
	StatusTriggered ReminderStatus = "triggered"
	StatusCancelled ReminderStatus = "cancelled"
	StatusSnoozed   ReminderStatus = "snoozed"
)

type Reminder struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	WorkspaceID string             `bson:"workspace_id" json:"workspace_id"`
	ChannelID   string             `bson:"channel_id,omitempty" json:"channel_id,omitempty"`
	MessageID   string             `bson:"message_id,omitempty" json:"message_id,omitempty"`
	Type        ReminderType       `bson:"type" json:"type"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	RemindAt    time.Time          `bson:"remind_at" json:"remind_at"`
	Status      ReminderStatus     `bson:"status" json:"status"`
	Recurrence  *Recurrence        `bson:"recurrence,omitempty" json:"recurrence,omitempty"`
	Metadata    map[string]any     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	TriggeredAt *time.Time         `bson:"triggered_at,omitempty" json:"triggered_at,omitempty"`
}

type Recurrence struct {
	Pattern   string `bson:"pattern" json:"pattern"` // daily, weekly, monthly, yearly
	Interval  int    `bson:"interval" json:"interval"`
	EndDate   *time.Time `bson:"end_date,omitempty" json:"end_date,omitempty"`
	DaysOfWeek []int  `bson:"days_of_week,omitempty" json:"days_of_week,omitempty"`
}

type CreateReminderRequest struct {
	UserID      string         `json:"user_id" binding:"required"`
	WorkspaceID string         `json:"workspace_id" binding:"required"`
	ChannelID   string         `json:"channel_id,omitempty"`
	MessageID   string         `json:"message_id,omitempty"`
	Type        ReminderType   `json:"type" binding:"required"`
	Title       string         `json:"title" binding:"required"`
	Description string         `json:"description,omitempty"`
	RemindAt    time.Time      `json:"remind_at" binding:"required"`
	Recurrence  *Recurrence    `json:"recurrence,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateReminderRequest struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	RemindAt    *time.Time     `json:"remind_at,omitempty"`
	Status      ReminderStatus `json:"status,omitempty"`
	Recurrence  *Recurrence    `json:"recurrence,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type SnoozeRequest struct {
	Duration string `json:"duration" binding:"required"` // e.g., "15m", "1h", "1d"
}
