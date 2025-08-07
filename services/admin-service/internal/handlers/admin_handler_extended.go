package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// System Alerts
func (h *AdminHandler) CreateSystemAlert(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	var req models.CreateSystemAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	alert, err := h.service.CreateSystemAlert(r.Context(), &req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create system alert", err)
		return
	}
	
	h.logActivity(r, "create", "system_alert", &alert.ID, req)
	utils.WriteJSONResponse(w, http.StatusCreated, alert)
}

func (h *AdminHandler) GetSystemAlert(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid alert ID", err)
		return
	}
	
	alert, err := h.service.GetSystemAlert(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "System alert not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get system alert", err)
		}
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, alert)
}

func (h *AdminHandler) ResolveSystemAlert(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid alert ID", err)
		return
	}
	
	var req models.ResolveAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	alert, err := h.service.ResolveSystemAlert(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "System alert not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to resolve system alert", err)
		}
		return
	}
	
	h.logActivity(r, "resolve", "system_alert", &id, req)
	utils.WriteJSONResponse(w, http.StatusOK, alert)
}

func (h *AdminHandler) DeleteSystemAlert(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid alert ID", err)
		return
	}
	
	if err := h.service.DeleteSystemAlert(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "System alert not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete system alert", err)
		}
		return
	}
	
	h.logActivity(r, "delete", "system_alert", &id, nil)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListSystemAlerts(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	var severity *string
	if s := r.URL.Query().Get("severity"); s != "" {
		severity = &s
	}
	
	var isResolved *bool
	if r := r.URL.Query().Get("resolved"); r != "" {
		resolved := r == "true"
		isResolved = &resolved
	}
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	alerts, total, err := h.service.ListSystemAlerts(r.Context(), severity, isResolved, page, limit)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list system alerts", err)
		return
	}
	
	response := map[string]interface{}{
		"data":  alerts,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// Dashboard Configs
func (h *AdminHandler) CreateDashboardConfig(w http.ResponseWriter, r *http.Request) {
	adminUserID, err := h.getAdminUserIDFromContext(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}
	
	var req struct {
		DashboardName string          `json:"dashboard_name"`
		Config        json.RawMessage `json:"config"`
		IsDefault     bool            `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	config, err := h.service.CreateDashboardConfig(r.Context(), adminUserID, req.DashboardName, req.Config, req.IsDefault)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.WriteErrorResponse(w, http.StatusConflict, "Dashboard config already exists", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create dashboard config", err)
		}
		return
	}
	
	h.logActivity(r, "create", "dashboard_config", &config.ID, req)
	utils.WriteJSONResponse(w, http.StatusCreated, config)
}

func (h *AdminHandler) GetDashboardConfig(w http.ResponseWriter, r *http.Request) {
	adminUserID, err := h.getAdminUserIDFromContext(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}
	
	vars := mux.Vars(r)
	dashboardName := vars["name"]
	
	config, err := h.service.GetDashboardConfig(r.Context(), adminUserID, dashboardName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Dashboard config not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get dashboard config", err)
		}
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, config)
}

func (h *AdminHandler) UpdateDashboardConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid config ID", err)
		return
	}
	
	var req models.UpdateDashboardConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	_, err = h.service.UpdateDashboardConfig(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Dashboard config not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update dashboard config", err)
		}
		return
	}
	
	h.logActivity(r, "update", "dashboard_config", &id, req)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) DeleteDashboardConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid config ID", err)
		return
	}
	
	if err := h.service.DeleteDashboardConfig(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Dashboard config not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete dashboard config", err)
		}
		return
	}
	
	h.logActivity(r, "delete", "dashboard_config", &id, nil)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListDashboardConfigs(w http.ResponseWriter, r *http.Request) {
	adminUserID, err := h.getAdminUserIDFromContext(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}
	
	configs, err := h.service.ListDashboardConfigs(r.Context(), adminUserID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list dashboard configs", err)
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, configs)
}

// Bulk Operations
func (h *AdminHandler) CreateBulkOperation(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	adminUserID, err := h.getAdminUserIDFromContext(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}
	
	var req models.CreateBulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	
	operation, err := h.service.CreateBulkOperation(r.Context(), adminUserID, &req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create bulk operation", err)
		return
	}
	
	h.logActivity(r, "create", "bulk_operation", &operation.ID, req)
	utils.WriteJSONResponse(w, http.StatusCreated, operation)
}

func (h *AdminHandler) GetBulkOperation(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid operation ID", err)
		return
	}
	
	operation, err := h.service.GetBulkOperation(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorResponse(w, http.StatusNotFound, "Bulk operation not found", err)
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get bulk operation", err)
		}
		return
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, operation)
}

func (h *AdminHandler) ListBulkOperations(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
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
	
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	operations, total, err := h.service.ListBulkOperations(r.Context(), adminUserID, status, page, limit)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list bulk operations", err)
		return
	}
	
	response := map[string]interface{}{
		"data":  operations,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// System Metrics
func (h *AdminHandler) UpdateSystemMetrics(w http.ResponseWriter, r *http.Request) {
	if err := h.requirePermission(r, models.PermissionSystemAdmin); err != nil {
		utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions", err)
		return
	}
	
	if err := h.service.UpdateSystemMetrics(r.Context()); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update system metrics", err)
		return
	}
	
	h.logActivity(r, "update", "system_metrics", nil, nil)
	w.WriteHeader(http.StatusNoContent)
}

// Health Check
func (h *AdminHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "admin-service",
		"timestamp": time.Now().UTC(),
	}
	
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
