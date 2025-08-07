package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
)

// Dashboard Configs
func (s *adminService) CreateDashboardConfig(ctx context.Context, adminUserID uuid.UUID, dashboardName string, config json.RawMessage, isDefault bool) (*models.DashboardConfig, error) {
	// Check if config already exists
	existingConfig, err := s.repo.GetDashboardConfig(ctx, adminUserID, dashboardName)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing dashboard config: %w", err)
	}
	if existingConfig != nil {
		return nil, fmt.Errorf("dashboard config with name '%s' already exists", dashboardName)
	}
	
	dashboardConfig := &models.DashboardConfig{
		ID:            uuid.New(),
		AdminUserID:   adminUserID,
		DashboardName: dashboardName,
		Config:        config,
		IsDefault:     isDefault,
	}
	
	if err := s.repo.CreateDashboardConfig(ctx, dashboardConfig); err != nil {
		s.logger.Error("Failed to create dashboard config", "error", err)
		return nil, fmt.Errorf("failed to create dashboard config: %w", err)
	}
	
	s.logger.Info("Dashboard config created successfully", "config_id", dashboardConfig.ID, "admin_user_id", adminUserID)
	return dashboardConfig, nil
}

func (s *adminService) GetDashboardConfig(ctx context.Context, adminUserID uuid.UUID, dashboardName string) (*models.DashboardConfig, error) {
	config, err := s.repo.GetDashboardConfig(ctx, adminUserID, dashboardName)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("dashboard config not found")
	}
	
	return config, nil
}

func (s *adminService) UpdateDashboardConfig(ctx context.Context, id uuid.UUID, req *models.UpdateDashboardConfigRequest) (*models.DashboardConfig, error) {
	updates := make(map[string]interface{})
	
	if req.Config != nil {
		updates["config"] = req.Config
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}
	
	if err := s.repo.UpdateDashboardConfig(ctx, id, updates); err != nil {
		s.logger.Error("Failed to update dashboard config", "error", err, "config_id", id)
		return nil, fmt.Errorf("failed to update dashboard config: %w", err)
	}
	
	s.logger.Info("Dashboard config updated successfully", "config_id", id)
	
	// Return updated config (we need to get it by ID, but we don't have that method)
	// For now, return nil and let the handler fetch it separately if needed
	return nil, nil
}

func (s *adminService) DeleteDashboardConfig(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteDashboardConfig(ctx, id); err != nil {
		s.logger.Error("Failed to delete dashboard config", "error", err, "config_id", id)
		return fmt.Errorf("failed to delete dashboard config: %w", err)
	}
	
	s.logger.Info("Dashboard config deleted successfully", "config_id", id)
	return nil
}

func (s *adminService) ListDashboardConfigs(ctx context.Context, adminUserID uuid.UUID) ([]*models.DashboardConfig, error) {
	configs, err := s.repo.ListDashboardConfigs(ctx, adminUserID)
	if err != nil {
		s.logger.Error("Failed to list dashboard configs", "error", err)
		return nil, fmt.Errorf("failed to list dashboard configs: %w", err)
	}
	
	return configs, nil
}

// System Alerts
func (s *adminService) CreateSystemAlert(ctx context.Context, req *models.CreateSystemAlertRequest) (*models.SystemAlert, error) {
	alert := &models.SystemAlert{
		ID:            uuid.New(),
		AlertType:     req.AlertType,
		Severity:      req.Severity,
		Title:         req.Title,
		Message:       req.Message,
		SourceService: req.SourceService,
		Metadata:      req.Metadata,
		IsResolved:    false,
	}
	
	if err := s.repo.CreateSystemAlert(ctx, alert); err != nil {
		s.logger.Error("Failed to create system alert", "error", err)
		return nil, fmt.Errorf("failed to create system alert: %w", err)
	}
	
	s.logger.Info("System alert created successfully", "alert_id", alert.ID, "severity", alert.Severity)
	return alert, nil
}

