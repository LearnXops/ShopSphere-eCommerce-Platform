package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/review-service/internal/service"
)

// ReviewHandler handles HTTP requests for review operations
type ReviewHandler struct {
	service service.ReviewService
	logger  *utils.StructuredLogger
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(service service.ReviewService, logger *utils.StructuredLogger) *ReviewHandler {
	return &ReviewHandler{
		service: service,
		logger:  logger,
	}
}

// CreateReview handles POST /reviews
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get user ID from context (would be set by auth middleware)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required")
		return
	}
	
	var request models.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error(ctx, "Failed to decode review request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	
	review, err := h.service.CreateReview(ctx, userID, &request)
	if err != nil {
		h.logger.Error(ctx, "Failed to create review", err, map[string]interface{}{
			"user_id": userID,
		})
		
		// Handle specific error types
		switch err.Error() {
		case "purchase not verified":
			utils.WriteErrorResponse(w, http.StatusForbidden, "PURCHASE_NOT_VERIFIED", "Purchase verification required")
		case "user has already reviewed this product":
			utils.WriteErrorResponse(w, http.StatusConflict, "DUPLICATE_REVIEW", "User has already reviewed this product")
		default:
			if err.Error()[:15] == "invalid request" {
				utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			} else {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create review")
			}
		}
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusCreated, review)
}

// GetReview handles GET /reviews/{id}
func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reviewID := vars["id"]
	
	review, err := h.service.GetReviewByID(ctx, reviewID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get review", err, map[string]interface{}{
			"review_id": reviewID,
		})
		
		if err.Error() == "review not found" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get review")
		}
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, review)
}

