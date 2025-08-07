package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
)

// System Metrics
func (r *PostgresAdminRepository) CreateSystemMetric(ctx context.Context, metric *models.SystemMetric) error {
	query := `
		INSERT INTO system_metrics (id, metric_name, metric_value, metric_type, tags, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`
	
	err := r.db.QueryRowContext(ctx, query,
		metric.ID,
		metric.MetricName,
		metric.MetricValue,
		metric.MetricType,
		metric.Tags,
		metric.RecordedAt,
	).Scan(&metric.CreatedAt)
	
	if err != nil {
		r.logger.Error("Failed to create system metric", "error", err)
		return fmt.Errorf("failed to create system metric: %w", err)
	}
	
	return nil
}

func (r *PostgresAdminRepository) GetSystemMetrics(ctx context.Context, metricNames []string, from, to *time.Time) ([]*models.SystemMetric, error) {
	query := "SELECT id, metric_name, metric_value, metric_type, tags, recorded_at, created_at FROM system_metrics WHERE 1=1"
	args := []interface{}{}
	argIndex := 1
	
	if len(metricNames) > 0 {
		placeholders := make([]string, len(metricNames))
		for i, name := range metricNames {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, name)
			argIndex++
		}
		query += fmt.Sprintf(" AND metric_name IN (%s)", strings.Join(placeholders, ","))
	}
	
	if from != nil {
		query += fmt.Sprintf(" AND recorded_at >= $%d", argIndex)
		args = append(args, *from)
		argIndex++
	}
	
	if to != nil {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argIndex)
		args = append(args, *to)
		argIndex++
	}
	
	query += " ORDER BY recorded_at DESC"
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get system metrics", "error", err)
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}
	defer rows.Close()
	
	var metrics []*models.SystemMetric
	for rows.Next() {
		metric := &models.SystemMetric{}
		
		err := rows.Scan(
			&metric.ID,
			&metric.MetricName,
			&metric.MetricValue,
			&metric.MetricType,
			&metric.Tags,
			&metric.RecordedAt,
			&metric.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan system metric", "error", err)
			return nil, fmt.Errorf("failed to scan system metric: %w", err)
		}
		
		metrics = append(metrics, metric)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating system metrics: %w", err)
	}
	
	return metrics, nil
}

func (r *PostgresAdminRepository) GetLatestMetrics(ctx context.Context, metricNames []string) ([]*models.SystemMetric, error) {
	if len(metricNames) == 0 {
		return []*models.SystemMetric{}, nil
	}
	
	placeholders := make([]string, len(metricNames))
	args := make([]interface{}, len(metricNames))
	for i, name := range metricNames {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = name
	}
	
	query := fmt.Sprintf(`
		SELECT DISTINCT ON (metric_name) id, metric_name, metric_value, metric_type, tags, recorded_at, created_at
		FROM system_metrics
		WHERE metric_name IN (%s)
		ORDER BY metric_name, recorded_at DESC`, strings.Join(placeholders, ","))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get latest metrics", "error", err)
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}
	defer rows.Close()
	
	var metrics []*models.SystemMetric
	for rows.Next() {
		metric := &models.SystemMetric{}
		
		err := rows.Scan(
			&metric.ID,
			&metric.MetricName,
			&metric.MetricValue,
			&metric.MetricType,
			&metric.Tags,
			&metric.RecordedAt,
			&metric.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan latest metric", "error", err)
			return nil, fmt.Errorf("failed to scan latest metric: %w", err)
		}
		
		metrics = append(metrics, metric)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating latest metrics: %w", err)
	}
	
	return metrics, nil
}

func (r *PostgresAdminRepository) UpdateSystemMetric(ctx context.Context, metricName string, value float64, tags map[string]interface{}) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	
	query := `
		INSERT INTO system_metrics (id, metric_name, metric_value, metric_type, tags, recorded_at)
		VALUES ($1, $2, $3, 'gauge', $4, $5)
		ON CONFLICT (metric_name) 
		DO UPDATE SET metric_value = $3, tags = $4, recorded_at = $5, created_at = CURRENT_TIMESTAMP`
	
	_, err = r.db.ExecContext(ctx, query,
		uuid.New(),
		metricName,
		value,
		tagsJSON,
		time.Now(),
	)
	
	if err != nil {
		r.logger.Error("Failed to update system metric", "error", err, "metric_name", metricName)
		return fmt.Errorf("failed to update system metric: %w", err)
	}
	
	return nil
}