func (s *adminService) GetSystemAlert(ctx context.Context, id uuid.UUID) (*models.SystemAlert, error) {
	alert, err := s.repo.GetSystemAlert(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get system alert: %w", err)
	}
	if alert == nil {
		return nil, fmt.Errorf("system alert not found")
	}
	
	return alert, nil
}

func (s *adminService) ResolveSystemAlert(ctx context.Context, id uuid.UUID, req *models.ResolveAlertRequest) (*models.SystemAlert, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"is_resolved": true,
		"resolved_by": req.ResolvedBy,
		"resolved_at": now,
	}
	
	if err := s.repo.UpdateSystemAlert(ctx, id, updates); err != nil {
		s.logger.Error("Failed to resolve system alert", "error", err, "alert_id", id)
		return nil, fmt.Errorf("failed to resolve system alert: %w", err)
	}
	
	s.logger.Info("System alert resolved successfully", "alert_id", id, "resolved_by", req.ResolvedBy)
	return s.GetSystemAlert(ctx, id)
}

func (s *adminService) DeleteSystemAlert(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteSystemAlert(ctx, id); err != nil {
		s.logger.Error("Failed to delete system alert", "error", err, "alert_id", id)
		return fmt.Errorf("failed to delete system alert: %w", err)
	}
	
	s.logger.Info("System alert deleted successfully", "alert_id", id)
	return nil
}

func (s *adminService) ListSystemAlerts(ctx context.Context, severity *string, isResolved *bool, page, limit int) ([]*models.SystemAlert, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	
	alerts, total, err := s.repo.ListSystemAlerts(ctx, severity, isResolved, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list system alerts", "error", err)
		return nil, 0, fmt.Errorf("failed to list system alerts: %w", err)
	}
	
	return alerts, total, nil
}

// Bulk Operations
func (s *adminService) CreateBulkOperation(ctx context.Context, adminUserID uuid.UUID, req *models.CreateBulkOperationRequest) (*models.BulkOperation, error) {
	operation := &models.BulkOperation{
		ID:            uuid.New(),
		AdminUserID:   adminUserID,
		OperationType: req.OperationType,
		ResourceType:  req.ResourceType,
		Status:        string(models.BulkOperationStatusPending),
		TotalItems:    0, // Will be updated when operation starts
		Parameters:    req.Parameters,
	}
	
	if err := s.repo.CreateBulkOperation(ctx, operation); err != nil {
		s.logger.Error("Failed to create bulk operation", "error", err)
		return nil, fmt.Errorf("failed to create bulk operation: %w", err)
	}
	
	s.logger.Info("Bulk operation created successfully", "operation_id", operation.ID, "type", operation.OperationType)
	return operation, nil
}

func (s *adminService) GetBulkOperation(ctx context.Context, id uuid.UUID) (*models.BulkOperation, error) {
	operation, err := s.repo.GetBulkOperation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk operation: %w", err)
	}
	if operation == nil {
		return nil, fmt.Errorf("bulk operation not found")
	}
	
	return operation, nil
}

func (s *adminService) UpdateBulkOperationProgress(ctx context.Context, id uuid.UUID, processedItems, failedItems int, results json.RawMessage, errorMessage *string) error {
	updates := map[string]interface{}{
		"processed_items": processedItems,
		"failed_items":    failedItems,
		"status":          string(models.BulkOperationStatusRunning),
	}
	
	if results != nil {
		updates["results"] = results
	}
	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	}
	
	if err := s.repo.UpdateBulkOperation(ctx, id, updates); err != nil {
		s.logger.Error("Failed to update bulk operation progress", "error", err, "operation_id", id)
		return fmt.Errorf("failed to update bulk operation progress: %w", err)
	}
	
	return nil
}

func (s *adminService) CompleteBulkOperation(ctx context.Context, id uuid.UUID, results json.RawMessage) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       string(models.BulkOperationStatusCompleted),
		"completed_at": now,
	}
	
	if results != nil {
		updates["results"] = results
	}
	
	if err := s.repo.UpdateBulkOperation(ctx, id, updates); err != nil {
		s.logger.Error("Failed to complete bulk operation", "error", err, "operation_id", id)
		return fmt.Errorf("failed to complete bulk operation: %w", err)
	}
	
	s.logger.Info("Bulk operation completed successfully", "operation_id", id)
	return nil
}

