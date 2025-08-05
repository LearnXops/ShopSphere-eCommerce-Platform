package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/shopsphere/product-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// Mock repository for testing
type mockProductRepository struct {
	products map[string]*models.Product
	nextID   int
}

func newMockProductRepository() *mockProductRepository {
	return &mockProductRepository{
		products: make(map[string]*models.Product),
		nextID:   1,
	}
}

func (m *mockProductRepository) Create(ctx context.Context, product *models.Product) error {
	if product.ID == "" {
		product.ID = "test-id-1"
	}
	
	// Check for duplicate SKU
	for _, p := range m.products {
		if p.SKU == product.SKU {
			return utils.NewConflictError("product with this SKU already exists")
		}
	}
	
	m.products[product.ID] = product
	return nil
}

func (m *mockProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	product, exists := m.products[id]
	if !exists {
		return nil, utils.NewNotFoundError("product")
	}
	return product, nil
}

func (m *mockProductRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	for _, product := range m.products {
		if product.SKU == sku {
			return product, nil
		}
	}
	return nil, utils.NewNotFoundError("product")
}

func (m *mockProductRepository) Update(ctx context.Context, product *models.Product) error {
	if _, exists := m.products[product.ID]; !exists {
		return utils.NewNotFoundError("product")
	}
	
	// Check for duplicate SKU (excluding current product)
	for id, p := range m.products {
		if id != product.ID && p.SKU == product.SKU {
			return utils.NewConflictError("product with this SKU already exists")
		}
	}
	
	m.products[product.ID] = product
	return nil
}

func (m *mockProductRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.products[id]; !exists {
		return utils.NewNotFoundError("product")
	}
	delete(m.products, id)
	return nil
}

func (m *mockProductRepository) List(ctx context.Context, filter repository.ProductFilter) ([]*models.Product, int, error) {
	var products []*models.Product
	for _, product := range m.products {
		products = append(products, product)
	}
	return products, len(products), nil
}

func (m *mockProductRepository) UpdateStock(ctx context.Context, productID string, quantity int) error {
	product, exists := m.products[productID]
	if !exists {
		return utils.NewNotFoundError("product")
	}
	product.Stock = quantity
	return nil
}

func (m *mockProductRepository) ReserveStock(ctx context.Context, productID string, quantity int) error {
	product, exists := m.products[productID]
	if !exists {
		return utils.NewNotFoundError("product")
	}
	if product.Stock < quantity {
		return utils.NewConflictError("insufficient stock")
	}
	return nil
}

func (m *mockProductRepository) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	_, exists := m.products[productID]
	if !exists {
		return utils.NewNotFoundError("product")
	}
	return nil
}

func (m *mockProductRepository) GetAvailableStock(ctx context.Context, productID string) (int, error) {
	product, exists := m.products[productID]
	if !exists {
		return 0, utils.NewNotFoundError("product")
	}
	return product.Stock, nil
}

func (m *mockProductRepository) BulkUpdateStock(ctx context.Context, updates []repository.StockUpdate) error {
	return nil
}

// Mock category repository
type mockCategoryRepository struct {
	categories map[string]*models.Category
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{
		categories: make(map[string]*models.Category),
	}
}

func (m *mockCategoryRepository) Create(ctx context.Context, category *models.Category) error {
	if category.ID == "" {
		category.ID = "test-cat-1"
	}
	m.categories[category.ID] = category
	return nil
}

func (m *mockCategoryRepository) GetByID(ctx context.Context, id string) (*models.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, utils.NewNotFoundError("category")
	}
	return category, nil
}

func (m *mockCategoryRepository) Update(ctx context.Context, category *models.Category) error {
	if _, exists := m.categories[category.ID]; !exists {
		return utils.NewNotFoundError("category")
	}
	m.categories[category.ID] = category
	return nil
}

func (m *mockCategoryRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.categories[id]; !exists {
		return utils.NewNotFoundError("category")
	}
	delete(m.categories, id)
	return nil
}

func (m *mockCategoryRepository) List(ctx context.Context, filter repository.CategoryFilter) ([]*models.Category, error) {
	var categories []*models.Category
	for _, category := range m.categories {
		categories = append(categories, category)
	}
	return categories, nil
}

func (m *mockCategoryRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Category, error) {
	var children []*models.Category
	for _, category := range m.categories {
		if category.ParentID != nil && *category.ParentID == parentID {
			children = append(children, category)
		}
	}
	return children, nil
}

