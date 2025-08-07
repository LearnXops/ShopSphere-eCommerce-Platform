package service

import (
	"context"
	"fmt"
	"strings"
	textTemplate "text/template"
	"time"

	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/notification-service/internal/gateway"
	"github.com/shopsphere/notification-service/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService struct {
	repo    repository.NotificationRepository
	gateway *gateway.NotificationGateway
	logger  *utils.StructuredLogger
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo repository.NotificationRepository, gateway *gateway.NotificationGateway, logger *utils.StructuredLogger) *NotificationService {
	return &NotificationService{
		repo:    repo,
		gateway: gateway,
		logger:  logger,
	}
}

// SendNotification sends a notification
func (ns *NotificationService) SendNotification(ctx context.Context, request *models.NotificationRequest) (*models.NotificationResponse, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		ns.logger.Error(ctx, "Invalid notification request", err, map[string]interface{}{
			"user_id": request.UserID,
			"channel": request.Channel,
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if user has enabled this notification channel and category
	category := ns.inferCategory(request)
	isEnabled, err := ns.repo.IsChannelEnabled(ctx, request.UserID, request.Channel, category)
	if err != nil {
		ns.logger.Error(ctx, "Failed to check channel enabled", err, map[string]interface{}{
			"user_id":  request.UserID,
			"channel":  request.Channel,
			"category": category,
		})
		return nil, fmt.Errorf("failed to check notification preferences: %w", err)
	}

	if !isEnabled {
		ns.logger.Info(ctx, "Notification channel disabled for user", map[string]interface{}{
			"user_id":  request.UserID,
			"channel":  request.Channel,
			"category": category,
		})
		return &models.NotificationResponse{
			Status: models.NotificationFailed,
			ErrorMessage: func() *string {
				msg := "notification channel disabled for user"
				return &msg
			}(),
		}, nil
	}

	// Validate recipient format
	if err := ns.gateway.ValidateRecipient(request.Channel, request.Recipient); err != nil {
		ns.logger.Error(ctx, "Invalid recipient", err, map[string]interface{}{
			"channel":   request.Channel,
			"recipient": request.Recipient,
		})
		return nil, fmt.Errorf("invalid recipient: %w", err)
	}

	// Create notification record
	notification := &models.Notification{
		UserID:     request.UserID,
		TemplateID: request.TemplateID,
		Channel:    request.Channel,
		Recipient:  request.Recipient,
		Subject:    request.Subject,
		Variables:  request.Variables,
		Status:     models.NotificationPending,
		MaxRetries: 3,
	}

	// If template ID is provided, render the template
	if request.TemplateID != nil {
		renderedContent, err := ns.renderTemplate(ctx, *request.TemplateID, request.Variables)
		if err != nil {
			ns.logger.Error(ctx, "Failed to render template", err, map[string]interface{}{
				"template_id": *request.TemplateID,
			})
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
		notification.Subject = renderedContent.Subject
		notification.Body = renderedContent.Body
	} else if request.Body != nil {
		notification.Body = *request.Body
	} else {
		return nil, fmt.Errorf("either template_id or body must be provided")
	}

	// Save notification to database
	if err := ns.repo.CreateNotification(ctx, notification); err != nil {
		ns.logger.Error(ctx, "Failed to create notification", err, map[string]interface{}{
			"user_id": request.UserID,
			"channel": request.Channel,
		})
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification asynchronously
	go ns.processNotification(context.Background(), notification)

	return &models.NotificationResponse{
		ID:        notification.ID,
		Status:    notification.Status,
		CreatedAt: notification.CreatedAt,
	}, nil
}

// processNotification processes a notification for sending
func (ns *NotificationService) processNotification(ctx context.Context, notification *models.Notification) {
	ns.logger.Info(ctx, "Processing notification", map[string]interface{}{
		"notification_id": notification.ID,
		"channel":         notification.Channel,
	})

	// Send notification via gateway
	response, err := ns.gateway.SendNotification(ctx, notification)
	if err != nil {
		ns.logger.Error(ctx, "Failed to send notification", err, map[string]interface{}{
			"notification_id": notification.ID,
		})

		// Update status to failed
		errorMsg := err.Error()
		if updateErr := ns.repo.UpdateNotificationStatus(ctx, notification.ID, models.NotificationFailed, nil, &errorMsg); updateErr != nil {
			ns.logger.Error(ctx, "Failed to update notification status", updateErr, map[string]interface{}{
				"notification_id": notification.ID,
			})
		}

		// Add to retry queue if retries are available
		if notification.RetryCount < notification.MaxRetries {
			ns.scheduleRetry(ctx, notification, err.Error())
		}
		return
	}

	// Update notification status to sent
	if err := ns.repo.UpdateNotificationStatus(ctx, notification.ID, models.NotificationSent, &response.MessageID, nil); err != nil {
		ns.logger.Error(ctx, "Failed to update notification status", err, map[string]interface{}{
			"notification_id": notification.ID,
		})
	}

	// Create delivery event
	deliveryEvent := &models.NotificationDeliveryEvent{
		NotificationID: notification.ID,
		EventType:      models.EventSent,
		EventData:      response.Metadata,
		OccurredAt:     time.Now(),
	}

	if err := ns.repo.CreateDeliveryEvent(ctx, deliveryEvent); err != nil {
		ns.logger.Error(ctx, "Failed to create delivery event", err, map[string]interface{}{
			"notification_id": notification.ID,
		})
	}

	ns.logger.Info(ctx, "Notification sent successfully", map[string]interface{}{
		"notification_id": notification.ID,
		"message_id":      response.MessageID,
	})
}

// scheduleRetry schedules a notification for retry
func (ns *NotificationService) scheduleRetry(ctx context.Context, notification *models.Notification, errorReason string) {
	// Calculate retry delay (exponential backoff)
	retryDelay := time.Duration(1<<notification.RetryCount) * time.Minute
	scheduledFor := time.Now().Add(retryDelay)

	retryEntry := &models.NotificationRetryQueue{
		NotificationID: notification.ID,
		RetryAttempt:   notification.RetryCount + 1,
		ScheduledFor:   scheduledFor,
		ErrorReason:    &errorReason,
	}

	if err := ns.repo.AddToRetryQueue(ctx, retryEntry); err != nil {
		ns.logger.Error(ctx, "Failed to add to retry queue", err, map[string]interface{}{
			"notification_id": notification.ID,
		})
	} else {
		ns.logger.Info(ctx, "Notification scheduled for retry", map[string]interface{}{
			"notification_id": notification.ID,
			"retry_attempt":   retryEntry.RetryAttempt,
			"scheduled_for":   scheduledFor,
		})
	}
}

// ProcessRetryQueue processes notifications in the retry queue
func (ns *NotificationService) ProcessRetryQueue(ctx context.Context) error {
	entries, err := ns.repo.GetRetryQueueEntries(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get retry queue entries: %w", err)
	}

	for _, entry := range entries {
		// Mark as processed first to avoid duplicate processing
		if err := ns.repo.MarkRetryProcessed(ctx, entry.ID); err != nil {
			ns.logger.Error(ctx, "Failed to mark retry processed", err, map[string]interface{}{
				"retry_id": entry.ID,
			})
			continue
		}

		// Get the notification
		notification, err := ns.repo.GetNotification(ctx, entry.NotificationID)
		if err != nil {
			ns.logger.Error(ctx, "Failed to get notification for retry", err, map[string]interface{}{
				"notification_id": entry.NotificationID,
			})
			continue
		}

		// Increment retry count
		if err := ns.repo.IncrementRetryCount(ctx, notification.ID); err != nil {
			ns.logger.Error(ctx, "Failed to increment retry count", err, map[string]interface{}{
				"notification_id": notification.ID,
			})
			continue
		}

		notification.RetryCount = entry.RetryAttempt

		// Process the notification
		go ns.processNotification(context.Background(), notification)
	}

	return nil
}

// renderTemplate renders a notification template with variables
func (ns *NotificationService) renderTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*models.TemplateRenderResponse, error) {
	// Get template from database
	tmpl, err := ns.repo.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if !tmpl.IsActive {
		return nil, fmt.Errorf("template is not active")
	}

	// Render body template
	bodyTemplate, err := textTemplate.New("body").Parse(tmpl.BodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body template: %w", err)
	}

	var bodyBuilder strings.Builder
	if err := bodyTemplate.Execute(&bodyBuilder, variables); err != nil {
		return nil, fmt.Errorf("failed to execute body template: %w", err)
	}

	response := &models.TemplateRenderResponse{
		Body: bodyBuilder.String(),
	}

	// Render subject template if present
	if tmpl.Subject != nil {
		subjectTmpl, err := textTemplate.New(tmpl.Name).Parse(*tmpl.Subject)
		if err != nil {
			return nil, fmt.Errorf("failed to parse subject template: %w", err)
		}

		var subjectBuilder strings.Builder
		if err := subjectTmpl.Execute(&subjectBuilder, variables); err != nil {
			return nil, fmt.Errorf("failed to execute subject template: %w", err)
		}

		subject := subjectBuilder.String()
		response.Subject = &subject
	}

	return response, nil
}

// RenderTemplate renders a template by name with variables
func (ns *NotificationService) RenderTemplate(ctx context.Context, request *models.TemplateRenderRequest) (*models.TemplateRenderResponse, error) {
	// Get template by name
	tmpl, err := ns.repo.GetTemplateByName(ctx, request.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return ns.renderTemplate(ctx, tmpl.ID, request.Variables)
}

// GetNotification retrieves a notification by ID
func (ns *NotificationService) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	return ns.repo.GetNotification(ctx, id)
}

// ListNotifications lists notifications for a user
func (ns *NotificationService) ListNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return ns.repo.ListNotifications(ctx, userID, limit, offset)
}

// GetUserPreferences retrieves notification preferences for a user
func (ns *NotificationService) GetUserPreferences(ctx context.Context, userID string) ([]*models.NotificationPreference, error) {
	return ns.repo.GetUserPreferences(ctx, userID)
}

// UpdateUserPreference updates a user's notification preference
func (ns *NotificationService) UpdateUserPreference(ctx context.Context, userID string, channel models.NotificationChannel, category string, enabled bool) error {
	// Check if preference exists
	preferences, err := ns.repo.GetUserPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Find existing preference
	var existingPreference *models.NotificationPreference
	for _, pref := range preferences {
		if pref.Channel == channel && pref.Category == category {
			existingPreference = pref
			break
		}
	}

	if existingPreference != nil {
		// Update existing preference
		existingPreference.IsEnabled = enabled
		return ns.repo.UpdatePreference(ctx, existingPreference)
	} else {
		// Create new preference
		newPreference := &models.NotificationPreference{
			UserID:    userID,
			Channel:   channel,
			Category:  category,
			IsEnabled: enabled,
		}
		return ns.repo.CreatePreference(ctx, newPreference)
	}
}

// CreateTemplate creates a new notification template
func (ns *NotificationService) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	// Validate template
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.BodyTemplate == "" {
		return fmt.Errorf("template body is required")
	}

	// Test template compilation
	if _, err := textTemplate.New("test").Parse(template.BodyTemplate); err != nil {
		return fmt.Errorf("invalid body template: %w", err)
	}

	if template.Subject != nil {
		if _, err := textTemplate.New("test").Parse(*template.Subject); err != nil {
			return fmt.Errorf("invalid subject template: %w", err)
		}
	}

	return ns.repo.CreateTemplate(ctx, template)
}

// GetTemplate retrieves a template by ID
func (ns *NotificationService) GetTemplate(ctx context.Context, id string) (*models.NotificationTemplate, error) {
	return ns.repo.GetTemplate(ctx, id)
}

// ListTemplates lists templates by channel
func (ns *NotificationService) ListTemplates(ctx context.Context, channel *models.NotificationChannel) ([]*models.NotificationTemplate, error) {
	return ns.repo.ListTemplates(ctx, channel)
}

// UpdateTemplate updates a notification template
func (ns *NotificationService) UpdateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	// Validate template
	if template.BodyTemplate == "" {
		return fmt.Errorf("template body is required")
	}

	// Test template compilation
	if _, err := textTemplate.New("test").Parse(template.BodyTemplate); err != nil {
		return fmt.Errorf("invalid body template: %w", err)
	}

	if template.Subject != nil {
		if _, err := textTemplate.New("test").Parse(*template.Subject); err != nil {
			return fmt.Errorf("invalid subject template: %w", err)
		}
	}

	return ns.repo.UpdateTemplate(ctx, template)
}

// DeleteTemplate deletes a notification template
func (ns *NotificationService) DeleteTemplate(ctx context.Context, id string) error {
	return ns.repo.DeleteTemplate(ctx, id)
}

// GetDeliveryEvents retrieves delivery events for a notification
func (ns *NotificationService) GetDeliveryEvents(ctx context.Context, notificationID string) ([]*models.NotificationDeliveryEvent, error) {
	return ns.repo.GetDeliveryEvents(ctx, notificationID)
}

// inferCategory infers the notification category from the request
func (ns *NotificationService) inferCategory(request *models.NotificationRequest) string {
	// Try to infer category from template ID or content
	if request.TemplateID != nil {
		templateName := *request.TemplateID
		if strings.Contains(templateName, "order") {
			return "order_updates"
		}
		if strings.Contains(templateName, "security") || strings.Contains(templateName, "password") {
			return "security"
		}
		if strings.Contains(templateName, "marketing") || strings.Contains(templateName, "promotion") {
			return "marketing"
		}
		if strings.Contains(templateName, "review") {
			return "reviews"
		}
	}

	// Default category
	return "general"
}
