package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopsphere/shared/models"
	"github.com/shopspring/decimal"
)

// MockCartRepository implements CartRepository for testing
type MockCartRepository struct {
	carts map[string]*models.Cart
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{
		carts: make(map[string]*models.Cart),
	}
}

func (m *MockCartRepository) GetCart(ctx context.Context, userID, sessionID string) (*models.Cart, error) {
	var key string
	if userID != "" {
		key = "user:" + userID
	} else {
		key = "session:" + sessionID
	}
	
	cart, exists := m.carts[key]
	if !exists {
		return nil, nil
	}
	
	if cart.IsExpired() {
		delete(m.carts, key)
		return nil, nil
	}
	
	return cart, nil
}

func (m *MockCartRepository) SaveCart(ctx context.Context, cart *models.Cart) error {
	var key string
	if cart.UserID != "" {
		key = "user:" + cart.UserID
	} else {
		key = "session:" + cart.SessionID
	}
	
	m.carts[key] = cart
	return nil
}

func (m *MockCartRepository) DeleteCart(ctx context.Context, cartID string) error {
	for key, cart := range m.carts {
		if cart.ID == cartID {
			delete(m.carts, key)
			return nil
		}
	}
	return nil
}

func (m *MockCartRepository) GetCartByID(ctx context.Context, cartID string) (*models.Cart, error) {
	for _, cart := range m.carts {
		if cart.ID == cartID {
			if cart.IsExpired() {
				m.DeleteCart(ctx, cartID)
				return nil, nil
			}
			return cart, nil
		}
	}
	return nil, nil
}

func (m *MockCartRepository) UpdateCartExpiry(ctx context.Context, cartID string, expiresAt time.Time) error {
	cart, err := m.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil
	}
	
	cart.ExpiresAt = expiresAt
	cart.UpdatedAt = time.Now()
	return m.SaveCart(ctx, cart)
}

func (m *MockCartRepository) GetExpiredCarts(ctx context.Context) ([]*models.Cart, error) {
	var expired []*models.Cart
	for _, cart := range m.carts {
		if cart.IsExpired() {
			expired = append(expired, cart)
		}
	}
	return expired, nil
}

func (m *MockCartRepository) DeleteExpiredCarts(ctx context.Context) error {
	expired, err := m.GetExpiredCarts(ctx)
	if err != nil {
		return err
	}
	
	for _, cart := range expired {
		m.DeleteCart(ctx, cart.ID)
	}
	return nil
}

func (m *MockCartRepository) MigrateGuestCartToUser(ctx context.Context, sessionID, userID string) error {
	guestKey := "session:" + sessionID
	userKey := "user:" + userID
	
	guestCart, exists := m.carts[guestKey]
	if !exists {
		return nil
	}
	
	userCart, userExists := m.carts[userKey]
	if userExists {
		// Merge carts
		for _, guestItem := range guestCart.Items {
			found := false
			for i, userItem := range userCart.Items {
				if userItem.ProductID == guestItem.ProductID {
					userCart.Items[i].Quantity += guestItem.Quantity
					userCart.Items[i].Total = userItem.Price.Mul(decimal.NewFromInt(int64(userCart.Items[i].Quantity)))
					found = true
					break
				}
			}
			if !found {
				userCart.Items = append(userCart.Items, guestItem)
			}
		}
		userCart.CalculateSubtotal()
	} else {
		// Convert guest cart to user cart
		guestCart.UserID = userID
		m.carts[userKey] = guestCart
	}
	
	delete(m.carts, guestKey)
	return nil
}

