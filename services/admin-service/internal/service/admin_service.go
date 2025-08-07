package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/admin-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

type AdminService interface {
	// Admin Users
	CreateAdminUser(ctx context.Context, req *models.CreateAdminUserRequest) (*models.AdminUser, error)
	GetAdminUser(ctx context.Context, id uuid.UUID) (*models.AdminUser, error)
	GetAdminUserByUserID(ctx context.Context, userID uuid.UUID) (*models.AdminUser, error)
	UpdateAdminUser(ctx context.Context, id uuid.UUID, req *models.UpdateAdminUserRequest) (*models.AdminUser, error)
	DeleteAdminUser(ctx context.Context, id uuid.UUID) error
	ListAdminUsers(ctx context.Context, page, limit int) ([]*models.AdminUser, int64, error)
	
	// Activity Logging
	LogActivity(ctx context.Context, adminUserID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent *string) error
	GetActivityLogs(ctx context.Context, adminUserID *uuid.UUID, page, limit int) ([]*models.AdminActivityLog, int64, error)
	
	// System Metrics
	RecordMetric(ctx context.Context, metricName string, value float64, metricType string, tags map[string]interface{}) error
	GetMetrics(ctx context.Context, metricNames []string, from, to *time.Time) ([]*models.SystemMetric, error)
	GetLatestMetrics(ctx context.Context, metricNames []string) ([]*models.SystemMetric, error)
	UpdateSystemMetrics(ctx context.Context) error
	
	// Dashboard Configs
	CreateDashboardConfig(ctx context.Context, adminUserID uuid.UUID, dashboardName string, config json.RawMessage, isDefault bool) (*models.DashboardConfig, error)
	GetDashboardConfig(ctx context.Context, adminUserID uuid.UUID, dashboardName string) (*models.DashboardConfig, error)
	UpdateDashboardConfig(ctx context.Context, id uuid.UUID, req *models.UpdateDashboardConfigRequest) (*models.DashboardConfig, error)
	DeleteDashboardConfig(ctx context.Context, id uuid.UUID) error
	ListDashboardConfigs(ctx context.Context, adminUserID uuid.UUID) ([]*models.DashboardConfig, error)
	
	// System Alerts
	CreateSystemAlert(ctx context.Context, req *models.CreateSystemAlertRequest) (*models.SystemAlert, error)
	GetSystemAlert(ctx context.Context, id uuid.UUID) (*models.SystemAlert, error)
	ResolveSystemAlert(ctx context.Context, id uuid.UUID, req *models.ResolveAlertRequest) (*models.SystemAlert, error)
	DeleteSystemAlert(ctx context.Context, id uuid.UUID) error
	ListSystemAlerts(ctx context.Context, severity *string, isResolved *bool, page, limit int) ([]*models.SystemAlert, int64, error)
	
	// Bulk Operations
	CreateBulkOperation(ctx context.Context, adminUserID uuid.UUID, req *models.CreateBulkOperationRequest) (*models.BulkOperation, error)
	GetBulkOperation(ctx context.Context, id uuid.UUID) (*models.BulkOperation, error)
	UpdateBulkOperationProgress(ctx context.Context, id uuid.UUID, processedItems, failedItems int, results json.RawMessage, errorMessage *string) error
	CompleteBulkOperation(ctx context.Context, id uuid.UUID, results json.RawMessage) error
	FailBulkOperation(ctx context.Context, id uuid.UUID, errorMessage string) error
	ListBulkOperations(ctx context.Context, adminUserID *uuid.UUID, status *string, page, limit int) ([]*models.BulkOperation, int64, error)
	
	// Analytics
	GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error)
	GetRevenueData(ctx context.Context, days int) ([]*models.RevenueData, error)
	GetOrderStatusDistribution(ctx context.Context) ([]*models.OrderStatusDistribution, error)
	GetTopProducts(ctx context.Context, limit int) ([]*models.TopProduct, error)
	GetUserGrowthData(ctx context.Context, days int) ([]*models.UserGrowthData, error)
	
	// Permission Checks
	HasPermission(ctx context.Context, adminUserID uuid.UUID, permission models.Permission) (bool, error)
	RequirePermission(ctx context.Context, adminUserID uuid.UUID, permission models.Permission) error
}

type adminService struct {
	repo   repository.AdminRepository
	logger *utils.StructuredLogger
}

func NewAdminService(repo repository.AdminRepository, logger *utils.StructuredLogger) AdminService {
	return &adminService{
		repo:   repo,
		logger: logger,
	}
}

// Admin Users
func (s *adminService) CreateAdminUser(ctx context.Context, req *models.CreateAdminUserRequest) (*models.AdminUser, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Check if admin user already exists for this user
	existingAdmin, err := s.repo.GetAdminUserByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing admin user: %w", err)
	}
	if existingAdmin != nil {
		return nil, fmt.Errorf("admin user already exists for user ID %s", req.UserID)
	}
	
	adminUser := &models.AdminUser{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Role:        req.Role,
		Permissions: req.Permissions,
		Department:  req.Department,
	}
	
	if err := s.repo.CreateAdminUser(ctx, adminUser); err != nil {
		s.logger.Error("Failed to create admin user", "error", err, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}
	
	s.logger.Info("Admin user created successfully", "admin_user_id", adminUser.ID, "user_id", req.UserID, "role", req.Role)
	return adminUser, nil
}

func (s *adminService) GetAdminUser(ctx context.Context, id uuid.UUID) (*models.AdminUser, error) {
	adminUser, err := s.repo.GetAdminUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}
	if adminUser == nil {
		return nil, fmt.Errorf("admin user not found")
	}
	
	return adminUser, nil
}

