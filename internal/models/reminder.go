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
	StatusCompleted ReminderStatus = "completed"
	StatusCancelled ReminderStatus = "cancelled"
	StatusSnoozed   ReminderStatus = "snoozed"
)

type ReminderPriority string

const (
	PriorityLow    ReminderPriority = "low"
	PriorityMedium ReminderPriority = "medium"
	PriorityHigh   ReminderPriority = "high"
	PriorityUrgent ReminderPriority = "urgent"
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
	Priority    ReminderPriority   `bson:"priority,omitempty" json:"priority,omitempty"`
	Recurrence  *Recurrence        `bson:"recurrence,omitempty" json:"recurrence,omitempty"`
	Metadata    map[string]any     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	TriggeredAt *time.Time         `bson:"triggered_at,omitempty" json:"triggered_at,omitempty"`
}

type Recurrence struct {
	Pattern    string     `bson:"pattern" json:"pattern"`
	Interval   int        `bson:"interval" json:"interval"`
	EndDate    *time.Time `bson:"end_date,omitempty" json:"end_date,omitempty"`
	DaysOfWeek []int      `bson:"days_of_week,omitempty" json:"days_of_week,omitempty"`
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
	Duration string `json:"duration" binding:"required"`
}

type PaginationParams struct {
	Page    int `form:"page" json:"page"`
	PerPage int `form:"per_page" json:"per_page"`
}

func (p *PaginationParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 20
	}
}

func (p *PaginationParams) Skip() int64 {
	return int64((p.Page - 1) * p.PerPage)
}

func (p *PaginationParams) Limit() int64 {
	return int64(p.PerPage)
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

type BulkCreateRequest struct {
	Reminders []CreateReminderRequest `json:"reminders" binding:"required"`
}

type BulkActionRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type BulkActionResponse struct {
	Successful int      `json:"successful"`
	Failed     int      `json:"failed"`
	Errors     []string `json:"errors,omitempty"`
}

type BulkSnoozeRequest struct {
	IDs      []string `json:"ids" binding:"required"`
	Duration string   `json:"duration" binding:"required"`
}

type BulkCompleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type BulkTagRequest struct {
	ReminderIDs []string `json:"reminder_ids" binding:"required"`
	TagIDs      []string `json:"tag_ids" binding:"required"`
}

type ReminderStats struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"by_status"`
	ByType   map[string]int64 `json:"by_type"`
	Upcoming int64            `json:"upcoming"`
	Overdue  int64            `json:"overdue"`
}

type ReminderTag struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	WorkspaceID string             `bson:"workspace_id" json:"workspace_id"`
	Name        string             `bson:"name" json:"name"`
	Color       string             `bson:"color,omitempty" json:"color,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color,omitempty"`
}

type UpdateTagRequest struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

type TagReminderRequest struct {
	TagIDs []string `json:"tag_ids" binding:"required"`
}

type ReminderTemplate struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	WorkspaceID string             `bson:"workspace_id" json:"workspace_id"`
	Name        string             `bson:"name" json:"name"`
	Type        ReminderType       `bson:"type" json:"type"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Recurrence  *Recurrence        `bson:"recurrence,omitempty" json:"recurrence,omitempty"`
	Metadata    map[string]any     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	IsShared    bool               `bson:"is_shared" json:"is_shared"`
	UsageCount  int64              `bson:"usage_count" json:"usage_count"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateTemplateRequest struct {
	Name        string         `json:"name" binding:"required"`
	Type        ReminderType   `json:"type" binding:"required"`
	Title       string         `json:"title" binding:"required"`
	Description string         `json:"description,omitempty"`
	Recurrence  *Recurrence    `json:"recurrence,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	IsShared    bool           `json:"is_shared"`
}

type UpdateTemplateRequest struct {
	Name        string         `json:"name,omitempty"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Recurrence  *Recurrence    `json:"recurrence,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	IsShared    bool           `json:"is_shared"`
}

type CreateFromTemplateRequest struct {
	TemplateID  string    `json:"template_id" binding:"required"`
	UserID      string    `json:"user_id" binding:"required"`
	WorkspaceID string    `json:"workspace_id" binding:"required"`
	ChannelID   string    `json:"channel_id,omitempty"`
	RemindAt    time.Time `json:"remind_at" binding:"required"`
}

