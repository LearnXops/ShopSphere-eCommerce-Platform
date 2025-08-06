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

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus, reason string, changedBy string) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Order, error)
	GetStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error)
}

// PostgresOrderRepository implements OrderRepository using PostgreSQL
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository creates a new PostgreSQL order repository
func NewPostgresOrderRepository(db *sql.DB) OrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Create creates a new order
func (r *PostgresOrderRepository) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate order number if not set
	if order.OrderNumber == "" {
		order.OrderNumber = generateOrderNumber()
	}

	// Marshal addresses and payment method to JSON
	shippingAddr, _ := json.Marshal(order.ShippingAddress)
	billingAddr, _ := json.Marshal(order.BillingAddress)
	paymentMethod, _ := json.Marshal(order.PaymentMethod)

	// Insert order
	query := `
		INSERT INTO orders (
			id, order_number, user_id, status, subtotal, tax, shipping, discount, total, currency,
			shipping_address, billing_address, payment_method, payment_status, payment_reference,
			shipping_method, tracking_number, estimated_delivery_date, notes, source, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`

	_, err = tx.ExecContext(ctx, query,
		order.ID, order.OrderNumber, order.UserID, order.Status, order.Subtotal, order.Tax,
		order.Shipping, order.Discount, order.Total, order.Currency,
		shippingAddr, billingAddr, paymentMethod, order.PaymentStatus, order.PaymentReference,
		order.ShippingMethod, order.TrackingNumber, order.EstimatedDeliveryDate, order.Notes,
		order.Source, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	for _, item := range order.Items {
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.OrderID = order.ID

		attrs, _ := json.Marshal(item.ProductAttributes)
		itemQuery := `
			INSERT INTO order_items (
				id, order_id, product_id, variant_id, sku, name, description, price, quantity, total,
				product_attributes, image_url, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID, item.OrderID, item.ProductID, item.VariantID, item.SKU, item.Name,
			item.Description, item.Price, item.Quantity, item.Total, attrs, item.ImageURL, item.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Record initial status
	if err := r.recordStatusChange(ctx, tx, order.ID, "", string(order.Status), "Order created", order.UserID); err != nil {
		return fmt.Errorf("failed to record status change: %w", err)
	}

	return tx.Commit()
}

// GetByID retrieves an order by ID
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	query := `
		SELECT id, order_number, user_id, status, subtotal, tax, shipping, discount, total, currency,
			   shipping_address, billing_address, payment_method, payment_status, payment_reference,
			   shipping_method, tracking_number, estimated_delivery_date, actual_delivery_date,
			   notes, internal_notes, source, confirmed_at, shipped_at, delivered_at, cancelled_at,
			   created_at, updated_at
		FROM orders WHERE id = $1`

	var order models.Order
	var shippingAddr, billingAddr, paymentMethod []byte
	var confirmedAt, shippedAt, deliveredAt, cancelledAt sql.NullTime
	var actualDeliveryDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.Status, &order.Subtotal,
		&order.Tax, &order.Shipping, &order.Discount, &order.Total, &order.Currency,
		&shippingAddr, &billingAddr, &paymentMethod, &order.PaymentStatus, &order.PaymentReference,
		&order.ShippingMethod, &order.TrackingNumber, &order.EstimatedDeliveryDate, &actualDeliveryDate,
		&order.Notes, &order.InternalNotes, &order.Source, &confirmedAt, &shippedAt,
		&deliveredAt, &cancelledAt, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Unmarshal JSON fields
	json.Unmarshal(shippingAddr, &order.ShippingAddress)
	json.Unmarshal(billingAddr, &order.BillingAddress)
	json.Unmarshal(paymentMethod, &order.PaymentMethod)

	// Handle nullable timestamps
	if confirmedAt.Valid {
		order.ConfirmedAt = &confirmedAt.Time
	}
	if shippedAt.Valid {
		order.ShippedAt = &shippedAt.Time
	}
	if deliveredAt.Valid {
		order.DeliveredAt = &deliveredAt.Time
	}
	if cancelledAt.Valid {
		order.CancelledAt = &cancelledAt.Time
	}
	if actualDeliveryDate.Valid {
		order.ActualDeliveryDate = &actualDeliveryDate.Time
	}

	// Load order items
	items, err := r.getOrderItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}
	order.Items = items

	return &order, nil
}

// GetByUserID retrieves orders for a specific user
func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, order_number, user_id, status, subtotal, tax, shipping, discount, total, currency,
			   shipping_address, billing_address, payment_method, payment_status, payment_reference,
			   shipping_method, tracking_number, estimated_delivery_date, actual_delivery_date,
			   notes, internal_notes, source, confirmed_at, shipped_at, delivered_at, cancelled_at,
			   created_at, updated_at
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var shippingAddr, billingAddr, paymentMethod []byte
		var confirmedAt, shippedAt, deliveredAt, cancelledAt sql.NullTime
		var actualDeliveryDate sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.Status, &order.Subtotal,
			&order.Tax, &order.Shipping, &order.Discount, &order.Total, &order.Currency,
			&shippingAddr, &billingAddr, &paymentMethod, &order.PaymentStatus, &order.PaymentReference,
			&order.ShippingMethod, &order.TrackingNumber, &order.EstimatedDeliveryDate, &actualDeliveryDate,
			&order.Notes, &order.InternalNotes, &order.Source, &confirmedAt, &shippedAt,
			&deliveredAt, &cancelledAt, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal JSON fields
		json.Unmarshal(shippingAddr, &order.ShippingAddress)
		json.Unmarshal(billingAddr, &order.BillingAddress)
		json.Unmarshal(paymentMethod, &order.PaymentMethod)

		// Handle nullable timestamps
		if confirmedAt.Valid {
			order.ConfirmedAt = &confirmedAt.Time
		}
		if shippedAt.Valid {
			order.ShippedAt = &shippedAt.Time
		}
		if deliveredAt.Valid {
			order.DeliveredAt = &deliveredAt.Time
		}
		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}
		if actualDeliveryDate.Valid {
			order.ActualDeliveryDate = &actualDeliveryDate.Time
		}

		// Load order items
		items, err := r.getOrderItems(ctx, order.ID)
		if err != nil {
			utils.Logger.Error(ctx, "Failed to load items for order", err, map[string]interface{}{"order_id": order.ID})
			continue
		}
		order.Items = items

		orders = append(orders, &order)
	}

	return orders, nil
}

// Update updates an existing order
func (r *PostgresOrderRepository) Update(ctx context.Context, order *models.Order) error {
	order.UpdatedAt = time.Now()

	shippingAddr, _ := json.Marshal(order.ShippingAddress)
	billingAddr, _ := json.Marshal(order.BillingAddress)
	paymentMethod, _ := json.Marshal(order.PaymentMethod)

	query := `
		UPDATE orders SET
			status = $2, subtotal = $3, tax = $4, shipping = $5, discount = $6, total = $7,
			shipping_address = $8, billing_address = $9, payment_method = $10, payment_status = $11,
			payment_reference = $12, shipping_method = $13, tracking_number = $14,
			estimated_delivery_date = $15, notes = $16, updated_at = $17
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.Status, order.Subtotal, order.Tax, order.Shipping, order.Discount,
		order.Total, shippingAddr, billingAddr, paymentMethod, order.PaymentStatus,
		order.PaymentReference, order.ShippingMethod, order.TrackingNumber,
		order.EstimatedDeliveryDate, order.Notes, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// UpdateStatus updates order status and records the change
func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus, reason string, changedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current status
	var currentStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1", orderID).Scan(&currentStatus)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	// Update order status and timestamp
	now := time.Now()
	var timestampField string
	switch status {
	case models.OrderConfirmed:
		timestampField = "confirmed_at"
	case models.OrderShipped:
		timestampField = "shipped_at"
	case models.OrderDelivered:
		timestampField = "delivered_at"
	case models.OrderCancelled:
		timestampField = "cancelled_at"
	}

	query := "UPDATE orders SET status = $1, updated_at = $2"
	args := []interface{}{status, now}
	if timestampField != "" {
		query += ", " + timestampField + " = $3"
		args = append(args, now)
	}
	query += " WHERE id = $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, orderID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Record status change
	if err := r.recordStatusChange(ctx, tx, orderID, currentStatus, string(status), reason, changedBy); err != nil {
		return fmt.Errorf("failed to record status change: %w", err)
	}

	return tx.Commit()
}

// Delete soft deletes an order (marks as cancelled)
func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.OrderCancelled, "Order deleted", "system")
}

