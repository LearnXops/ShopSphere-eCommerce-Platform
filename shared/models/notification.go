package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NotificationChannel represents the delivery channel for notifications
type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationPending   NotificationStatus = "pending"
	NotificationSent      NotificationStatus = "sent"
	NotificationDelivered NotificationStatus = "delivered"
	NotificationFailed    NotificationStatus = "failed"
	NotificationBounced   NotificationStatus = "bounced"
)

// NotificationEventType represents delivery event types
type NotificationEventType string

const (
	EventSent         NotificationEventType = "sent"
	EventDelivered    NotificationEventType = "delivered"
	EventBounced      NotificationEventType = "bounced"
	EventOpened       NotificationEventType = "opened"
	EventClicked      NotificationEventType = "clicked"
	EventUnsubscribed NotificationEventType = "unsubscribed"
)

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID           string                 `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Channel      NotificationChannel    `json:"channel" db:"channel"`
	Subject      *string                `json:"subject" db:"subject"`
	BodyTemplate string                 `json:"body_template" db:"body_template"`
	Variables    map[string]interface{} `json:"variables" db:"variables"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// Notification represents a notification to be sent
type Notification struct {
	ID                 string                 `json:"id" db:"id"`
	UserID             string                 `json:"user_id" db:"user_id"`
	TemplateID         *string                `json:"template_id" db:"template_id"`
	Channel            NotificationChannel    `json:"channel" db:"channel"`
	Recipient          string                 `json:"recipient" db:"recipient"`
	Subject            *string                `json:"subject" db:"subject"`
	Body               string                 `json:"body" db:"body"`
	Variables          map[string]interface{} `json:"variables" db:"variables"`
	Status             NotificationStatus     `json:"status" db:"status"`
	Provider           *string                `json:"provider" db:"provider"`
	ProviderMessageID  *string                `json:"provider_message_id" db:"provider_message_id"`
	ErrorMessage       *string                `json:"error_message" db:"error_message"`
	RetryCount         int                    `json:"retry_count" db:"retry_count"`
	MaxRetries         int                    `json:"max_retries" db:"max_retries"`
	ScheduledAt        *time.Time             `json:"scheduled_at" db:"scheduled_at"`
	SentAt             *time.Time             `json:"sent_at" db:"sent_at"`
	DeliveredAt        *time.Time             `json:"delivered_at" db:"delivered_at"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// NewNotification creates a new notification with default values
func NewNotification(userID string, channel NotificationChannel, recipient, body string) *Notification {
	var subject *string
	return &Notification{
		ID:         uuid.New().String(),
		UserID:     userID,
		Channel:    channel,
		Recipient:  recipient,
		Subject:    subject,
		Body:       body,
		Status:     NotificationPending,
		Variables:  make(map[string]interface{}),
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID        string              `json:"id" db:"id"`
	UserID    string              `json:"user_id" db:"user_id"`
	Channel   NotificationChannel `json:"channel" db:"channel"`
	Category  string              `json:"category" db:"category"`
	IsEnabled bool                `json:"is_enabled" db:"is_enabled"`
	CreatedAt time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" db:"updated_at"`
}

// NotificationDeliveryEvent represents a delivery event
type NotificationDeliveryEvent struct {
	ID                string                `json:"id" db:"id"`
	NotificationID    string                `json:"notification_id" db:"notification_id"`
	EventType         NotificationEventType `json:"event_type" db:"event_type"`
	EventData         map[string]interface{} `json:"event_data" db:"event_data"`
	ProviderEventID   *string               `json:"provider_event_id" db:"provider_event_id"`
	OccurredAt        time.Time             `json:"occurred_at" db:"occurred_at"`
	CreatedAt         time.Time             `json:"created_at" db:"created_at"`
}

// NotificationRetryQueue represents a retry queue entry
type NotificationRetryQueue struct {
	ID             string     `json:"id" db:"id"`
	NotificationID string     `json:"notification_id" db:"notification_id"`
	RetryAttempt   int        `json:"retry_attempt" db:"retry_attempt"`
	ScheduledFor   time.Time  `json:"scheduled_for" db:"scheduled_for"`
	ErrorReason    *string    `json:"error_reason" db:"error_reason"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	ProcessedAt    *time.Time `json:"processed_at" db:"processed_at"`
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	UserID     string                 `json:"user_id" validate:"required,uuid"`
	Channel    NotificationChannel    `json:"channel" validate:"required,oneof=email sms push"`
	TemplateID *string                `json:"template_id" validate:"omitempty,uuid"`
	Recipient  string                 `json:"recipient" validate:"required"`
	Subject    *string                `json:"subject"`
	Body       *string                `json:"body"`
	Variables  map[string]interface{} `json:"variables"`
}

// Validate validates the notification request
func (nr *NotificationRequest) Validate() error {
	if nr.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if nr.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	if nr.Recipient == "" {
		return fmt.Errorf("recipient is required")
	}
	if nr.TemplateID == nil && nr.Body == nil {
		return fmt.Errorf("either template_id or body is required")
	}
	return nil
}

// NotificationResponse represents the response after sending a notification
type NotificationResponse struct {
	ID                string                 `json:"id"`
	Status            NotificationStatus     `json:"status"`
	ProviderMessageID *string                `json:"provider_message_id,omitempty"`
	ErrorMessage      *string                `json:"error_message,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// TemplateRenderRequest represents a request to render a template
type TemplateRenderRequest struct {
	TemplateName string                 `json:"template_name" validate:"required"`
	Variables    map[string]interface{} `json:"variables"`
}

// TemplateRenderResponse represents the rendered template
type TemplateRenderResponse struct {
	Subject *string `json:"subject"`
	Body    string  `json:"body"`
}