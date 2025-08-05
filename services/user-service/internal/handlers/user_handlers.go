package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/user-service/internal/service"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	User              *models.User `json:"user"`
	VerificationToken string       `json:"verification_token,omitempty"`
	Message           string       `json:"message"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty"`
	Username  *string `json:"username,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetResponse represents a password reset response
type PasswordResetResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"` // Only for testing/development
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// UpdateStatusRequest represents a status update request
type UpdateStatusRequest struct {
	Status models.UserStatus `json:"status"`
}

// ListUsersResponse represents a list users response
type ListUsersResponse struct {
	Users  []*models.User `json:"users"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// Register handles user registration
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, utils.NewValidationError("invalid request body"), "")
		return
	}
	
	user, verificationToken, err := h.userService.RegisterUser(
		req.Email, req.Username, req.FirstName, req.LastName, req.Password,
	)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	response := RegisterResponse{
		User:    user,
		Message: "User registered successfully. Please check your email for verification.",
	}
	
	// Include verification token in development mode
	// In production, this should be sent via email
	if strings.ToLower(r.Header.Get("X-Environment")) == "development" {
		response.VerificationToken = verificationToken
	}
	
	h.writeJSONResponse(w, http.StatusCreated, response)
}

// GetUser handles getting a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	
	user, err := h.userService.GetUser(userID)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, user)
}

// UpdateUser handles updating user profile
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, utils.NewValidationError("invalid request body"), "")
		return
	}
	
	// Convert request to map for service layer
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	
	user, err := h.userService.UpdateUser(userID, updates)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, user)
}

// ChangePassword handles password changes
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, utils.NewValidationError("invalid request body"), "")
		return
	}
	
	err := h.userService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// RequestPasswordReset handles password reset requests
func (h *UserHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, utils.NewValidationError("invalid request body"), "")
		return
	}
	
	token, err := h.userService.RequestPasswordReset(req.Email)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	response := PasswordResetResponse{
		Message: "If the email exists, a password reset link has been sent.",
	}
	
	// Include token in development mode
	// In production, this should be sent via email
	if strings.ToLower(r.Header.Get("X-Environment")) == "development" {
		response.Token = token
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// ResetPassword handles password reset with token
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, utils.NewValidationError("invalid request body"), "")
		return
	}
	
	err := h.userService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

// VerifyEmail handles email verification
func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, utils.NewValidationError("invalid request body"), "")
		return
	}
	
	err := h.userService.VerifyEmail(req.Token)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

// DeleteUser handles user deletion (soft delete)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	
	err := h.userService.DeleteUser(userID)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

// ListUsers handles listing users with pagination and filtering (admin operation)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	status := models.UserStatus(r.URL.Query().Get("status"))
	role := models.UserRole(r.URL.Query().Get("role"))
	
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}
	
	users, total, err := h.userService.ListUsers(limit, offset, status, role)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	response := ListUsersResponse{
		Users:  users,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// SearchUsers handles searching users (admin operation)
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}
	
	users, total, err := h.userService.SearchUsers(query, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	response := ListUsersResponse{
		Users:  users,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateUserStatus handles updating user status (admin operation)
func (h *UserHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	
	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, utils.NewValidationError("invalid request body"), "")
		return
	}
	
	err := h.userService.UpdateUserStatus(userID, req.Status)
	if err != nil {
		h.writeErrorResponse(w, err, "")
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "User status updated successfully",
	})
}

// writeJSONResponse writes a JSON response
func (h *UserHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, err error, traceID string) {
	var appErr *utils.AppError
	var statusCode int
	
	if e, ok := err.(*utils.AppError); ok {
		appErr = e
		statusCode = e.HTTPStatusCode()
	} else {
		appErr = utils.NewInternalError("internal server error", err)
		statusCode = http.StatusInternalServerError
	}
	
	response := utils.ErrorResponse{
		Error: utils.ErrorDetail{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		},
		TraceID: traceID,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}