func (m *mockCategoryRepository) GetPath(ctx context.Context, categoryID string) ([]*models.Category, error) {
	category, exists := m.categories[categoryID]
	if !exists {
		return nil, utils.NewNotFoundError("category")
	}
	return []*models.Category{category}, nil
}

// Test functions

func TestProductService_CreateProduct(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Test successful product creation
	req := CreateProductRequest{
		SKU:         "TEST-001",
		Name:        "Test Product",
		Description: "A test product",
		Price:       decimal.NewFromFloat(99.99),
		Currency:    "USD",
		Stock:       10,
		Status:      "active",
	}
	
	product, err := service.CreateProduct(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if product.SKU != req.SKU {
		t.Errorf("Expected SKU %s, got %s", req.SKU, product.SKU)
	}
	
	if product.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, product.Name)
	}
	
	if !product.Price.Equal(req.Price) {
		t.Errorf("Expected price %s, got %s", req.Price.String(), product.Price.String())
	}
}

func TestProductService_CreateProduct_ValidationError(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Test validation error - empty SKU
	req := CreateProductRequest{
		SKU:         "",
		Name:        "Test Product",
		Description: "A test product",
		Price:       decimal.NewFromFloat(99.99),
	}
	
	_, err := service.CreateProduct(ctx, req)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
	
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	
	if appErr.Code != utils.ErrValidation {
		t.Errorf("Expected validation error, got %s", appErr.Code)
	}
}

func TestProductService_GetProduct(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Create a test product
	testProduct := models.NewProduct("TEST-001", "Test Product", "Description", "", decimal.NewFromFloat(99.99))
	productRepo.Create(ctx, testProduct)
	
	// Test successful retrieval
	product, err := service.GetProduct(ctx, testProduct.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if product.ID != testProduct.ID {
		t.Errorf("Expected ID %s, got %s", testProduct.ID, product.ID)
	}
}

func TestProductService_GetProduct_NotFound(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Test product not found
	_, err := service.GetProduct(ctx, "non-existent-id")
	if err == nil {
		t.Fatal("Expected not found error, got nil")
	}
	
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	
	if appErr.Code != utils.ErrNotFound {
		t.Errorf("Expected not found error, got %s", appErr.Code)
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Create a test product
	testProduct := models.NewProduct("TEST-001", "Test Product", "Description", "", decimal.NewFromFloat(99.99))
	productRepo.Create(ctx, testProduct)
	
	// Test successful update
	newName := "Updated Product"
	newPrice := decimal.NewFromFloat(149.99)
	req := UpdateProductRequest{
		Name:  &newName,
		Price: &newPrice,
	}
	
	product, err := service.UpdateProduct(ctx, testProduct.ID, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if product.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, product.Name)
	}
	
	if !product.Price.Equal(newPrice) {
		t.Errorf("Expected price %s, got %s", newPrice.String(), product.Price.String())
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Create a test product
	testProduct := models.NewProduct("TEST-001", "Test Product", "Description", "", decimal.NewFromFloat(99.99))
	productRepo.Create(ctx, testProduct)
	
	// Test successful deletion
	err := service.DeleteProduct(ctx, testProduct.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify product is deleted
	_, err = service.GetProduct(ctx, testProduct.ID)
	if err == nil {
		t.Fatal("Expected not found error after deletion, got nil")
	}
}

func TestProductService_ReserveStock(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Create a test product with stock
	testProduct := models.NewProduct("TEST-001", "Test Product", "Description", "", decimal.NewFromFloat(99.99))
	testProduct.Stock = 10
	productRepo.Create(ctx, testProduct)
	
	// Test successful stock reservation
	err := service.ReserveStock(ctx, testProduct.ID, 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestProductService_ReserveStock_InsufficientStock(t *testing.T) {
	productRepo := newMockProductRepository()
	categoryRepo := newMockCategoryRepository()
	service := NewProductService(productRepo, categoryRepo)
	
	ctx := context.Background()
	
	// Create a test product with limited stock
	testProduct := models.NewProduct("TEST-001", "Test Product", "Description", "", decimal.NewFromFloat(99.99))
	testProduct.Stock = 5
	productRepo.Create(ctx, testProduct)
	
	// Test insufficient stock error
	err := service.ReserveStock(ctx, testProduct.ID, 10)
	if err == nil {
		t.Fatal("Expected insufficient stock error, got nil")
	}
	
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	
	if appErr.Code != utils.ErrConflict {
		t.Errorf("Expected conflict error, got %s", appErr.Code)
	}
}