func TestCartService_GetCart(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCartRepository()
	service := NewCartService(repo, nil)
	
	// Test getting a new cart
	cart, err := service.GetCart(ctx, "user1", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cart == nil {
		t.Fatal("Expected cart to be created")
	}
	if cart.UserID != "user1" {
		t.Errorf("Expected UserID to be 'user1', got %s", cart.UserID)
	}
	
	// Test getting existing cart
	existingCart, err := service.GetCart(ctx, "user1", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if existingCart.ID != cart.ID {
		t.Errorf("Expected same cart ID, got different cart")
	}
}

func TestCartService_AddItem(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCartRepository()
	service := NewCartService(repo, nil)
	
	price := decimal.NewFromFloat(19.99)
	
	// Test adding item to new cart
	cart, err := service.AddItem(ctx, "user1", "", "prod1", "SKU1", "Product 1", price, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(cart.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(cart.Items))
	}
	
	item := cart.Items[0]
	if item.ProductID != "prod1" {
		t.Errorf("Expected ProductID 'prod1', got %s", item.ProductID)
	}
	if item.Quantity != 2 {
		t.Errorf("Expected quantity 2, got %d", item.Quantity)
	}
	
	expectedTotal := price.Mul(decimal.NewFromInt(2))
	if !item.Total.Equal(expectedTotal) {
		t.Errorf("Expected total %s, got %s", expectedTotal.String(), item.Total.String())
	}
	
	// Test adding same item again (should update quantity)
	cart, err = service.AddItem(ctx, "user1", "", "prod1", "SKU1", "Product 1", price, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(cart.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(cart.Items))
	}
	
	if cart.Items[0].Quantity != 3 {
		t.Errorf("Expected quantity 3, got %d", cart.Items[0].Quantity)
	}
}

func TestCartService_UpdateItem(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCartRepository()
	service := NewCartService(repo, nil)
	
	price := decimal.NewFromFloat(19.99)
	
	// Add item first
	cart, err := service.AddItem(ctx, "user1", "", "prod1", "SKU1", "Product 1", price, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Update quantity
	cart, err = service.UpdateItem(ctx, "user1", "", "prod1", 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if cart.Items[0].Quantity != 5 {
		t.Errorf("Expected quantity 5, got %d", cart.Items[0].Quantity)
	}
	
	// Test removing item by setting quantity to 0
	cart, err = service.UpdateItem(ctx, "user1", "", "prod1", 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(cart.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(cart.Items))
	}
}

func TestCartService_RemoveItem(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCartRepository()
	service := NewCartService(repo, nil)
	
	price := decimal.NewFromFloat(19.99)
	
	// Add item first
	cart, err := service.AddItem(ctx, "user1", "", "prod1", "SKU1", "Product 1", price, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Remove item
	cart, err = service.RemoveItem(ctx, "user1", "", "prod1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(cart.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(cart.Items))
	}
	
	// Test removing non-existent item
	_, err = service.RemoveItem(ctx, "user1", "", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent item")
	}
}

func TestCartService_ClearCart(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCartRepository()
	service := NewCartService(repo, nil)
	
	price := decimal.NewFromFloat(19.99)
	
	// Add items first
	service.AddItem(ctx, "user1", "", "prod1", "SKU1", "Product 1", price, 2)
	service.AddItem(ctx, "user1", "", "prod2", "SKU2", "Product 2", price, 1)
	
	// Clear cart
	err := service.ClearCart(ctx, "user1", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify cart is empty
	cart, err := service.GetCart(ctx, "user1", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(cart.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(cart.Items))
	}
	
	if !cart.Subtotal.Equal(decimal.Zero) {
		t.Errorf("Expected subtotal to be 0, got %s", cart.Subtotal.String())
	}
}

func TestCartService_MigrateGuestCart(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCartRepository()
	service := NewCartService(repo, nil)
	
	price := decimal.NewFromFloat(19.99)
	
	// Add item to guest cart
	service.AddItem(ctx, "", "session1", "prod1", "SKU1", "Product 1", price, 2)
	
	// Migrate to user cart
	cart, err := service.MigrateGuestCart(ctx, "session1", "user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if cart.UserID != "user1" {
		t.Errorf("Expected UserID 'user1', got %s", cart.UserID)
	}
	
	if len(cart.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(cart.Items))
	}
	
	// Verify guest cart is gone
	guestCart, err := service.GetCart(ctx, "", "session1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(guestCart.Items) != 0 {
		t.Errorf("Expected guest cart to be empty, got %d items", len(guestCart.Items))
	}
}

func TestCartService_ExtendCartExpiry(t *testing.T) {
	ctx := context.Background()
	repo := NewMockCartRepository()
	service := NewCartService(repo, nil)
	
	// Create cart
	cart, err := service.GetCart(ctx, "user1", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	originalExpiry := cart.ExpiresAt
	
	// Extend expiry
	cart, err = service.ExtendCartExpiry(ctx, "user1", "", 2*time.Hour)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Allow for small time differences due to processing time
	if cart.ExpiresAt.Sub(originalExpiry) < time.Hour {
		t.Errorf("Expected expiry to be extended by at least 1 hour, got %v", cart.ExpiresAt.Sub(originalExpiry))
	}
}