type NotificationPreference struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	WorkspaceID string             `bson:"workspace_id" json:"workspace_id"`
	Channels    []string           `bson:"channels" json:"channels"`
	QuietStart  string             `bson:"quiet_start,omitempty" json:"quiet_start,omitempty"`
	QuietEnd    string             `bson:"quiet_end,omitempty" json:"quiet_end,omitempty"`
	Timezone    string             `bson:"timezone,omitempty" json:"timezone,omitempty"`
	AdvanceLead int                `bson:"advance_lead" json:"advance_lead"`
	Enabled     bool               `bson:"enabled" json:"enabled"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type UpdateNotificationPrefRequest struct {
	Channels    []string `json:"channels,omitempty"`
	QuietStart  string   `json:"quiet_start,omitempty"`
	QuietEnd    string   `json:"quiet_end,omitempty"`
	Timezone    string   `json:"timezone,omitempty"`
	AdvanceLead *int     `json:"advance_lead,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

// -- Sharing Models --

type ReminderShare struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	SharedBy   string             `bson:"shared_by" json:"shared_by"`
	SharedWith string             `bson:"shared_with" json:"shared_with"`
	Permission string             `bson:"permission" json:"permission"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type ShareReminderRequest struct {
	SharedWith string `json:"shared_with" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

// -- Note Models --

type ReminderNote struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Content    string             `bson:"content" json:"content"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateNoteRequest struct {
	Content string `json:"content" binding:"required"`
}

type UpdateNoteRequest struct {
	Content string `json:"content" binding:"required"`
}

// -- Activity Models --

type ReminderActivity struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Action     string             `bson:"action" json:"action"`
	Details    map[string]any     `bson:"details,omitempty" json:"details,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// -- Export/Import Models --

type ExportRequest struct {
	UserID string `json:"user_id"`
	Format string `json:"format"`
	Status string `json:"status,omitempty"`
}

type ExportResponse struct {
	Data       []*Reminder `json:"data"`
	Format     string      `json:"format"`
	Count      int         `json:"count"`
	ExportedAt time.Time   `json:"exported_at"`
}

type ImportRequest struct {
	Reminders []CreateReminderRequest `json:"reminders" binding:"required"`
}

type ImportResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

// -- Analytics Models --

type WorkspaceReminderAnalytics struct {
	TotalReminders      int64            `json:"total_reminders"`
	ActiveReminders     int64            `json:"active_reminders"`
	CompletionRate      float64          `json:"completion_rate"`
	ByStatus            map[string]int64 `json:"by_status"`
	ByType              map[string]int64 `json:"by_type"`
	TopUsers            []UserActivity   `json:"top_users"`
	AvgRemindersPerUser float64          `json:"avg_reminders_per_user"`
}

type UserActivity struct {
	UserID string `json:"user_id" bson:"_id"`
	Count  int64  `json:"count" bson:"count"`
}

// -- Search Models --

type ReminderSearchParams struct {
	UserID      string `form:"user_id" json:"user_id"`
	WorkspaceID string `form:"workspace_id" json:"workspace_id"`
	Query       string `form:"q" json:"q"`
	Status      string `form:"status" json:"status"`
	Type        string `form:"type" json:"type"`
	DateFrom    string `form:"date_from" json:"date_from"`
	DateTo      string `form:"date_to" json:"date_to"`
	Page        int    `form:"page" json:"page"`
	PerPage     int    `form:"per_page" json:"per_page"`
}

func (p *ReminderSearchParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 20
	}
}

type UpcomingRemindersRequest struct {
	Days  int `form:"days" json:"days"`
	Limit int `form:"limit" json:"limit"`
}

// -- Priority Models --

type SetPriorityRequest struct {
	Priority ReminderPriority `json:"priority" binding:"required"`
}

type PriorityDistribution struct {
	Low    int64 `json:"low"`
	Medium int64 `json:"medium"`
	High   int64 `json:"high"`
	Urgent int64 `json:"urgent"`
}

// -- Subtask Models --

type ReminderSubtask struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Title      string             `bson:"title" json:"title"`
	Completed  bool               `bson:"completed" json:"completed"`
	SortOrder  int                `bson:"sort_order" json:"sort_order"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateSubtaskRequest struct {
	Title     string `json:"title" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

type UpdateSubtaskRequest struct {
	Title     string `json:"title,omitempty"`
	SortOrder *int   `json:"sort_order,omitempty"`
}

