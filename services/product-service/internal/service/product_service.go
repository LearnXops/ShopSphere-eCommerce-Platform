package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopsphere/product-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// ProductService handles product business logic
type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

// NewProductService creates a new product service
func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(ctx context.Context, req CreateProductRequest) (*models.Product, error) {
	// Validate request
	if err := s.validateCreateProductRequest(req); err != nil {
		return nil, err
	}
	
	// Validate category exists if provided
	if req.CategoryID != "" {
		_, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
		if err != nil {
			return nil, utils.NewValidationError("invalid category_id")
		}
	}
	
	// Create product
	product := models.NewProduct(req.SKU, req.Name, req.Description, req.CategoryID, req.Price)
	product.Currency = req.Currency
	product.Stock = req.Stock
	product.Status = models.ProductStatus(req.Status)
	product.Images = req.Images
	
	// Set attributes
	if req.Attributes != nil {
		product.Attributes.Brand = req.Attributes.Brand
		product.Attributes.Color = req.Attributes.Color
		product.Attributes.Size = req.Attributes.Size
		product.Attributes.Weight = req.Attributes.Weight
		product.Attributes.Dimensions = models.Dimensions{
			Length: req.Attributes.Dimensions.Length,
			Width:  req.Attributes.Dimensions.Width,
			Height: req.Attributes.Dimensions.Height,
			Unit:   req.Attributes.Dimensions.Unit,
		}
		product.Attributes.Custom = req.Attributes.Custom
	}
	
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}
	
	return product, nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	if id == "" {
		return nil, utils.NewValidationError("product ID is required")
	}
	
	return s.productRepo.GetByID(ctx, id)
}

// GetProductBySKU retrieves a product by SKU
func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
	if sku == "" {
		return nil, utils.NewValidationError("product SKU is required")
	}
	
	return s.productRepo.GetBySKU(ctx, sku)
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(ctx context.Context, id string, req UpdateProductRequest) (*models.Product, error) {
	if id == "" {
		return nil, utils.NewValidationError("product ID is required")
	}
	
	// Get existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Validate request
	if err := s.validateUpdateProductRequest(req); err != nil {
		return nil, err
	}
	
	// Validate category exists if provided
	if req.CategoryID != nil && *req.CategoryID != "" {
		_, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, utils.NewValidationError("invalid category_id")
		}
		product.CategoryID = *req.CategoryID
	}
	
	// Update fields
	if req.SKU != nil {
		product.SKU = *req.SKU
	}
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Currency != nil {
		product.Currency = *req.Currency
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Status != nil {
		product.Status = models.ProductStatus(*req.Status)
	}
	if req.Images != nil {
		product.Images = req.Images
	}
	
	// Update attributes
	if req.Attributes != nil {
		if req.Attributes.Brand != nil {
			product.Attributes.Brand = *req.Attributes.Brand
		}
		if req.Attributes.Color != nil {
			product.Attributes.Color = *req.Attributes.Color
		}
		if req.Attributes.Size != nil {
			product.Attributes.Size = *req.Attributes.Size
		}
		if req.Attributes.Weight != nil {
			product.Attributes.Weight = *req.Attributes.Weight
		}
		if req.Attributes.Dimensions != nil {
			if req.Attributes.Dimensions.Length != nil {
				product.Attributes.Dimensions.Length = *req.Attributes.Dimensions.Length
			}
			if req.Attributes.Dimensions.Width != nil {
				product.Attributes.Dimensions.Width = *req.Attributes.Dimensions.Width
			}
			if req.Attributes.Dimensions.Height != nil {
				product.Attributes.Dimensions.Height = *req.Attributes.Dimensions.Height
			}
			if req.Attributes.Dimensions.Unit != nil {
				product.Attributes.Dimensions.Unit = *req.Attributes.Dimensions.Unit
			}
		}
		if req.Attributes.Custom != nil {
			product.Attributes.Custom = req.Attributes.Custom
		}
	}
	
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}
	
	return product, nil
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if id == "" {
		return utils.NewValidationError("product ID is required")
	}
	
	return s.productRepo.Delete(ctx, id)
}

