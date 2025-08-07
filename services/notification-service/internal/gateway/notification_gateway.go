package gateway

import (
	"context"
	"fmt"

	"github.com/shopsphere/shared/models"
)

// EmailProvider defines the interface for email notification providers
type EmailProvider interface {
	SendEmail(ctx context.Context, to, subject, body string) (*ProviderResponse, error)
	SendTemplateEmail(ctx context.Context, to, templateID string, variables map[string]interface{}) (*ProviderResponse, error)
	ValidateEmail(email string) error
}

// SMSProvider defines the interface for SMS notification providers
type SMSProvider interface {
	SendSMS(ctx context.Context, to, message string) (*ProviderResponse, error)
	ValidatePhoneNumber(phoneNumber string) error
}

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]interface{}) (*ProviderResponse, error)
	SendTopicNotification(ctx context.Context, topic, title, body string, data map[string]interface{}) (*ProviderResponse, error)
	ValidateDeviceToken(deviceToken string) error
}

// ProviderResponse represents the response from a notification provider
type ProviderResponse struct {
	MessageID    string                 `json:"message_id"`
	Status       string                 `json:"status"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationGateway manages different notification providers
type NotificationGateway struct {
	emailProvider EmailProvider
	smsProvider   SMSProvider
	pushProvider  PushProvider
}

// NewNotificationGateway creates a new notification gateway
func NewNotificationGateway(emailProvider EmailProvider, smsProvider SMSProvider, pushProvider PushProvider) *NotificationGateway {
	return &NotificationGateway{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
		pushProvider:  pushProvider,
	}
}

// SendNotification sends a notification using the appropriate provider
func (ng *NotificationGateway) SendNotification(ctx context.Context, notification *models.Notification) (*ProviderResponse, error) {
	switch notification.Channel {
	case models.ChannelEmail:
		if ng.emailProvider == nil {
			return nil, fmt.Errorf("email provider not configured")
		}
		subject := ""
		if notification.Subject != nil {
			subject = *notification.Subject
		}
		return ng.emailProvider.SendEmail(ctx, notification.Recipient, subject, notification.Body)

	case models.ChannelSMS:
		if ng.smsProvider == nil {
			return nil, fmt.Errorf("SMS provider not configured")
		}
		return ng.smsProvider.SendSMS(ctx, notification.Recipient, notification.Body)

	case models.ChannelPush:
		if ng.pushProvider == nil {
			return nil, fmt.Errorf("push provider not configured")
		}
		title := ""
		if notification.Subject != nil {
			title = *notification.Subject
		}
		return ng.pushProvider.SendPushNotification(ctx, notification.Recipient, title, notification.Body, notification.Variables)

	default:
		return nil, fmt.Errorf("unsupported notification channel: %s", notification.Channel)
	}
}

// ValidateRecipient validates the recipient based on the notification channel
func (ng *NotificationGateway) ValidateRecipient(channel models.NotificationChannel, recipient string) error {
	switch channel {
	case models.ChannelEmail:
		if ng.emailProvider == nil {
			return fmt.Errorf("email provider not configured")
		}
		return ng.emailProvider.ValidateEmail(recipient)

	case models.ChannelSMS:
		if ng.smsProvider == nil {
			return fmt.Errorf("SMS provider not configured")
		}
		return ng.smsProvider.ValidatePhoneNumber(recipient)

	case models.ChannelPush:
		if ng.pushProvider == nil {
			return fmt.Errorf("push provider not configured")
		}
		return ng.pushProvider.ValidateDeviceToken(recipient)

	default:
		return fmt.Errorf("unsupported notification channel: %s", channel)
	}
}