// Dashboard Configs
func (r *PostgresAdminRepository) CreateDashboardConfig(ctx context.Context, config *models.DashboardConfig) error {
	query := `
		INSERT INTO dashboard_configs (id, admin_user_id, dashboard_name, config, is_default)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		config.ID,
		config.AdminUserID,
		config.DashboardName,
		config.Config,
		config.IsDefault,
	).Scan(&config.CreatedAt, &config.UpdatedAt)
	
	if err != nil {
		r.logger.Error("Failed to create dashboard config", "error", err)
		return fmt.Errorf("failed to create dashboard config: %w", err)
	}
	
	return nil
}

func (r *PostgresAdminRepository) GetDashboardConfig(ctx context.Context, adminUserID uuid.UUID, dashboardName string) (*models.DashboardConfig, error) {
	query := `
		SELECT id, admin_user_id, dashboard_name, config, is_default, created_at, updated_at
		FROM dashboard_configs
		WHERE admin_user_id = $1 AND dashboard_name = $2`
	
	config := &models.DashboardConfig{}
	
	err := r.db.QueryRowContext(ctx, query, adminUserID, dashboardName).Scan(
		&config.ID,
		&config.AdminUserID,
		&config.DashboardName,
		&config.Config,
		&config.IsDefault,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get dashboard config", "error", err)
		return nil, fmt.Errorf("failed to get dashboard config: %w", err)
	}
	
	return config, nil
}

func (r *PostgresAdminRepository) UpdateDashboardConfig(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1
	
	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	query := fmt.Sprintf("UPDATE dashboard_configs SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)
	
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update dashboard config", "error", err, "id", id)
		return fmt.Errorf("failed to update dashboard config: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("dashboard config not found")
	}
	
	return nil
}

func (r *PostgresAdminRepository) DeleteDashboardConfig(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM dashboard_configs WHERE id = $1"
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete dashboard config", "error", err, "id", id)
		return fmt.Errorf("failed to delete dashboard config: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("dashboard config not found")
	}
	
	return nil
}

func (r *PostgresAdminRepository) ListDashboardConfigs(ctx context.Context, adminUserID uuid.UUID) ([]*models.DashboardConfig, error) {
	query := `
		SELECT id, admin_user_id, dashboard_name, config, is_default, created_at, updated_at
		FROM dashboard_configs
		WHERE admin_user_id = $1
		ORDER BY is_default DESC, dashboard_name ASC`
	
	rows, err := r.db.QueryContext(ctx, query, adminUserID)
	if err != nil {
		r.logger.Error("Failed to list dashboard configs", "error", err)
		return nil, fmt.Errorf("failed to list dashboard configs: %w", err)
	}
	defer rows.Close()
	
	var configs []*models.DashboardConfig
	for rows.Next() {
		config := &models.DashboardConfig{}
		
		err := rows.Scan(
			&config.ID,
			&config.AdminUserID,
			&config.DashboardName,
			&config.Config,
			&config.IsDefault,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan dashboard config", "error", err)
			return nil, fmt.Errorf("failed to scan dashboard config: %w", err)
		}
		
		configs = append(configs, config)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dashboard configs: %w", err)
	}
	
	return configs, nil
}

// System Alerts
func (r *PostgresAdminRepository) CreateSystemAlert(ctx context.Context, alert *models.SystemAlert) error {
	query := `
		INSERT INTO system_alerts (id, alert_type, severity, title, message, source_service, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		alert.ID,
		alert.AlertType,
		alert.Severity,
		alert.Title,
		alert.Message,
		alert.SourceService,
		alert.Metadata,
	).Scan(&alert.CreatedAt, &alert.UpdatedAt)
	
	if err != nil {
		r.logger.Error("Failed to create system alert", "error", err)
		return fmt.Errorf("failed to create system alert: %w", err)
	}
	
	return nil
}

