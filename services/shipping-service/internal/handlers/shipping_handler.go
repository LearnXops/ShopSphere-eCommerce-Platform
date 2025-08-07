package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/shipping-service/internal/service"
)

// ShippingHandler handles HTTP requests for shipping operations
type ShippingHandler struct {
	service *service.ShippingService
	logger  *utils.StructuredLogger
}

// NewShippingHandler creates a new shipping handler
func NewShippingHandler(service *service.ShippingService, logger *utils.StructuredLogger) *ShippingHandler {
	return &ShippingHandler{
		service: service,
		logger:  logger,
	}
}

// GetShippingQuotes handles POST /quotes
func (h *ShippingHandler) GetShippingQuotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request models.ShippingQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error(ctx, "Failed to decode shipping quote request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate request
	if err := h.validateShippingQuoteRequest(&request); err != nil {
		h.logger.Error(ctx, "Invalid shipping quote request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	quotes, err := h.service.GetShippingQuotes(ctx, &request)
	if err != nil {
		h.logger.Error(ctx, "Failed to get shipping quotes", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "QUOTE_ERROR", "Failed to calculate shipping quotes")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"quotes": quotes,
	})
}

// CreateShipment handles POST /shipments
func (h *ShippingHandler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		OrderID          string          `json:"order_id"`
		UserID           string          `json:"user_id"`
		ShippingMethodID string          `json:"shipping_method_id"`
		FromAddress      models.Address  `json:"from_address"`
		ToAddress        models.Address  `json:"to_address"`
		WeightKg         decimal.Decimal `json:"weight_kg"`
		DeclaredValue    decimal.Decimal `json:"declared_value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error(ctx, "Failed to decode create shipment request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate request
	if request.OrderID == "" || request.UserID == "" || request.ShippingMethodID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Missing required fields")
		return
	}

	if request.WeightKg.LessThanOrEqual(decimal.Zero) {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Weight must be greater than zero")
		return
	}

	shipment, err := h.service.CreateShipment(ctx, request.OrderID, request.UserID, request.ShippingMethodID, 
		request.FromAddress, request.ToAddress, request.WeightKg, request.DeclaredValue)
	if err != nil {
		h.logger.Error(ctx, "Failed to create shipment", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "SHIPMENT_ERROR", "Failed to create shipment")
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, shipment)
}

// GetShipment handles GET /shipments/{id}
func (h *ShippingHandler) GetShipment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Shipment ID is required")
		return
	}

	shipment, err := h.service.GetShipment(ctx, id)
	if err != nil {
		h.logger.Error(ctx, "Failed to get shipment", err)
		utils.WriteErrorResponse(w, http.StatusNotFound, "SHIPMENT_NOT_FOUND", "Shipment not found")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, shipment)
}

// GetShipmentsByOrder handles GET /orders/{order_id}/shipments
func (h *ShippingHandler) GetShipmentsByOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	if orderID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Order ID is required")
		return
	}

	shipments, err := h.service.GetShipmentsByOrder(ctx, orderID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get shipments by order", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "SHIPMENT_ERROR", "Failed to get shipments")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"shipments": shipments,
	})
}

// TrackShipment handles GET /track/{tracking_number}
func (h *ShippingHandler) TrackShipment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	trackingNumber := vars["tracking_number"]

	if trackingNumber == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Tracking number is required")
		return
	}

	trackingInfo, err := h.service.TrackShipment(ctx, trackingNumber)
	if err != nil {
		h.logger.Error(ctx, "Failed to track shipment", err)
		utils.WriteErrorResponse(w, http.StatusNotFound, "TRACKING_ERROR", "Failed to track shipment")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, trackingInfo)
}

// UpdateShipmentStatus handles PUT /shipments/{id}/status
func (h *ShippingHandler) UpdateShipmentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Shipment ID is required")
		return
	}

	var request struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error(ctx, "Failed to decode status update request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	status := models.ShipmentStatus(request.Status)
	if !status.IsValid() {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid shipment status")
		return
	}

	err := h.service.UpdateShipmentStatus(ctx, id, status)
	if err != nil {
		h.logger.Error(ctx, "Failed to update shipment status", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "STATUS_UPDATE_ERROR", "Failed to update shipment status")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Shipment status updated successfully",
		"status":  status,
	})
}

// GetShippingMethods handles GET /methods
func (h *ShippingHandler) GetShippingMethods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	methods, err := h.service.GetShippingMethods(ctx)
	if err != nil {
		h.logger.Error(ctx, "Failed to get shipping methods", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "METHODS_ERROR", "Failed to get shipping methods")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"methods": methods,
	})
}

// ValidateAddress handles POST /validate-address
func (h *ShippingHandler) ValidateAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var address models.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		h.logger.Error(ctx, "Failed to decode address validation request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	response, err := h.service.ValidateAddress(ctx, &address)
	if err != nil {
		h.logger.Error(ctx, "Failed to validate address", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "VALIDATION_ERROR", "Failed to validate address")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetShipmentByTrackingNumber handles GET /shipments/tracking/{tracking_number}
func (h *ShippingHandler) GetShipmentByTrackingNumber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	trackingNumber := vars["tracking_number"]

	if trackingNumber == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Tracking number is required")
		return
	}

	shipment, err := h.service.GetShipmentByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		h.logger.Error(ctx, "Failed to get shipment by tracking number", err)
		utils.WriteErrorResponse(w, http.StatusNotFound, "SHIPMENT_NOT_FOUND", "Shipment not found")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, shipment)
}

// HealthCheck handles GET /health
func (h *ShippingHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "shipping-service",
	})
}

// RegisterRoutes registers all shipping routes
func (h *ShippingHandler) RegisterRoutes(router *mux.Router) {
	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// Shipping quotes
	router.HandleFunc("/quotes", h.GetShippingQuotes).Methods("POST")

	// Shipping methods
	router.HandleFunc("/methods", h.GetShippingMethods).Methods("GET")

	// Shipments
	router.HandleFunc("/shipments", h.CreateShipment).Methods("POST")
	router.HandleFunc("/shipments/{id}", h.GetShipment).Methods("GET")
	router.HandleFunc("/shipments/{id}/status", h.UpdateShipmentStatus).Methods("PUT")
	router.HandleFunc("/shipments/tracking/{tracking_number}", h.GetShipmentByTrackingNumber).Methods("GET")

	// Order shipments
	router.HandleFunc("/orders/{order_id}/shipments", h.GetShipmentsByOrder).Methods("GET")

	// Tracking
	router.HandleFunc("/track/{tracking_number}", h.TrackShipment).Methods("GET")

	// Address validation
	router.HandleFunc("/validate-address", h.ValidateAddress).Methods("POST")
}

// Helper methods

// validateShippingQuoteRequest validates a shipping quote request
func (h *ShippingHandler) validateShippingQuoteRequest(request *models.ShippingQuoteRequest) error {
	if request.WeightKg.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("weight must be greater than zero")
	}

	if request.FromAddress.Country == "" || request.ToAddress.Country == "" {
		return fmt.Errorf("from and to addresses must include country")
	}

	if request.FromAddress.PostalCode == "" || request.ToAddress.PostalCode == "" {
		return fmt.Errorf("from and to addresses must include postal code")
	}

	if request.DeclaredValue.LessThan(decimal.Zero) {
		return fmt.Errorf("declared value cannot be negative")
	}

	if request.OrderValue.LessThan(decimal.Zero) {
		return fmt.Errorf("order value cannot be negative")
	}

	return nil
}

// parseQueryParams parses common query parameters
func (h *ShippingHandler) parseQueryParams(r *http.Request) (limit, offset int) {
	limit = 20 // default
	offset = 0 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}
