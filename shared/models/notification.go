package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationEmail NotificationType = "email"
	NotificationSMS   NotificationType = "sms"
	NotificationPush  NotificationType = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationPending   NotificationStatus = "pending"
	NotificationSent      NotificationStatus = "sent"
	NotificationDelivered NotificationStatus = "delivered"
	NotificationFailed    NotificationStatus = "failed"
	NotificationRetrying  NotificationStatus = "retrying"
)

// Notification represents a notification to be sent
type Notification struct {
	ID        string             `json:"id" db:"id"`
	UserID    string             `json:"user_id" db:"user_id"`
	Type      NotificationType   `json:"type" db:"type"`
	Status    NotificationStatus `json:"status" db:"status"`
	Subject   string             `json:"subject" db:"subject"`
	Content   string             `json:"content" db:"content"`
	Recipient string             `json:"recipient" db:"recipient"` // email, phone, device_token
	Template  string             `json:"template" db:"template"`
	Variables map[string]interface{} `json:"variables" db:"variables"`
	Priority  int                `json:"priority" db:"priority"` // 1-5, 5 being highest
	RetryCount int               `json:"retry_count" db:"retry_count"`
	MaxRetries int               `json:"max_retries" db:"max_retries"`
	ScheduledAt *time.Time        `json:"scheduled_at" db:"scheduled_at"`
	SentAt     *time.Time         `json:"sent_at" db:"sent_at"`
	CreatedAt  time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" db:"updated_at"`
}

// NewNotification creates a new notification with default values
func NewNotification(userID string, notificationType NotificationType, recipient, subject, content string) *Notification {
	return &Notification{
		ID:         uuid.New().String(),
		UserID:     userID,
		Type:       notificationType,
		Status:     NotificationPending,
		Subject:    subject,
		Content:    content,
		Recipient:  recipient,
		Variables:  make(map[string]interface{}),
		Priority:   3, // medium priority
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID                string    `json:"id" db:"id"`
	UserID            string    `json:"user_id" db:"user_id"`
	EmailEnabled      bool      `json:"email_enabled" db:"email_enabled"`
	SMSEnabled        bool      `json:"sms_enabled" db:"sms_enabled"`
	PushEnabled       bool      `json:"push_enabled" db:"push_enabled"`
	OrderUpdates      bool      `json:"order_updates" db:"order_updates"`
	PromotionalEmails bool      `json:"promotional_emails" db:"promotional_emails"`
	SecurityAlerts    bool      `json:"security_alerts" db:"security_alerts"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}