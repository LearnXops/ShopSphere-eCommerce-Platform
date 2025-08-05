package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// EventType represents the type of domain event
type EventType string

const (
	// User events
	EventUserRegistered EventType = "user.registered"
	EventUserUpdated    EventType = "user.updated"
	EventUserDeleted    EventType = "user.deleted"

	// Product events
	EventProductCreated EventType = "product.created"
	EventProductUpdated EventType = "product.updated"
	EventProductDeleted EventType = "product.deleted"
	EventInventoryUpdated EventType = "inventory.updated"

	// Order events
	EventOrderCreated   EventType = "order.created"
	EventOrderConfirmed EventType = "order.confirmed"
	EventOrderShipped   EventType = "order.shipped"
	EventOrderDelivered EventType = "order.delivered"
	EventOrderCancelled EventType = "order.cancelled"

	// Payment events
	EventPaymentProcessed EventType = "payment.processed"
	EventPaymentFailed    EventType = "payment.failed"
	EventPaymentRefunded  EventType = "payment.refunded"

	// Cart events
	EventCartItemAdded   EventType = "cart.item.added"
	EventCartItemRemoved EventType = "cart.item.removed"
	EventCartAbandoned   EventType = "cart.abandoned"

	// Review events
	EventReviewCreated EventType = "review.created"
	EventReviewUpdated EventType = "review.updated"
)

// DomainEvent represents a domain event
type DomainEvent struct {
	ID          string          `json:"id"`
	EventType   EventType       `json:"event_type"`
	AggregateID string          `json:"aggregate_id"`
	Version     int             `json:"version"`
	Timestamp   time.Time       `json:"timestamp"`
	Data        json.RawMessage `json:"data"`
	Metadata    EventMetadata   `json:"metadata"`
}

// EventMetadata contains event metadata
type EventMetadata struct {
	UserID      string `json:"user_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	TraceID     string `json:"trace_id,omitempty"`
	Source      string `json:"source,omitempty"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType EventType, aggregateID string, data interface{}, metadata EventMetadata) (*DomainEvent, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &DomainEvent{
		ID:          uuid.New().String(),
		EventType:   eventType,
		AggregateID: aggregateID,
		Version:     1,
		Timestamp:   time.Now().UTC(),
		Data:        dataBytes,
		Metadata:    metadata,
	}, nil
}

// UnmarshalData unmarshals the event data into the provided interface
func (e *DomainEvent) UnmarshalData(v interface{}) error {
	return json.Unmarshal(e.Data, v)
}

// User Events Data Structures

// UserRegisteredData represents data for user registered event
type UserRegisteredData struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserUpdatedData represents data for user updated event
type UserUpdatedData struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Changes   map[string]interface{} `json:"changes"`
}

// Product Events Data Structures

// ProductCreatedData represents data for product created event
type ProductCreatedData struct {
	ProductID   string          `json:"product_id"`
	SKU         string          `json:"sku"`
	Name        string          `json:"name"`
	CategoryID  string          `json:"category_id"`
	Price       decimal.Decimal `json:"price"`
	Currency    string          `json:"currency"`
	Stock       int             `json:"stock"`
}

// InventoryUpdatedData represents data for inventory updated event
type InventoryUpdatedData struct {
	ProductID    string `json:"product_id"`
	SKU          string `json:"sku"`
	PreviousStock int   `json:"previous_stock"`
	NewStock     int    `json:"new_stock"`
	Reason       string `json:"reason"` // sale, restock, adjustment
}

// Order Events Data Structures

// OrderCreatedData represents data for order created event
type OrderCreatedData struct {
	OrderID         string          `json:"order_id"`
	UserID          string          `json:"user_id"`
	Items           []OrderItem     `json:"items"`
	Total           decimal.Decimal `json:"total"`
	Currency        string          `json:"currency"`
	ShippingAddress Address         `json:"shipping_address"`
	BillingAddress  Address         `json:"billing_address"`
}

// OrderStatusChangedData represents data for order status change events
type OrderStatusChangedData struct {
	OrderID       string      `json:"order_id"`
	UserID        string      `json:"user_id"`
	PreviousStatus OrderStatus `json:"previous_status"`
	NewStatus     OrderStatus `json:"new_status"`
	Reason        string      `json:"reason,omitempty"`
}

// Payment Events Data Structures

// PaymentProcessedData represents data for payment processed event
type PaymentProcessedData struct {
	PaymentID     string          `json:"payment_id"`
	OrderID       string          `json:"order_id"`
	UserID        string          `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	PaymentMethod PaymentMethod   `json:"payment_method"`
	TransactionID string          `json:"transaction_id"`
}

// PaymentFailedData represents data for payment failed event
type PaymentFailedData struct {
	PaymentID     string          `json:"payment_id"`
	OrderID       string          `json:"order_id"`
	UserID        string          `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Reason        string          `json:"reason"`
	ErrorCode     string          `json:"error_code"`
}

// Cart Events Data Structures

// CartItemAddedData represents data for cart item added event
type CartItemAddedData struct {
	CartID    string          `json:"cart_id"`
	UserID    string          `json:"user_id"`
	ProductID string          `json:"product_id"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int             `json:"quantity"`
}

// CartAbandonedData represents data for cart abandoned event
type CartAbandonedData struct {
	CartID       string          `json:"cart_id"`
	UserID       string          `json:"user_id"`
	ItemCount    int             `json:"item_count"`
	TotalValue   decimal.Decimal `json:"total_value"`
	Currency     string          `json:"currency"`
	AbandonedAt  time.Time       `json:"abandoned_at"`
}

// Review Events Data Structures

// ReviewCreatedData represents data for review created event
type ReviewCreatedData struct {
	ReviewID  string `json:"review_id"`
	ProductID string `json:"product_id"`
	UserID    string `json:"user_id"`
	OrderID   string `json:"order_id"`
	Rating    int    `json:"rating"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}