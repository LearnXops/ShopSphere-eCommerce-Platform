package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/shopsphere/order-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// OrderService defines the interface for order business logic
type OrderService interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*models.Order, error)
	GetOrder(ctx context.Context, id string) (*models.Order, error)
	GetOrdersByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
	UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus, reason string, changedBy string) error
	CancelOrder(ctx context.Context, orderID string, reason string, cancelledBy string) error
	SearchOrders(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Order, error)
	GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error)
	ValidateOrderItems(ctx context.Context, items []OrderItemRequest) error
	CalculateOrderTotals(ctx context.Context, req *CreateOrderRequest) (*OrderTotals, error)
}

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	UserID          string              `json:"user_id" validate:"required"`
	Items           []OrderItemRequest  `json:"items" validate:"required"`
	ShippingAddress models.Address      `json:"shipping_address" validate:"required"`
	BillingAddress  models.Address      `json:"billing_address" validate:"required"`
	PaymentMethod   models.PaymentMethod `json:"payment_method" validate:"required"`
	ShippingMethod  string              `json:"shipping_method"`
	Notes           string              `json:"notes"`
	Source          string              `json:"source"`
}

// OrderItemRequest represents an item in an order request
type OrderItemRequest struct {
	ProductID string          `json:"product_id" validate:"required"`
	VariantID string          `json:"variant_id"`
	Quantity  int             `json:"quantity" validate:"required,min=1"`
	Price     decimal.Decimal `json:"price" validate:"required"`
}

// OrderTotals represents calculated order totals
type OrderTotals struct {
	Subtotal decimal.Decimal `json:"subtotal"`
	Tax      decimal.Decimal `json:"tax"`
	Shipping decimal.Decimal `json:"shipping"`
	Discount decimal.Decimal `json:"discount"`
	Total    decimal.Decimal `json:"total"`
}

// ProductService interface for product validation
type ProductService interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	ValidateStock(ctx context.Context, productID string, quantity int) error
}

// InventoryService interface for inventory management
type InventoryService interface {
	ReserveStock(ctx context.Context, items []models.OrderItem) error
	ReleaseStock(ctx context.Context, items []models.OrderItem) error
}

// orderService implements OrderService
type orderService struct {
	repo             repository.OrderRepository
	productService   ProductService
	inventoryService InventoryService
}

// NewOrderService creates a new order service
func NewOrderService(repo repository.OrderRepository, productService ProductService, inventoryService InventoryService) OrderService {
	return &orderService{
		repo:             repo,
		productService:   productService,
		inventoryService: inventoryService,
	}
}

// CreateOrder creates a new order
func (s *orderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*models.Order, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate order items
	if err := s.ValidateOrderItems(ctx, req.Items); err != nil {
		return nil, fmt.Errorf("invalid order items: %w", err)
	}

	// Calculate totals
	totals, err := s.CalculateOrderTotals(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate totals: %w", err)
	}

	// Create order
	order := &models.Order{
		ID:              uuid.New().String(),
		OrderNumber:     generateOrderNumber(),
		UserID:          req.UserID,
		Status:          models.OrderPending,
		Subtotal:        totals.Subtotal,
		Tax:             totals.Tax,
		Shipping:        totals.Shipping,
		Discount:        totals.Discount,
		Total:           totals.Total,
		Currency:        "USD",
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   "pending",
		ShippingMethod:  req.ShippingMethod,
		Notes:           req.Notes,
		Source:          req.Source,
		Items:           make([]models.OrderItem, 0, len(req.Items)),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if order.Source == "" {
		order.Source = "web"
	}

	// Convert request items to order items
	for _, itemReq := range req.Items {
		// Get product details
		product, err := s.productService.GetProduct(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %s: %w", itemReq.ProductID, err)
		}

		item := models.OrderItem{
			ID:          uuid.New().String(),
			OrderID:     order.ID,
			ProductID:   itemReq.ProductID,
			VariantID:   itemReq.VariantID,
			SKU:         product.SKU,
			Name:        product.Name,
			Description: product.Description,
			Price:       itemReq.Price,
			Quantity:    itemReq.Quantity,
			Total:       itemReq.Price.Mul(decimal.NewFromInt(int64(itemReq.Quantity))),
			CreatedAt:   time.Now(),
		}

		// Add product attributes snapshot
		attributesMap := map[string]interface{}{
			"brand":      product.Attributes.Brand,
			"color":      product.Attributes.Color,
			"size":       product.Attributes.Size,
			"weight":     product.Attributes.Weight,
			"dimensions": product.Attributes.Dimensions,
		}
		for k, v := range product.Attributes.Custom {
			attributesMap[k] = v
		}
		item.ProductAttributes = attributesMap

		// Add primary image
		if len(product.Images) > 0 {
			item.ImageURL = product.Images[0]
		}

		order.Items = append(order.Items, item)
	}

	// Reserve inventory
	if s.inventoryService != nil {
		if err := s.inventoryService.ReserveStock(ctx, order.Items); err != nil {
			return nil, fmt.Errorf("failed to reserve stock: %w", err)
		}
	}

	// Create order in database
	if err := s.repo.Create(ctx, order); err != nil {
		// Release reserved stock on failure
		if s.inventoryService != nil {
			s.inventoryService.ReleaseStock(ctx, order.Items)
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	utils.Logger.Info(ctx, "Order created successfully", nil, map[string]interface{}{
		"order_id":     order.ID,
		"order_number": order.OrderNumber,
		"user_id":      order.UserID,
		"total":        order.Total,
		"item_count":   len(order.Items),
	})

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *orderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	if id == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetOrdersByUser retrieves orders for a specific user
func (s *orderService) GetOrdersByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders for user: %w", err)
	}

	return orders, nil
}

// UpdateOrder updates an existing order
func (s *orderService) UpdateOrder(ctx context.Context, order *models.Order) error {
	if order == nil || order.ID == "" {
		return fmt.Errorf("invalid order")
	}

	// Validate that order exists
	existing, err := s.repo.GetByID(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Prevent updates to cancelled or completed orders
	if existing.Status == models.OrderCancelled || existing.Status == models.OrderDelivered {
		return fmt.Errorf("cannot update order in status: %s", existing.Status)
	}

	order.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	utils.Logger.Info(ctx, "Order updated successfully", nil, map[string]interface{}{
		"order_id": order.ID,
		"status":   order.Status,
	})

	return nil
}

// UpdateOrderStatus updates the status of an order
func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus, reason string, changedBy string) error {
	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}

	// Validate status transition
	if err := s.validateStatusTransition(ctx, orderID, status); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, orderID, status, reason, changedBy); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	utils.Logger.Info(ctx, "Order status updated", nil, map[string]interface{}{
		"order_id":   orderID,
		"new_status": status,
		"reason":     reason,
		"changed_by": changedBy,
	})

	return nil
}

