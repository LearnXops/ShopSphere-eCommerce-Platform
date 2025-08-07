package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/notification-service/internal/service"
)

// NotificationHandler handles HTTP requests for notification operations
type NotificationHandler struct {
	service *service.NotificationService
	logger  *utils.StructuredLogger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service *service.NotificationService, logger *utils.StructuredLogger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers notification routes
func (h *NotificationHandler) RegisterRoutes(router *mux.Router) {
	// Notification operations
	router.HandleFunc("/notifications", h.SendNotification).Methods("POST")
	router.HandleFunc("/notifications/{id}", h.GetNotification).Methods("GET")
	router.HandleFunc("/notifications", h.ListNotifications).Methods("GET")
	router.HandleFunc("/notifications/{id}/events", h.GetDeliveryEvents).Methods("GET")

	// Template operations
	router.HandleFunc("/templates", h.CreateTemplate).Methods("POST")
	router.HandleFunc("/templates/{id}", h.GetTemplate).Methods("GET")
	router.HandleFunc("/templates", h.ListTemplates).Methods("GET")
	router.HandleFunc("/templates/{id}", h.UpdateTemplate).Methods("PUT")
	router.HandleFunc("/templates/{id}", h.DeleteTemplate).Methods("DELETE")
	router.HandleFunc("/templates/render", h.RenderTemplate).Methods("POST")

	// User preference operations
	router.HandleFunc("/users/{userId}/preferences", h.GetUserPreferences).Methods("GET")
	router.HandleFunc("/users/{userId}/preferences", h.UpdateUserPreferences).Methods("PUT")

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
}

// SendNotification handles POST /notifications
func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	var request models.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	response, err := h.service.SendNotification(r.Context(), &request)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to send notification", err, map[string]interface{}{
			"user_id": request.UserID,
			"channel": request.Channel,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "SEND_FAILED", "Failed to send notification")
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetNotification handles GET /notifications/{id}
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	notification, err := h.service.GetNotification(r.Context(), id)
	if err != nil {
		if err.Error() == "notification not found" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
			return
		}
		h.logger.Error(r.Context(), "Failed to get notification", err, map[string]interface{}{
			"notification_id": id,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get notification")
		return
	}

	// Check if user has permission to view this notification
	userID := r.Header.Get("X-User-ID")
	if userID != notification.UserID {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, notification)
}

// ListNotifications handles GET /notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required")
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

	notifications, err := h.service.ListNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to list notifications", err, map[string]interface{}{
			"user_id": userID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list notifications")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetDeliveryEvents handles GET /notifications/{id}/events
func (h *NotificationHandler) GetDeliveryEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notificationID := vars["id"]

	// First check if notification exists and user has permission
	notification, err := h.service.GetNotification(r.Context(), notificationID)
	if err != nil {
		if err.Error() == "notification not found" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get notification")
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID != notification.UserID {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	events, err := h.service.GetDeliveryEvents(r.Context(), notificationID)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to get delivery events", err, map[string]interface{}{
			"notification_id": notificationID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get delivery events")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"events": events,
	})
}

// CreateTemplate handles POST /templates
func (h *NotificationHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	// Check admin permission
	userRole := r.Header.Get("X-User-Role")
	if userRole != "admin" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Admin access required")
		return
	}

	var template models.NotificationTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.CreateTemplate(r.Context(), &template); err != nil {
		h.logger.Error(r.Context(), "Failed to create template", err, map[string]interface{}{
			"template_name": template.Name,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create template")
		return
	}

	utils.WriteJSONResponse(w, http.StatusCreated, template)
}

// GetTemplate handles GET /templates/{id}
func (h *NotificationHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	template, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		if err.Error() == "notification template not found" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Template not found")
			return
		}
		h.logger.Error(r.Context(), "Failed to get template", err, map[string]interface{}{
			"template_id": id,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get template")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, template)
}

// ListTemplates handles GET /templates
func (h *NotificationHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	channelStr := r.URL.Query().Get("channel")
	var channel *models.NotificationChannel

	if channelStr != "" {
		ch := models.NotificationChannel(channelStr)
		if ch == models.ChannelEmail || ch == models.ChannelSMS || ch == models.ChannelPush {
			channel = &ch
		}
	}

	templates, err := h.service.ListTemplates(r.Context(), channel)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to list templates", err, map[string]interface{}{
			"channel": channel,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list templates")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

// UpdateTemplate handles PUT /templates/{id}
func (h *NotificationHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	// Check admin permission
	userRole := r.Header.Get("X-User-Role")
	if userRole != "admin" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Admin access required")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var template models.NotificationTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	template.ID = id

	if err := h.service.UpdateTemplate(r.Context(), &template); err != nil {
		if err.Error() == "notification template not found" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Template not found")
			return
		}
		h.logger.Error(r.Context(), "Failed to update template", err, map[string]interface{}{
			"template_id": id,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update template")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, template)
}

// DeleteTemplate handles DELETE /templates/{id}
func (h *NotificationHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	// Check admin permission
	userRole := r.Header.Get("X-User-Role")
	if userRole != "admin" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Admin access required")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteTemplate(r.Context(), id); err != nil {
		if err.Error() == "notification template not found" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Template not found")
			return
		}
		h.logger.Error(r.Context(), "Failed to delete template", err, map[string]interface{}{
			"template_id": id,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete template")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RenderTemplate handles POST /templates/render
func (h *NotificationHandler) RenderTemplate(w http.ResponseWriter, r *http.Request) {
	var request models.TemplateRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	response, err := h.service.RenderTemplate(r.Context(), &request)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to render template", err, map[string]interface{}{
			"template_name": request.TemplateName,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "RENDER_FAILED", "Failed to render template")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetUserPreferences handles GET /users/{userId}/preferences
func (h *NotificationHandler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	// Check permission - users can only access their own preferences
	requestingUserID := r.Header.Get("X-User-ID")
	userRole := r.Header.Get("X-User-Role")
	if requestingUserID != userID && userRole != "admin" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	preferences, err := h.service.GetUserPreferences(r.Context(), userID)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to get user preferences", err, map[string]interface{}{
			"user_id": userID,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get preferences")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"preferences": preferences,
	})
}

// UpdateUserPreferences handles PUT /users/{userId}/preferences
func (h *NotificationHandler) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	// Check permission - users can only update their own preferences
	requestingUserID := r.Header.Get("X-User-ID")
	userRole := r.Header.Get("X-User-Role")
	if requestingUserID != userID && userRole != "admin" {
		utils.WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	var updateRequest struct {
		Channel   models.NotificationChannel `json:"channel"`
		Category  string                     `json:"category"`
		IsEnabled bool                       `json:"is_enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.UpdateUserPreference(r.Context(), userID, updateRequest.Channel, updateRequest.Category, updateRequest.IsEnabled); err != nil {
		h.logger.Error(r.Context(), "Failed to update user preference", err, map[string]interface{}{
			"user_id":  userID,
			"channel":  updateRequest.Channel,
			"category": updateRequest.Category,
		})
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update preference")
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Preference updated successfully",
	})
}

// HealthCheck handles GET /health
func (h *NotificationHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "notification-service",
	})
}