func (s *adminService) GetAdminUserByUserID(ctx context.Context, userID uuid.UUID) (*models.AdminUser, error) {
	adminUser, err := s.repo.GetAdminUserByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user by user ID: %w", err)
	}
	if adminUser == nil {
		return nil, fmt.Errorf("admin user not found")
	}
	
	return adminUser, nil
}

func (s *adminService) UpdateAdminUser(ctx context.Context, id uuid.UUID, req *models.UpdateAdminUserRequest) (*models.AdminUser, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	updates := make(map[string]interface{})
	
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Permissions != nil {
		updates["permissions"] = req.Permissions
	}
	if req.Department != nil {
		updates["department"] = *req.Department
	}
	
	if len(updates) == 0 {
		return s.GetAdminUser(ctx, id)
	}
	
	if err := s.repo.UpdateAdminUser(ctx, id, updates); err != nil {
		s.logger.Error("Failed to update admin user", "error", err, "admin_user_id", id)
		return nil, fmt.Errorf("failed to update admin user: %w", err)
	}
	
	s.logger.Info("Admin user updated successfully", "admin_user_id", id)
	return s.GetAdminUser(ctx, id)
}

func (s *adminService) DeleteAdminUser(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteAdminUser(ctx, id); err != nil {
		s.logger.Error("Failed to delete admin user", "error", err, "admin_user_id", id)
		return fmt.Errorf("failed to delete admin user: %w", err)
	}
	
	s.logger.Info("Admin user deleted successfully", "admin_user_id", id)
	return nil
}

func (s *adminService) ListAdminUsers(ctx context.Context, page, limit int) ([]*models.AdminUser, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	
	adminUsers, total, err := s.repo.ListAdminUsers(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list admin users", "error", err)
		return nil, 0, fmt.Errorf("failed to list admin users: %w", err)
	}
	
	return adminUsers, total, nil
}

// Activity Logging
func (s *adminService) LogActivity(ctx context.Context, adminUserID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details interface{}, ipAddress, userAgent *string) error {
	var detailsJSON json.RawMessage
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			s.logger.Error("Failed to marshal activity details", "error", err)
			return fmt.Errorf("failed to marshal activity details: %w", err)
		}
		detailsJSON = detailsBytes
	}
	
	log := &models.AdminActivityLog{
		ID:           uuid.New(),
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      detailsJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}
	
	if err := s.repo.CreateActivityLog(ctx, log); err != nil {
		s.logger.Error("Failed to create activity log", "error", err)
		return fmt.Errorf("failed to create activity log: %w", err)
	}
	
	return nil
}

func (s *adminService) GetActivityLogs(ctx context.Context, adminUserID *uuid.UUID, page, limit int) ([]*models.AdminActivityLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	
	logs, total, err := s.repo.GetActivityLogs(ctx, adminUserID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get activity logs", "error", err)
		return nil, 0, fmt.Errorf("failed to get activity logs: %w", err)
	}
	
	return logs, total, nil
}

// System Metrics
func (s *adminService) RecordMetric(ctx context.Context, metricName string, value float64, metricType string, tags map[string]interface{}) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	
	metric := &models.SystemMetric{
		ID:          uuid.New(),
		MetricName:  metricName,
		MetricValue: value,
		MetricType:  metricType,
		Tags:        tagsJSON,
		RecordedAt:  time.Now(),
	}
	
	if err := s.repo.CreateSystemMetric(ctx, metric); err != nil {
		s.logger.Error("Failed to record metric", "error", err, "metric_name", metricName)
		return fmt.Errorf("failed to record metric: %w", err)
	}
	
	return nil
}

func (s *adminService) GetMetrics(ctx context.Context, metricNames []string, from, to *time.Time) ([]*models.SystemMetric, error) {
	metrics, err := s.repo.GetSystemMetrics(ctx, metricNames, from, to)
	if err != nil {
		s.logger.Error("Failed to get metrics", "error", err)
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	
	return metrics, nil
}

func (s *adminService) GetLatestMetrics(ctx context.Context, metricNames []string) ([]*models.SystemMetric, error) {
	metrics, err := s.repo.GetLatestMetrics(ctx, metricNames)
	if err != nil {
		s.logger.Error("Failed to get latest metrics", "error", err)
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}
	
	return metrics, nil
}

func (s *adminService) UpdateSystemMetrics(ctx context.Context) error {
	// Get dashboard metrics and update system metrics table
	dashboardMetrics, err := s.repo.GetDashboardMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dashboard metrics: %w", err)
	}
	
	// Update individual metrics
	metrics := map[string]float64{
		"total_users":            float64(dashboardMetrics.TotalUsers),
		"total_products":         float64(dashboardMetrics.TotalProducts),
		"total_orders":           float64(dashboardMetrics.TotalOrders),
		"daily_revenue":          dashboardMetrics.DailyRevenue,
		"monthly_revenue":        dashboardMetrics.MonthlyRevenue,
		"active_sessions":        float64(dashboardMetrics.ActiveSessions),
		"cart_abandonment_rate":  dashboardMetrics.CartAbandonmentRate,
		"average_order_value":    dashboardMetrics.AverageOrderValue,
	}
	
	for metricName, value := range metrics {
		tags := map[string]interface{}{
			"updated_at": time.Now().Format(time.RFC3339),
		}
		
		if err := s.repo.UpdateSystemMetric(ctx, metricName, value, tags); err != nil {
			s.logger.Error("Failed to update system metric", "error", err, "metric_name", metricName)
			continue
		}
	}
	
	s.logger.Info("System metrics updated successfully")
	return nil
}
