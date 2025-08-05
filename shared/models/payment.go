package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentProcessing PaymentStatus = "processing"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentCancelled PaymentStatus = "cancelled"
	PaymentRefunded  PaymentStatus = "refunded"
)

// PaymentType represents the type of payment
type PaymentType string

const (
	PaymentTypeCard   PaymentType = "card"
	PaymentTypePayPal PaymentType = "paypal"
	PaymentTypeApplePay PaymentType = "apple_pay"
	PaymentTypeGooglePay PaymentType = "google_pay"
	PaymentTypeBankTransfer PaymentType = "bank_transfer"
)

// Payment represents a payment transaction
type Payment struct {
	ID              string          `json:"id" db:"id"`
	OrderID         string          `json:"order_id" db:"order_id"`
	UserID          string          `json:"user_id" db:"user_id"`
	Amount          decimal.Decimal `json:"amount" db:"amount"`
	Currency        string          `json:"currency" db:"currency"`
	Status          PaymentStatus   `json:"status" db:"status"`
	Type            PaymentType     `json:"type" db:"type"`
	PaymentMethodID string          `json:"payment_method_id" db:"payment_method_id"`
	TransactionID   string          `json:"transaction_id" db:"transaction_id"`
	GatewayResponse GatewayResponse `json:"gateway_response" db:"gateway_response"`
	FailureReason   string          `json:"failure_reason" db:"failure_reason"`
	ProcessedAt     *time.Time      `json:"processed_at" db:"processed_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// GatewayResponse represents the response from payment gateway
type GatewayResponse struct {
	GatewayID       string                 `json:"gateway_id"`
	TransactionID   string                 `json:"transaction_id"`
	Status          string                 `json:"status"`
	ResponseCode    string                 `json:"response_code"`
	ResponseMessage string                 `json:"response_message"`
	RawResponse     map[string]interface{} `json:"raw_response"`
	ProcessedAt     time.Time              `json:"processed_at"`
}

// NewPayment creates a new payment with default values
func NewPayment(orderID, userID string, amount decimal.Decimal, currency string, paymentType PaymentType) *Payment {
	return &Payment{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Status:    PaymentPending,
		Type:      paymentType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// PaymentMethodInfo represents stored payment method information
type PaymentMethodInfo struct {
	ID          string      `json:"id" db:"id"`
	UserID      string      `json:"user_id" db:"user_id"`
	Type        PaymentType `json:"type" db:"type"`
	IsDefault   bool        `json:"is_default" db:"is_default"`
	CardInfo    *CardInfo   `json:"card_info,omitempty" db:"card_info"`
	PayPalInfo  *PayPalInfo `json:"paypal_info,omitempty" db:"paypal_info"`
	BankInfo    *BankInfo   `json:"bank_info,omitempty" db:"bank_info"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// CardInfo represents credit/debit card information
type CardInfo struct {
	Last4       string `json:"last4"`
	Brand       string `json:"brand"` // visa, mastercard, amex, etc.
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	HolderName  string `json:"holder_name"`
	Fingerprint string `json:"fingerprint"` // unique identifier for the card
}

// PayPalInfo represents PayPal account information
type PayPalInfo struct {
	Email       string `json:"email"`
	AccountID   string `json:"account_id"`
	Verified    bool   `json:"verified"`
}

// BankInfo represents bank account information
type BankInfo struct {
	AccountNumber string `json:"account_number"` // masked
	RoutingNumber string `json:"routing_number"`
	BankName      string `json:"bank_name"`
	AccountType   string `json:"account_type"` // checking, savings
}

// Refund represents a payment refund
type Refund struct {
	ID              string          `json:"id" db:"id"`
	PaymentID       string          `json:"payment_id" db:"payment_id"`
	OrderID         string          `json:"order_id" db:"order_id"`
	Amount          decimal.Decimal `json:"amount" db:"amount"`
	Currency        string          `json:"currency" db:"currency"`
	Reason          string          `json:"reason" db:"reason"`
	Status          PaymentStatus   `json:"status" db:"status"`
	TransactionID   string          `json:"transaction_id" db:"transaction_id"`
	GatewayResponse GatewayResponse `json:"gateway_response" db:"gateway_response"`
	ProcessedAt     *time.Time      `json:"processed_at" db:"processed_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// NewRefund creates a new refund with default values
func NewRefund(paymentID, orderID string, amount decimal.Decimal, currency, reason string) *Refund {
	return &Refund{
		ID:        uuid.New().String(),
		PaymentID: paymentID,
		OrderID:   orderID,
		Amount:    amount,
		Currency:  currency,
		Reason:    reason,
		Status:    PaymentPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// PaymentIntent represents a payment intent for processing
type PaymentIntent struct {
	ID                string                 `json:"id"`
	Amount            decimal.Decimal        `json:"amount"`
	Currency          string                 `json:"currency"`
	PaymentMethodID   string                 `json:"payment_method_id"`
	CustomerID        string                 `json:"customer_id"`
	Description       string                 `json:"description"`
	Metadata          map[string]interface{} `json:"metadata"`
	ConfirmationMethod string                `json:"confirmation_method"`
	Status            string                 `json:"status"`
	ClientSecret      string                 `json:"client_secret"`
	CreatedAt         time.Time              `json:"created_at"`
}

// PaymentWebhook represents a webhook event from payment gateway
type PaymentWebhook struct {
	ID          string                 `json:"id" db:"id"`
	EventType   string                 `json:"event_type" db:"event_type"`
	PaymentID   string                 `json:"payment_id" db:"payment_id"`
	Status      string                 `json:"status" db:"status"`
	Data        map[string]interface{} `json:"data" db:"data"`
	Signature   string                 `json:"signature" db:"signature"`
	Processed   bool                   `json:"processed" db:"processed"`
	ProcessedAt *time.Time             `json:"processed_at" db:"processed_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}