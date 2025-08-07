package gateway

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/utils"
)

// MockSMSProvider implements SMSProvider for testing and development
type MockSMSProvider struct {
	logger *utils.StructuredLogger
}

// NewMockSMSProvider creates a new mock SMS provider
func NewMockSMSProvider(logger *utils.StructuredLogger) *MockSMSProvider {
	return &MockSMSProvider{
		logger: logger,
	}
}

// SendSMS sends an SMS message (mock implementation)
func (m *MockSMSProvider) SendSMS(ctx context.Context, to, message string) (*ProviderResponse, error) {
	// Simulate sending SMS
	messageID := uuid.New().String()
	
	m.logger.Info(ctx, "Mock SMS sent", map[string]interface{}{
		"to":         to,
		"message":    message,
		"message_id": messageID,
	})

	return &ProviderResponse{
		MessageID: messageID,
		Status:    "sent",
		Metadata: map[string]interface{}{
			"provider": "mock_sms",
			"to":       to,
		},
	}, nil
}

// ValidatePhoneNumber validates a phone number format
func (m *MockSMSProvider) ValidatePhoneNumber(phoneNumber string) error {
	// Basic phone number validation (E.164 format)
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phoneNumber) {
		return fmt.Errorf("invalid phone number format, expected E.164 format (e.g., +1234567890)")
	}
	return nil
}

// MockPushProvider implements PushProvider for testing and development
type MockPushProvider struct {
	logger *utils.StructuredLogger
}

// NewMockPushProvider creates a new mock push notification provider
func NewMockPushProvider(logger *utils.StructuredLogger) *MockPushProvider {
	return &MockPushProvider{
		logger: logger,
	}
}

// SendPushNotification sends a push notification (mock implementation)
func (m *MockPushProvider) SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]interface{}) (*ProviderResponse, error) {
	messageID := uuid.New().String()
	
	m.logger.Info(ctx, "Mock push notification sent", map[string]interface{}{
		"device_token": deviceToken,
		"title":        title,
		"body":         body,
		"data":         data,
		"message_id":   messageID,
	})

	return &ProviderResponse{
		MessageID: messageID,
		Status:    "sent",
		Metadata: map[string]interface{}{
			"provider":     "mock_push",
			"device_token": deviceToken,
			"title":        title,
		},
	}, nil
}

// SendTopicNotification sends a push notification to a topic (mock implementation)
func (m *MockPushProvider) SendTopicNotification(ctx context.Context, topic, title, body string, data map[string]interface{}) (*ProviderResponse, error) {
	messageID := uuid.New().String()
	
	m.logger.Info(ctx, "Mock topic push notification sent", map[string]interface{}{
		"topic":      topic,
		"title":      title,
		"body":       body,
		"data":       data,
		"message_id": messageID,
	})

	return &ProviderResponse{
		MessageID: messageID,
		Status:    "sent",
		Metadata: map[string]interface{}{
			"provider": "mock_push",
			"topic":    topic,
			"title":    title,
		},
	}, nil
}

// ValidateDeviceToken validates a device token format
func (m *MockPushProvider) ValidateDeviceToken(deviceToken string) error {
	if deviceToken == "" {
		return fmt.Errorf("device token cannot be empty")
	}
	
	// Basic validation - device tokens are typically long alphanumeric strings
	if len(deviceToken) < 10 {
		return fmt.Errorf("device token too short")
	}
	
	// Check for valid characters (alphanumeric, hyphens, underscores)
	validTokenRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validTokenRegex.MatchString(deviceToken) {
		return fmt.Errorf("device token contains invalid characters")
	}
	
	return nil
}

// TwilioSMSProvider implements SMSProvider using Twilio (placeholder for real implementation)
type TwilioSMSProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	logger     *utils.StructuredLogger
}

// NewTwilioSMSProvider creates a new Twilio SMS provider
func NewTwilioSMSProvider(accountSID, authToken, fromNumber string, logger *utils.StructuredLogger) *TwilioSMSProvider {
	return &TwilioSMSProvider{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		logger:     logger,
	}
}

// SendSMS sends an SMS message using Twilio (placeholder implementation)
func (t *TwilioSMSProvider) SendSMS(ctx context.Context, to, message string) (*ProviderResponse, error) {
	// TODO: Implement actual Twilio API call
	// For now, return a mock response
	messageID := fmt.Sprintf("twilio_%s", uuid.New().String())
	
	t.logger.Info(ctx, "Twilio SMS sent (mock)", map[string]interface{}{
		"to":         to,
		"from":       t.fromNumber,
		"message":    message,
		"message_id": messageID,
	})

	return &ProviderResponse{
		MessageID: messageID,
		Status:    "sent",
		Metadata: map[string]interface{}{
			"provider": "twilio",
			"to":       to,
			"from":     t.fromNumber,
		},
	}, nil
}

