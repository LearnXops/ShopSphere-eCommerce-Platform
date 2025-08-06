package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/shopsphere/shared/models"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"
)

// PaymentGateway defines the interface for payment gateway operations
type PaymentGateway interface {
	// Payment processing
	CreatePaymentIntent(ctx context.Context, req *CreatePaymentIntentRequest) (*models.PaymentIntent, error)
	ConfirmPayment(ctx context.Context, paymentIntentID string) (*PaymentResult, error)
	CapturePayment(ctx context.Context, paymentIntentID string, amount decimal.Decimal) (*PaymentResult, error)
	
	// Payment methods
	CreatePaymentMethod(ctx context.Context, req *CreatePaymentMethodRequest) (*PaymentMethodResult, error)
	AttachPaymentMethodToCustomer(ctx context.Context, paymentMethodID, customerID string) error
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
	
	// Customer management
	CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CustomerResult, error)
	GetCustomer(ctx context.Context, customerID string) (*CustomerResult, error)
	
	// Refunds
	CreateRefund(ctx context.Context, req *CreateRefundRequest) (*RefundResult, error)
	
	// Webhooks
	VerifyWebhookSignature(payload []byte, signature, secret string) error
	ParseWebhookEvent(payload []byte) (*WebhookEvent, error)
}

// Request/Response types
type CreatePaymentIntentRequest struct {
	Amount              decimal.Decimal        `json:"amount"`
	Currency            string                 `json:"currency"`
	CustomerID          string                 `json:"customer_id,omitempty"`
	PaymentMethodID     string                 `json:"payment_method_id,omitempty"`
	Description         string                 `json:"description,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	ConfirmationMethod  string                 `json:"confirmation_method,omitempty"`
	AutomaticCapture    bool                   `json:"automatic_capture"`
}

type CreatePaymentMethodRequest struct {
	Type     string                 `json:"type"`
	Card     *CardDetails           `json:"card,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CardDetails struct {
	Number    string `json:"number"`
	ExpMonth  int    `json:"exp_month"`
	ExpYear   int    `json:"exp_year"`
	CVC       string `json:"cvc"`
	Name      string `json:"name,omitempty"`
}

type CreateCustomerRequest struct {
	Email       string                 `json:"email"`
	Name        string                 `json:"name,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type CreateRefundRequest struct {
	PaymentIntentID string          `json:"payment_intent_id"`
	Amount          decimal.Decimal `json:"amount,omitempty"`
	Reason          string          `json:"reason,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type PaymentResult struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"`
	Amount          decimal.Decimal        `json:"amount"`
	Currency        string                 `json:"currency"`
	PaymentMethodID string                 `json:"payment_method_id"`
	ClientSecret    string                 `json:"client_secret,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

type PaymentMethodResult struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Card         *CardInfo  `json:"card,omitempty"`
	CustomerID   string     `json:"customer_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CardInfo struct {
	Brand       string `json:"brand"`
	Last4       string `json:"last4"`
	ExpMonth    int    `json:"exp_month"`
	ExpYear     int    `json:"exp_year"`
	Fingerprint string `json:"fingerprint"`
}

type CustomerResult struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email"`
	Name        string                 `json:"name"`
	Phone       string                 `json:"phone"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

type RefundResult struct {
	ID              string          `json:"id"`
	PaymentIntentID string          `json:"payment_intent_id"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	Reason          string          `json:"reason"`
	CreatedAt       time.Time       `json:"created_at"`
}

type WebhookEvent struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
	Created time.Time              `json:"created"`
}

// StripeGateway implements PaymentGateway using Stripe
type StripeGateway struct {
	secretKey     string
	webhookSecret string
}

