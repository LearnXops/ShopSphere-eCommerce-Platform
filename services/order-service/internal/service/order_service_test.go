package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/shopsphere/shared/models"
)

// MockOrderRepository implements OrderRepository for testing
type MockOrderRepository struct {
	orders        map[string]*models.Order
	statusHistory map[string][]models.OrderStatusHistory
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders:        make(map[string]*models.Order),
		statusHistory: make(map[string][]models.OrderStatusHistory),
	}
}

func (m *MockOrderRepository) Create(ctx context.Context, order *models.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	order, exists := m.orders[id]
	if !exists {
		return nil, &NotFoundError{Resource: "order", ID: id}
	}
	return order, nil
}

func (m *MockOrderRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	var orders []*models.Order
	for _, order := range m.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (m *MockOrderRepository) Update(ctx context.Context, order *models.Order) error {
	if _, exists := m.orders[order.ID]; !exists {
		return &NotFoundError{Resource: "order", ID: order.ID}
	}
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus, reason string, changedBy string) error {
	order, exists := m.orders[orderID]
	if !exists {
		return &NotFoundError{Resource: "order", ID: orderID}
	}
	
	oldStatus := order.Status
	order.Status = status
	order.UpdatedAt = time.Now()
	
	// Record status change
	history := models.OrderStatusHistory{
		ID:        "hist-" + orderID + "-" + string(status),
		OrderID:   orderID,
		FromStatus: string(oldStatus),
		ToStatus:  string(status),
		Reason:    reason,
		ChangedBy: changedBy,
		CreatedAt: time.Now(),
	}
	
	m.statusHistory[orderID] = append(m.statusHistory[orderID], history)
	return nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
	return m.UpdateStatus(ctx, id, models.OrderCancelled, "Order deleted", "system")
}

func (m *MockOrderRepository) Search(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Order, error) {
	var orders []*models.Order
	for _, order := range m.orders {
		match := true
		
		if userID, ok := filters["user_id"]; ok && order.UserID != userID {
			match = false
		}
		if status, ok := filters["status"]; ok && string(order.Status) != status {
			match = false
		}
		
		if match {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (m *MockOrderRepository) GetStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error) {
	history, exists := m.statusHistory[orderID]
	if !exists {
		return []models.OrderStatusHistory{}, nil
	}
	return history, nil
}

// MockProductService implements ProductService for testing
type MockProductService struct {
	products map[string]*models.Product
}

func NewMockProductService() *MockProductService {
	return &MockProductService{
		products: map[string]*models.Product{
			"prod1": {
				ID:          "prod1",
				SKU:         "SKU001",
				Name:        "Test Product 1",
				Description: "Test product description",
				Price:       decimal.NewFromFloat(99.99),
				Stock:       10,
				Images:      []string{"image1.jpg"},
				Attributes: models.ProductAttributes{
					Brand: "TestBrand",
					Color: "Red",
					Size:  "M",
					Custom: map[string]interface{}{
						"material": "cotton",
					},
				},
			},
		},
	}
}

func (m *MockProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	product, exists := m.products[id]
	if !exists {
		return nil, &NotFoundError{Resource: "product", ID: id}
	}
	return product, nil
}

func (m *MockProductService) ValidateStock(ctx context.Context, productID string, quantity int) error {
	product, exists := m.products[productID]
	if !exists {
		return &NotFoundError{Resource: "product", ID: productID}
	}
	if product.Stock < quantity {
		return &InsufficientStockError{ProductID: productID, Available: product.Stock, Requested: quantity}
	}
	return nil
}

// Error types for testing
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return e.Resource + " not found: " + e.ID
}

type InsufficientStockError struct {
	ProductID string
	Available int
	Requested int
}

func (e *InsufficientStockError) Error() string {
	return "insufficient stock for product " + e.ProductID
}

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	productService := NewMockProductService()
	service := NewOrderService(repo, productService, nil)

	req := &CreateOrderRequest{
		UserID: "user1",
		Items: []OrderItemRequest{
			{
				ProductID: "prod1",
				Quantity:  2,
				Price:     decimal.NewFromFloat(99.99),
			},
		},
		ShippingAddress: models.Address{
			Street:     "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "US",
		},
		BillingAddress: models.Address{
			Street:     "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "US",
		},
		PaymentMethod: models.PaymentMethod{
			Type:  "card",
			Last4: "1234",
			Brand: "visa",
		},
	}

	order, err := service.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if order.ID == "" {
		t.Error("Expected order ID to be set")
	}

	if order.Status != models.OrderPending {
		t.Errorf("Expected status to be pending, got %s", order.Status)
	}

	if len(order.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(order.Items))
	}

	// Subtotal: 99.99 * 2 = 199.98
	// Tax: 199.98 * 0.10 = 19.998
	// Shipping: 10.00 (since subtotal > 100, shipping should be 0, but let's check actual logic)
	// Total: 199.98 + 19.998 + 10.00 = 229.978
	expectedTotal := decimal.NewFromFloat(219.978) // Based on actual calculation: 199.98 + 19.998 + 0 (free shipping)
	if !order.Total.Equal(expectedTotal) {
		t.Errorf("Expected total %s, got %s", expectedTotal, order.Total)
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	service := NewOrderService(repo, nil, nil)

	// Create test order
	testOrder := &models.Order{
		ID:     "order1",
		UserID: "user1",
		Status: models.OrderPending,
		Total:  decimal.NewFromFloat(100.00),
	}
	repo.Create(ctx, testOrder)

	order, err := service.GetOrder(ctx, "order1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if order.ID != "order1" {
		t.Errorf("Expected order ID order1, got %s", order.ID)
	}
}

func TestOrderService_GetOrder_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	service := NewOrderService(repo, nil, nil)

	_, err := service.GetOrder(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent order")
	}
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	service := NewOrderService(repo, nil, nil)

	// Create test order
	testOrder := &models.Order{
		ID:     "order1",
		UserID: "user1",
		Status: models.OrderPending,
		Total:  decimal.NewFromFloat(100.00),
	}
	repo.Create(ctx, testOrder)

	err := service.UpdateOrderStatus(ctx, "order1", models.OrderConfirmed, "Payment confirmed", "system")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify status was updated
	order, _ := repo.GetByID(ctx, "order1")
	if order.Status != models.OrderConfirmed {
		t.Errorf("Expected status to be confirmed, got %s", order.Status)
	}

	// Verify history was recorded
	history, _ := repo.GetStatusHistory(ctx, "order1")
	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}
}

