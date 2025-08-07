package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
)

// Bulk Operations
func (r *PostgresAdminRepository) CreateBulkOperation(ctx context.Context, operation *models.BulkOperation) error {
	query := `
		INSERT INTO bulk_operations (id, admin_user_id, operation_type, resource_type, status, total_items, parameters)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		operation.ID,
		operation.AdminUserID,
		operation.OperationType,
		operation.ResourceType,
		operation.Status,
		operation.TotalItems,
		operation.Parameters,
	).Scan(&operation.CreatedAt, &operation.UpdatedAt)
	
	if err != nil {
		r.logger.Error("Failed to create bulk operation", "error", err)
		return fmt.Errorf("failed to create bulk operation: %w", err)
	}
	
	return nil
}

func (r *PostgresAdminRepository) GetBulkOperation(ctx context.Context, id uuid.UUID) (*models.BulkOperation, error) {
	query := `
		SELECT bo.id, bo.admin_user_id, bo.operation_type, bo.resource_type, bo.status, 
		       bo.total_items, bo.processed_items, bo.failed_items, bo.parameters, bo.results, 
		       bo.error_message, bo.started_at, bo.completed_at, bo.created_at, bo.updated_at,
		       au.role, u.email, u.first_name, u.last_name
		FROM bulk_operations bo
		JOIN admin_users au ON bo.admin_user_id = au.id
		JOIN users u ON au.user_id = u.id
		WHERE bo.id = $1`
	
	operation := &models.BulkOperation{AdminUser: &models.AdminUser{User: &models.User{}}}
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&operation.ID,
		&operation.AdminUserID,
		&operation.OperationType,
		&operation.ResourceType,
		&operation.Status,
		&operation.TotalItems,
		&operation.ProcessedItems,
		&operation.FailedItems,
		&operation.Parameters,
		&operation.Results,
		&operation.ErrorMessage,
		&operation.StartedAt,
		&operation.CompletedAt,
		&operation.CreatedAt,
		&operation.UpdatedAt,
		&operation.AdminUser.Role,
		&operation.AdminUser.User.Email,
		&operation.AdminUser.User.FirstName,
		&operation.AdminUser.User.LastName,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get bulk operation", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get bulk operation: %w", err)
	}
	
	return operation, nil
}

func (r *PostgresAdminRepository) UpdateBulkOperation(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
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
	
	query := fmt.Sprintf("UPDATE bulk_operations SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)
	
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update bulk operation", "error", err, "id", id)
		return fmt.Errorf("failed to update bulk operation: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("bulk operation not found")
	}
	
	return nil
}

func (r *PostgresAdminRepository) ListBulkOperations(ctx context.Context, adminUserID *uuid.UUID, status *string, limit, offset int) ([]*models.BulkOperation, int64, error) {
	// Build query conditions
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1
	
	if adminUserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("bo.admin_user_id = $%d", argIndex))
		args = append(args, *adminUserID)
		argIndex++
	}
	
	if status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("bo.status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}
	
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}
	
	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bulk_operations bo %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("Failed to count bulk operations", "error", err)
		return nil, 0, fmt.Errorf("failed to count bulk operations: %w", err)
	}
	
	// Get bulk operations with pagination
	query := fmt.Sprintf(`
		SELECT bo.id, bo.admin_user_id, bo.operation_type, bo.resource_type, bo.status, 
		       bo.total_items, bo.processed_items, bo.failed_items, bo.parameters, bo.results, 
		       bo.error_message, bo.started_at, bo.completed_at, bo.created_at, bo.updated_at,
		       au.role, u.email, u.first_name, u.last_name
		FROM bulk_operations bo
		JOIN admin_users au ON bo.admin_user_id = au.id
		JOIN users u ON au.user_id = u.id
		%s
		ORDER BY bo.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list bulk operations", "error", err)
		return nil, 0, fmt.Errorf("failed to list bulk operations: %w", err)
	}
	defer rows.Close()
	
	var operations []*models.BulkOperation
	for rows.Next() {
		operation := &models.BulkOperation{AdminUser: &models.AdminUser{User: &models.User{}}}
		
		err := rows.Scan(
			&operation.ID,
			&operation.AdminUserID,
			&operation.OperationType,
			&operation.ResourceType,
			&operation.Status,
			&operation.TotalItems,
			&operation.ProcessedItems,
			&operation.FailedItems,
			&operation.Parameters,
			&operation.Results,
			&operation.ErrorMessage,
			&operation.StartedAt,
			&operation.CompletedAt,
			&operation.CreatedAt,
			&operation.UpdatedAt,
			&operation.AdminUser.Role,
			&operation.AdminUser.User.Email,
			&operation.AdminUser.User.FirstName,
			&operation.AdminUser.User.LastName,
		)
		if err != nil {
			r.logger.Error("Failed to scan bulk operation", "error", err)
			return nil, 0, fmt.Errorf("failed to scan bulk operation: %w", err)
		}
		
		operations = append(operations, operation)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating bulk operations: %w", err)
	}
	
	return operations, total, nil
}