// NewStripeGateway creates a new Stripe payment gateway
func NewStripeGateway(secretKey, webhookSecret string) PaymentGateway {
	stripe.Key = secretKey
	return &StripeGateway{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

// CreatePaymentIntent creates a payment intent in Stripe
func (s *StripeGateway) CreatePaymentIntent(ctx context.Context, req *CreatePaymentIntentRequest) (*models.PaymentIntent, error) {
	// Convert decimal amount to cents (Stripe expects integer cents)
	amountCents := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(req.Currency),
	}

	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}

	if req.PaymentMethodID != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethodID)
	}

	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	if req.ConfirmationMethod != "" {
		params.ConfirmationMethod = stripe.String(req.ConfirmationMethod)
	} else {
		params.ConfirmationMethod = stripe.String("manual")
	}

	if req.AutomaticCapture {
		params.CaptureMethod = stripe.String("automatic")
	} else {
		params.CaptureMethod = stripe.String("manual")
	}

	// Add metadata
	if req.Metadata != nil {
		params.Metadata = make(map[string]string)
		for k, v := range req.Metadata {
			params.Metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return &models.PaymentIntent{
		ID:                pi.ID,
		Amount:            decimal.NewFromInt(pi.Amount).Div(decimal.NewFromInt(100)),
		Currency:          string(pi.Currency),
		PaymentMethodID:   getStringValue(pi.PaymentMethod),
		CustomerID:        getStringValue(pi.Customer),
		Description:       getStringValue(pi.Description),
		Metadata:          req.Metadata,
		ConfirmationMethod: string(pi.ConfirmationMethod),
		Status:            string(pi.Status),
		ClientSecret:      pi.ClientSecret,
		CreatedAt:         time.Unix(pi.Created, 0),
	}, nil
}

// ConfirmPayment confirms a payment intent
func (s *StripeGateway) ConfirmPayment(ctx context.Context, paymentIntentID string) (*PaymentResult, error) {
	params := &stripe.PaymentIntentConfirmParams{}
	pi, err := paymentintent.Confirm(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm payment: %w", err)
	}

	return &PaymentResult{
		ID:              pi.ID,
		Status:          string(pi.Status),
		Amount:          decimal.NewFromInt(pi.Amount).Div(decimal.NewFromInt(100)),
		Currency:        string(pi.Currency),
		PaymentMethodID: getStringValue(pi.PaymentMethod),
		ClientSecret:    pi.ClientSecret,
		CreatedAt:       time.Unix(pi.Created, 0),
	}, nil
}

// CapturePayment captures a payment intent
func (s *StripeGateway) CapturePayment(ctx context.Context, paymentIntentID string, amount decimal.Decimal) (*PaymentResult, error) {
	params := &stripe.PaymentIntentCaptureParams{}
	
	if !amount.IsZero() {
		amountCents := amount.Mul(decimal.NewFromInt(100)).IntPart()
		params.AmountToCapture = stripe.Int64(amountCents)
	}

	pi, err := paymentintent.Capture(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to capture payment: %w", err)
	}

	return &PaymentResult{
		ID:              pi.ID,
		Status:          string(pi.Status),
		Amount:          decimal.NewFromInt(pi.Amount).Div(decimal.NewFromInt(100)),
		Currency:        string(pi.Currency),
		PaymentMethodID: getStringValue(pi.PaymentMethod),
		CreatedAt:       time.Unix(pi.Created, 0),
	}, nil
}

// CreatePaymentMethod creates a payment method in Stripe
func (s *StripeGateway) CreatePaymentMethod(ctx context.Context, req *CreatePaymentMethodRequest) (*PaymentMethodResult, error) {
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(req.Type),
	}

	if req.Card != nil {
		params.Card = &stripe.PaymentMethodCardParams{
			Number:   stripe.String(req.Card.Number),
			ExpMonth: stripe.Int64(int64(req.Card.ExpMonth)),
			ExpYear:  stripe.Int64(int64(req.Card.ExpYear)),
			CVC:      stripe.String(req.Card.CVC),
		}
	}

	pm, err := paymentmethod.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	result := &PaymentMethodResult{
		ID:        pm.ID,
		Type:      string(pm.Type),
		CreatedAt: time.Unix(pm.Created, 0),
	}

	if pm.Card != nil {
		result.Card = &CardInfo{
			Brand:       string(pm.Card.Brand),
			Last4:       pm.Card.Last4,
			ExpMonth:    int(pm.Card.ExpMonth),
			ExpYear:     int(pm.Card.ExpYear),
			Fingerprint: pm.Card.Fingerprint,
		}
	}

	return result, nil
}

