package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	// Template operations
	CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error
	GetTemplate(ctx context.Context, id string) (*models.NotificationTemplate, error)
	GetTemplateByName(ctx context.Context, name string) (*models.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, template *models.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, id string) error
	ListTemplates(ctx context.Context, channel *models.NotificationChannel) ([]*models.NotificationTemplate, error)

	// Notification operations
	CreateNotification(ctx context.Context, notification *models.Notification) error
	GetNotification(ctx context.Context, id string) (*models.Notification, error)
	UpdateNotificationStatus(ctx context.Context, id string, status models.NotificationStatus, providerMessageID *string, errorMessage *string) error
	ListNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
	GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error)
	IncrementRetryCount(ctx context.Context, id string) error

	// Preference operations
	CreatePreference(ctx context.Context, preference *models.NotificationPreference) error
	GetUserPreferences(ctx context.Context, userID string) ([]*models.NotificationPreference, error)
	UpdatePreference(ctx context.Context, preference *models.NotificationPreference) error
	IsChannelEnabled(ctx context.Context, userID string, channel models.NotificationChannel, category string) (bool, error)

	// Delivery event operations
	CreateDeliveryEvent(ctx context.Context, event *models.NotificationDeliveryEvent) error
	GetDeliveryEvents(ctx context.Context, notificationID string) ([]*models.NotificationDeliveryEvent, error)

	// Retry queue operations
	AddToRetryQueue(ctx context.Context, retryEntry *models.NotificationRetryQueue) error
	GetRetryQueueEntries(ctx context.Context, limit int) ([]*models.NotificationRetryQueue, error)
	MarkRetryProcessed(ctx context.Context, id string) error
}

// PostgresNotificationRepository implements NotificationRepository using PostgreSQL
type PostgresNotificationRepository struct {
	db     *sql.DB
	logger *utils.StructuredLogger
}

// NewPostgresNotificationRepository creates a new PostgreSQL notification repository
func NewPostgresNotificationRepository(db *sql.DB, logger *utils.StructuredLogger) NotificationRepository {
	return &PostgresNotificationRepository{
		db:     db,
		logger: logger,
	}
}

// CreateTemplate creates a new notification template
func (r *PostgresNotificationRepository) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	template.ID = uuid.New().String()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	variablesJSON, err := json.Marshal(template.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `
		INSERT INTO notification_templates (id, name, channel, subject, body_template, variables, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Channel, template.Subject, template.BodyTemplate,
		variablesJSON, template.IsActive, template.CreatedAt, template.UpdatedAt)
	if err != nil {
		r.logger.Error(ctx, "Failed to create notification template", err, map[string]interface{}{
			"template_name": template.Name,
			"channel":       template.Channel,
		})
		return fmt.Errorf("failed to create notification template: %w", err)
	}

	r.logger.Info(ctx, "Created notification template", map[string]interface{}{
		"template_id":   template.ID,
		"template_name": template.Name,
		"channel":       template.Channel,
	})

	return nil
}

// GetTemplate retrieves a notification template by ID
func (r *PostgresNotificationRepository) GetTemplate(ctx context.Context, id string) (*models.NotificationTemplate, error) {
	template := &models.NotificationTemplate{}
	var variablesJSON []byte

	query := `
		SELECT id, name, channel, subject, body_template, variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID, &template.Name, &template.Channel, &template.Subject,
		&template.BodyTemplate, &variablesJSON, &template.IsActive,
		&template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification template not found")
		}
		r.logger.Error(ctx, "Failed to get notification template", err, map[string]interface{}{
			"template_id": id,
		})
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}

	if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return template, nil
}

// GetTemplateByName retrieves a notification template by name
func (r *PostgresNotificationRepository) GetTemplateByName(ctx context.Context, name string) (*models.NotificationTemplate, error) {
	template := &models.NotificationTemplate{}
	var variablesJSON []byte

	query := `
		SELECT id, name, channel, subject, body_template, variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE name = $1 AND is_active = true`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&template.ID, &template.Name, &template.Channel, &template.Subject,
		&template.BodyTemplate, &variablesJSON, &template.IsActive,
		&template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification template not found")
		}
		r.logger.Error(ctx, "Failed to get notification template by name", err, map[string]interface{}{
			"template_name": name,
		})
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}

	if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return template, nil
}

// UpdateTemplate updates a notification template
func (r *PostgresNotificationRepository) UpdateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	template.UpdatedAt = time.Now()

	variablesJSON, err := json.Marshal(template.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `
		UPDATE notification_templates
		SET name = $2, channel = $3, subject = $4, body_template = $5, variables = $6, is_active = $7, updated_at = $8
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Channel, template.Subject, template.BodyTemplate,
		variablesJSON, template.IsActive, template.UpdatedAt)
	if err != nil {
		r.logger.Error(ctx, "Failed to update notification template", err, map[string]interface{}{
			"template_id": template.ID,
		})
		return fmt.Errorf("failed to update notification template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification template not found")
	}

	r.logger.Info(ctx, "Updated notification template", map[string]interface{}{
		"template_id": template.ID,
	})

	return nil
}

// DeleteTemplate deletes a notification template
func (r *PostgresNotificationRepository) DeleteTemplate(ctx context.Context, id string) error {
	query := `DELETE FROM notification_templates WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to delete notification template", err, map[string]interface{}{
			"template_id": id,
		})
		return fmt.Errorf("failed to delete notification template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification template not found")
	}

	r.logger.Info(ctx, "Deleted notification template", map[string]interface{}{
		"template_id": id,
	})

	return nil
}

// ListTemplates lists notification templates
func (r *PostgresNotificationRepository) ListTemplates(ctx context.Context, channel *models.NotificationChannel) ([]*models.NotificationTemplate, error) {
	var query string
	var args []interface{}

	if channel != nil {
		query = `
			SELECT id, name, channel, subject, body_template, variables, is_active, created_at, updated_at
			FROM notification_templates
			WHERE channel = $1 AND is_active = true
			ORDER BY name`
		args = append(args, *channel)
	} else {
		query = `
			SELECT id, name, channel, subject, body_template, variables, is_active, created_at, updated_at
			FROM notification_templates
			WHERE is_active = true
			ORDER BY channel, name`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "Failed to list notification templates", err, map[string]interface{}{
			"channel": channel,
		})
		return nil, fmt.Errorf("failed to list notification templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.NotificationTemplate
	for rows.Next() {
		template := &models.NotificationTemplate{}
		var variablesJSON []byte

		err := rows.Scan(
			&template.ID, &template.Name, &template.Channel, &template.Subject,
			&template.BodyTemplate, &variablesJSON, &template.IsActive,
			&template.CreatedAt, &template.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification template: %w", err)
		}

		if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification templates: %w", err)
	}

	return templates, nil
}