// ListProducts retrieves products with filtering and pagination
func (s *ProductService) ListProducts(ctx context.Context, req ListProductsRequest) (*ListProductsResponse, error) {
	// Validate request
	if err := s.validateListProductsRequest(req); err != nil {
		return nil, err
	}
	
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	
	// Build filter
	filter := repository.ProductFilter{
		CategoryID:  req.CategoryID,
		Status:      req.Status,
		MinPrice:    req.MinPrice,
		MaxPrice:    req.MaxPrice,
		SearchTerm:  req.SearchTerm,
		Featured:    req.Featured,
		InStock:     req.InStock,
		Limit:       req.Limit,
		Offset:      req.Offset,
		SortBy:      req.SortBy,
		SortOrder:   req.SortOrder,
	}
	
	products, total, err := s.productRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return &ListProductsResponse{
		Products: products,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

// ReserveStock reserves stock for a product (used by cart service)
func (s *ProductService) ReserveStock(ctx context.Context, productID string, quantity int) error {
	if productID == "" {
		return utils.NewValidationError("product ID is required")
	}
	if quantity <= 0 {
		return utils.NewValidationError("quantity must be positive")
	}
	
	// Check available stock
	availableStock, err := s.productRepo.GetAvailableStock(ctx, productID)
	if err != nil {
		return err
	}
	
	if availableStock < quantity {
		return utils.NewConflictError("insufficient stock available")
	}
	
	return s.productRepo.ReserveStock(ctx, productID, quantity)
}

// ReleaseStock releases reserved stock
func (s *ProductService) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	if productID == "" {
		return utils.NewValidationError("product ID is required")
	}
	if quantity <= 0 {
		return utils.NewValidationError("quantity must be positive")
	}
	
	return s.productRepo.ReleaseStock(ctx, productID, quantity)
}

// UpdateStock updates product stock
func (s *ProductService) UpdateStock(ctx context.Context, productID string, quantity int) error {
	if productID == "" {
		return utils.NewValidationError("product ID is required")
	}
	
	return s.productRepo.UpdateStock(ctx, productID, quantity)
}

// BulkUpdateStock updates stock for multiple products
func (s *ProductService) BulkUpdateStock(ctx context.Context, updates []BulkStockUpdate) error {
	if len(updates) == 0 {
		return utils.NewValidationError("no updates provided")
	}
	
	// Validate updates
	for i, update := range updates {
		if update.ProductID == "" {
			return utils.NewValidationError(fmt.Sprintf("product_id is required for update %d", i))
		}
		if update.Type == "" {
			return utils.NewValidationError(fmt.Sprintf("type is required for update %d", i))
		}
		if update.Type != "in" && update.Type != "out" && update.Type != "adjustment" {
			return utils.NewValidationError(fmt.Sprintf("invalid type for update %d", i))
		}
	}
	
	// Convert to repository format
	repoUpdates := make([]repository.StockUpdate, len(updates))
	for i, update := range updates {
		repoUpdates[i] = repository.StockUpdate{
			ProductID: update.ProductID,
			Quantity:  update.Quantity,
			Type:      update.Type,
			Reason:    update.Reason,
		}
	}
	
	return s.productRepo.BulkUpdateStock(ctx, repoUpdates)
}

// SearchProducts performs advanced product search
func (s *ProductService) SearchProducts(ctx context.Context, req SearchProductsRequest) (*ListProductsResponse, error) {
	// Validate request
	if req.Query == "" {
		return nil, utils.NewValidationError("search query is required")
	}
	
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	
	// Build filter for search
	filter := repository.ProductFilter{
		SearchTerm: req.Query,
		CategoryID: req.CategoryID,
		Status:     req.Status,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Featured:   req.Featured,
		InStock:    req.InStock,
		Limit:      req.Limit,
		Offset:     req.Offset,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}
	
	products, total, err := s.productRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return &ListProductsResponse{
		Products: products,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

// validateCreateProductRequest validates create product request
func (s *ProductService) validateCreateProductRequest(req CreateProductRequest) error {
	v := utils.NewValidator()
	
	v.Required("sku", req.SKU).SKU("sku", req.SKU)
	v.Required("name", req.Name).MaxLength("name", req.Name, 255)
	v.Required("description", req.Description).MaxLength("description", req.Description, 2000)
	v.Required("price", req.Price).DecimalPositive("price", req.Price)
	
	if req.Currency != "" {
		v.MaxLength("currency", req.Currency, 3)
	}
	
	if req.Status != "" {
		validStatuses := []interface{}{"active", "inactive", "out_of_stock", "discontinued"}
		v.OneOf("status", req.Status, validStatuses)
	}
	
	// Additional validation
	if req.Stock < 0 {
		return utils.NewValidationError("stock must be non-negative")
	}
	
	if v.HasErrors() {
		return utils.NewValidationError(v.Errors().Error())
	}
	
	return nil
}

// validateUpdateProductRequest validates update product request
func (s *ProductService) validateUpdateProductRequest(req UpdateProductRequest) error {
	v := utils.NewValidator()
	
	if req.SKU != nil {
		v.SKU("sku", *req.SKU)
	}
	
	if req.Name != nil {
		v.Required("name", *req.Name).MaxLength("name", *req.Name, 255)
	}
	
	if req.Description != nil {
		v.MaxLength("description", *req.Description, 2000)
	}
	
	if req.Price != nil {
		v.DecimalPositive("price", *req.Price)
	}
	
	if req.Currency != nil {
		v.MaxLength("currency", *req.Currency, 3)
	}
	
	// Additional validation
	if req.Stock != nil && *req.Stock < 0 {
		return utils.NewValidationError("stock must be non-negative")
	}
	
	if req.Status != nil {
		validStatuses := []interface{}{"active", "inactive", "out_of_stock", "discontinued"}
		v.OneOf("status", *req.Status, validStatuses)
	}
	
	if v.HasErrors() {
		return utils.NewValidationError(v.Errors().Error())
	}
	
	return nil
}

// validateListProductsRequest validates list products request
func (s *ProductService) validateListProductsRequest(req ListProductsRequest) error {
	v := utils.NewValidator()
	
	if req.SortBy != "" {
		validSortFields := []interface{}{"name", "price", "created_at", "updated_at", "stock"}
		v.OneOf("sort_by", req.SortBy, validSortFields)
	}
	
	if req.SortOrder != "" {
		validSortOrders := []interface{}{"asc", "desc"}
		v.OneOf("sort_order", strings.ToLower(req.SortOrder), validSortOrders)
	}
	
	// Additional validation
	if req.Limit < 0 {
		return utils.NewValidationError("limit must be non-negative")
	}
	
	if req.Offset < 0 {
		return utils.NewValidationError("offset must be non-negative")
	}
	
	if v.HasErrors() {
		return utils.NewValidationError(v.Errors().Error())
	}
	
	return nil
}