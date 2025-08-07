package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/utils"
)

// AdminUser represents an admin user with extended privileges
type AdminUser struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	Role         string          `json:"role" db:"role"`
	Permissions  []string        `json:"permissions" db:"permissions"`
	Department   *string         `json:"department,omitempty" db:"department"`
	LastLoginAt  *time.Time      `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
	
	// Joined fields from users table
	User         *User           `json:"user,omitempty"`
}

// AdminRole represents different admin roles
type AdminRole string

const (
	AdminRoleSuperAdmin    AdminRole = "super_admin"
	AdminRoleAdmin         AdminRole = "admin"
	AdminRoleModerator     AdminRole = "moderator"
	AdminRoleSupport       AdminRole = "support"
	AdminRoleAnalyst       AdminRole = "analyst"
)

// Permission represents admin permissions
type Permission string

const (
	// User permissions
	PermissionUsersRead   Permission = "users:read"
	PermissionUsersWrite  Permission = "users:write"
	PermissionUsersDelete Permission = "users:delete"
	
	// Product permissions
	PermissionProductsRead   Permission = "products:read"
	PermissionProductsWrite  Permission = "products:write"
	PermissionProductsDelete Permission = "products:delete"
	
	// Order permissions
	PermissionOrdersRead   Permission = "orders:read"
	PermissionOrdersWrite  Permission = "orders:write"
	PermissionOrdersDelete Permission = "orders:delete"
	
	// Analytics permissions
	PermissionAnalyticsRead Permission = "analytics:read"
	
	// System permissions
	PermissionSystemAdmin Permission = "system:admin"
)

// AdminActivityLog represents admin activity tracking
type AdminActivityLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	AdminUserID  uuid.UUID       `json:"admin_user_id" db:"admin_user_id"`
	Action       string          `json:"action" db:"action"`
	ResourceType string          `json:"resource_type" db:"resource_type"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty" db:"resource_id"`
	Details      json.RawMessage `json:"details,omitempty" db:"details"`
	IPAddress    *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string         `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	
	// Joined fields
	AdminUser    *AdminUser      `json:"admin_user,omitempty"`
}

// SystemMetric represents system metrics for dashboard
type SystemMetric struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	MetricName  string          `json:"metric_name" db:"metric_name"`
	MetricValue float64         `json:"metric_value" db:"metric_value"`
	MetricType  string          `json:"metric_type" db:"metric_type"`
	Tags        json.RawMessage `json:"tags,omitempty" db:"tags"`
	RecordedAt  time.Time       `json:"recorded_at" db:"recorded_at"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// DashboardConfig represents dashboard configuration for admin users
type DashboardConfig struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	AdminUserID   uuid.UUID       `json:"admin_user_id" db:"admin_user_id"`
	DashboardName string          `json:"dashboard_name" db:"dashboard_name"`
	Config        json.RawMessage `json:"config" db:"config"`
	IsDefault     bool            `json:"is_default" db:"is_default"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// SystemAlert represents system alerts for admin dashboard
type SystemAlert struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	AlertType     string     `json:"alert_type" db:"alert_type"`
	Severity      string     `json:"severity" db:"severity"`
	Title         string     `json:"title" db:"title"`
	Message       string     `json:"message" db:"message"`
	SourceService *string    `json:"source_service,omitempty" db:"source_service"`
	Metadata      json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	IsResolved    bool       `json:"is_resolved" db:"is_resolved"`
	ResolvedBy    *uuid.UUID `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	ResolvedByUser *AdminUser `json:"resolved_by_user,omitempty"`
}

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// BulkOperation represents bulk admin operations
type BulkOperation struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	AdminUserID     uuid.UUID       `json:"admin_user_id" db:"admin_user_id"`
	OperationType   string          `json:"operation_type" db:"operation_type"`
	ResourceType    string          `json:"resource_type" db:"resource_type"`
	Status          string          `json:"status" db:"status"`
	TotalItems      int             `json:"total_items" db:"total_items"`
	ProcessedItems  int             `json:"processed_items" db:"processed_items"`
	FailedItems     int             `json:"failed_items" db:"failed_items"`
	Parameters      json.RawMessage `json:"parameters,omitempty" db:"parameters"`
	Results         json.RawMessage `json:"results,omitempty" db:"results"`
	ErrorMessage    *string         `json:"error_message,omitempty" db:"error_message"`
	StartedAt       *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	AdminUser       *AdminUser      `json:"admin_user,omitempty"`
}

// BulkOperationStatus represents bulk operation status
type BulkOperationStatus string

