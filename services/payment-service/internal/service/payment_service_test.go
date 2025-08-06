package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/shopsphere/payment-service/internal/gateway"
	"github.com/shopsphere/payment-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPaymentRepository is a mock implementation of PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPaymentByID(ctx context.Context, paymentID string) (*models.Payment, error) {
	args := m.Called(ctx, paymentID)
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentsByOrderID(ctx context.Context, orderID string) ([]*models.Payment, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Payment, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdatePayment(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, status models.PaymentStatus, transactionID, failureReason string, gatewayResponse *models.GatewayResponse) error {
	args := m.Called(ctx, paymentID, status, transactionID, failureReason, gatewayResponse)
	return args.Error(0)
}

func (m *MockPaymentRepository) CreatePaymentMethod(ctx context.Context, method *models.PaymentMethodInfo) error {
	args := m.Called(ctx, method)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPaymentMethodByID(ctx context.Context, methodID string) (*models.PaymentMethodInfo, error) {
	args := m.Called(ctx, methodID)
	return args.Get(0).(*models.PaymentMethodInfo), args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentMethodsByUserID(ctx context.Context, userID string) ([]*models.PaymentMethodInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.PaymentMethodInfo), args.Error(1)
}

func (m *MockPaymentRepository) UpdatePaymentMethod(ctx context.Context, method *models.PaymentMethodInfo) error {
	args := m.Called(ctx, method)
	return args.Error(0)
}

func (m *MockPaymentRepository) DeletePaymentMethod(ctx context.Context, methodID string) error {
	args := m.Called(ctx, methodID)
	return args.Error(0)
}

func (m *MockPaymentRepository) SetDefaultPaymentMethod(ctx context.Context, userID, methodID string) error {
	args := m.Called(ctx, userID, methodID)
	return args.Error(0)
}

func (m *MockPaymentRepository) CreateRefund(ctx context.Context, refund *models.Refund) error {
	args := m.Called(ctx, refund)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetRefundByID(ctx context.Context, refundID string) (*models.Refund, error) {
	args := m.Called(ctx, refundID)
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockPaymentRepository) GetRefundsByPaymentID(ctx context.Context, paymentID string) ([]*models.Refund, error) {
	args := m.Called(ctx, paymentID)
	return args.Get(0).([]*models.Refund), args.Error(1)
}

func (m *MockPaymentRepository) UpdateRefund(ctx context.Context, refund *models.Refund) error {
	args := m.Called(ctx, refund)
	return args.Error(0)
}

func (m *MockPaymentRepository) CreateWebhook(ctx context.Context, webhook *models.PaymentWebhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetWebhookByID(ctx context.Context, webhookID string) (*models.PaymentWebhook, error) {
	args := m.Called(ctx, webhookID)
	return args.Get(0).(*models.PaymentWebhook), args.Error(1)
}

func (m *MockPaymentRepository) GetUnprocessedWebhooks(ctx context.Context, limit int) ([]*models.PaymentWebhook, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*models.PaymentWebhook), args.Error(1)
}

func (m *MockPaymentRepository) MarkWebhookProcessed(ctx context.Context, webhookID string) error {
	args := m.Called(ctx, webhookID)
	return args.Error(0)
}

func (m *MockPaymentRepository) CreatePaymentAttempt(ctx context.Context, paymentID string, attemptNumber int, status, failureReason string, gatewayResponse *models.GatewayResponse) error {
	args := m.Called(ctx, paymentID, attemptNumber, status, failureReason, gatewayResponse)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPaymentAttempts(ctx context.Context, paymentID string) ([]repository.PaymentAttempt, error) {
	args := m.Called(ctx, paymentID)
	return args.Get(0).([]repository.PaymentAttempt), args.Error(1)
}

// MockPaymentGateway is a mock implementation of PaymentGateway
type MockPaymentGateway struct {
	mock.Mock
}

func (m *MockPaymentGateway) CreatePaymentIntent(ctx context.Context, req *gateway.CreatePaymentIntentRequest) (*models.PaymentIntent, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.PaymentIntent), args.Error(1)
}

func (m *MockPaymentGateway) ConfirmPayment(ctx context.Context, paymentIntentID string) (*gateway.PaymentResult, error) {
	args := m.Called(ctx, paymentIntentID)
	return args.Get(0).(*gateway.PaymentResult), args.Error(1)
}

func (m *MockPaymentGateway) CapturePayment(ctx context.Context, paymentIntentID string, amount decimal.Decimal) (*gateway.PaymentResult, error) {
	args := m.Called(ctx, paymentIntentID, amount)
	return args.Get(0).(*gateway.PaymentResult), args.Error(1)
}

func (m *MockPaymentGateway) CreatePaymentMethod(ctx context.Context, req *gateway.CreatePaymentMethodRequest) (*gateway.PaymentMethodResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*gateway.PaymentMethodResult), args.Error(1)
}

func (m *MockPaymentGateway) AttachPaymentMethodToCustomer(ctx context.Context, paymentMethodID, customerID string) error {
	args := m.Called(ctx, paymentMethodID, customerID)
	return args.Error(0)
}

func (m *MockPaymentGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	args := m.Called(ctx, paymentMethodID)
	return args.Error(0)
}

func (m *MockPaymentGateway) CreateCustomer(ctx context.Context, req *gateway.CreateCustomerRequest) (*gateway.CustomerResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*gateway.CustomerResult), args.Error(1)
}

func (m *MockPaymentGateway) GetCustomer(ctx context.Context, customerID string) (*gateway.CustomerResult, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).(*gateway.CustomerResult), args.Error(1)
}