// -- Calendar Integration Models --

type CalendarSync struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       string             `bson:"user_id" json:"user_id"`
	WorkspaceID  string             `bson:"workspace_id" json:"workspace_id"`
	Provider     string             `bson:"provider" json:"provider"`
	ExternalID   string             `bson:"external_id,omitempty" json:"external_id,omitempty"`
	FeedURL      string             `bson:"feed_url,omitempty" json:"feed_url,omitempty"`
	FeedToken    string             `bson:"feed_token,omitempty" json:"feed_token,omitempty"`
	SyncEnabled  bool               `bson:"sync_enabled" json:"sync_enabled"`
	LastSyncedAt *time.Time         `bson:"last_synced_at,omitempty" json:"last_synced_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type CalendarEvent struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	ReminderID  string    `json:"reminder_id"`
	Status      string    `json:"status"`
}

type CalendarFeedRequest struct {
	Provider string `json:"provider" binding:"required"`
}

type CalendarViewRequest struct {
	Start string `form:"start" json:"start" binding:"required"`
	End   string `form:"end" json:"end" binding:"required"`
}

// -- Escalation Rule Models --

type EscalationAction struct {
	Type      string `bson:"type" json:"type"`
	Target    string `bson:"target" json:"target"`
	DelayMins int    `bson:"delay_mins" json:"delay_mins"`
	Message   string `bson:"message,omitempty" json:"message,omitempty"`
}

type EscalationRule struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	WorkspaceID string             `bson:"workspace_id" json:"workspace_id"`
	Name        string             `bson:"name" json:"name"`
	Priority    ReminderPriority   `bson:"priority,omitempty" json:"priority,omitempty"`
	Enabled     bool               `bson:"enabled" json:"enabled"`
	Actions     []EscalationAction `bson:"actions" json:"actions"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateEscalationRequest struct {
	Name     string             `json:"name" binding:"required"`
	Priority ReminderPriority   `json:"priority,omitempty"`
	Enabled  bool               `json:"enabled"`
	Actions  []EscalationAction `json:"actions" binding:"required"`
}

type UpdateEscalationRequest struct {
	Name     string             `json:"name,omitempty"`
	Priority ReminderPriority   `json:"priority,omitempty"`
	Enabled  *bool              `json:"enabled,omitempty"`
	Actions  []EscalationAction `json:"actions,omitempty"`
}

type EscalationEvent struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID       string             `bson:"reminder_id" json:"reminder_id"`
	EscalationRuleID string             `bson:"escalation_rule_id" json:"escalation_rule_id"`
	Action           EscalationAction   `bson:"action" json:"action"`
	Status           string             `bson:"status" json:"status"`
	ExecutedAt       time.Time          `bson:"executed_at" json:"executed_at"`
}

// -- Category Models --

type ReminderCategory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	WorkspaceID string             `bson:"workspace_id" json:"workspace_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Color       string             `bson:"color,omitempty" json:"color,omitempty"`
	Icon        string             `bson:"icon,omitempty" json:"icon,omitempty"`
	SortOrder   int                `bson:"sort_order" json:"sort_order"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	SortOrder   *int   `json:"sort_order,omitempty"`
}

// -- Delegation/Assignment Models --

type DelegationStatus string

const (
	DelegationPending  DelegationStatus = "pending"
	DelegationAccepted DelegationStatus = "accepted"
	DelegationRejected DelegationStatus = "rejected"
)