// Search searches orders based on filters
func (r *PostgresOrderRepository) Search(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, order_number, user_id, status, subtotal, tax, shipping, discount, total, currency,
			   shipping_address, billing_address, payment_method, payment_status, payment_reference,
			   shipping_method, tracking_number, estimated_delivery_date, actual_delivery_date,
			   notes, internal_notes, source, confirmed_at, shipped_at, delivered_at, cancelled_at,
			   created_at, updated_at
		FROM orders WHERE 1=1`

	var args []interface{}
	argCount := 0

	// Add filters
	if userID, ok := filters["user_id"]; ok {
		argCount++
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, userID)
	}

	if status, ok := filters["status"]; ok {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if orderNumber, ok := filters["order_number"]; ok {
		argCount++
		query += fmt.Sprintf(" AND order_number ILIKE $%d", argCount)
		args = append(args, "%"+orderNumber.(string)+"%")
	}

	if dateFrom, ok := filters["date_from"]; ok {
		argCount++
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, dateFrom)
	}

	if dateTo, ok := filters["date_to"]; ok {
		argCount++
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, dateTo)
	}

	query += " ORDER BY created_at DESC"
	
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)
	
	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var shippingAddr, billingAddr, paymentMethod []byte
		var confirmedAt, shippedAt, deliveredAt, cancelledAt sql.NullTime
		var actualDeliveryDate sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.UserID, &order.Status, &order.Subtotal,
			&order.Tax, &order.Shipping, &order.Discount, &order.Total, &order.Currency,
			&shippingAddr, &billingAddr, &paymentMethod, &order.PaymentStatus, &order.PaymentReference,
			&order.ShippingMethod, &order.TrackingNumber, &order.EstimatedDeliveryDate, &actualDeliveryDate,
			&order.Notes, &order.InternalNotes, &order.Source, &confirmedAt, &shippedAt,
			&deliveredAt, &cancelledAt, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal JSON fields
		json.Unmarshal(shippingAddr, &order.ShippingAddress)
		json.Unmarshal(billingAddr, &order.BillingAddress)
		json.Unmarshal(paymentMethod, &order.PaymentMethod)

		// Handle nullable timestamps
		if confirmedAt.Valid {
			order.ConfirmedAt = &confirmedAt.Time
		}
		if shippedAt.Valid {
			order.ShippedAt = &shippedAt.Time
		}
		if deliveredAt.Valid {
			order.DeliveredAt = &deliveredAt.Time
		}
		if cancelledAt.Valid {
			order.CancelledAt = &cancelledAt.Time
		}
		if actualDeliveryDate.Valid {
			order.ActualDeliveryDate = &actualDeliveryDate.Time
		}

		// Load order items
		items, err := r.getOrderItems(ctx, order.ID)
		if err != nil {
			utils.Logger.Error(ctx, "Failed to load items for order", err, map[string]interface{}{"order_id": order.ID})
			continue
		}
		order.Items = items

		orders = append(orders, &order)
	}

	return orders, nil
}

// GetStatusHistory retrieves the status change history for an order
func (r *PostgresOrderRepository) GetStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error) {
	query := `
		SELECT id, order_id, from_status, to_status, reason, notes, changed_by, created_at
		FROM order_status_history
		WHERE order_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query status history: %w", err)
	}
	defer rows.Close()

	var history []models.OrderStatusHistory
	for rows.Next() {
		var h models.OrderStatusHistory
		var fromStatus sql.NullString

		err := rows.Scan(&h.ID, &h.OrderID, &fromStatus, &h.ToStatus, &h.Reason, &h.Notes, &h.ChangedBy, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status history: %w", err)
		}

		if fromStatus.Valid {
			h.FromStatus = fromStatus.String
		}

		history = append(history, h)
	}

	return history, nil
}

