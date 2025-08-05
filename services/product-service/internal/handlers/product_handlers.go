package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/product-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productService  *service.ProductService
	categoryService *service.CategoryService
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *service.ProductService, categoryService *service.CategoryService) *ProductHandler {
	return &ProductHandler{
		productService:  productService,
		categoryService: categoryService,
	}
}

// CreateProduct handles POST /products
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req service.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	product, err := h.productService.CreateProduct(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusCreated, product)
}

// GetProduct handles GET /products/{id}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	product, err := h.productService.GetProduct(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, product)
}

// GetProductBySKU handles GET /products/sku/{sku}
func (h *ProductHandler) GetProductBySKU(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sku := vars["sku"]
	
	product, err := h.productService.GetProductBySKU(r.Context(), sku)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, product)
}

// UpdateProduct handles PUT /products/{id}
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var req service.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	product, err := h.productService.UpdateProduct(r.Context(), id, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, product)
}

// DeleteProduct handles DELETE /products/{id}
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	err := h.productService.DeleteProduct(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// ListProducts handles GET /products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	req := h.parseListProductsRequest(r)
	
	response, err := h.productService.ListProducts(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// SearchProducts handles GET /products/search
func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	req := h.parseSearchProductsRequest(r)
	
	response, err := h.productService.SearchProducts(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, response)
}

// ReserveStock handles POST /products/{id}/reserve-stock
func (h *ProductHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]
	
	var req service.StockReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	req.ProductID = productID
	
	err := h.productService.ReserveStock(r.Context(), req.ProductID, req.Quantity)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ReleaseStock handles POST /products/{id}/release-stock
func (h *ProductHandler) ReleaseStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]
	
	var req service.StockReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	req.ProductID = productID
	
	err := h.productService.ReleaseStock(r.Context(), req.ProductID, req.Quantity)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// UpdateStock handles PUT /products/{id}/stock
func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]
	
	var req service.StockUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	err := h.productService.UpdateStock(r.Context(), productID, req.Quantity)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// BulkUpdateStock handles POST /products/bulk-stock-update
func (h *ProductHandler) BulkUpdateStock(w http.ResponseWriter, r *http.Request) {
	var req []service.BulkStockUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "")
		return
	}
	
	err := h.productService.BulkUpdateStock(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// parseListProductsRequest parses query parameters for listing products
func (h *ProductHandler) parseListProductsRequest(r *http.Request) service.ListProductsRequest {
	query := r.URL.Query()
	
	req := service.ListProductsRequest{
		CategoryID: query.Get("category_id"),
		Status:     query.Get("status"),
		SearchTerm: query.Get("search_term"),
		SortBy:     query.Get("sort_by"),
		SortOrder:  query.Get("sort_order"),
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
	
	if minPrice := query.Get("min_price"); minPrice != "" {
		if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
			req.MinPrice = &val
		}
	}
	
	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			req.MaxPrice = &val
		}
	}
	
	// Parse boolean parameters
	if featured := query.Get("featured"); featured != "" {
		if val, err := strconv.ParseBool(featured); err == nil {
			req.Featured = &val
		}
	}
	
	if inStock := query.Get("in_stock"); inStock != "" {
		if val, err := strconv.ParseBool(inStock); err == nil {
			req.InStock = &val
		}
	}
	
	return req
}

// parseSearchProductsRequest parses query parameters for searching products
func (h *ProductHandler) parseSearchProductsRequest(r *http.Request) service.SearchProductsRequest {
	query := r.URL.Query()
	
	req := service.SearchProductsRequest{
		Query:      query.Get("q"),
		CategoryID: query.Get("category_id"),
		Status:     query.Get("status"),
		SortBy:     query.Get("sort_by"),
		SortOrder:  query.Get("sort_order"),
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
	
	if minPrice := query.Get("min_price"); minPrice != "" {
		if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
			req.MinPrice = &val
		}
	}
	
	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			req.MaxPrice = &val
		}
	}
	
	// Parse boolean parameters
	if featured := query.Get("featured"); featured != "" {
		if val, err := strconv.ParseBool(featured); err == nil {
			req.Featured = &val
		}
	}
	
	if inStock := query.Get("in_stock"); inStock != "" {
		if val, err := strconv.ParseBool(inStock); err == nil {
			req.InStock = &val
		}
	}
	
	return req
}

// handleServiceError handles service layer errors and converts them to HTTP responses
func (h *ProductHandler) handleServiceError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*utils.AppError); ok {
		h.writeErrorResponse(w, appErr.HTTPStatusCode(), string(appErr.Code), appErr.Message, appErr.Details)
	} else {
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "")
	}
}

// writeJSONResponse writes a JSON response
func (h *ProductHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error without context since we don't have access to request here
		fmt.Printf("Failed to encode JSON response: %v\n", err)
	}
}

// writeErrorResponse writes an error response
func (h *ProductHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResponse := service.ErrorResponse{
		Error: service.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		TraceID: h.getTraceID(w),
	}
	
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		fmt.Printf("Failed to encode error response: %v\n", err)
	}
}

// getTraceID extracts trace ID from response headers or generates one
func (h *ProductHandler) getTraceID(w http.ResponseWriter) string {
	// In a real implementation, this would extract from request context or headers
	// For now, return a placeholder
	return "trace-12345"
}