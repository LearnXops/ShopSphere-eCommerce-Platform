package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/shopsphere/admin-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

type AdminHandler struct {
	service service.AdminService
	logger  *utils.StructuredLogger
}

func NewAdminHandler(service service.AdminService, logger *utils.StructuredLogger) *AdminHandler {
	return &AdminHandler{
		service: service,
		logger:  logger,
	}
}

// Helper methods
func (h *AdminHandler) getAdminUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	// In a real implementation, this would extract the admin user ID from JWT token or session
	// For now, we'll use a header for demonstration
	adminUserIDStr := r.Header.Get("X-Admin-User-ID")
	if adminUserIDStr == "" {
		return uuid.Nil, models.NewValidationError("admin_user_id", "admin user ID required")
	}
	
	adminUserID, err := uuid.Parse(adminUserIDStr)
	if err != nil {
		return uuid.Nil, models.NewValidationError("admin_user_id", "invalid admin user ID format")
	}
	
	return adminUserID, nil
}

func (h *AdminHandler) requirePermission(r *http.Request, permission models.Permission) error {
	adminUserID, err := h.getAdminUserIDFromContext(r)
	if err != nil {
		return err
	}
	
	return h.service.RequirePermission(r.Context(), adminUserID, permission)
}

func (h *AdminHandler) logActivity(r *http.Request, action, resourceType string, resourceID *uuid.UUID, details interface{}) {
	adminUserID, err := h.getAdminUserIDFromContext(r)
	if err != nil {
		return
	}
	
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	
	err = h.service.LogActivity(r.Context(), adminUserID, action, resourceType, resourceID, details, &ipAddress, &userAgent)
	if err != nil {
		h.logger.Error("Failed to log admin activity", "error", err)
	}
}

// Admin User Management
func (h *AdminHandler) CreateAdminUser(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	var req models.CreateAdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	adminUser, err := h.service.CreateAdminUser(r.Context(), &req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create admin user", err)
		return
	}
	
	h.logActivity(r, "create", "admin_user", &adminUser.ID, req)
	utils.WriteJSONResponse(w, http.StatusCreated, adminUser)
}

func (h *AdminHandler) GetAdminUser(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionUsersRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid admin user ID", err)
		return
	}
	
	adminUser, err := h.service.GetAdminUser(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Admin user not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get admin user", err)
		}
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, adminUser)
}

func (h *AdminHandler) UpdateAdminUser(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid admin user ID", err)
		return
	}
	
	var req models.UpdateAdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	adminUser, err := h.service.UpdateAdminUser(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Admin user not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update admin user", err)
		}
		return
	}
	
	h.logActivity(r, "update", "admin_user", &id, req)
	utils.WriteJSONResponse(w, http.StatusOK, adminUser)
}

func (h *AdminHandler) DeleteAdminUser(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid admin user ID", err)
		return
	}
	
	if err := h.service.DeleteAdminUser(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Admin user not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete admin user", err)
		}
		return
	}
	
	h.logActivity(r, "delete", "admin_user", &id, nil)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListAdminUsers(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionUsersRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	adminUsers, total, err := h.service.ListAdminUsers(r.Context(), page, limit)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list admin users", err)
		return
	}
	
	response := map[string]interface{}{
		"data":  adminUsers,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// Activity Logs
func (h *AdminHandler) GetActivityLogs(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionUsersRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	var adminUserID *uuid.UUID
	if adminUserIDStr := r.URL.Query().Get("admin_user_id"); adminUserIDStr != "" {
		id, err := uuid.Parse(adminUserIDStr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid admin user ID", err)
			return
		}
		adminUserID = &id
	}
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	logs, total, err := h.service.GetActivityLogs(r.Context(), adminUserID, page, limit)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get activity logs", err)
		return
	}
	
	response := map[string]interface{}{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// Dashboard and Analytics
func (h *AdminHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionAnalyticsRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	metrics, err := h.service.GetDashboardMetrics(r.Context())
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get dashboard metrics", err)
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, metrics)
}

func (h *AdminHandler) GetRevenueData(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionAnalyticsRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}
	
	revenueData, err := h.service.GetRevenueData(r.Context(), days)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get revenue data", err)
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, revenueData)
}

func (h *AdminHandler) GetOrderStatusDistribution(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionAnalyticsRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	distribution, err := h.service.GetOrderStatusDistribution(r.Context())
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get order status distribution", err)
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, distribution)
}

func (h *AdminHandler) GetTopProducts(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionAnalyticsRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	
	topProducts, err := h.service.GetTopProducts(r.Context(), limit)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get top products", err)
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, topProducts)
}

func (h *AdminHandler) GetUserGrowthData(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionAnalyticsRead); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}
	
	growthData, err := h.service.GetUserGrowthData(r.Context(), days)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user growth data", err)
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, growthData)
}
