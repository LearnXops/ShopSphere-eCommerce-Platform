package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopsphere/cart-service/internal/service"
	"github.com/shopsphere/shared/utils"
	"github.com/shopspring/decimal"
)

// CartHandler handles HTTP requests for cart operations
type CartHandler struct {
	cartService service.CartService
}

// NewCartHandler creates a new cart handler
func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// AddItemRequest represents the request to add an item to cart
type AddItemRequest struct {
	ProductID string          `json:"product_id" validate:"required"`
	SKU       string          `json:"sku" validate:"required"`
	Name      string          `json:"name" validate:"required"`
	Price     decimal.Decimal `json:"price" validate:"required"`
	Quantity  int             `json:"quantity" validate:"required,min=1"`
}

// UpdateItemRequest represents the request to update an item in cart
type UpdateItemRequest struct {
	Quantity int `json:"quantity" validate:"min=0"`
}

// MigrateCartRequest represents the request to migrate a guest cart
type MigrateCartRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	UserID    string `json:"user_id" validate:"required"`
}

// ExtendExpiryRequest represents the request to extend cart expiry
type ExtendExpiryRequest struct {
	Hours int `json:"hours" validate:"required,min=1,max=168"` // Max 7 days
}

// GetCart retrieves the current cart for a user or session
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	cart, err := h.cartService.GetCart(ctx, userID, sessionID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get cart", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "CART_RETRIEVAL_FAILED", "Failed to retrieve cart")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, cart)
}

// AddItem adds an item to the cart
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	cart, err := h.cartService.AddItem(ctx, userID, sessionID, req.ProductID, req.SKU, req.Name, req.Price, req.Quantity)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to add item to cart", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
			"product_id": req.ProductID,
			"quantity":   req.Quantity,
		})
		
		if err.Error() == "insufficient stock for product "+req.ProductID {
			utils.WriteErrorResponse(w, http.StatusConflict, "INSUFFICIENT_STOCK", "Insufficient stock for the requested quantity")
			return
		}
		
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "ADD_ITEM_FAILED", "Failed to add item to cart")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, cart)
}

// UpdateItem updates the quantity of an item in the cart
func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")
	
	vars := mux.Vars(r)
	productID := vars["productId"]

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	cart, err := h.cartService.UpdateItem(ctx, userID, sessionID, productID, req.Quantity)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to update cart item", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
			"product_id": productID,
			"quantity":   req.Quantity,
		})
		
		if err.Error() == "item not found in cart" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found in cart")
			return
		}
		
		if err.Error() == "insufficient stock for product "+productID {
			utils.WriteErrorResponse(w, http.StatusConflict, "INSUFFICIENT_STOCK", "Insufficient stock for the requested quantity")
			return
		}
		
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "UPDATE_ITEM_FAILED", "Failed to update cart item")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, cart)
}

// RemoveItem removes an item from the cart
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")
	
	vars := mux.Vars(r)
	productID := vars["productId"]

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	cart, err := h.cartService.RemoveItem(ctx, userID, sessionID, productID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to remove cart item", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
			"product_id": productID,
		})
		
		if err.Error() == "item not found in cart" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found in cart")
			return
		}
		
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "REMOVE_ITEM_FAILED", "Failed to remove cart item")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, cart)
}

// ClearCart removes all items from the cart
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	if err := h.cartService.ClearCart(ctx, userID, sessionID); err != nil {
		utils.Logger.Error(ctx, "Failed to clear cart", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "CLEAR_CART_FAILED", "Failed to clear cart")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Cart cleared successfully"})
}

// MigrateGuestCart migrates a guest cart to a user cart
func (h *CartHandler) MigrateGuestCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req MigrateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	cart, err := h.cartService.MigrateGuestCart(ctx, req.SessionID, req.UserID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to migrate guest cart", err, map[string]interface{}{
			"user_id":    req.UserID,
			"session_id": req.SessionID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "MIGRATION_FAILED", "Failed to migrate guest cart")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, cart)
}

// ValidateCart validates all items in the cart
func (h *CartHandler) ValidateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	cart, err := h.cartService.GetCart(ctx, userID, sessionID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get cart for validation", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "CART_RETRIEVAL_FAILED", "Failed to retrieve cart")
		return
	}

	validation, err := h.cartService.ValidateCart(ctx, cart)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to validate cart", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
			"cart_id":    cart.ID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "VALIDATION_FAILED", "Failed to validate cart")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, validation)
}

// ExtendExpiry extends the expiry time of the cart
func (h *CartHandler) ExtendExpiry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	var req ExtendExpiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	duration := time.Duration(req.Hours) * time.Hour
	cart, err := h.cartService.ExtendCartExpiry(ctx, userID, sessionID, duration)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to extend cart expiry", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
			"hours":      req.Hours,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "EXTEND_EXPIRY_FAILED", "Failed to extend cart expiry")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, cart)
}

// GetCartSummary returns a summary of the cart (item count, total, etc.)
func (h *CartHandler) GetCartSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-ID")
	sessionID := r.Header.Get("X-Session-ID")

	if userID == "" && sessionID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "USER_ID_OR_SESSION_REQUIRED", "Either user ID or session ID is required")
		return
	}

	cart, err := h.cartService.GetCart(ctx, userID, sessionID)
	if err != nil {
		utils.Logger.Error(ctx, "Failed to get cart for summary", err, map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "CART_RETRIEVAL_FAILED", "Failed to retrieve cart")
		return
	}

	summary := map[string]interface{}{
		"cart_id":    cart.ID,
		"item_count": cart.GetItemCount(),
		"subtotal":   cart.Subtotal,
		"currency":   cart.Currency,
		"expires_at": cart.ExpiresAt,
		"updated_at": cart.UpdatedAt,
	}

	utils.WriteJSONResponse(w, http.StatusOK, summary)
}

// CleanupExpiredCarts removes expired carts (admin endpoint)
func (h *CartHandler) CleanupExpiredCarts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// This should be protected by admin authentication middleware
	if err := h.cartService.CleanupExpiredCarts(ctx); err != nil {
		utils.Logger.Error(ctx, "Failed to cleanup expired carts", err, nil)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "CLEANUP_FAILED", "Failed to cleanup expired carts")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Expired carts cleaned up successfully"})
}