type ReminderDelegation struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReminderID  string             `bson:"reminder_id" json:"reminder_id"`
	DelegatedBy string             `bson:"delegated_by" json:"delegated_by"`
	DelegatedTo string             `bson:"delegated_to" json:"delegated_to"`
	Status      DelegationStatus   `bson:"status" json:"status"`
	Message     string             `bson:"message,omitempty" json:"message,omitempty"`
	RespondedAt *time.Time         `bson:"responded_at,omitempty" json:"responded_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type DelegateRequest struct {
	DelegatedTo string `json:"delegated_to" binding:"required"`
	Message     string `json:"message,omitempty"`
}

// -- Recurring Patterns --

type RecurringPattern struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	WorkspaceID     string             `bson:"workspace_id" json:"workspace_id"`
	Name            string             `bson:"name" json:"name"`
	Pattern         string             `bson:"pattern" json:"pattern"`
	Interval        int                `bson:"interval" json:"interval"`
	DaysOfWeek      []int              `bson:"days_of_week,omitempty" json:"days_of_week,omitempty"`
	DaysOfMonth     []int              `bson:"days_of_month,omitempty" json:"days_of_month,omitempty"`
	MonthsOfYear    []int              `bson:"months_of_year,omitempty" json:"months_of_year,omitempty"`
	TimeOfDay       string             `bson:"time_of_day,omitempty" json:"time_of_day,omitempty"`
	Timezone        string             `bson:"timezone,omitempty" json:"timezone,omitempty"`
	StartDate       time.Time          `bson:"start_date" json:"start_date"`
	EndDate         *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty"`
	MaxOccurrences  int                `bson:"max_occurrences" json:"max_occurrences"`
	OccurrenceCount int                `bson:"occurrence_count" json:"occurrence_count"`
	IsActive        bool               `bson:"is_active" json:"is_active"`
	ReminderTitle   string             `bson:"reminder_title" json:"reminder_title"`
	ReminderType    ReminderType       `bson:"reminder_type" json:"reminder_type"`
	ReminderDesc    string             `bson:"reminder_desc,omitempty" json:"reminder_desc,omitempty"`
	LastTriggered   *time.Time         `bson:"last_triggered,omitempty" json:"last_triggered,omitempty"`
	NextOccurrence  *time.Time         `bson:"next_occurrence,omitempty" json:"next_occurrence,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateRecurringPatternRequest struct {
	Name           string       `json:"name" binding:"required"`
	Pattern        string       `json:"pattern" binding:"required"`
	Interval       int          `json:"interval" binding:"required"`
	DaysOfWeek     []int        `json:"days_of_week,omitempty"`
	DaysOfMonth    []int        `json:"days_of_month,omitempty"`
	MonthsOfYear   []int        `json:"months_of_year,omitempty"`
	TimeOfDay      string       `json:"time_of_day,omitempty"`
	Timezone       string       `json:"timezone,omitempty"`
	StartDate      time.Time    `json:"start_date" binding:"required"`
	EndDate        *time.Time   `json:"end_date,omitempty"`
	MaxOccurrences int          `json:"max_occurrences"`
	ReminderTitle  string       `json:"reminder_title" binding:"required"`
	ReminderType   ReminderType `json:"reminder_type" binding:"required"`
	ReminderDesc   string       `json:"reminder_desc,omitempty"`
}

type UpdateRecurringPatternRequest struct {
	Name           string     `json:"name,omitempty"`
	Pattern        string     `json:"pattern,omitempty"`
	Interval       *int       `json:"interval,omitempty"`
	DaysOfWeek     []int      `json:"days_of_week,omitempty"`
	DaysOfMonth    []int      `json:"days_of_month,omitempty"`
	MonthsOfYear   []int      `json:"months_of_year,omitempty"`
	TimeOfDay      string     `json:"time_of_day,omitempty"`
	Timezone       string     `json:"timezone,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	MaxOccurrences *int       `json:"max_occurrences,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	ReminderTitle  string     `json:"reminder_title,omitempty"`
	ReminderDesc   string     `json:"reminder_desc,omitempty"`
}

type RecurringOccurrence struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PatternID  string             `bson:"pattern_id" json:"pattern_id"`
	ReminderID string             `bson:"reminder_id" json:"reminder_id"`
	Occurrence int                `bson:"occurrence" json:"occurrence"`
	ScheduledAt time.Time         `bson:"scheduled_at" json:"scheduled_at"`
	Status     string             `bson:"status" json:"status"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// -- Timezone Handling --

type UserTimezone struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Timezone    string             `bson:"timezone" json:"timezone"`
	UTCOffset   int                `bson:"utc_offset" json:"utc_offset"`
	AutoDetect  bool               `bson:"auto_detect" json:"auto_detect"`
	DSTEnabled  bool               `bson:"dst_enabled" json:"dst_enabled"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type UpdateTimezoneRequest struct {
	Timezone   string `json:"timezone" binding:"required"`
	AutoDetect *bool  `json:"auto_detect,omitempty"`
	DSTEnabled *bool  `json:"dst_enabled,omitempty"`
}

