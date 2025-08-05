package service

import (
	"github.com/shopspring/decimal"
	"github.com/shopsphere/shared/models"
)

// Product Service DTOs

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	SKU         string                     `json:"sku" validate:"required"`
	Name        string                     `json:"name" validate:"required"`
	Description string                     `json:"description" validate:"required"`
	CategoryID  string                     `json:"category_id"`
	Price       decimal.Decimal            `json:"price" validate:"required"`
	Currency    string                     `json:"currency"`
	Stock       int                        `json:"stock"`
	Status      string                     `json:"status"`
	Images      []string                   `json:"images"`
	Attributes  *ProductAttributesRequest  `json:"attributes"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	SKU         *string                    `json:"sku"`
	Name        *string                    `json:"name"`
	Description *string                    `json:"description"`
	CategoryID  *string                    `json:"category_id"`
	Price       *decimal.Decimal           `json:"price"`
	Currency    *string                    `json:"currency"`
	Stock       *int                       `json:"stock"`
	Status      *string                    `json:"status"`
	Images      []string                   `json:"images"`
	Attributes  *UpdateProductAttributesRequest `json:"attributes"`
}

// ProductAttributesRequest represents product attributes in requests
type ProductAttributesRequest struct {
	Brand      string                 `json:"brand"`
	Color      string                 `json:"color"`
	Size       string                 `json:"size"`
	Weight     float64                `json:"weight"`
	Dimensions DimensionsRequest      `json:"dimensions"`
	Custom     map[string]interface{} `json:"custom"`
}

// UpdateProductAttributesRequest represents product attributes in update requests
type UpdateProductAttributesRequest struct {
	Brand      *string                `json:"brand"`
	Color      *string                `json:"color"`
	Size       *string                `json:"size"`
	Weight     *float64               `json:"weight"`
	Dimensions *UpdateDimensionsRequest `json:"dimensions"`
	Custom     map[string]interface{} `json:"custom"`
}

// DimensionsRequest represents product dimensions in requests
type DimensionsRequest struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

// UpdateDimensionsRequest represents product dimensions in update requests
type UpdateDimensionsRequest struct {
	Length *float64 `json:"length"`
	Width  *float64 `json:"width"`
	Height *float64 `json:"height"`
	Unit   *string  `json:"unit"`
}

// ListProductsRequest represents a request to list products
type ListProductsRequest struct {
	CategoryID  string   `json:"category_id"`
	Status      string   `json:"status"`
	MinPrice    *float64 `json:"min_price"`
	MaxPrice    *float64 `json:"max_price"`
	SearchTerm  string   `json:"search_term"`
	Featured    *bool    `json:"featured"`
	InStock     *bool    `json:"in_stock"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
	SortBy      string   `json:"sort_by"`
	SortOrder   string   `json:"sort_order"`
}

// ListProductsResponse represents a response to list products
type ListProductsResponse struct {
	Products []*models.Product `json:"products"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// SearchProductsRequest represents a request to search products
type SearchProductsRequest struct {
	Query      string   `json:"query" validate:"required"`
	CategoryID string   `json:"category_id"`
	Status     string   `json:"status"`
	MinPrice   *float64 `json:"min_price"`
	MaxPrice   *float64 `json:"max_price"`
	Featured   *bool    `json:"featured"`
	InStock    *bool    `json:"in_stock"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	SortBy     string   `json:"sort_by"`
	SortOrder  string   `json:"sort_order"`
}

// BulkStockUpdate represents a bulk stock update operation
type BulkStockUpdate struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
	Type      string `json:"type" validate:"required"` // "in", "out", "adjustment"
	Reason    string `json:"reason"`
}

// StockReservationRequest represents a request to reserve stock
type StockReservationRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
}

// StockUpdateRequest represents a request to update stock
type StockUpdateRequest struct {
	Quantity int    `json:"quantity" validate:"required"`
	Reason   string `json:"reason"`
}

// Category Service DTOs

// CreateCategoryRequest represents a request to create a category
type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"`
	IsActive    bool    `json:"is_active"`
}

// UpdateCategoryRequest represents a request to update a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
	IsActive    *bool   `json:"is_active"`
}

// ListCategoriesRequest represents a request to list categories
type ListCategoriesRequest struct {
	ParentID *string `json:"parent_id"`
	IsActive *bool   `json:"is_active"`
	Level    *int    `json:"level"`
	Limit    int     `json:"limit"`
	Offset   int     `json:"offset"`
}

// ListCategoriesResponse represents a response to list categories
type ListCategoriesResponse struct {
	Categories []*models.Category `json:"categories"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// CategoryTreeNode represents a node in the category tree
type CategoryTreeNode struct {
	Category *models.Category    `json:"category"`
	Children []*CategoryTreeNode `json:"children"`
}

// Image Management DTOs

// UploadImageRequest represents a request to upload an image
type UploadImageRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	ImageData []byte `json:"image_data" validate:"required"`
	FileName  string `json:"file_name" validate:"required"`
	AltText   string `json:"alt_text"`
	IsPrimary bool   `json:"is_primary"`
}

// UploadImageResponse represents a response to upload an image
type UploadImageResponse struct {
	ImageURL string `json:"image_url"`
	ImageID  string `json:"image_id"`
}

// DeleteImageRequest represents a request to delete an image
type DeleteImageRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	ImageID   string `json:"image_id" validate:"required"`
}

// Inventory DTOs

// InventoryMovement represents an inventory movement record
type InventoryMovement struct {
	ID            string `json:"id"`
	ProductID     string `json:"product_id"`
	VariantID     *string `json:"variant_id"`
	MovementType  string `json:"movement_type"`
	Quantity      int    `json:"quantity"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   string `json:"reference_id"`
	Reason        string `json:"reason"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
}

// GetInventoryMovementsRequest represents a request to get inventory movements
type GetInventoryMovementsRequest struct {
	ProductID     string `json:"product_id"`
	MovementType  string `json:"movement_type"`
	ReferenceType string `json:"reference_type"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
}

// GetInventoryMovementsResponse represents a response to get inventory movements
type GetInventoryMovementsResponse struct {
	Movements []*InventoryMovement `json:"movements"`
	Total     int                  `json:"total"`
	Limit     int                  `json:"limit"`
	Offset    int                  `json:"offset"`
}

// ProductStockInfo represents product stock information
type ProductStockInfo struct {
	ProductID       string `json:"product_id"`
	TotalStock      int    `json:"total_stock"`
	ReservedStock   int    `json:"reserved_stock"`
	AvailableStock  int    `json:"available_stock"`
	LowStockThreshold int  `json:"low_stock_threshold"`
	IsLowStock      bool   `json:"is_low_stock"`
}

// Error Response DTOs

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   ErrorDetail `json:"error"`
	TraceID string      `json:"trace_id"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Validation Error Response

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error  string             `json:"error"`
	Fields []ValidationError  `json:"fields"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}