func TestOrderService_UpdateOrderStatus_InvalidTransition(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	service := NewOrderService(repo, nil, nil)

	// Create test order in cancelled status
	testOrder := &models.Order{
		ID:     "order1",
		UserID: "user1",
		Status: models.OrderCancelled,
		Total:  decimal.NewFromFloat(100.00),
	}
	repo.Create(ctx, testOrder)

	err := service.UpdateOrderStatus(ctx, "order1", models.OrderConfirmed, "Invalid transition", "system")
	if err == nil {
		t.Error("Expected error for invalid status transition")
	}
}

func TestOrderService_CancelOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	service := NewOrderService(repo, nil, nil)

	// Create test order
	testOrder := &models.Order{
		ID:     "order1",
		UserID: "user1",
		Status: models.OrderPending,
		Total:  decimal.NewFromFloat(100.00),
	}
	repo.Create(ctx, testOrder)

	err := service.CancelOrder(ctx, "order1", "Customer request", "user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify order was cancelled
	order, _ := repo.GetByID(ctx, "order1")
	if order.Status != models.OrderCancelled {
		t.Errorf("Expected status to be cancelled, got %s", order.Status)
	}
}

func TestOrderService_CancelOrder_AlreadyCancelled(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	service := NewOrderService(repo, nil, nil)

	// Create test order in cancelled status
	testOrder := &models.Order{
		ID:     "order1",
		UserID: "user1",
		Status: models.OrderCancelled,
		Total:  decimal.NewFromFloat(100.00),
	}
	repo.Create(ctx, testOrder)

	err := service.CancelOrder(ctx, "order1", "Customer request", "user1")
	if err == nil {
		t.Error("Expected error when cancelling already cancelled order")
	}
}

func TestOrderService_ValidateOrderItems(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	productService := NewMockProductService()
	service := NewOrderService(repo, productService, nil)

	items := []OrderItemRequest{
		{
			ProductID: "prod1",
			Quantity:  2,
			Price:     decimal.NewFromFloat(99.99),
		},
	}

	err := service.ValidateOrderItems(ctx, items)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestOrderService_ValidateOrderItems_InvalidProduct(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	productService := NewMockProductService()
	service := NewOrderService(repo, productService, nil)

	items := []OrderItemRequest{
		{
			ProductID: "nonexistent",
			Quantity:  2,
			Price:     decimal.NewFromFloat(99.99),
		},
	}

	err := service.ValidateOrderItems(ctx, items)
	if err == nil {
		t.Error("Expected error for invalid product")
	}
}

func TestOrderService_CalculateOrderTotals(t *testing.T) {
	ctx := context.Background()
	repo := NewMockOrderRepository()
	service := NewOrderService(repo, nil, nil)

	req := &CreateOrderRequest{
		Items: []OrderItemRequest{
			{
				ProductID: "prod1",
				Quantity:  2,
				Price:     decimal.NewFromFloat(50.00),
			},
		},
	}

	totals, err := service.CalculateOrderTotals(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedSubtotal := decimal.NewFromFloat(100.00)
	if !totals.Subtotal.Equal(expectedSubtotal) {
		t.Errorf("Expected subtotal %s, got %s", expectedSubtotal, totals.Subtotal)
	}

	expectedTax := decimal.NewFromFloat(10.00) // 10% of subtotal
	if !totals.Tax.Equal(expectedTax) {
		t.Errorf("Expected tax %s, got %s", expectedTax, totals.Tax)
	}

	expectedShipping := decimal.NewFromFloat(10.00) // Flat rate shipping for orders under $100
	if !totals.Shipping.Equal(expectedShipping) {
		t.Errorf("Expected shipping %s, got %s", expectedShipping, totals.Shipping)
	}

	expectedTotal := decimal.NewFromFloat(120.00) // subtotal + tax + shipping
	if !totals.Total.Equal(expectedTotal) {
		t.Errorf("Expected total %s, got %s", expectedTotal, totals.Total)
	}
}