// ValidatePhoneNumber validates a phone number format for Twilio
func (t *TwilioSMSProvider) ValidatePhoneNumber(phoneNumber string) error {
	// Twilio expects E.164 format
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phoneNumber) {
		return fmt.Errorf("invalid phone number format for Twilio, expected E.164 format (e.g., +1234567890)")
	}
	return nil
}

// FirebasePushProvider implements PushProvider using Firebase Cloud Messaging (placeholder)
type FirebasePushProvider struct {
	projectID   string
	credentials string
	logger      *utils.StructuredLogger
}

// NewFirebasePushProvider creates a new Firebase push notification provider
func NewFirebasePushProvider(projectID, credentials string, logger *utils.StructuredLogger) *FirebasePushProvider {
	return &FirebasePushProvider{
		projectID:   projectID,
		credentials: credentials,
		logger:      logger,
	}
}

// SendPushNotification sends a push notification using Firebase (placeholder implementation)
func (f *FirebasePushProvider) SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]interface{}) (*ProviderResponse, error) {
	// TODO: Implement actual Firebase FCM API call
	messageID := fmt.Sprintf("fcm_%s", uuid.New().String())
	
	f.logger.Info(ctx, "Firebase push notification sent (mock)", map[string]interface{}{
		"device_token": deviceToken,
		"title":        title,
		"body":         body,
		"data":         data,
		"message_id":   messageID,
	})

	return &ProviderResponse{
		MessageID: messageID,
		Status:    "sent",
		Metadata: map[string]interface{}{
			"provider":     "firebase",
			"device_token": deviceToken,
			"project_id":   f.projectID,
		},
	}, nil
}

// SendTopicNotification sends a push notification to a topic using Firebase
func (f *FirebasePushProvider) SendTopicNotification(ctx context.Context, topic, title, body string, data map[string]interface{}) (*ProviderResponse, error) {
	// TODO: Implement actual Firebase FCM topic messaging
	messageID := fmt.Sprintf("fcm_topic_%s", uuid.New().String())
	
	f.logger.Info(ctx, "Firebase topic push notification sent (mock)", map[string]interface{}{
		"topic":      topic,
		"title":      title,
		"body":       body,
		"data":       data,
		"message_id": messageID,
	})

	return &ProviderResponse{
		MessageID: messageID,
		Status:    "sent",
		Metadata: map[string]interface{}{
			"provider":   "firebase",
			"topic":      topic,
			"project_id": f.projectID,
		},
	}, nil
}

// ValidateDeviceToken validates a Firebase device token
func (f *FirebasePushProvider) ValidateDeviceToken(deviceToken string) error {
	if deviceToken == "" {
		return fmt.Errorf("Firebase device token cannot be empty")
	}
	
	// Firebase tokens are typically 152+ characters long
	if len(deviceToken) < 100 {
		return fmt.Errorf("Firebase device token appears to be too short")
	}
	
	// Firebase tokens contain alphanumeric characters, hyphens, and underscores
	validTokenRegex := regexp.MustCompile(`^[a-zA-Z0-9_:-]+$`)
	if !validTokenRegex.MatchString(deviceToken) {
		return fmt.Errorf("Firebase device token contains invalid characters")
	}
	
	return nil
}

// ProviderFactory creates notification providers based on configuration
type ProviderFactory struct {
	logger *utils.StructuredLogger
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(logger *utils.StructuredLogger) *ProviderFactory {
	return &ProviderFactory{
		logger: logger,
	}
}

// CreateEmailProvider creates an email provider based on configuration
func (pf *ProviderFactory) CreateEmailProvider() (EmailProvider, error) {
	provider := strings.ToLower(getEnv("EMAIL_PROVIDER", "sendgrid"))
	
	switch provider {
	case "sendgrid":
		return NewSendGridProvider(pf.logger)
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", provider)
	}
}

// CreateSMSProvider creates an SMS provider based on configuration
func (pf *ProviderFactory) CreateSMSProvider() SMSProvider {
	provider := strings.ToLower(getEnv("SMS_PROVIDER", "mock"))
	
	switch provider {
	case "twilio":
		accountSID := getEnv("TWILIO_ACCOUNT_SID", "")
		authToken := getEnv("TWILIO_AUTH_TOKEN", "")
		fromNumber := getEnv("TWILIO_FROM_NUMBER", "")
		return NewTwilioSMSProvider(accountSID, authToken, fromNumber, pf.logger)
	case "mock":
		fallthrough
	default:
		return NewMockSMSProvider(pf.logger)
	}
}

// CreatePushProvider creates a push notification provider based on configuration
func (pf *ProviderFactory) CreatePushProvider() PushProvider {
	provider := strings.ToLower(getEnv("PUSH_PROVIDER", "mock"))
	
	switch provider {
	case "firebase":
		projectID := getEnv("FIREBASE_PROJECT_ID", "")
		credentials := getEnv("FIREBASE_CREDENTIALS", "")
		return NewFirebasePushProvider(projectID, credentials, pf.logger)
	case "mock":
		fallthrough
	default:
		return NewMockPushProvider(pf.logger)
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
