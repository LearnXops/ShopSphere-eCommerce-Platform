package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/payment-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

// PaymentHandler handles HTTP requests for payment operations
type PaymentHandler struct {
	service service.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(service service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

// RegisterRoutes registers payment routes
func (h *PaymentHandler) RegisterRoutes(router *mux.Router) {
	// Payment routes
	router.HandleFunc("/payments", h.CreatePayment).Methods("POST")
	router.HandleFunc("/payments/{id}", h.GetPayment).Methods("GET")
	router.HandleFunc("/payments/{id}/process", h.ProcessPayment).Methods("POST")
	router.HandleFunc("/payments/{id}/cancel", h.CancelPayment).Methods("POST")
	router.HandleFunc("/payments/{id}/retry", h.RetryPayment).Methods("POST")
	router.HandleFunc("/orders/{order_id}/payments", h.GetPaymentsByOrder).Methods("GET")
	router.HandleFunc("/users/{user_id}/payments", h.GetUserPayments).Methods("GET")

	// Payment method routes
	router.HandleFunc("/payment-methods", h.CreatePaymentMethod).Methods("POST")
	router.HandleFunc("/payment-methods/{id}", h.GetPaymentMethod).Methods("GET")
	router.HandleFunc("/payment-methods/{id}", h.UpdatePaymentMethod).Methods("PUT")
	router.HandleFunc("/payment-methods/{id}", h.DeletePaymentMethod).Methods("DELETE")
	router.HandleFunc("/users/{user_id}/payment-methods", h.GetUserPaymentMethods).Methods("GET")
	router.HandleFunc("/users/{user_id}/payment-methods/{method_id}/default", h.SetDefaultPaymentMethod).Methods("POST")

	// Refund routes
	router.HandleFunc("/refunds", h.CreateRefund).Methods("POST")
	router.HandleFunc("/refunds/{id}", h.GetRefund).Methods("GET")
	router.HandleFunc("/payments/{payment_id}/refunds", h.GetPaymentRefunds).Methods("GET")

	// Webhook routes
	router.HandleFunc("/webhooks/stripe", h.HandleStripeWebhook).Methods("POST")
}

// CreatePayment creates a new payment
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req service.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	payment, err := h.service.CreatePayment(ctx, &req)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to create payment", err, map[string]interface{}{
			"order_id": req.OrderID,
			"user_id":  req.UserID,
			"amount":   req.Amount,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_CREATION_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, payment)
}

// GetPayment retrieves a payment by ID
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	paymentID := vars["id"]

	payment, err := h.service.GetPayment(ctx, paymentID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get payment", err, map[string]interface{}{
			"payment_id": paymentID,
		})
		utils.WriteErrorResponse(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, payment)
}

// ProcessPayment processes a payment
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	paymentID := vars["id"]

	payment, err := h.service.ProcessPayment(ctx, paymentID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to process payment", err, map[string]interface{}{
			"payment_id": paymentID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_PROCESSING_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, payment)
}

// CancelPayment cancels a payment
func (h *PaymentHandler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	paymentID := vars["id"]

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.CancelPayment(ctx, paymentID, req.Reason); err != nil {
		utils.Logger.Error(ctx, "Failed to cancel payment", err, map[string]interface{}{
			"payment_id": paymentID,
			"reason":     req.Reason,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_CANCELLATION_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Payment cancelled successfully",
	})
}

// RetryPayment retries a failed payment
func (h *PaymentHandler) RetryPayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	paymentID := vars["id"]

	payment, err := h.service.RetryFailedPayment(ctx, paymentID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to retry payment", err, map[string]interface{}{
			"payment_id": paymentID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_RETRY_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, payment)
}

// GetPaymentsByOrder retrieves payments for an order
func (h *PaymentHandler) GetPaymentsByOrder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	payments, err := h.service.GetPaymentsByOrder(ctx, orderID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get payments by order", err, map[string]interface{}{
			"order_id": orderID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENTS_RETRIEVAL_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"count":    len(payments),
	})
}

// GetUserPayments retrieves payments for a user with pagination
func (h *PaymentHandler) GetUserPayments(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	userID := vars["user_id"]

	// Parse pagination parameters
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	payments, err := h.service.GetUserPayments(ctx, userID, limit, offset)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get user payments", err, map[string]interface{}{
			"user_id": userID,
			"limit":   limit,
			"offset":  offset,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENTS_RETRIEVAL_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"count":    len(payments),
		"limit":    limit,
		"offset":   offset,
	})
}