type TimezoneConvertRequest struct {
	Time     time.Time `json:"time" binding:"required"`
	FromTZ   string    `json:"from_tz" binding:"required"`
	ToTZ     string    `json:"to_tz" binding:"required"`
}

type TimezoneConvertResponse struct {
	OriginalTime  time.Time `json:"original_time"`
	ConvertedTime time.Time `json:"converted_time"`
	FromTimezone  string    `json:"from_timezone"`
	ToTimezone    string    `json:"to_timezone"`
}

type WorldClockEntry struct {
	Timezone    string    `json:"timezone"`
	CurrentTime time.Time `json:"current_time"`
	UTCOffset   string    `json:"utc_offset"`
	Label       string    `json:"label,omitempty"`
}

// -- Habit Tracking --

type HabitStatus string

const (
	HabitActive   HabitStatus = "active"
	HabitPaused   HabitStatus = "paused"
	HabitArchived HabitStatus = "archived"
)

type Habit struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           string             `bson:"user_id" json:"user_id"`
	WorkspaceID      string             `bson:"workspace_id" json:"workspace_id"`
	Name             string             `bson:"name" json:"name"`
	Description      string             `bson:"description,omitempty" json:"description,omitempty"`
	Icon             string             `bson:"icon,omitempty" json:"icon,omitempty"`
	Color            string             `bson:"color,omitempty" json:"color,omitempty"`
	Frequency        string             `bson:"frequency" json:"frequency"`
	TargetDays       []int              `bson:"target_days,omitempty" json:"target_days,omitempty"`
	TargetCount      int                `bson:"target_count" json:"target_count"`
	ReminderTime     string             `bson:"reminder_time,omitempty" json:"reminder_time,omitempty"`
	Timezone         string             `bson:"timezone,omitempty" json:"timezone,omitempty"`
	Status           HabitStatus        `bson:"status" json:"status"`
	CurrentStreak    int                `bson:"current_streak" json:"current_streak"`
	LongestStreak    int                `bson:"longest_streak" json:"longest_streak"`
	TotalCompletions int                `bson:"total_completions" json:"total_completions"`
	StartDate        time.Time          `bson:"start_date" json:"start_date"`
	LastCompletedAt  *time.Time         `bson:"last_completed_at,omitempty" json:"last_completed_at,omitempty"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateHabitRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Color        string `json:"color,omitempty"`
	Frequency    string `json:"frequency" binding:"required"`
	TargetDays   []int  `json:"target_days,omitempty"`
	TargetCount  int    `json:"target_count"`
	ReminderTime string `json:"reminder_time,omitempty"`
	Timezone     string `json:"timezone,omitempty"`
}

type UpdateHabitRequest struct {
	Name         string      `json:"name,omitempty"`
	Description  string      `json:"description,omitempty"`
	Icon         string      `json:"icon,omitempty"`
	Color        string      `json:"color,omitempty"`
	Frequency    string      `json:"frequency,omitempty"`
	TargetDays   []int       `json:"target_days,omitempty"`
	TargetCount  *int        `json:"target_count,omitempty"`
	ReminderTime string      `json:"reminder_time,omitempty"`
	Status       HabitStatus `json:"status,omitempty"`
}

type HabitCompletion struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HabitID     string             `bson:"habit_id" json:"habit_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	CompletedAt time.Time          `bson:"completed_at" json:"completed_at"`
	Note        string             `bson:"note,omitempty" json:"note,omitempty"`
	Count       int                `bson:"count" json:"count"`
}

type HabitCompletionRequest struct {
	Note  string `json:"note,omitempty"`
	Count int    `json:"count"`
}

type HabitStats struct {
	TotalCompletions int       `json:"total_completions"`
	CurrentStreak    int       `json:"current_streak"`
	LongestStreak    int       `json:"longest_streak"`
	CompletionRate   float64   `json:"completion_rate"`
	ThisWeek         int       `json:"this_week"`
	ThisMonth        int       `json:"this_month"`
	LastCompleted    *time.Time `json:"last_completed,omitempty"`
}

type HabitSummary struct {
	TotalHabits  int     `json:"total_habits"`
	ActiveHabits int     `json:"active_habits"`
	AvgStreak    float64 `json:"avg_streak"`
	TopStreak    int     `json:"top_streak"`
	TodayDone    int     `json:"today_done"`
	TodayTotal   int     `json:"today_total"`
}
