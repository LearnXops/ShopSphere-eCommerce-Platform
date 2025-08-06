package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/shopsphere/order-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	service service.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req service.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	order, err := h.service.CreateOrder(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_ORDER", err.Error())
		} else if strings.Contains(err.Error(), "stock") {
			utils.WriteErrorResponse(w, http.StatusConflict, "INSUFFICIENT_STOCK", err.Error())
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, order)
}

// GetOrder handles GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Order ID is required", "Request validation failed")
		return
	}

	order, err := h.service.GetOrder(ctx, orderID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Order not found", err.Error())
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get order", err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, order)
}

// GetUserOrders handles GET /orders/user/{userId}
func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["userId"]

	if userID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "User ID is required", "Request validation failed")
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := h.service.GetOrdersByUser(ctx, userID, limit, offset)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user orders", err.Error())
		return
	}

	response := map[string]interface{}{
		"orders": orders,
		"limit":  limit,
		"offset": offset,
		"count":  len(orders),
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdateOrder handles PUT /orders/{id}
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Order ID is required", "Request validation failed")
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Ensure the ID matches the URL parameter
	order.ID = orderID

	if err := h.service.UpdateOrder(ctx, &order); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Order not found", err.Error())
		} else if strings.Contains(err.Error(), "cannot update") {
			utils.WriteErrorResponse(w, http.StatusConflict, "Cannot update order", err.Error())
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update order", err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Order updated successfully"})
}

// UpdateOrderStatus handles PATCH /orders/{id}/status
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Order ID is required", "Request validation failed")
		return
	}

	var req struct {
		Status    models.OrderStatus `json:"status" validate:"required"`
		Reason    string            `json:"reason"`
		ChangedBy string            `json:"changed_by" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.service.UpdateOrderStatus(ctx, orderID, req.Status, req.Reason, req.ChangedBy); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Order not found", err.Error())
		} else if strings.Contains(err.Error(), "invalid status transition") {
			utils.WriteErrorResponse(w, http.StatusConflict, "Invalid status transition", err.Error())
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update order status", err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Order status updated successfully"})
}

// CancelOrder handles POST /orders/{id}/cancel
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Order ID is required", "Request validation failed")
		return
	}

	var req struct {
		Reason      string `json:"reason"`
		CancelledBy string `json:"cancelled_by" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := h.service.CancelOrder(ctx, orderID, req.Reason, req.CancelledBy); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Order not found", err.Error())
		} else if strings.Contains(err.Error(), "cannot cancel") {
			utils.WriteErrorResponse(w, http.StatusConflict, "Cannot cancel order", err.Error())
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to cancel order", err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Order cancelled successfully"})
}

// SearchOrders handles GET /orders/search
func (h *OrderHandler) SearchOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	// Build filters from query parameters
	filters := make(map[string]interface{})

	if userID := query.Get("user_id"); userID != "" {
		filters["user_id"] = userID
	}

	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}

	if orderNumber := query.Get("order_number"); orderNumber != "" {
		filters["order_number"] = orderNumber
	}

	if dateFrom := query.Get("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}

	if dateTo := query.Get("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := h.service.SearchOrders(ctx, filters, limit, offset)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to search orders", err.Error())
		return
	}

	response := map[string]interface{}{
		"orders":  orders,
		"filters": filters,
		"limit":   limit,
		"offset":  offset,
		"count":   len(orders),
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetOrderStatusHistory handles GET /orders/{id}/history
func (h *OrderHandler) GetOrderStatusHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Order ID is required", "Request validation failed")
		return
	}

	history, err := h.service.GetOrderStatusHistory(ctx, orderID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Order not found", err.Error())
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get order history", err.Error())
		}
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"order_id": orderID,
		"history":  history,
	})
}

// ValidateOrder handles POST /orders/validate
func (h *OrderHandler) ValidateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req service.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate order items
	if err := h.service.ValidateOrderItems(ctx, req.Items); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Order validation failed", err.Error())
		return
	}

	// Calculate totals
	totals, err := h.service.CalculateOrderTotals(ctx, &req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to calculate totals", err.Error())
		return
	}

	response := map[string]interface{}{
		"valid":  true,
		"totals": totals,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetOrderSummary handles GET /orders/{id}/summary
func (h *OrderHandler) GetOrderSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	orderID := vars["id"]

	if orderID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Order ID is required", "Request validation failed")
		return
	}

	order, err := h.service.GetOrder(ctx, orderID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Order not found", err.Error())
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get order", err.Error())
		}
		return
	}

	// Create summary with essential information
	summary := map[string]interface{}{
		"id":           order.ID,
		"order_number": order.OrderNumber,
		"status":       order.Status,
		"total":        order.Total,
		"currency":     order.Currency,
		"item_count":   len(order.Items),
		"created_at":   order.CreatedAt,
		"updated_at":   order.UpdatedAt,
	}

	// Add status-specific timestamps
	if order.ConfirmedAt != nil {
		summary["confirmed_at"] = order.ConfirmedAt
	}
	if order.ShippedAt != nil {
		summary["shipped_at"] = order.ShippedAt
	}
	if order.DeliveredAt != nil {
		summary["delivered_at"] = order.DeliveredAt
	}
	if order.CancelledAt != nil {
		summary["cancelled_at"] = order.CancelledAt
	}

	utils.WriteJSONResponse(w, http.StatusOK, summary)
}

// HealthCheck handles GET /health
func (h *OrderHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "order-service",
	})
}
