package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/shopsphere/payment-service/internal/gateway"
	"github.com/shopsphere/payment-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// PaymentService defines the interface for payment business logic
type PaymentService interface {
	// Payment processing
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*models.Payment, error)
	ProcessPayment(ctx context.Context, paymentID string) (*models.Payment, error)
	GetPayment(ctx context.Context, paymentID string) (*models.Payment, error)
	GetPaymentsByOrder(ctx context.Context, orderID string) ([]*models.Payment, error)
	GetUserPayments(ctx context.Context, userID string, limit, offset int) ([]*models.Payment, error)
	CancelPayment(ctx context.Context, paymentID string, reason string) error
	
	// Payment methods
	CreatePaymentMethod(ctx context.Context, req *CreatePaymentMethodRequest) (*models.PaymentMethodInfo, error)
	GetPaymentMethod(ctx context.Context, methodID string) (*models.PaymentMethodInfo, error)
	GetUserPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethodInfo, error)
	UpdatePaymentMethod(ctx context.Context, methodID string, req *UpdatePaymentMethodRequest) (*models.PaymentMethodInfo, error)
	DeletePaymentMethod(ctx context.Context, methodID string) error
	SetDefaultPaymentMethod(ctx context.Context, userID, methodID string) error
	
	// Refunds
	CreateRefund(ctx context.Context, req *CreateRefundRequest) (*models.Refund, error)
	ProcessRefund(ctx context.Context, refundID string) (*models.Refund, error)
	GetRefund(ctx context.Context, refundID string) (*models.Refund, error)
	GetPaymentRefunds(ctx context.Context, paymentID string) ([]*models.Refund, error)
	
	// Webhooks
	ProcessWebhook(ctx context.Context, eventType string, payload []byte, signature string) error
	
	// Retry logic
	RetryFailedPayment(ctx context.Context, paymentID string) (*models.Payment, error)
	
	// Fraud detection
	ValidatePayment(ctx context.Context, payment *models.Payment) error
}

// Request types
type CreatePaymentRequest struct {
	OrderID         string                 `json:"order_id" validate:"required"`
	UserID          string                 `json:"user_id" validate:"required"`
	Amount          decimal.Decimal        `json:"amount" validate:"required,gt=0"`
	Currency        string                 `json:"currency" validate:"required,len=3"`
	PaymentMethodID string                 `json:"payment_method_id"`
	Description     string                 `json:"description"`
	Metadata        map[string]interface{} `json:"metadata"`
	AutoCapture     bool                   `json:"auto_capture"`
}