// CreatePaymentMethod creates a new payment method
func (h *PaymentHandler) CreatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req service.CreatePaymentMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	method, err := h.service.CreatePaymentMethod(ctx, &req)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to create payment method", err, map[string]interface{}{
			"user_id": req.UserID,
			"type":    req.Type,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_METHOD_CREATION_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, method)
}

// GetPaymentMethod retrieves a payment method by ID
func (h *PaymentHandler) GetPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	methodID := vars["id"]

	method, err := h.service.GetPaymentMethod(ctx, methodID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get payment method", err, map[string]interface{}{
			"method_id": methodID,
		})
		utils.WriteErrorResponse(w, http.StatusNotFound, "PAYMENT_METHOD_NOT_FOUND", "Payment method not found")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, method)
}

// UpdatePaymentMethod updates a payment method
func (h *PaymentHandler) UpdatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	methodID := vars["id"]

	var req service.UpdatePaymentMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	method, err := h.service.UpdatePaymentMethod(ctx, methodID, &req)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to update payment method", err, map[string]interface{}{
			"method_id": methodID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_METHOD_UPDATE_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, method)
}

// DeletePaymentMethod deletes a payment method
func (h *PaymentHandler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	methodID := vars["id"]

	if err := h.service.DeletePaymentMethod(ctx, methodID); err != nil {
		utils.Logger.Error(ctx, "Failed to delete payment method", err, map[string]interface{}{
			"method_id": methodID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_METHOD_DELETION_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Payment method deleted successfully",
	})
}

// GetUserPaymentMethods retrieves payment methods for a user
func (h *PaymentHandler) GetUserPaymentMethods(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	userID := vars["user_id"]

	methods, err := h.service.GetUserPaymentMethods(ctx, userID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get user payment methods", err, map[string]interface{}{
			"user_id": userID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "PAYMENT_METHODS_RETRIEVAL_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"payment_methods": methods,
		"count":           len(methods),
	})
}

// SetDefaultPaymentMethod sets a payment method as default
func (h *PaymentHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	userID := vars["user_id"]
	methodID := vars["method_id"]

	if err := h.service.SetDefaultPaymentMethod(ctx, userID, methodID); err != nil {
		utils.Logger.Error(ctx, "Failed to set default payment method", err, map[string]interface{}{
			"user_id":   userID,
			"method_id": methodID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "DEFAULT_PAYMENT_METHOD_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Default payment method set successfully",
	})
}

// CreateRefund creates a new refund
func (h *PaymentHandler) CreateRefund(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req service.CreateRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	refund, err := h.service.CreateRefund(ctx, &req)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to create refund", err, map[string]interface{}{
			"payment_id": req.PaymentID,
			"amount":     req.Amount,
			"reason":     req.Reason,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "REFUND_CREATION_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, refund)
}

// GetRefund retrieves a refund by ID
func (h *PaymentHandler) GetRefund(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	refundID := vars["id"]

	refund, err := h.service.GetRefund(ctx, refundID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get refund", err, map[string]interface{}{
			"refund_id": refundID,
		})
		utils.WriteErrorResponse(w, http.StatusNotFound, "REFUND_NOT_FOUND", "Refund not found")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, refund)
}

// GetPaymentRefunds retrieves refunds for a payment
func (h *PaymentHandler) GetPaymentRefunds(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	paymentID := vars["payment_id"]

	refunds, err := h.service.GetPaymentRefunds(ctx, paymentID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get payment refunds", err, map[string]interface{}{
			"payment_id": paymentID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "REFUNDS_RETRIEVAL_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"refunds": refunds,
		"count":   len(refunds),
	})
}

// HandleStripeWebhook handles Stripe webhook events
func (h *PaymentHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to read webhook payload", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to read request body")
		return
	}

	// Get Stripe signature header
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "MISSING_SIGNATURE", "Missing Stripe signature")
		return
	}

	// Process webhook
	if err := h.service.ProcessWebhook(ctx, "stripe", payload, signature); err != nil {
		utils.Logger.Error(ctx, "Failed to process webhook", err, map[string]interface{}{
			"signature": signature,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "WEBHOOK_PROCESSING_FAILED", err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Webhook processed successfully",
	})
}