const (
	BulkOperationStatusPending   BulkOperationStatus = "pending"
	BulkOperationStatusRunning   BulkOperationStatus = "running"
	BulkOperationStatusCompleted BulkOperationStatus = "completed"
	BulkOperationStatusFailed    BulkOperationStatus = "failed"
)

// Dashboard Analytics DTOs
type DashboardMetrics struct {
	TotalUsers        int64   `json:"total_users"`
	TotalProducts     int64   `json:"total_products"`
	TotalOrders       int64   `json:"total_orders"`
	DailyRevenue      float64 `json:"daily_revenue"`
	MonthlyRevenue    float64 `json:"monthly_revenue"`
	ActiveSessions    int64   `json:"active_sessions"`
	CartAbandonmentRate float64 `json:"cart_abandonment_rate"`
	AverageOrderValue float64 `json:"average_order_value"`
}

type RevenueData struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type OrderStatusDistribution struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type TopProduct struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Sales    int64     `json:"sales"`
	Revenue  float64   `json:"revenue"`
}

type UserGrowthData struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// Request/Response DTOs
type CreateAdminUserRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Role        string    `json:"role" validate:"required,oneof=super_admin admin moderator support analyst"`
	Permissions []string  `json:"permissions" validate:"required"`
	Department  *string   `json:"department,omitempty"`
}

type UpdateAdminUserRequest struct {
	Role        *string  `json:"role,omitempty" validate:"omitempty,oneof=super_admin admin moderator support analyst"`
	Permissions []string `json:"permissions,omitempty"`
	Department  *string  `json:"department,omitempty"`
}

type CreateSystemAlertRequest struct {
	AlertType     string          `json:"alert_type" validate:"required"`
	Severity      string          `json:"severity" validate:"required,oneof=critical warning info"`
	Title         string          `json:"title" validate:"required,max=200"`
	Message       string          `json:"message" validate:"required"`
	SourceService *string         `json:"source_service,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

type ResolveAlertRequest struct {
	ResolvedBy uuid.UUID `json:"resolved_by" validate:"required"`
}

type CreateBulkOperationRequest struct {
	OperationType string          `json:"operation_type" validate:"required"`
	ResourceType  string          `json:"resource_type" validate:"required"`
	Parameters    json.RawMessage `json:"parameters" validate:"required"`
}

type UpdateDashboardConfigRequest struct {
	Config    json.RawMessage `json:"config" validate:"required"`
	IsDefault *bool           `json:"is_default,omitempty"`
}

// Validation methods
func (r *CreateAdminUserRequest) Validate() error {
	// Validate permissions
	validPermissions := map[Permission]bool{
		PermissionUsersRead: true, PermissionUsersWrite: true, PermissionUsersDelete: true,
		PermissionProductsRead: true, PermissionProductsWrite: true, PermissionProductsDelete: true,
		PermissionOrdersRead: true, PermissionOrdersWrite: true, PermissionOrdersDelete: true,
		PermissionAnalyticsRead: true, PermissionSystemAdmin: true,
	}
	
	for _, perm := range r.Permissions {
		if !validPermissions[Permission(perm)] {
			return utils.NewValidationError("invalid permission: "+perm)
		}
	}
	
	return nil
}

func (r *UpdateAdminUserRequest) Validate() error {
	if r.Permissions != nil {
		validPermissions := map[Permission]bool{
			PermissionUsersRead: true, PermissionUsersWrite: true, PermissionUsersDelete: true,
			PermissionProductsRead: true, PermissionProductsWrite: true, PermissionProductsDelete: true,
			PermissionOrdersRead: true, PermissionOrdersWrite: true, PermissionOrdersDelete: true,
			PermissionAnalyticsRead: true, PermissionSystemAdmin: true,
		}
		
		for _, perm := range r.Permissions {
			if !validPermissions[Permission(perm)] {
				return utils.NewValidationError("invalid permission: "+perm)
			}
		}
	}
	
	return nil
}

// Helper methods
func (au *AdminUser) HasPermission(permission Permission) bool {
	for _, perm := range au.Permissions {
		if perm == string(permission) {
			return true
		}
	}
	return false
}

func (au *AdminUser) IsSuperAdmin() bool {
	return au.Role == string(AdminRoleSuperAdmin)
}

func (bo *BulkOperation) GetProgress() float64 {
	if bo.TotalItems == 0 {
		return 0.0
	}
	return float64(bo.ProcessedItems) / float64(bo.TotalItems) * 100.0
}