type CreatePaymentMethodRequest struct {
	UserID   string                 `json:"user_id" validate:"required"`
	Type     models.PaymentType     `json:"type" validate:"required"`
	Card     *CardDetailsRequest    `json:"card,omitempty"`
	PayPal   *PayPalDetailsRequest  `json:"paypal,omitempty"`
	Bank     *BankDetailsRequest    `json:"bank,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

type CardDetailsRequest struct {
	Number      string `json:"number" validate:"required"`
	ExpMonth    int    `json:"exp_month" validate:"required,min=1,max=12"`
	ExpYear     int    `json:"exp_year" validate:"required"`
	CVC         string `json:"cvc" validate:"required"`
	HolderName  string `json:"holder_name" validate:"required"`
}

type PayPalDetailsRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type BankDetailsRequest struct {
	AccountNumber string `json:"account_number" validate:"required"`
	RoutingNumber string `json:"routing_number" validate:"required"`
	BankName      string `json:"bank_name" validate:"required"`
	AccountType   string `json:"account_type" validate:"required"`
}

type UpdatePaymentMethodRequest struct {
	IsDefault bool                   `json:"is_default"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type CreateRefundRequest struct {
	PaymentID string          `json:"payment_id" validate:"required"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	Reason    string          `json:"reason" validate:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// paymentService implements PaymentService
type paymentService struct {
	repo    repository.PaymentRepository
	gateway gateway.PaymentGateway
}

// NewPaymentService creates a new payment service
func NewPaymentService(repo repository.PaymentRepository, gateway gateway.PaymentGateway) PaymentService {
	return &paymentService{
		repo:    repo,
		gateway: gateway,
	}
}

// CreatePayment creates a new payment
func (s *paymentService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*models.Payment, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// Create payment record
	payment := models.NewPayment(req.OrderID, req.UserID, req.Amount, req.Currency, models.PaymentTypeCard)
	payment.PaymentMethodID = req.PaymentMethodID

	// Create payment intent in gateway
	gatewayReq := &gateway.CreatePaymentIntentRequest{
		Amount:             req.Amount,
		Currency:           req.Currency,
		PaymentMethodID:    req.PaymentMethodID,
		Description:        req.Description,
		Metadata:           req.Metadata,
		ConfirmationMethod: "manual",
		AutomaticCapture:   req.AutoCapture,
	}

	paymentIntent, err := s.gateway.CreatePaymentIntent(ctx, gatewayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Update payment with gateway response
	payment.TransactionID = paymentIntent.ID
	payment.GatewayResponse = models.GatewayResponse{
		GatewayID:       "stripe",
		TransactionID:   paymentIntent.ID,
		Status:          paymentIntent.Status,
		ResponseCode:    "200",
		ResponseMessage: "Payment intent created",
		RawResponse: map[string]interface{}{
			"client_secret": paymentIntent.ClientSecret,
			"status":        paymentIntent.Status,
		},
		ProcessedAt: time.Now(),
	}

	// Save payment to database
	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	utils.Logger.Info(ctx, "Payment created successfully", map[string]interface{}{
		"payment_id": payment.ID,
		"order_id":   payment.OrderID,
		"amount":     payment.Amount,
	})

	return payment, nil
}

// ProcessPayment processes a payment by confirming it with the gateway
func (s *paymentService) ProcessPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	// Get payment from database
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Check if payment is in correct state
	if payment.Status != models.PaymentPending {
		return nil, fmt.Errorf("payment cannot be processed in current state: %s", payment.Status)
	}

	// Update status to processing
	payment.Status = models.PaymentProcessing
	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		utils.Logger.Error(ctx, "Failed to update payment status to processing", err)
	}

	// Record payment attempt
	if err := s.repo.CreatePaymentAttempt(ctx, paymentID, 1, "processing", "", nil); err != nil {
		utils.Logger.Error(ctx, "Failed to record payment attempt", err)
	}

	// Confirm payment with gateway
	result, err := s.gateway.ConfirmPayment(ctx, payment.TransactionID)
	if err != nil {
		// Update payment status to failed
		gatewayResponse := &models.GatewayResponse{
			GatewayID:       "stripe",
			TransactionID:   payment.TransactionID,
			Status:          "failed",
			ResponseCode:    "400",
			ResponseMessage: err.Error(),
			ProcessedAt:     time.Now(),
		}

		if updateErr := s.repo.UpdatePaymentStatus(ctx, paymentID, models.PaymentFailed, payment.TransactionID, err.Error(), gatewayResponse); updateErr != nil {
			utils.Logger.Error(ctx, "Failed to update payment status to failed", updateErr)
		}

		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	// Update payment status based on gateway response
	var status models.PaymentStatus
	switch result.Status {
	case "succeeded":
		status = models.PaymentCompleted
	case "requires_capture":
		status = models.PaymentCompleted // Will be captured later if needed
	case "processing":
		status = models.PaymentProcessing
	default:
		status = models.PaymentFailed
	}

	gatewayResponse := &models.GatewayResponse{
		GatewayID:       "stripe",
		TransactionID:   result.ID,
		Status:          result.Status,
		ResponseCode:    "200",
		ResponseMessage: "Payment processed successfully",
		RawResponse: map[string]interface{}{
			"status":            result.Status,
			"payment_method_id": result.PaymentMethodID,
		},
		ProcessedAt: time.Now(),
	}

	if err := s.repo.UpdatePaymentStatus(ctx, paymentID, status, result.ID, "", gatewayResponse); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Get updated payment
	payment, err = s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated payment: %w", err)
	}

	utils.Logger.Info(ctx, "Payment processed successfully", map[string]interface{}{
		"payment_id": payment.ID,
		"status":     payment.Status,
		"amount":     payment.Amount,
	})

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (s *paymentService) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	return s.repo.GetPaymentByID(ctx, paymentID)
}

// GetPaymentsByOrder retrieves all payments for an order
func (s *paymentService) GetPaymentsByOrder(ctx context.Context, orderID string) ([]*models.Payment, error) {
	return s.repo.GetPaymentsByOrderID(ctx, orderID)
}

// GetUserPayments retrieves payments for a user with pagination
func (s *paymentService) GetUserPayments(ctx context.Context, userID string, limit, offset int) ([]*models.Payment, error) {
	return s.repo.GetPaymentsByUserID(ctx, userID, limit, offset)
}

// CancelPayment cancels a payment
func (s *paymentService) CancelPayment(ctx context.Context, paymentID string, reason string) error {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != models.PaymentPending && payment.Status != models.PaymentProcessing {
		return fmt.Errorf("payment cannot be cancelled in current state: %s", payment.Status)
	}

	// Update payment status
	gatewayResponse := &models.GatewayResponse{
		GatewayID:       "stripe",
		TransactionID:   payment.TransactionID,
		Status:          "cancelled",
		ResponseCode:    "200",
		ResponseMessage: reason,
		ProcessedAt:     time.Now(),
	}

	if err := s.repo.UpdatePaymentStatus(ctx, paymentID, models.PaymentCancelled, payment.TransactionID, reason, gatewayResponse); err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}

	utils.Logger.Info(ctx, "Payment cancelled", map[string]interface{}{
		"payment_id": paymentID,
		"reason":     reason,
	})

	return nil
}