// Helper methods

func (r *PostgresOrderRepository) getOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, variant_id, sku, name, description, price, quantity, total,
			   product_attributes, image_url, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		var variantID sql.NullString
		var description sql.NullString
		var attrs []byte
		var imageURL sql.NullString

		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &variantID, &item.SKU, &item.Name,
			&description, &item.Price, &item.Quantity, &item.Total, &attrs, &imageURL, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		if variantID.Valid {
			item.VariantID = variantID.String
		}
		if description.Valid {
			item.Description = description.String
		}
		if imageURL.Valid {
			item.ImageURL = imageURL.String
		}

		// Unmarshal product attributes
		if len(attrs) > 0 {
			json.Unmarshal(attrs, &item.ProductAttributes)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *PostgresOrderRepository) recordStatusChange(ctx context.Context, tx *sql.Tx, orderID, fromStatus, toStatus, reason, changedBy string) error {
	query := `
		INSERT INTO order_status_history (id, order_id, from_status, to_status, reason, changed_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var fromStatusPtr *string
	if fromStatus != "" {
		fromStatusPtr = &fromStatus
	}

	_, err := tx.ExecContext(ctx, query,
		uuid.New().String(), orderID, fromStatusPtr, toStatus, reason, changedBy, time.Now(),
	)
	return err
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}