func (m *MockPaymentGateway) CreateRefund(ctx context.Context, req *gateway.CreateRefundRequest) (*gateway.RefundResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*gateway.RefundResult), args.Error(1)
}

func (m *MockPaymentGateway) VerifyWebhookSignature(payload []byte, signature, secret string) error {
	args := m.Called(payload, signature, secret)
	return args.Error(0)
}

func (m *MockPaymentGateway) ParseWebhookEvent(payload []byte) (*gateway.WebhookEvent, error) {
	args := m.Called(payload)
	return args.Get(0).(*gateway.WebhookEvent), args.Error(1)
}

// Test cases
func TestPaymentService_CreatePayment(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	mockGateway := new(MockPaymentGateway)
	service := NewPaymentService(mockRepo, mockGateway)

	ctx := context.Background()
	req := &CreatePaymentRequest{
		OrderID:         "order-123",
		UserID:          "user-123",
		Amount:          decimal.NewFromFloat(100.00),
		Currency:        "USD",
		PaymentMethodID: "pm_123",
		Description:     "Test payment",
		AutoCapture:     true,
	}

	// Mock payment intent creation
	mockGateway.On("CreatePaymentIntent", ctx, mock.AnythingOfType("*gateway.CreatePaymentIntentRequest")).Return(&models.PaymentIntent{
		ID:                "pi_123",
		Amount:            decimal.NewFromFloat(100.00),
		Currency:          "USD",
		PaymentMethodID:   "pm_123",
		Status:            "requires_confirmation",
		ClientSecret:      "pi_123_secret",
		CreatedAt:         time.Now(),
	}, nil)

	// Mock payment creation
	mockRepo.On("CreatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)

	payment, err := service.CreatePayment(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, req.OrderID, payment.OrderID)
	assert.Equal(t, req.UserID, payment.UserID)
	assert.Equal(t, req.Amount, payment.Amount)
	assert.Equal(t, req.Currency, payment.Currency)
	assert.Equal(t, "pi_123", payment.TransactionID)

	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}

func TestPaymentService_ProcessPayment(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	mockGateway := new(MockPaymentGateway)
	service := NewPaymentService(mockRepo, mockGateway)

	ctx := context.Background()
	paymentID := "payment-123"

	// Mock payment retrieval
	payment := &models.Payment{
		ID:            paymentID,
		OrderID:       "order-123",
		UserID:        "user-123",
		Amount:        decimal.NewFromFloat(100.00),
		Currency:      "USD",
		Status:        models.PaymentPending,
		TransactionID: "pi_123",
		CreatedAt:     time.Now(),
	}
	mockRepo.On("GetPaymentByID", ctx, paymentID).Return(payment, nil).Once()

	// Mock payment status update
	mockRepo.On("UpdatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)

	// Mock payment attempt creation
	mockRepo.On("CreatePaymentAttempt", ctx, paymentID, 1, "processing", "", mock.AnythingOfType("*models.GatewayResponse")).Return(nil)

	// Mock payment confirmation
	mockGateway.On("ConfirmPayment", ctx, "pi_123").Return(&gateway.PaymentResult{
		ID:              "pi_123",
		Status:          "succeeded",
		Amount:          decimal.NewFromFloat(100.00),
		Currency:        "USD",
		PaymentMethodID: "pm_123",
		CreatedAt:       time.Now(),
	}, nil)

	// Mock payment status update to completed
	mockRepo.On("UpdatePaymentStatus", ctx, paymentID, models.PaymentCompleted, "pi_123", "", mock.AnythingOfType("*models.GatewayResponse")).Return(nil)

	// Mock final payment retrieval
	completedPayment := *payment
	completedPayment.Status = models.PaymentCompleted
	mockRepo.On("GetPaymentByID", ctx, paymentID).Return(&completedPayment, nil).Once()

	result, err := service.ProcessPayment(ctx, paymentID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.PaymentCompleted, result.Status)

	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}

func TestPaymentService_CreatePaymentMethod(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	mockGateway := new(MockPaymentGateway)
	service := NewPaymentService(mockRepo, mockGateway)

	ctx := context.Background()
	req := &CreatePaymentMethodRequest{
		UserID: "user-123",
		Type:   models.PaymentTypeCard,
		Card: &CardDetailsRequest{
			Number:     "4242424242424242",
			ExpMonth:   12,
			ExpYear:    2025,
			CVC:        "123",
			HolderName: "John Doe",
		},
	}

	// Mock gateway payment method creation
	mockGateway.On("CreatePaymentMethod", ctx, mock.AnythingOfType("*gateway.CreatePaymentMethodRequest")).Return(&gateway.PaymentMethodResult{
		ID:   "pm_123",
		Type: "card",
		Card: &gateway.CardInfo{
			Brand:       "visa",
			Last4:       "4242",
			ExpMonth:    12,
			ExpYear:     2025,
			Fingerprint: "fingerprint123",
		},
		CreatedAt: time.Now(),
	}, nil)

	// Mock repository payment method creation
	mockRepo.On("CreatePaymentMethod", ctx, mock.AnythingOfType("*models.PaymentMethodInfo")).Return(nil)

	method, err := service.CreatePaymentMethod(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, method)
	assert.Equal(t, req.UserID, method.UserID)
	assert.Equal(t, req.Type, method.Type)
	assert.NotNil(t, method.CardInfo)
	assert.Equal(t, "4242", method.CardInfo.Last4)
	assert.Equal(t, "visa", method.CardInfo.Brand)

	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}

func TestPaymentService_CreateRefund(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	mockGateway := new(MockPaymentGateway)
	service := NewPaymentService(mockRepo, mockGateway)

	ctx := context.Background()
	req := &CreateRefundRequest{
		PaymentID: "payment-123",
		Amount:    decimal.NewFromFloat(50.00),
		Reason:    "Customer request",
	}

	// Mock payment retrieval
	payment := &models.Payment{
		ID:            "payment-123",
		OrderID:       "order-123",
		Amount:        decimal.NewFromFloat(100.00),
		Currency:      "USD",
		Status:        models.PaymentCompleted,
		TransactionID: "pi_123",
	}
	mockRepo.On("GetPaymentByID", ctx, req.PaymentID).Return(payment, nil)

	// Mock gateway refund creation
	mockGateway.On("CreateRefund", ctx, mock.AnythingOfType("*gateway.CreateRefundRequest")).Return(&gateway.RefundResult{
		ID:              "re_123",
		PaymentIntentID: "pi_123",
		Amount:          decimal.NewFromFloat(50.00),
		Currency:        "USD",
		Status:          "succeeded",
		Reason:          "requested_by_customer",
		CreatedAt:       time.Now(),
	}, nil)

	// Mock repository refund creation
	mockRepo.On("CreateRefund", ctx, mock.AnythingOfType("*models.Refund")).Return(nil)

	refund, err := service.CreateRefund(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, refund)
	assert.Equal(t, req.PaymentID, refund.PaymentID)
	assert.Equal(t, req.Amount, refund.Amount)
	assert.Equal(t, req.Reason, refund.Reason)
	assert.Equal(t, "re_123", refund.TransactionID)

	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}

func TestPaymentService_CancelPayment(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	mockGateway := new(MockPaymentGateway)
	service := NewPaymentService(mockRepo, mockGateway)

	ctx := context.Background()
	paymentID := "payment-123"
	reason := "Order cancelled"

	// Mock payment retrieval
	payment := &models.Payment{
		ID:            paymentID,
		Status:        models.PaymentPending,
		TransactionID: "pi_123",
	}
	mockRepo.On("GetPaymentByID", ctx, paymentID).Return(payment, nil)

	// Mock payment status update
	mockRepo.On("UpdatePaymentStatus", ctx, paymentID, models.PaymentCancelled, "pi_123", reason, mock.AnythingOfType("*models.GatewayResponse")).Return(nil)

	err := service.CancelPayment(ctx, paymentID, reason)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}

func TestPaymentService_ProcessWebhook(t *testing.T) {
	mockRepo := new(MockPaymentRepository)
	mockGateway := new(MockPaymentGateway)
	service := NewPaymentService(mockRepo, mockGateway)

	ctx := context.Background()
	eventType := "stripe"
	payload := []byte(`{"id": "evt_123", "type": "payment_intent.succeeded"}`)
	signature := "test_signature"

	// Mock webhook signature verification
	mockGateway.On("VerifyWebhookSignature", payload, signature, "").Return(nil)

	// Mock webhook event parsing
	mockGateway.On("ParseWebhookEvent", payload).Return(&gateway.WebhookEvent{
		ID:      "evt_123",
		Type:    "payment_intent.succeeded",
		Data:    map[string]interface{}{"id": "pi_123"},
		Created: time.Now(),
	}, nil)

	// Mock webhook creation
	mockRepo.On("CreateWebhook", ctx, mock.AnythingOfType("*models.PaymentWebhook")).Return(nil)

	// Mock webhook processed marking
	mockRepo.On("MarkWebhookProcessed", ctx, mock.AnythingOfType("string")).Return(nil)

	err := service.ProcessWebhook(ctx, eventType, payload, signature)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}