// CreatePaymentMethod creates a new payment method
func (s *paymentService) CreatePaymentMethod(ctx context.Context, req *CreatePaymentMethodRequest) (*models.PaymentMethodInfo, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("invalid payment method request: %w", err)
	}

	// Create payment method in gateway
	var gatewayReq *gateway.CreatePaymentMethodRequest
	if req.Type == models.PaymentTypeCard && req.Card != nil {
		gatewayReq = &gateway.CreatePaymentMethodRequest{
			Type: "card",
			Card: &gateway.CardDetails{
				Number:   req.Card.Number,
				ExpMonth: req.Card.ExpMonth,
				ExpYear:  req.Card.ExpYear,
				CVC:      req.Card.CVC,
				Name:     req.Card.HolderName,
			},
			Metadata: req.Metadata,
		}
	}

	if gatewayReq != nil {
		gatewayResult, err := s.gateway.CreatePaymentMethod(ctx, gatewayReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment method in gateway: %w", err)
		}

		// Create payment method record
		method := &models.PaymentMethodInfo{
			ID:        uuid.New().String(),
			UserID:    req.UserID,
			Type:      req.Type,
			IsDefault: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Set card info if available
		if gatewayResult.Card != nil {
			method.CardInfo = &models.CardInfo{
				Last4:       gatewayResult.Card.Last4,
				Brand:       gatewayResult.Card.Brand,
				ExpiryMonth: gatewayResult.Card.ExpMonth,
				ExpiryYear:  gatewayResult.Card.ExpYear,
				HolderName:  req.Card.HolderName,
				Fingerprint: gatewayResult.Card.Fingerprint,
			}
		}

		// Save to database
		if err := s.repo.CreatePaymentMethod(ctx, method); err != nil {
			return nil, fmt.Errorf("failed to save payment method: %w", err)
		}

		utils.Logger.Info(ctx, "Payment method created successfully", map[string]interface{}{
			"method_id": method.ID,
			"user_id":   method.UserID,
			"type":      method.Type,
		})

		return method, nil
	}

	return nil, fmt.Errorf("unsupported payment method type: %s", req.Type)
}

// GetPaymentMethod retrieves a payment method by ID
func (s *paymentService) GetPaymentMethod(ctx context.Context, methodID string) (*models.PaymentMethodInfo, error) {
	return s.repo.GetPaymentMethodByID(ctx, methodID)
}

// GetUserPaymentMethods retrieves all payment methods for a user
func (s *paymentService) GetUserPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethodInfo, error) {
	return s.repo.GetPaymentMethodsByUserID(ctx, userID)
}

// UpdatePaymentMethod updates a payment method
func (s *paymentService) UpdatePaymentMethod(ctx context.Context, methodID string, req *UpdatePaymentMethodRequest) (*models.PaymentMethodInfo, error) {
	method, err := s.repo.GetPaymentMethodByID(ctx, methodID)
	if err != nil {
		return nil, fmt.Errorf("payment method not found: %w", err)
	}

	method.IsDefault = req.IsDefault
	method.UpdatedAt = time.Now()

	if err := s.repo.UpdatePaymentMethod(ctx, method); err != nil {
		return nil, fmt.Errorf("failed to update payment method: %w", err)
	}

	return method, nil
}

// DeletePaymentMethod deletes a payment method
func (s *paymentService) DeletePaymentMethod(ctx context.Context, methodID string) error {
	return s.repo.DeletePaymentMethod(ctx, methodID)
}

// SetDefaultPaymentMethod sets a payment method as default
func (s *paymentService) SetDefaultPaymentMethod(ctx context.Context, userID, methodID string) error {
	return s.repo.SetDefaultPaymentMethod(ctx, userID, methodID)
}

