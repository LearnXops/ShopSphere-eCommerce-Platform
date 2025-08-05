package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/shopsphere/auth-service/internal/jwt"
	"github.com/shopsphere/auth-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

// AuthHandlers handles authentication HTTP requests
type AuthHandlers struct {
	authService *service.AuthService
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService *service.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// Login handles user login
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, utils.NewValidationError("Invalid request body"))
		return
	}

	response, err := h.authService.Login(ctx, &req)
	if err != nil {
		h.writeErrorResponse(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req service.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, utils.NewValidationError("Invalid request body"))
		return
	}

	tokens, err := h.authService.RefreshToken(ctx, &req)
	if err != nil {
		h.writeErrorResponse(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, tokens)
}

// Logout handles user logout
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req service.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, utils.NewValidationError("Invalid request body"))
		return
	}

	if err := h.authService.Logout(ctx, &req); err != nil {
		h.writeErrorResponse(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// ValidateToken handles token validation
func (h *AuthHandlers) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	tokenString, err := jwt.ExtractTokenFromHeader(authHeader)
	if err != nil {
		h.writeErrorResponse(w, r, err)
		return
	}

	user, err := h.authService.ValidateToken(ctx, tokenString)
	if err != nil {
		h.writeErrorResponse(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"valid": true,
		"user":  user,
	})
}

// Me returns current user information
func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	tokenString, err := jwt.ExtractTokenFromHeader(authHeader)
	if err != nil {
		h.writeErrorResponse(w, r, err)
		return
	}

	user, err := h.authService.ValidateToken(ctx, tokenString)
	if err != nil {
		h.writeErrorResponse(w, r, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, user)
}

// writeJSONResponse writes a JSON response
func (h *AuthHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		utils.Logger.Error(context.Background(), "Failed to encode JSON response", err)
	}
}

// writeErrorResponse writes an error response
func (h *AuthHandlers) writeErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *utils.AppError
	var ok bool
	
	if appErr, ok = err.(*utils.AppError); !ok {
		appErr = utils.NewInternalError("Internal server error", err)
	}

	response := utils.ErrorResponse{
		Error: utils.ErrorDetail{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		},
		TraceID: utils.GetTraceID(r.Context()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatusCode())
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.Logger.Error(r.Context(), "Failed to encode error response", err)
	}
}