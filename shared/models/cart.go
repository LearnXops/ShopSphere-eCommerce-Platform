package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CartStatus represents the status of a cart
type CartStatus string

const (
	CartActive    CartStatus = "active"
	CartAbandoned CartStatus = "abandoned"
	CartConverted CartStatus = "converted"
)

// Cart represents a shopping cart
type Cart struct {
	ID        string          `json:"id" db:"id"`
	UserID    string          `json:"user_id" db:"user_id"`
	SessionID string          `json:"session_id" db:"session_id"`
	Status    CartStatus      `json:"status" db:"status"`
	Items     []CartItem      `json:"items"`
	Subtotal  decimal.Decimal `json:"subtotal" db:"subtotal"`
	Currency  string          `json:"currency" db:"currency"`
	ExpiresAt time.Time       `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// CartItem represents an item in a cart
type CartItem struct {
	ID        string          `json:"id" db:"id"`
	CartID    string          `json:"cart_id" db:"cart_id"`
	ProductID string          `json:"product_id" db:"product_id"`
	SKU       string          `json:"sku" db:"sku"`
	Name      string          `json:"name" db:"name"`
	Price     decimal.Decimal `json:"price" db:"price"`
	Quantity  int             `json:"quantity" db:"quantity"`
	Total     decimal.Decimal `json:"total" db:"total"`
	AddedAt   time.Time       `json:"added_at" db:"added_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// NewCart creates a new cart with default values
func NewCart(userID, sessionID string) *Cart {
	return &Cart{
		ID:        uuid.New().String(),
		UserID:    userID,
		SessionID: sessionID,
		Status:    CartActive,
		Items:     []CartItem{},
		Currency:  "USD",
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiry
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AddItem adds an item to the cart
func (c *Cart) AddItem(productID, sku, name string, price decimal.Decimal, quantity int) {
	item := CartItem{
		ID:        uuid.New().String(),
		CartID:    c.ID,
		ProductID: productID,
		SKU:       sku,
		Name:      name,
		Price:     price,
		Quantity:  quantity,
		Total:     price.Mul(decimal.NewFromInt(int64(quantity))),
		AddedAt:   time.Now(),
		UpdatedAt: time.Now(),
	}
	c.Items = append(c.Items, item)
	c.UpdatedAt = time.Now()
	c.calculateSubtotal()
}

// CalculateSubtotal calculates the cart subtotal
func (c *Cart) CalculateSubtotal() {
	subtotal := decimal.Zero
	for _, item := range c.Items {
		subtotal = subtotal.Add(item.Total)
	}
	c.Subtotal = subtotal
}

// calculateSubtotal calculates the cart subtotal (private method for backward compatibility)
func (c *Cart) calculateSubtotal() {
	c.CalculateSubtotal()
}

// UpdateItem updates an existing item in the cart
func (c *Cart) UpdateItem(productID string, quantity int) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			if quantity <= 0 {
				// Remove item if quantity is 0 or negative
				c.Items = append(c.Items[:i], c.Items[i+1:]...)
			} else {
				c.Items[i].Quantity = quantity
				c.Items[i].Total = item.Price.Mul(decimal.NewFromInt(int64(quantity)))
				c.Items[i].UpdatedAt = time.Now()
			}
			c.UpdatedAt = time.Now()
			c.CalculateSubtotal()
			return true
		}
	}
	return false
}

// RemoveItem removes an item from the cart
func (c *Cart) RemoveItem(productID string) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.UpdatedAt = time.Now()
			c.CalculateSubtotal()
			return true
		}
	}
	return false
}

// GetItemCount returns the total number of items in the cart
func (c *Cart) GetItemCount() int {
	count := 0
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}

// IsExpired checks if the cart has expired
func (c *Cart) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// ExtendExpiry extends the cart expiry by the given duration
func (c *Cart) ExtendExpiry(duration time.Duration) {
	c.ExpiresAt = c.ExpiresAt.Add(duration)
	c.UpdatedAt = time.Now()
}