// CreateRefund creates a new refund
func (s *paymentService) CreateRefund(ctx context.Context, req *CreateRefundRequest) (*models.Refund, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("invalid refund request: %w", err)
	}

	// Get original payment
	payment, err := s.repo.GetPaymentByID(ctx, req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != models.PaymentCompleted {
		return nil, fmt.Errorf("payment cannot be refunded in current state: %s", payment.Status)
	}

	// Determine refund amount
	refundAmount := req.Amount
	if refundAmount.IsZero() {
		refundAmount = payment.Amount
	}

	// Validate refund amount
	if refundAmount.GreaterThan(payment.Amount) {
		return nil, fmt.Errorf("refund amount cannot exceed payment amount")
	}

	// Create refund record
	refund := models.NewRefund(req.PaymentID, payment.OrderID, refundAmount, payment.Currency, req.Reason)

	// Create refund in gateway
	gatewayReq := &gateway.CreateRefundRequest{
		PaymentIntentID: payment.TransactionID,
		Amount:          refundAmount,
		Reason:          req.Reason,
		Metadata:        req.Metadata,
	}

	gatewayResult, err := s.gateway.CreateRefund(ctx, gatewayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund in gateway: %w", err)
	}

	// Update refund with gateway response
	refund.TransactionID = gatewayResult.ID
	refund.Status = models.PaymentStatus(gatewayResult.Status)
	refund.GatewayResponse = models.GatewayResponse{
		GatewayID:       "stripe",
		TransactionID:   gatewayResult.ID,
		Status:          gatewayResult.Status,
		ResponseCode:    "200",
		ResponseMessage: "Refund created successfully",
		ProcessedAt:     time.Now(),
	}

	// Save refund to database
	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to save refund: %w", err)
	}

	utils.Logger.Info(ctx, "Refund created successfully", map[string]interface{}{
		"refund_id":  refund.ID,
		"payment_id": req.PaymentID,
		"amount":     refundAmount,
	})

	return refund, nil
}

// ProcessRefund processes a refund (placeholder for async processing)
func (s *paymentService) ProcessRefund(ctx context.Context, refundID string) (*models.Refund, error) {
	return s.repo.GetRefundByID(ctx, refundID)
}

// GetRefund retrieves a refund by ID
func (s *paymentService) GetRefund(ctx context.Context, refundID string) (*models.Refund, error) {
	return s.repo.GetRefundByID(ctx, refundID)
}

// GetPaymentRefunds retrieves all refunds for a payment
func (s *paymentService) GetPaymentRefunds(ctx context.Context, paymentID string) ([]*models.Refund, error) {
	return s.repo.GetRefundsByPaymentID(ctx, paymentID)
}

// ProcessWebhook processes webhook events from payment gateway
func (s *paymentService) ProcessWebhook(ctx context.Context, eventType string, payload []byte, signature string) error {
	// Verify webhook signature
	if err := s.gateway.VerifyWebhookSignature(payload, signature, ""); err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	// Parse webhook event
	event, err := s.gateway.ParseWebhookEvent(payload)
	if err != nil {
		return fmt.Errorf("failed to parse webhook event: %w", err)
	}

	// Create webhook record
	webhook := &models.PaymentWebhook{
		ID:        uuid.New().String(),
		EventType: event.Type,
		Status:    "received",
		Data:      event.Data,
		Signature: signature,
		Processed: false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateWebhook(ctx, webhook); err != nil {
		return fmt.Errorf("failed to save webhook: %w", err)
	}

	// Process webhook based on event type
	switch event.Type {
	case "payment_intent.succeeded":
		// Handle successful payment
		utils.Logger.Info(ctx, "Payment succeeded webhook received", map[string]interface{}{
			"event_id": event.ID,
		})
	case "payment_intent.payment_failed":
		// Handle failed payment
		utils.Logger.Info(ctx, "Payment failed webhook received", map[string]interface{}{
			"event_id": event.ID,
		})
	default:
		utils.Logger.Info(ctx, "Unhandled webhook event type", map[string]interface{}{
			"event_type": event.Type,
			"event_id":   event.ID,
		})
	}

	// Mark webhook as processed
	if err := s.repo.MarkWebhookProcessed(ctx, webhook.ID); err != nil {
		utils.Logger.Error(ctx, "Failed to mark webhook as processed", err)
	}

	return nil
}

// RetryFailedPayment retries a failed payment
func (s *paymentService) RetryFailedPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != models.PaymentFailed {
		return nil, fmt.Errorf("payment is not in failed state")
	}

	// Get previous attempts
	attempts, err := s.repo.GetPaymentAttempts(ctx, paymentID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get payment attempts", err)
	}

	attemptNumber := len(attempts) + 1
	if attemptNumber > 3 {
		return nil, fmt.Errorf("maximum retry attempts exceeded")
	}

	// Reset payment status and retry
	payment.Status = models.PaymentPending
	payment.FailureReason = ""

	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to reset payment status: %w", err)
	}

	// Process payment again
	return s.ProcessPayment(ctx, paymentID)
}

// ValidatePayment performs fraud detection and validation
func (s *paymentService) ValidatePayment(ctx context.Context, payment *models.Payment) error {
	// Basic validation rules
	if payment.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("invalid payment amount")
	}

	if payment.Amount.GreaterThan(decimal.NewFromInt(10000)) {
		utils.Logger.Info(ctx, "High-value payment detected", map[string]interface{}{
			"payment_id": payment.ID,
			"amount":     payment.Amount,
		})
		// Could trigger additional verification
	}

	// Additional fraud detection logic would go here
	// - Check user payment history
	// - Validate payment method
	// - Check for suspicious patterns
	// - Integration with fraud detection services

	return nil
}