// Analytics
func (r *PostgresAdminRepository) GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error) {
	metrics := &models.DashboardMetrics{}
	
	// Get user count
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&metrics.TotalUsers)
	if err != nil {
		r.logger.Error("Failed to get user count", "error", err)
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}
	
	// Get product count
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE status = 'active'").Scan(&metrics.TotalProducts)
	if err != nil {
		r.logger.Error("Failed to get product count", "error", err)
		return nil, fmt.Errorf("failed to get product count: %w", err)
	}
	
	// Get order count
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&metrics.TotalOrders)
	if err != nil {
		r.logger.Error("Failed to get order count", "error", err)
		return nil, fmt.Errorf("failed to get order count: %w", err)
	}
	
	// Get daily revenue
	today := time.Now().Format("2006-01-02")
	err = r.db.QueryRowContext(ctx, 
		"SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE DATE(created_at) = $1 AND status IN ('completed', 'delivered')", 
		today).Scan(&metrics.DailyRevenue)
	if err != nil {
		r.logger.Error("Failed to get daily revenue", "error", err)
		return nil, fmt.Errorf("failed to get daily revenue: %w", err)
	}
	
	// Get monthly revenue
	firstDayOfMonth := time.Now().Format("2006-01-01")
	err = r.db.QueryRowContext(ctx, 
		"SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE DATE(created_at) >= $1 AND status IN ('completed', 'delivered')", 
		firstDayOfMonth).Scan(&metrics.MonthlyRevenue)
	if err != nil {
		r.logger.Error("Failed to get monthly revenue", "error", err)
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}
	
	// Get active sessions (mock data for now)
	metrics.ActiveSessions = 0
	
	// Get cart abandonment rate (mock calculation)
	var totalCarts, convertedCarts int64
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM carts WHERE created_at >= NOW() - INTERVAL '30 days'").Scan(&totalCarts)
	if err == nil && totalCarts > 0 {
		err = r.db.QueryRowContext(ctx, 
			"SELECT COUNT(DISTINCT c.user_id) FROM carts c JOIN orders o ON c.user_id = o.user_id WHERE c.created_at >= NOW() - INTERVAL '30 days'").Scan(&convertedCarts)
		if err == nil {
			metrics.CartAbandonmentRate = float64(totalCarts-convertedCarts) / float64(totalCarts) * 100
		}
	}
	
	// Get average order value
	err = r.db.QueryRowContext(ctx, 
		"SELECT COALESCE(AVG(total_amount), 0) FROM orders WHERE status IN ('completed', 'delivered')").Scan(&metrics.AverageOrderValue)
	if err != nil {
		r.logger.Error("Failed to get average order value", "error", err)
		return nil, fmt.Errorf("failed to get average order value: %w", err)
	}
	
	return metrics, nil
}