// GetProductReviews handles GET /products/{productId}/reviews
func (h *ReviewHandler) GetProductReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	productID := vars["productId"]
	
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 20
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	reviews, err := h.service.GetReviewsByProductID(ctx, productID, limit, offset)
	if err != nil {
		h.logger.Error(ctx, "Failed to get product reviews", err, map[string]interface{}{
			"product_id": productID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get reviews")
		return
	}
	
	response := map[string]interface{}{
		"reviews": reviews,
		"limit":   limit,
		"offset":  offset,
		"count":   len(reviews),
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetUserReviews handles GET /users/{userId}/reviews
func (h *ReviewHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["userId"]
	
	// Check if requesting user can access these reviews
	requestingUserID := r.Header.Get("X-User-ID")
	if requestingUserID != userID {
		// For now, allow anyone to view user reviews. In production, you might want to restrict this.
		// utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Cannot access other user's reviews")
		// return
	}
	
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 20
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	reviews, err := h.service.GetReviewsByUserID(ctx, userID, limit, offset)
	if err != nil {
		h.logger.Error(ctx, "Failed to get user reviews", err, map[string]interface{}{
			"user_id": userID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get reviews")
		return
	}
	
	response := map[string]interface{}{
		"reviews": reviews,
		"limit":   limit,
		"offset":  offset,
		"count":   len(reviews),
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdateReview handles PUT /reviews/{id}
func (h *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reviewID := vars["id"]
	
	// Get user ID from context
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required")
		return
	}
	
	var request models.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error(ctx, "Failed to decode review update request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	
	review, err := h.service.UpdateReview(ctx, userID, reviewID, &request)
	if err != nil {
		h.logger.Error(ctx, "Failed to update review", err, map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
		})
		
		switch {
		case err.Error() == "review not found":
			utils.WriteErrorResponse(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
		case err.Error()[:12] == "unauthorized":
			utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Not authorized to update this review")
		case err.Error()[:15] == "invalid request":
			utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update review")
		}
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, review)
}

// DeleteReview handles DELETE /reviews/{id}
func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reviewID := vars["id"]
	
	// Get user ID from context
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required")
		return
	}
	
	err := h.service.DeleteReview(ctx, userID, reviewID)
	if err != nil {
		h.logger.Error(ctx, "Failed to delete review", err, map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
		})
		
		switch {
		case err.Error() == "review not found":
			utils.WriteErrorResponse(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
		case err.Error()[:12] == "unauthorized":
			utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Not authorized to delete this review")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete review")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// VoteOnReview handles POST /reviews/{id}/vote
func (h *ReviewHandler) VoteOnReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reviewID := vars["id"]
	
	// Get user ID from context
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required")
		return
	}
	
	var request models.ReviewVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error(ctx, "Failed to decode vote request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	
	err := h.service.VoteOnReview(ctx, userID, reviewID, &request)
	if err != nil {
		h.logger.Error(ctx, "Failed to vote on review", err, map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
		})
		
		switch {
		case err.Error() == "review not found":
			utils.WriteErrorResponse(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
		case err.Error() == "users cannot vote on their own reviews":
			utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Cannot vote on your own review")
		case err.Error()[:15] == "invalid request":
			utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to vote on review")
		}
		return
	}
	
	response := map[string]string{"message": "Vote recorded successfully"}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// RemoveVote handles DELETE /reviews/{id}/vote
func (h *ReviewHandler) RemoveVote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reviewID := vars["id"]
	
	// Get user ID from context
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID required")
		return
	}
	
	err := h.service.RemoveVote(ctx, userID, reviewID)
	if err != nil {
		h.logger.Error(ctx, "Failed to remove vote", err, map[string]interface{}{
			"user_id":   userID,
			"review_id": reviewID,
		})
		
		switch {
		case err.Error() == "no vote found to remove":
			utils.WriteErrorResponse(w, http.StatusNotFound, "VOTE_NOT_FOUND", "No vote found to remove")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove vote")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetProductRatingSummary handles GET /products/{productId}/rating-summary
func (h *ReviewHandler) GetProductRatingSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	productID := vars["productId"]
	
	summary, err := h.service.GetProductRatingSummary(ctx, productID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get product rating summary", err, map[string]interface{}{
			"product_id": productID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get rating summary")
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, summary)
}

// ModerateReview handles POST /reviews/{id}/moderate
func (h *ReviewHandler) ModerateReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reviewID := vars["id"]
	
	// Get moderator ID from context
	moderatorID := r.Header.Get("X-User-ID")
	if moderatorID == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Moderator ID required")
		return
	}
	
	// Check if user has moderation permissions (this would typically be done by middleware)
	role := r.Header.Get("X-User-Role")
	if role != "admin" && role != "moderator" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		return
	}
	
	var request models.ModerationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error(ctx, "Failed to decode moderation request", err)
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	
	err := h.service.ModerateReview(ctx, moderatorID, reviewID, &request)
	if err != nil {
		h.logger.Error(ctx, "Failed to moderate review", err, map[string]interface{}{
			"moderator_id": moderatorID,
			"review_id":    reviewID,
		})
		
		switch {
		case err.Error() == "review not found":
			utils.WriteErrorResponse(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found")
		case err.Error()[:15] == "invalid request":
			utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to moderate review")
		}
		return
	}
	
	response := map[string]string{"message": "Review moderated successfully"}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetPendingReviews handles GET /admin/reviews/pending
func (h *ReviewHandler) GetPendingReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Check if user has moderation permissions
	role := r.Header.Get("X-User-Role")
	if role != "admin" && role != "moderator" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		return
	}
	
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 20
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	reviews, err := h.service.GetPendingReviews(ctx, limit, offset)
	if err != nil {
		h.logger.Error(ctx, "Failed to get pending reviews", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get pending reviews")
		return
	}
	
	response := map[string]interface{}{
		"reviews": reviews,
		"limit":   limit,
		"offset":  offset,
		"count":   len(reviews),
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetModerationLogs handles GET /reviews/{id}/moderation-logs
func (h *ReviewHandler) GetModerationLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	reviewID := vars["id"]
	
	// Check if user has moderation permissions
	role := r.Header.Get("X-User-Role")
	if role != "admin" && role != "moderator" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		return
	}
	
	logs, err := h.service.GetModerationLogs(ctx, reviewID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get moderation logs", err, map[string]interface{}{
			"review_id": reviewID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get moderation logs")
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, logs)
}

// HealthCheck handles GET /health
func (h *ReviewHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "review-service",
		"timestamp": "2024-01-01T00:00:00Z", // In production, use time.Now()
	}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// RegisterRoutes registers all review routes
func (h *ReviewHandler) RegisterRoutes(router *mux.Router) {
	// Review CRUD operations
	router.HandleFunc("/reviews", h.CreateReview).Methods("POST")
	router.HandleFunc("/reviews/{id}", h.GetReview).Methods("GET")
	router.HandleFunc("/reviews/{id}", h.UpdateReview).Methods("PUT")
	router.HandleFunc("/reviews/{id}", h.DeleteReview).Methods("DELETE")
	
	// Product reviews
	router.HandleFunc("/products/{productId}/reviews", h.GetProductReviews).Methods("GET")
	router.HandleFunc("/products/{productId}/rating-summary", h.GetProductRatingSummary).Methods("GET")
	
	// User reviews
	router.HandleFunc("/users/{userId}/reviews", h.GetUserReviews).Methods("GET")
	
	// Review voting
	router.HandleFunc("/reviews/{id}/vote", h.VoteOnReview).Methods("POST")
	router.HandleFunc("/reviews/{id}/vote", h.RemoveVote).Methods("DELETE")
	
	// Moderation endpoints
	router.HandleFunc("/reviews/{id}/moderate", h.ModerateReview).Methods("POST")
	router.HandleFunc("/admin/reviews/pending", h.GetPendingReviews).Methods("GET")
	router.HandleFunc("/reviews/{id}/moderation-logs", h.GetModerationLogs).Methods("GET")
	
	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
}
