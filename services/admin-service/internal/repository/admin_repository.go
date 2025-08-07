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
	"github.com/shopsphere/shared/utils"
)

type AdminRepository interface {
	// Admin Users
	CreateAdminUser(ctx context.Context, adminUser *models.AdminUser) error
	GetAdminUserByID(ctx context.Context, id uuid.UUID) (*models.AdminUser, error)
	GetAdminUserByUserID(ctx context.Context, userID uuid.UUID) (*models.AdminUser, error)
	UpdateAdminUser(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteAdminUser(ctx context.Context, id uuid.UUID) error
	ListAdminUsers(ctx context.Context, limit, offset int) ([]*models.AdminUser, int64, error)
	
	// Activity Logs
	CreateActivityLog(ctx context.Context, log *models.AdminActivityLog) error
	GetActivityLogs(ctx context.Context, adminUserID *uuid.UUID, limit, offset int) ([]*models.AdminActivityLog, int64, error)
	
	// System Metrics
	CreateSystemMetric(ctx context.Context, metric *models.SystemMetric) error
	GetSystemMetrics(ctx context.Context, metricNames []string, from, to *time.Time) ([]*models.SystemMetric, error)
	GetLatestMetrics(ctx context.Context, metricNames []string) ([]*models.SystemMetric, error)
	UpdateSystemMetric(ctx context.Context, metricName string, value float64, tags map[string]interface{}) error
	
	// Dashboard Configs
	CreateDashboardConfig(ctx context.Context, config *models.DashboardConfig) error
	GetDashboardConfig(ctx context.Context, adminUserID uuid.UUID, dashboardName string) (*models.DashboardConfig, error)
	UpdateDashboardConfig(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteDashboardConfig(ctx context.Context, id uuid.UUID) error
	ListDashboardConfigs(ctx context.Context, adminUserID uuid.UUID) ([]*models.DashboardConfig, error)
	
	// System Alerts
	CreateSystemAlert(ctx context.Context, alert *models.SystemAlert) error
	GetSystemAlert(ctx context.Context, id uuid.UUID) (*models.SystemAlert, error)
	UpdateSystemAlert(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteSystemAlert(ctx context.Context, id uuid.UUID) error
	ListSystemAlerts(ctx context.Context, severity *string, isResolved *bool, limit, offset int) ([]*models.SystemAlert, int64, error)
	
	// Bulk Operations
	CreateBulkOperation(ctx context.Context, operation *models.BulkOperation) error
	GetBulkOperation(ctx context.Context, id uuid.UUID) (*models.BulkOperation, error)
	UpdateBulkOperation(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	ListBulkOperations(ctx context.Context, adminUserID *uuid.UUID, status *string, limit, offset int) ([]*models.BulkOperation, int64, error)
	
	// Analytics
	GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error)
	GetRevenueData(ctx context.Context, days int) ([]*models.RevenueData, error)
	GetOrderStatusDistribution(ctx context.Context) ([]*models.OrderStatusDistribution, error)
	GetTopProducts(ctx context.Context, limit int) ([]*models.TopProduct, error)
	GetUserGrowthData(ctx context.Context, days int) ([]*models.UserGrowthData, error)
}

type PostgresAdminRepository struct {
	db     *sql.DB
	logger *utils.StructuredLogger
}

func NewPostgresAdminRepository(db *sql.DB, logger *utils.StructuredLogger) AdminRepository {
	return &PostgresAdminRepository{
		db:     db,
		logger: logger,
	}
}

// Admin Users
func (r *PostgresAdminRepository) CreateAdminUser(ctx context.Context, adminUser *models.AdminUser) error {
	query := `
		INSERT INTO admin_users (id, user_id, role, permissions, department)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`
	
	permissionsJSON, err := json.Marshal(adminUser.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	
	err = r.db.QueryRowContext(ctx, query,
		adminUser.ID,
		adminUser.UserID,
		adminUser.Role,
		permissionsJSON,
		adminUser.Department,
	).Scan(&adminUser.CreatedAt, &adminUser.UpdatedAt)
	
	if err != nil {
		r.logger.Error("Failed to create admin user", "error", err)
		return fmt.Errorf("failed to create admin user: %w", err)
	}
	
	return nil
}

func (r *PostgresAdminRepository) GetAdminUserByID(ctx context.Context, id uuid.UUID) (*models.AdminUser, error) {
	query := `
		SELECT au.id, au.user_id, au.role, au.permissions, au.department, 
		       au.last_login_at, au.created_at, au.updated_at,
		       u.email, u.first_name, u.last_name, u.status
		FROM admin_users au
		JOIN users u ON au.user_id = u.id
		WHERE au.id = $1`
	
	adminUser := &models.AdminUser{User: &models.User{}}
	var permissionsJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&adminUser.ID,
		&adminUser.UserID,
		&adminUser.Role,
		&permissionsJSON,
		&adminUser.Department,
		&adminUser.LastLoginAt,
		&adminUser.CreatedAt,
		&adminUser.UpdatedAt,
		&adminUser.User.Email,
		&adminUser.User.FirstName,
		&adminUser.User.LastName,
		&adminUser.User.Status,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get admin user by ID", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}
	
	if err := json.Unmarshal(permissionsJSON, &adminUser.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}
	
	adminUser.User.ID = adminUser.UserID
	return adminUser, nil
}

func (r *PostgresAdminRepository) GetAdminUserByUserID(ctx context.Context, userID uuid.UUID) (*models.AdminUser, error) {
	query := `
		SELECT au.id, au.user_id, au.role, au.permissions, au.department, 
		       au.last_login_at, au.created_at, au.updated_at,
		       u.email, u.first_name, u.last_name, u.status
		FROM admin_users au
		JOIN users u ON au.user_id = u.id
		WHERE au.user_id = $1`
	
	adminUser := &models.AdminUser{User: &models.User{}}
	var permissionsJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&adminUser.ID,
		&adminUser.UserID,
		&adminUser.Role,
		&permissionsJSON,
		&adminUser.Department,
		&adminUser.LastLoginAt,
		&adminUser.CreatedAt,
		&adminUser.UpdatedAt,
		&adminUser.User.Email,
		&adminUser.User.FirstName,
		&adminUser.User.LastName,
		&adminUser.User.Status,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get admin user by user ID", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}
	
	if err := json.Unmarshal(permissionsJSON, &adminUser.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}
	
	adminUser.User.ID = adminUser.UserID
	return adminUser, nil
}

func (r *PostgresAdminRepository) UpdateAdminUser(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1
	
	for field, value := range updates {
		if field == "permissions" {
			if permissions, ok := value.([]string); ok {
				permissionsJSON, err := json.Marshal(permissions)
				if err != nil {
					return fmt.Errorf("failed to marshal permissions: %w", err)
				}
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, permissionsJSON)
				argIndex++
			}
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}
	
	query := fmt.Sprintf("UPDATE admin_users SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)
	
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update admin user", "error", err, "id", id)
		return fmt.Errorf("failed to update admin user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("admin user not found")
	}
	
	return nil
}

func (r *PostgresAdminRepository) DeleteAdminUser(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM admin_users WHERE id = $1"
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete admin user", "error", err, "id", id)
		return fmt.Errorf("failed to delete admin user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("admin user not found")
	}
	
	return nil
}

func (r *PostgresAdminRepository) ListAdminUsers(ctx context.Context, limit, offset int) ([]*models.AdminUser, int64, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM admin_users"
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count admin users", "error", err)
		return nil, 0, fmt.Errorf("failed to count admin users: %w", err)
	}
	
	// Get admin users with pagination
	query := `
		SELECT au.id, au.user_id, au.role, au.permissions, au.department, 
		       au.last_login_at, au.created_at, au.updated_at,
		       u.email, u.first_name, u.last_name, u.status
		FROM admin_users au
		JOIN users u ON au.user_id = u.id
		ORDER BY au.created_at DESC
		LIMIT $1 OFFSET $2`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list admin users", "error", err)
		return nil, 0, fmt.Errorf("failed to list admin users: %w", err)
	}
	defer rows.Close()
	
	var adminUsers []*models.AdminUser
	for rows.Next() {
		adminUser := &models.AdminUser{User: &models.User{}}
		var permissionsJSON []byte
		
		err := rows.Scan(
			&adminUser.ID,
			&adminUser.UserID,
			&adminUser.Role,
			&permissionsJSON,
			&adminUser.Department,
			&adminUser.LastLoginAt,
			&adminUser.CreatedAt,
			&adminUser.UpdatedAt,
			&adminUser.User.Email,
			&adminUser.User.FirstName,
			&adminUser.User.LastName,
			&adminUser.User.Status,
		)
		if err != nil {
			r.logger.Error("Failed to scan admin user", "error", err)
			return nil, 0, fmt.Errorf("failed to scan admin user: %w", err)
		}
		
		if err := json.Unmarshal(permissionsJSON, &adminUser.Permissions); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
		
		adminUser.User.ID = adminUser.UserID
		adminUsers = append(adminUsers, adminUser)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating admin users: %w", err)
	}
	
	return adminUsers, total, nil
}

// Activity Logs
func (r *PostgresAdminRepository) CreateActivityLog(ctx context.Context, log *models.AdminActivityLog) error {
	query := `
		INSERT INTO admin_activity_logs (id, admin_user_id, action, resource_type, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`
	
	err := r.db.QueryRowContext(ctx, query,
		log.ID,
		log.AdminUserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.Details,
		log.IPAddress,
		log.UserAgent,
	).Scan(&log.CreatedAt)
	
	if err != nil {
		r.logger.Error("Failed to create activity log", "error", err)
		return fmt.Errorf("failed to create activity log: %w", err)
	}
	
	return nil
}

func (r *PostgresAdminRepository) GetActivityLogs(ctx context.Context, adminUserID *uuid.UUID, limit, offset int) ([]*models.AdminActivityLog, int64, error) {
	// Build query conditions
	whereClause := ""
	args := []interface{}{}
	argIndex := 1
	
	if adminUserID != nil {
		whereClause = "WHERE aal.admin_user_id = $1"
		args = append(args, *adminUserID)
		argIndex++
	}
	
	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_activity_logs aal %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count activity logs", "error", err)
		return nil, 0, fmt.Errorf("failed to count activity logs: %w", err)
	}
	
	// Get activity logs with pagination
	query := fmt.Sprintf(`
		SELECT aal.id, aal.admin_user_id, aal.action, aal.resource_type, aal.resource_id, 
		       aal.details, aal.ip_address, aal.user_agent, aal.created_at,
		       au.role, u.email, u.first_name, u.last_name
		FROM admin_activity_logs aal
		JOIN admin_users au ON aal.admin_user_id = au.id
		JOIN users u ON au.user_id = u.id
		%s
		ORDER BY aal.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get activity logs", "error", err)
		return nil, 0, fmt.Errorf("failed to get activity logs: %w", err)
	}
	defer rows.Close()
	
	var logs []*models.AdminActivityLog
	for rows.Next() {
		log := &models.AdminActivityLog{AdminUser: &models.AdminUser{User: &models.User{}}}
		
		err := rows.Scan(
			&log.ID,
			&log.AdminUserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.Details,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
			&log.AdminUser.Role,
			&log.AdminUser.User.Email,
			&log.AdminUser.User.FirstName,
			&log.AdminUser.User.LastName,
		)
		if err != nil {
			r.logger.Error("Failed to scan activity log", "error", err)
			return nil, 0, fmt.Errorf("failed to scan activity log: %w", err)
		}
		
		logs = append(logs, log)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating activity logs: %w", err)
	}
	
	return logs, total, nil
}