func (r *PostgresAdminRepository) GetRevenueData(ctx context.Context, days int) ([]*models.RevenueData, error) {
	query := `
		SELECT DATE(created_at) as date, COALESCE(SUM(total_amount), 0) as amount
		FROM orders
		WHERE created_at >= NOW() - INTERVAL '%d days'
		  AND status IN ('completed', 'delivered')
		GROUP BY DATE(created_at)
		ORDER BY date DESC`
	
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		r.logger.Error("Failed to get revenue data", "error", err)
		return nil, fmt.Errorf("failed to get revenue data: %w", err)
	}
	defer rows.Close()
	
	var revenueData []*models.RevenueData
	for rows.Next() {
		data := &models.RevenueData{}
		
		err := rows.Scan(&data.Date, &data.Amount)
		if err != nil {
			r.logger.Error("Failed to scan revenue data", "error", err)
			return nil, fmt.Errorf("failed to scan revenue data: %w", err)
		}
		
		revenueData = append(revenueData, data)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating revenue data: %w", err)
	}
	
	return revenueData, nil
}

func (r *PostgresAdminRepository) GetOrderStatusDistribution(ctx context.Context) ([]*models.OrderStatusDistribution, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM orders
		GROUP BY status
		ORDER BY count DESC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get order status distribution", "error", err)
		return nil, fmt.Errorf("failed to get order status distribution: %w", err)
	}
	defer rows.Close()
	
	var distribution []*models.OrderStatusDistribution
	for rows.Next() {
		data := &models.OrderStatusDistribution{}
		
		err := rows.Scan(&data.Status, &data.Count)
		if err != nil {
			r.logger.Error("Failed to scan order status distribution", "error", err)
			return nil, fmt.Errorf("failed to scan order status distribution: %w", err)
		}
		
		distribution = append(distribution, data)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order status distribution: %w", err)
	}
	
	return distribution, nil
}

func (r *PostgresAdminRepository) GetTopProducts(ctx context.Context, limit int) ([]*models.TopProduct, error) {
	query := `
		SELECT p.id, p.name, 
		       COALESCE(SUM(oi.quantity), 0) as sales,
		       COALESCE(SUM(oi.price * oi.quantity), 0) as revenue
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id
		WHERE o.status IN ('completed', 'delivered') OR o.status IS NULL
		GROUP BY p.id, p.name
		ORDER BY sales DESC, revenue DESC
		LIMIT $1`
	
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		r.logger.Error("Failed to get top products", "error", err)
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}
	defer rows.Close()
	
	var topProducts []*models.TopProduct
	for rows.Next() {
		product := &models.TopProduct{}
		
		err := rows.Scan(&product.ID, &product.Name, &product.Sales, &product.Revenue)
		if err != nil {
			r.logger.Error("Failed to scan top product", "error", err)
			return nil, fmt.Errorf("failed to scan top product: %w", err)
		}
		
		topProducts = append(topProducts, product)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top products: %w", err)
	}
	
	return topProducts, nil
}

func (r *PostgresAdminRepository) GetUserGrowthData(ctx context.Context, days int) ([]*models.UserGrowthData, error) {
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users
		WHERE created_at >= NOW() - INTERVAL '%d days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC`
	
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		r.logger.Error("Failed to get user growth data", "error", err)
		return nil, fmt.Errorf("failed to get user growth data: %w", err)
	}
	defer rows.Close()
	
	var growthData []*models.UserGrowthData
	for rows.Next() {
		data := &models.UserGrowthData{}
		
		err := rows.Scan(&data.Date, &data.Count)
		if err != nil {
			r.logger.Error("Failed to scan user growth data", "error", err)
			return nil, fmt.Errorf("failed to scan user growth data: %w", err)
		}
		
		growthData = append(growthData, data)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user growth data: %w", err)
	}
	
	return growthData, nil
}
