package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
)

// CreateNotification creates a new notification
func (r *PostgresNotificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	variablesJSON, err := json.Marshal(notification.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `
		INSERT INTO notifications (id, user_id, template_id, channel, recipient, subject, body, variables, 
		                          status, provider, provider_message_id, error_message, retry_count, max_retries, 
		                          scheduled_at, sent_at, delivered_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	_, err = r.db.ExecContext(ctx, query,
		notification.ID, notification.UserID, notification.TemplateID, notification.Channel,
		notification.Recipient, notification.Subject, notification.Body, variablesJSON,
		notification.Status, notification.Provider, notification.ProviderMessageID,
		notification.ErrorMessage, notification.RetryCount, notification.MaxRetries,
		notification.ScheduledAt, notification.SentAt, notification.DeliveredAt,
		notification.CreatedAt, notification.UpdatedAt)
	if err != nil {
		r.logger.Error(ctx, "Failed to create notification", err, map[string]interface{}{
			"user_id": notification.UserID,
			"channel": notification.Channel,
		})
		return fmt.Errorf("failed to create notification: %w", err)
	}

	r.logger.Info(ctx, "Created notification", map[string]interface{}{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"channel":         notification.Channel,
	})

	return nil
}

// GetNotification retrieves a notification by ID
func (r *PostgresNotificationRepository) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	notification := &models.Notification{}
	var variablesJSON []byte

	query := `
		SELECT id, user_id, template_id, channel, recipient, subject, body, variables, 
		       status, provider, provider_message_id, error_message, retry_count, max_retries, 
		       scheduled_at, sent_at, delivered_at, created_at, updated_at
		FROM notifications
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID, &notification.UserID, &notification.TemplateID, &notification.Channel,
		&notification.Recipient, &notification.Subject, &notification.Body, &variablesJSON,
		&notification.Status, &notification.Provider, &notification.ProviderMessageID,
		&notification.ErrorMessage, &notification.RetryCount, &notification.MaxRetries,
		&notification.ScheduledAt, &notification.SentAt, &notification.DeliveredAt,
		&notification.CreatedAt, &notification.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		r.logger.Error(ctx, "Failed to get notification", err, map[string]interface{}{
			"notification_id": id,
		})
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if err := json.Unmarshal(variablesJSON, &notification.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return notification, nil
}

// UpdateNotificationStatus updates the status of a notification
func (r *PostgresNotificationRepository) UpdateNotificationStatus(ctx context.Context, id string, status models.NotificationStatus, providerMessageID *string, errorMessage *string) error {
	var sentAt, deliveredAt *time.Time
	now := time.Now()

	if status == models.NotificationSent {
		sentAt = &now
	} else if status == models.NotificationDelivered {
		deliveredAt = &now
	}

	query := `
		UPDATE notifications
		SET status = $2, provider_message_id = $3, error_message = $4, sent_at = $5, delivered_at = $6, updated_at = $7
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, status, providerMessageID, errorMessage, sentAt, deliveredAt, now)
	if err != nil {
		r.logger.Error(ctx, "Failed to update notification status", err, map[string]interface{}{
			"notification_id": id,
			"status":          status,
		})
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	r.logger.Info(ctx, "Updated notification status", map[string]interface{}{
		"notification_id": id,
		"status":          status,
	})

	return nil
}

// ListNotifications lists notifications for a user
func (r *PostgresNotificationRepository) ListNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, template_id, channel, recipient, subject, body, variables, 
		       status, provider, provider_message_id, error_message, retry_count, max_retries, 
		       scheduled_at, sent_at, delivered_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error(ctx, "Failed to list notifications", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		notification := &models.Notification{}
		var variablesJSON []byte

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.TemplateID, &notification.Channel,
			&notification.Recipient, &notification.Subject, &notification.Body, &variablesJSON,
			&notification.Status, &notification.Provider, &notification.ProviderMessageID,
			&notification.ErrorMessage, &notification.RetryCount, &notification.MaxRetries,
			&notification.ScheduledAt, &notification.SentAt, &notification.DeliveredAt,
			&notification.CreatedAt, &notification.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if err := json.Unmarshal(variablesJSON, &notification.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// GetPendingNotifications retrieves pending notifications for processing
func (r *PostgresNotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, template_id, channel, recipient, subject, body, variables, 
		       status, provider, provider_message_id, error_message, retry_count, max_retries, 
		       scheduled_at, sent_at, delivered_at, created_at, updated_at
		FROM notifications
		WHERE status = $1 AND (scheduled_at IS NULL OR scheduled_at <= $2)
		ORDER BY created_at ASC
		LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, models.NotificationPending, time.Now(), limit)
	if err != nil {
		r.logger.Error(ctx, "Failed to get pending notifications", err)
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		notification := &models.Notification{}
		var variablesJSON []byte

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.TemplateID, &notification.Channel,
			&notification.Recipient, &notification.Subject, &notification.Body, &variablesJSON,
			&notification.Status, &notification.Provider, &notification.ProviderMessageID,
			&notification.ErrorMessage, &notification.RetryCount, &notification.MaxRetries,
			&notification.ScheduledAt, &notification.SentAt, &notification.DeliveredAt,
			&notification.CreatedAt, &notification.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if err := json.Unmarshal(variablesJSON, &notification.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// IncrementRetryCount increments the retry count for a notification
func (r *PostgresNotificationRepository) IncrementRetryCount(ctx context.Context, id string) error {
	query := `
		UPDATE notifications
		SET retry_count = retry_count + 1, updated_at = $2
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		r.logger.Error(ctx, "Failed to increment retry count", err, map[string]interface{}{
			"notification_id": id,
		})
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// CreatePreference creates a new notification preference
func (r *PostgresNotificationRepository) CreatePreference(ctx context.Context, preference *models.NotificationPreference) error {
	preference.ID = uuid.New().String()
	preference.CreatedAt = time.Now()
	preference.UpdatedAt = time.Now()

	query := `
		INSERT INTO notification_preferences (id, user_id, channel, category, is_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, channel, category) 
		DO UPDATE SET is_enabled = $5, updated_at = $7`

	_, err := r.db.ExecContext(ctx, query,
		preference.ID, preference.UserID, preference.Channel, preference.Category,
		preference.IsEnabled, preference.CreatedAt, preference.UpdatedAt)
	if err != nil {
		r.logger.Error(ctx, "Failed to create notification preference", err, map[string]interface{}{
			"user_id":  preference.UserID,
			"channel":  preference.Channel,
			"category": preference.Category,
		})
		return fmt.Errorf("failed to create notification preference: %w", err)
	}

	r.logger.Info(ctx, "Created notification preference", map[string]interface{}{
		"preference_id": preference.ID,
		"user_id":       preference.UserID,
		"channel":       preference.Channel,
		"category":      preference.Category,
	})

	return nil
}

// GetUserPreferences retrieves all notification preferences for a user
func (r *PostgresNotificationRepository) GetUserPreferences(ctx context.Context, userID string) ([]*models.NotificationPreference, error) {
	query := `
		SELECT id, user_id, channel, category, is_enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
		ORDER BY channel, category`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get user preferences", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}
	defer rows.Close()

	var preferences []*models.NotificationPreference
	for rows.Next() {
		preference := &models.NotificationPreference{}

		err := rows.Scan(
			&preference.ID, &preference.UserID, &preference.Channel, &preference.Category,
			&preference.IsEnabled, &preference.CreatedAt, &preference.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification preference: %w", err)
		}

		preferences = append(preferences, preference)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification preferences: %w", err)
	}

	return preferences, nil
}

// UpdatePreference updates a notification preference
func (r *PostgresNotificationRepository) UpdatePreference(ctx context.Context, preference *models.NotificationPreference) error {
	preference.UpdatedAt = time.Now()

	query := `
		UPDATE notification_preferences
		SET is_enabled = $2, updated_at = $3
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, preference.ID, preference.IsEnabled, preference.UpdatedAt)
	if err != nil {
		r.logger.Error(ctx, "Failed to update notification preference", err, map[string]interface{}{
			"preference_id": preference.ID,
		})
		return fmt.Errorf("failed to update notification preference: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification preference not found")
	}

	r.logger.Info(ctx, "Updated notification preference", map[string]interface{}{
		"preference_id": preference.ID,
	})

	return nil
}

// IsChannelEnabled checks if a notification channel is enabled for a user and category
func (r *PostgresNotificationRepository) IsChannelEnabled(ctx context.Context, userID string, channel models.NotificationChannel, category string) (bool, error) {
	var isEnabled bool

	query := `
		SELECT is_enabled
		FROM notification_preferences
		WHERE user_id = $1 AND channel = $2 AND category = $3`

	err := r.db.QueryRowContext(ctx, query, userID, channel, category).Scan(&isEnabled)
	if err != nil {
		if err == sql.ErrNoRows {
			// Default to enabled if no preference is set
			return true, nil
		}
		r.logger.Error(ctx, "Failed to check channel enabled", err, map[string]interface{}{
			"user_id":  userID,
			"channel":  channel,
			"category": category,
		})
		return false, fmt.Errorf("failed to check channel enabled: %w", err)
	}

	return isEnabled, nil
}

// CreateDeliveryEvent creates a new delivery event
func (r *PostgresNotificationRepository) CreateDeliveryEvent(ctx context.Context, event *models.NotificationDeliveryEvent) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()

	eventDataJSON, err := json.Marshal(event.EventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	query := `
		INSERT INTO notification_delivery_events (id, notification_id, event_type, event_data, provider_event_id, occurred_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.NotificationID, event.EventType, eventDataJSON,
		event.ProviderEventID, event.OccurredAt, event.CreatedAt)
	if err != nil {
		r.logger.Error(ctx, "Failed to create delivery event", err, map[string]interface{}{
			"notification_id": event.NotificationID,
			"event_type":      event.EventType,
		})
		return fmt.Errorf("failed to create delivery event: %w", err)
	}

	r.logger.Info(ctx, "Created delivery event", map[string]interface{}{
		"event_id":        event.ID,
		"notification_id": event.NotificationID,
		"event_type":      event.EventType,
	})

	return nil
}

// GetDeliveryEvents retrieves delivery events for a notification
func (r *PostgresNotificationRepository) GetDeliveryEvents(ctx context.Context, notificationID string) ([]*models.NotificationDeliveryEvent, error) {
	query := `
		SELECT id, notification_id, event_type, event_data, provider_event_id, occurred_at, created_at
		FROM notification_delivery_events
		WHERE notification_id = $1
		ORDER BY occurred_at ASC`

	rows, err := r.db.QueryContext(ctx, query, notificationID)
	if err != nil {
		r.logger.Error(ctx, "Failed to get delivery events", err, map[string]interface{}{
			"notification_id": notificationID,
		})
		return nil, fmt.Errorf("failed to get delivery events: %w", err)
	}
	defer rows.Close()

	var events []*models.NotificationDeliveryEvent
	for rows.Next() {
		event := &models.NotificationDeliveryEvent{}
		var eventDataJSON []byte

		err := rows.Scan(
			&event.ID, &event.NotificationID, &event.EventType, &eventDataJSON,
			&event.ProviderEventID, &event.OccurredAt, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery event: %w", err)
		}

		if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating delivery events: %w", err)
	}

	return events, nil
}

// AddToRetryQueue adds a notification to the retry queue
func (r *PostgresNotificationRepository) AddToRetryQueue(ctx context.Context, retryEntry *models.NotificationRetryQueue) error {
	retryEntry.ID = uuid.New().String()
	retryEntry.CreatedAt = time.Now()

	query := `
		INSERT INTO notification_retry_queue (id, notification_id, retry_attempt, scheduled_for, error_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		retryEntry.ID, retryEntry.NotificationID, retryEntry.RetryAttempt,
		retryEntry.ScheduledFor, retryEntry.ErrorReason, retryEntry.CreatedAt)
	if err != nil {
		r.logger.Error(ctx, "Failed to add to retry queue", err, map[string]interface{}{
			"notification_id": retryEntry.NotificationID,
			"retry_attempt":   retryEntry.RetryAttempt,
		})
		return fmt.Errorf("failed to add to retry queue: %w", err)
	}

	r.logger.Info(ctx, "Added to retry queue", map[string]interface{}{
		"retry_id":        retryEntry.ID,
		"notification_id": retryEntry.NotificationID,
		"retry_attempt":   retryEntry.RetryAttempt,
	})

	return nil
}

// GetRetryQueueEntries retrieves entries from the retry queue that are ready for processing
func (r *PostgresNotificationRepository) GetRetryQueueEntries(ctx context.Context, limit int) ([]*models.NotificationRetryQueue, error) {
	query := `
		SELECT id, notification_id, retry_attempt, scheduled_for, error_reason, created_at, processed_at
		FROM notification_retry_queue
		WHERE processed_at IS NULL AND scheduled_for <= $1
		ORDER BY scheduled_for ASC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, time.Now(), limit)
	if err != nil {
		r.logger.Error(ctx, "Failed to get retry queue entries", err)
		return nil, fmt.Errorf("failed to get retry queue entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.NotificationRetryQueue
	for rows.Next() {
		entry := &models.NotificationRetryQueue{}

		err := rows.Scan(
			&entry.ID, &entry.NotificationID, &entry.RetryAttempt, &entry.ScheduledFor,
			&entry.ErrorReason, &entry.CreatedAt, &entry.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retry queue entry: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating retry queue entries: %w", err)
	}

	return entries, nil
}

// MarkRetryProcessed marks a retry queue entry as processed
func (r *PostgresNotificationRepository) MarkRetryProcessed(ctx context.Context, id string) error {
	query := `
		UPDATE notification_retry_queue
		SET processed_at = $2
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		r.logger.Error(ctx, "Failed to mark retry processed", err, map[string]interface{}{
			"retry_id": id,
		})
		return fmt.Errorf("failed to mark retry processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("retry queue entry not found")
	}

	return nil
}
