package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderConfirmed  OrderStatus = "confirmed"
	OrderProcessing OrderStatus = "processing"
	OrderShipped    OrderStatus = "shipped"
	OrderDelivered  OrderStatus = "delivered"
	OrderCancelled  OrderStatus = "cancelled"
	OrderRefunded   OrderStatus = "refunded"
)

// Order represents an order in the system
type Order struct {
	ID              string          `json:"id" db:"id"`
	UserID          string          `json:"user_id" db:"user_id"`
	Status          OrderStatus     `json:"status" db:"status"`
	Items           []OrderItem     `json:"items"`
	Subtotal        decimal.Decimal `json:"subtotal" db:"subtotal"`
	Tax             decimal.Decimal `json:"tax" db:"tax"`
	Shipping        decimal.Decimal `json:"shipping" db:"shipping"`
	Total           decimal.Decimal `json:"total" db:"total"`
	Currency        string          `json:"currency" db:"currency"`
	ShippingAddress Address         `json:"shipping_address"`
	BillingAddress  Address         `json:"billing_address"`
	PaymentMethod   PaymentMethod   `json:"payment_method"`
	TrackingNumber  string          `json:"tracking_number" db:"tracking_number"`
	Notes           string          `json:"notes" db:"notes"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        string          `json:"id" db:"id"`
	OrderID   string          `json:"order_id" db:"order_id"`
	ProductID string          `json:"product_id" db:"product_id"`
	SKU       string          `json:"sku" db:"sku"`
	Name      string          `json:"name" db:"name"`
	Price     decimal.Decimal `json:"price" db:"price"`
	Quantity  int             `json:"quantity" db:"quantity"`
	Total     decimal.Decimal `json:"total" db:"total"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	Type        string `json:"type"`        // card, paypal, etc.
	Last4       string `json:"last4"`       // last 4 digits for cards
	Brand       string `json:"brand"`       // visa, mastercard, etc.
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

// NewOrder creates a new order with default values
func NewOrder(userID string) *Order {
	return &Order{
		ID:       uuid.New().String(),
		UserID:   userID,
		Status:   OrderPending,
		Items:    []OrderItem{},
		Currency: "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}