// AttachPaymentMethodToCustomer attaches a payment method to a customer
func (s *StripeGateway) AttachPaymentMethodToCustomer(ctx context.Context, paymentMethodID, customerID string) error {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	_, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return fmt.Errorf("failed to attach payment method: %w", err)
	}

	return nil
}

// DetachPaymentMethod detaches a payment method from a customer
func (s *StripeGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}

	return nil
}

// CreateCustomer creates a customer in Stripe
func (s *StripeGateway) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CustomerResult, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(req.Email),
	}

	if req.Name != "" {
		params.Name = stripe.String(req.Name)
	}

	if req.Phone != "" {
		params.Phone = stripe.String(req.Phone)
	}

	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	if req.Metadata != nil {
		params.Metadata = make(map[string]string)
		for k, v := range req.Metadata {
			params.Metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &CustomerResult{
		ID:          cust.ID,
		Email:       cust.Email,
		Name:        getStringValue(cust.Name),
		Phone:       getStringValue(cust.Phone),
		Description: getStringValue(cust.Description),
		Metadata:    req.Metadata,
		CreatedAt:   time.Unix(cust.Created, 0),
	}, nil
}

// GetCustomer retrieves a customer from Stripe
func (s *StripeGateway) GetCustomer(ctx context.Context, customerID string) (*CustomerResult, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &CustomerResult{
		ID:          cust.ID,
		Email:       cust.Email,
		Name:        getStringValue(cust.Name),
		Phone:       getStringValue(cust.Phone),
		Description: getStringValue(cust.Description),
		CreatedAt:   time.Unix(cust.Created, 0),
	}, nil
}

// CreateRefund creates a refund in Stripe
func (s *StripeGateway) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*RefundResult, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentIntentID),
	}

	if !req.Amount.IsZero() {
		amountCents := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()
		params.Amount = stripe.Int64(amountCents)
	}

	if req.Reason != "" {
		params.Reason = stripe.String(req.Reason)
	}

	if req.Metadata != nil {
		params.Metadata = make(map[string]string)
		for k, v := range req.Metadata {
			params.Metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	r, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return &RefundResult{
		ID:              r.ID,
		PaymentIntentID: getStringValue(r.PaymentIntent),
		Amount:          decimal.NewFromInt(r.Amount).Div(decimal.NewFromInt(100)),
		Currency:        string(r.Currency),
		Status:          string(r.Status),
		Reason:          getStringValue(r.Reason),
		CreatedAt:       time.Unix(r.Created, 0),
	}, nil
}

// VerifyWebhookSignature verifies a Stripe webhook signature
func (s *StripeGateway) VerifyWebhookSignature(payload []byte, signature, secret string) error {
	_, err := webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		return fmt.Errorf("failed to verify webhook signature: %w", err)
	}
	return nil
}

// ParseWebhookEvent parses a Stripe webhook event
func (s *StripeGateway) ParseWebhookEvent(payload []byte) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, "", s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}

	return &WebhookEvent{
		ID:      event.ID,
		Type:    string(event.Type),
		Data:    event.Data.Object,
		Created: time.Unix(event.Created, 0),
	}, nil
}

// Helper function to safely get string values from Stripe objects
func getStringValue(s interface{}) string {
	if s == nil {
		return ""
	}
	
	switch v := s.(type) {
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case string:
		return v
	case *stripe.PaymentMethod:
		if v == nil {
			return ""
		}
		return v.ID
	case *stripe.Customer:
		if v == nil {
			return ""
		}
		return v.ID
	case *stripe.PaymentIntent:
		if v == nil {
			return ""
		}
		return v.ID
	default:
		return fmt.Sprintf("%v", s)
	}
}