func (s *adminService) FailBulkOperation(ctx context.Context, id uuid.UUID, errorMessage string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        string(models.BulkOperationStatusFailed),
		"error_message": errorMessage,
		"completed_at":  now,
	}
	
	if err := s.repo.UpdateBulkOperation(ctx, id, updates); err != nil {
		s.logger.Error("Failed to fail bulk operation", "error", err, "operation_id", id)
		return fmt.Errorf("failed to fail bulk operation: %w", err)
	}
	
	s.logger.Info("Bulk operation marked as failed", "operation_id", id, "error", errorMessage)
	return nil
}

func (s *adminService) ListBulkOperations(ctx context.Context, adminUserID *uuid.UUID, status *string, page, limit int) ([]*models.BulkOperation, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	
	operations, total, err := s.repo.ListBulkOperations(ctx, adminUserID, status, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list bulk operations", "error", err)
		return nil, 0, fmt.Errorf("failed to list bulk operations: %w", err)
	}
	
	return operations, total, nil
}

// Analytics
func (s *adminService) GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error) {
	metrics, err := s.repo.GetDashboardMetrics(ctx)
	if err != nil {
		s.logger.Error("Failed to get dashboard metrics", "error", err)
		return nil, fmt.Errorf("failed to get dashboard metrics: %w", err)
	}
	
	return metrics, nil
}

func (s *adminService) GetRevenueData(ctx context.Context, days int) ([]*models.RevenueData, error) {
	if days <= 0 || days > 365 {
		days = 30 // Default to 30 days
	}
	
	revenueData, err := s.repo.GetRevenueData(ctx, days)
	if err != nil {
		s.logger.Error("Failed to get revenue data", "error", err)
		return nil, fmt.Errorf("failed to get revenue data: %w", err)
	}
	
	return revenueData, nil
}

func (s *adminService) GetOrderStatusDistribution(ctx context.Context) ([]*models.OrderStatusDistribution, error) {
	distribution, err := s.repo.GetOrderStatusDistribution(ctx)
	if err != nil {
		s.logger.Error("Failed to get order status distribution", "error", err)
		return nil, fmt.Errorf("failed to get order status distribution: %w", err)
	}
	
	return distribution, nil
}

func (s *adminService) GetTopProducts(ctx context.Context, limit int) ([]*models.TopProduct, error) {
	if limit <= 0 || limit > 100 {
		limit = 10 // Default to top 10
	}
	
	topProducts, err := s.repo.GetTopProducts(ctx, limit)
	if err != nil {
		s.logger.Error("Failed to get top products", "error", err)
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}
	
	return topProducts, nil
}

func (s *adminService) GetUserGrowthData(ctx context.Context, days int) ([]*models.UserGrowthData, error) {
	if days <= 0 || days > 365 {
		days = 30 // Default to 30 days
	}
	
	growthData, err := s.repo.GetUserGrowthData(ctx, days)
	if err != nil {
		s.logger.Error("Failed to get user growth data", "error", err)
		return nil, fmt.Errorf("failed to get user growth data: %w", err)
	}
	
	return growthData, nil
}

// Permission Checks
func (s *adminService) HasPermission(ctx context.Context, adminUserID uuid.UUID, permission models.Permission) (bool, error) {
	adminUser, err := s.repo.GetAdminUserByID(ctx, adminUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get admin user: %w", err)
	}
	if adminUser == nil {
		return false, fmt.Errorf("admin user not found")
	}
	
	return adminUser.HasPermission(permission), nil
}

func (s *adminService) RequirePermission(ctx context.Context, adminUserID uuid.UUID, permission models.Permission) error {
	hasPermission, err := s.HasPermission(ctx, adminUserID, permission)
	if err != nil {
		return err
	}
	
	if !hasPermission {
		return fmt.Errorf("insufficient permissions: required %s", permission)
	}
	
	return nil
}
