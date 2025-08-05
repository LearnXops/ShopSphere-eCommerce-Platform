package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/product-service/internal/service"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// CreateCategory handles POST /categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req service.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	category, err := h.categoryService.CreateCategory(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusCreated, category)
}

// GetCategory handles GET /categories/{id}
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	category, err := h.categoryService.GetCategory(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, category)
}

// UpdateCategory handles PUT /categories/{id}
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var req service.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	category, err := h.categoryService.UpdateCategory(r.Context(), id, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, category)
}

// DeleteCategory handles DELETE /categories/{id}
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	err := h.categoryService.DeleteCategory(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// ListCategories handles GET /categories
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	req := h.parseListCategoriesRequest(r)
	
	response, err := h.categoryService.ListCategories(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetCategoryChildren handles GET /categories/{id}/children
func (h *CategoryHandler) GetCategoryChildren(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentID := vars["id"]
	
	children, err := h.categoryService.GetCategoryChildren(r.Context(), parentID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"children": children,
	})
}

// GetCategoryPath handles GET /categories/{id}/path
func (h *CategoryHandler) GetCategoryPath(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]
	
	path, err := h.categoryService.GetCategoryPath(r.Context(), categoryID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"path": path,
	})
}

// GetRootCategories handles GET /categories/root
func (h *CategoryHandler) GetRootCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.GetRootCategories(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

// GetCategoryTree handles GET /categories/tree
func (h *CategoryHandler) GetCategoryTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.categoryService.GetCategoryTree(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"tree": tree,
	})
}

// parseListCategoriesRequest parses query parameters for listing categories
func (h *CategoryHandler) parseListCategoriesRequest(r *http.Request) service.ListCategoriesRequest {
	query := r.URL.Query()
	
	req := service.ListCategoriesRequest{}
	
	// Parse parent_id
	if parentID := query.Get("parent_id"); parentID != "" {
		req.ParentID = &parentID
	}
	
	// Parse numeric parameters
	if limit := query.Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			req.Limit = val
		}
	}
	
	if offset := query.Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			req.Offset = val
		}
	}
	
	if level := query.Get("level"); level != "" {
		if val, err := strconv.Atoi(level); err == nil {
			req.Level = &val
		}
	}
	
	// Parse boolean parameters
	if isActive := query.Get("is_active"); isActive != "" {
		if val, err := strconv.ParseBool(isActive); err == nil {
			req.IsActive = &val
		}
	}
	
	return req
}

// Helper methods (reuse from ProductHandler)

// handleServiceError handles service layer errors and converts them to HTTP responses
func (h *CategoryHandler) handleServiceError(w http.ResponseWriter, err error) {
	// Create a temporary ProductHandler to reuse the error handling logic
	ph := &ProductHandler{}
	ph.handleServiceError(w, err)
}

// writeJSONResponse writes a JSON response
func (h *CategoryHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	// Create a temporary ProductHandler to reuse the JSON response logic
	ph := &ProductHandler{}
	ph.writeJSONResponse(w, statusCode, data)
}

// writeErrorResponse writes an error response
func (h *CategoryHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message, details string) {
	// Create a temporary ProductHandler to reuse the error response logic
	ph := &ProductHandler{}
	ph.writeErrorResponse(w, statusCode, code, message, details)
}