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
	ID                    string          `json:"id" db:"id"`
	OrderNumber           string          `json:"order_number" db:"order_number"`
	UserID                string          `json:"user_id" db:"user_id"`
	Status                OrderStatus     `json:"status" db:"status"`
	Items                 []OrderItem     `json:"items"`
	Subtotal              decimal.Decimal `json:"subtotal" db:"subtotal"`
	Tax                   decimal.Decimal `json:"tax" db:"tax"`
	Shipping              decimal.Decimal `json:"shipping" db:"shipping"`
	Discount              decimal.Decimal `json:"discount" db:"discount"`
	Total                 decimal.Decimal `json:"total" db:"total"`
	Currency              string          `json:"currency" db:"currency"`
	ShippingAddress       Address         `json:"shipping_address"`
	BillingAddress        Address         `json:"billing_address"`
	PaymentMethod         PaymentMethod   `json:"payment_method"`
	PaymentStatus         string          `json:"payment_status" db:"payment_status"`
	PaymentReference      string          `json:"payment_reference" db:"payment_reference"`
	ShippingMethod        string          `json:"shipping_method" db:"shipping_method"`
	TrackingNumber        string          `json:"tracking_number" db:"tracking_number"`
	EstimatedDeliveryDate *time.Time      `json:"estimated_delivery_date" db:"estimated_delivery_date"`
	ActualDeliveryDate    *time.Time      `json:"actual_delivery_date" db:"actual_delivery_date"`
	Notes                 string          `json:"notes" db:"notes"`
	InternalNotes         string          `json:"internal_notes" db:"internal_notes"`
	Source                string          `json:"source" db:"source"`
	ConfirmedAt           *time.Time      `json:"confirmed_at" db:"confirmed_at"`
	ShippedAt             *time.Time      `json:"shipped_at" db:"shipped_at"`
	DeliveredAt           *time.Time      `json:"delivered_at" db:"delivered_at"`
	CancelledAt           *time.Time      `json:"cancelled_at" db:"cancelled_at"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID                 string                 `json:"id" db:"id"`
	OrderID            string                 `json:"order_id" db:"order_id"`
	ProductID          string                 `json:"product_id" db:"product_id"`
	VariantID          string                 `json:"variant_id" db:"variant_id"`
	SKU                string                 `json:"sku" db:"sku"`
	Name               string                 `json:"name" db:"name"`
	Description        string                 `json:"description" db:"description"`
	Price              decimal.Decimal        `json:"price" db:"price"`
	Quantity           int                    `json:"quantity" db:"quantity"`
	Total              decimal.Decimal        `json:"total" db:"total"`
	ProductAttributes  map[string]interface{} `json:"product_attributes" db:"product_attributes"`
	ImageURL           string                 `json:"image_url" db:"image_url"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	Type        string `json:"type"`        // card, paypal, etc.
	Last4       string `json:"last4"`       // last 4 digits for cards
	Brand       string `json:"brand"`       // visa, mastercard, etc.
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

// OrderStatusHistory represents the history of status changes for an order
type OrderStatusHistory struct {
	ID        string    `json:"id" db:"id"`
	OrderID   string    `json:"order_id" db:"order_id"`
	FromStatus string   `json:"from_status" db:"from_status"`
	ToStatus  string    `json:"to_status" db:"to_status"`
	Reason    string    `json:"reason" db:"reason"`
	Notes     string    `json:"notes" db:"notes"`
	ChangedBy string    `json:"changed_by" db:"changed_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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