func (r *PostgresAdminRepository) GetSystemAlert(ctx context.Context, id uuid.UUID) (*models.SystemAlert, error) {
	query := `
		SELECT sa.id, sa.alert_type, sa.severity, sa.title, sa.message, sa.source_service, 
		       sa.metadata, sa.is_resolved, sa.resolved_by, sa.resolved_at, sa.created_at, sa.updated_at,
		       au.role, u.email, u.first_name, u.last_name
		FROM system_alerts sa
		LEFT JOIN admin_users au ON sa.resolved_by = au.id
		LEFT JOIN users u ON au.user_id = u.id
		WHERE sa.id = $1`
	
	alert := &models.SystemAlert{}
	var resolvedByUser *models.AdminUser
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID,
		&alert.AlertType,
		&alert.Severity,
		&alert.Title,
		&alert.Message,
		&alert.SourceService,
		&alert.Metadata,
		&alert.IsResolved,
		&alert.ResolvedBy,
		&alert.ResolvedAt,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get system alert", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get system alert: %w", err)
	}
	
	if alert.ResolvedBy != nil {
		resolvedByUser = &models.AdminUser{User: &models.User{}}
		alert.ResolvedByUser = resolvedByUser
	}
	
	return alert, nil
}

func (r *PostgresAdminRepository) UpdateSystemAlert(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1
	
	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	query := fmt.Sprintf("UPDATE system_alerts SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)
	
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update system alert", "error", err, "id", id)
		return fmt.Errorf("failed to update system alert: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("system alert not found")
	}
	
	return nil
}

func (r *PostgresAdminRepository) DeleteSystemAlert(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM system_alerts WHERE id = $1"
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete system alert", "error", err, "id", id)
		return fmt.Errorf("failed to delete system alert: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("system alert not found")
	}
	
	return nil
}

func (r *PostgresAdminRepository) ListSystemAlerts(ctx context.Context, severity *string, isResolved *bool, limit, offset int) ([]*models.SystemAlert, int64, error) {
	// Build query conditions
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1
	
	if severity != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("sa.severity = $%d", argIndex))
		args = append(args, *severity)
		argIndex++
	}
	
	if isResolved != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("sa.is_resolved = $%d", argIndex))
		args = append(args, *isResolved)
		argIndex++
	}
	
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}
	
	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM system_alerts sa %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count system alerts", "error", err)
		return nil, 0, fmt.Errorf("failed to count system alerts: %w", err)
	}
	
	// Get system alerts with pagination
	query := fmt.Sprintf(`
		SELECT sa.id, sa.alert_type, sa.severity, sa.title, sa.message, sa.source_service, 
		       sa.metadata, sa.is_resolved, sa.resolved_by, sa.resolved_at, sa.created_at, sa.updated_at,
		       au.role, u.email, u.first_name, u.last_name
		FROM system_alerts sa
		LEFT JOIN admin_users au ON sa.resolved_by = au.id
		LEFT JOIN users u ON au.user_id = u.id
		%s
		ORDER BY sa.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list system alerts", "error", err)
		return nil, 0, fmt.Errorf("failed to list system alerts: %w", err)
	}
	defer rows.Close()
	
	var alerts []*models.SystemAlert
	for rows.Next() {
		alert := &models.SystemAlert{}
		var resolvedByRole, resolvedByEmail, resolvedByFirstName, resolvedByLastName sql.NullString
		
		err := rows.Scan(
			&alert.ID,
			&alert.AlertType,
			&alert.Severity,
			&alert.Title,
			&alert.Message,
			&alert.SourceService,
			&alert.Metadata,
			&alert.IsResolved,
			&alert.ResolvedBy,
			&alert.ResolvedAt,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&resolvedByRole,
			&resolvedByEmail,
			&resolvedByFirstName,
			&resolvedByLastName,
		)
		if err != nil {
			r.logger.Error("Failed to scan system alert", "error", err)
			return nil, 0, fmt.Errorf("failed to scan system alert: %w", err)
		}
		
		if alert.ResolvedBy != nil && resolvedByRole.Valid {
			alert.ResolvedByUser = &models.AdminUser{
				Role: resolvedByRole.String,
				User: &models.User{
					Email:     resolvedByEmail.String,
					FirstName: resolvedByFirstName.String,
					LastName:  resolvedByLastName.String,
				},
			}
		}
		
		alerts = append(alerts, alert)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating system alerts: %w", err)
	}
	
	return alerts, total, nil
}