// CancelOrder cancels an order
func (s *orderService) CancelOrder(ctx context.Context, orderID string, reason string, cancelledBy string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Check if order can be cancelled
	if order.Status == models.OrderCancelled {
		return fmt.Errorf("order is already cancelled")
	}
	if order.Status == models.OrderDelivered {
		return fmt.Errorf("cannot cancel delivered order")
	}
	if order.Status == models.OrderShipped {
		return fmt.Errorf("cannot cancel shipped order")
	}

	// Release reserved stock
	if s.inventoryService != nil {
		if err := s.inventoryService.ReleaseStock(ctx, order.Items); err != nil {
			utils.Logger.Error(ctx, "Failed to release stock for cancelled order", err, map[string]interface{}{
				"order_id": orderID,
			})
		}
	}

	// Update status to cancelled
	if err := s.repo.UpdateStatus(ctx, orderID, models.OrderCancelled, reason, cancelledBy); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	utils.Logger.Info(ctx, "Order cancelled", nil, map[string]interface{}{
		"order_id":     orderID,
		"reason":       reason,
		"cancelled_by": cancelledBy,
	})

	return nil
}

// SearchOrders searches for orders based on filters
func (s *orderService) SearchOrders(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.repo.Search(ctx, filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search orders: %w", err)
	}

	return orders, nil
}

// GetOrderStatusHistory retrieves the status change history for an order
func (s *orderService) GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	history, err := s.repo.GetStatusHistory(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order status history: %w", err)
	}

	return history, nil
}

// ValidateOrderItems validates order items against product catalog
func (s *orderService) ValidateOrderItems(ctx context.Context, items []OrderItemRequest) error {
	if len(items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	for i, item := range items {
		if item.ProductID == "" {
			return fmt.Errorf("item %d: product ID is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be positive", i)
		}
		if item.Price.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("item %d: price must be positive", i)
		}

		// Validate product exists
		if s.productService != nil {
			product, err := s.productService.GetProduct(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("item %d: invalid product %s: %w", i, item.ProductID, err)
			}

			// Validate price matches current product price
			if !item.Price.Equal(product.Price) {
				return fmt.Errorf("item %d: price mismatch for product %s", i, item.ProductID)
			}

			// Validate stock availability
			if err := s.productService.ValidateStock(ctx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("item %d: stock validation failed for product %s: %w", i, item.ProductID, err)
			}
		}
	}

	return nil
}

// CalculateOrderTotals calculates order totals including tax and shipping
func (s *orderService) CalculateOrderTotals(ctx context.Context, req *CreateOrderRequest) (*OrderTotals, error) {
	subtotal := decimal.Zero
	
	// Calculate subtotal
	for _, item := range req.Items {
		itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		subtotal = subtotal.Add(itemTotal)
	}

	// Calculate tax (simplified - 10% for now)
	tax := subtotal.Mul(decimal.NewFromFloat(0.10))

	// Calculate shipping (simplified - flat rate for now)
	shipping := decimal.NewFromFloat(10.00)
	if subtotal.GreaterThan(decimal.NewFromFloat(100.00)) {
		shipping = decimal.Zero // Free shipping over $100
	}

	// Calculate discount (none for now)
	discount := decimal.Zero

	// Calculate total
	total := subtotal.Add(tax).Add(shipping).Sub(discount)

	return &OrderTotals{
		Subtotal: subtotal,
		Tax:      tax,
		Shipping: shipping,
		Discount: discount,
		Total:    total,
	}, nil
}

// validateStatusTransition validates if a status transition is allowed
func (s *orderService) validateStatusTransition(ctx context.Context, orderID string, newStatus models.OrderStatus) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	currentStatus := order.Status

	// Define valid transitions
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderPending:    {models.OrderConfirmed, models.OrderCancelled},
		models.OrderConfirmed:  {models.OrderProcessing, models.OrderCancelled},
		models.OrderProcessing: {models.OrderShipped, models.OrderCancelled},
		models.OrderShipped:    {models.OrderDelivered},
		models.OrderDelivered:  {models.OrderRefunded},
		models.OrderCancelled:  {}, // No transitions from cancelled
		models.OrderRefunded:   {}, // No transitions from refunded
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", currentStatus)
	}

	// Check if transition is allowed
	for _, allowed := range allowedStatuses {
		if newStatus == allowed {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", currentStatus, newStatus)
}

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}
