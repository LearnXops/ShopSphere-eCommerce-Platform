package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/shopsphere/product-service/internal/handlers"
	"github.com/shopsphere/product-service/internal/repository"
	"github.com/shopsphere/product-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Setup test database
	var err error
	testDB, err = setupTestDatabase()
	if err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestDatabase(testDB)
	os.Exit(code)
}

func setupTestDatabase() (*sql.DB, error) {
	// Use test database configuration
	dbConfig := utils.NewDatabaseConfig()
	dbConfig.DBName = "product_service_test"
	
	// Connect to postgres to create test database
	adminConfig := utils.NewDatabaseConfig()
	adminConfig.DBName = "postgres"
	
	adminDB, err := adminConfig.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer adminDB.Close()
	
	// Create test database
	_, err = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbConfig.DBName))
	if err != nil {
		return nil, fmt.Errorf("failed to drop test database: %w", err)
	}
	
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbConfig.DBName))
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}
	
	// Connect to test database
	testDB, err := dbConfig.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}
	
	// Create tables
	if err := createTestTables(testDB); err != nil {
		testDB.Close()
		return nil, fmt.Errorf("failed to create test tables: %w", err)
	}
	
	return testDB, nil
}

func createTestTables(db *sql.DB) error {
	// Create the update_updated_at_column function first
	createFunctionSQL := `
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';`
	
	if _, err := db.Exec(createFunctionSQL); err != nil {
		return fmt.Errorf("failed to create update function: %w", err)
	}
	
	// Read and execute the schema migration
	schemaSQL := `
	-- Categories table with hierarchical structure
	CREATE TABLE IF NOT EXISTS categories (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		parent_id VARCHAR(36) REFERENCES categories(id) ON DELETE SET NULL,
		path VARCHAR(500) NOT NULL,
		level INTEGER DEFAULT 0,
		sort_order INTEGER DEFAULT 0,
		is_active BOOLEAN DEFAULT TRUE,
		meta_title VARCHAR(255),
		meta_description TEXT,
		meta_keywords TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Products table
	CREATE TABLE IF NOT EXISTS products (
		id VARCHAR(36) PRIMARY KEY,
		sku VARCHAR(100) UNIQUE NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		short_description TEXT,
		category_id VARCHAR(36) REFERENCES categories(id) ON DELETE SET NULL,
		price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
		compare_price DECIMAL(10,2) CHECK (compare_price >= 0),
		cost_price DECIMAL(10,2) CHECK (cost_price >= 0),
		currency VARCHAR(3) DEFAULT 'USD',
		stock INTEGER DEFAULT 0 CHECK (stock >= 0),
		reserved_stock INTEGER DEFAULT 0 CHECK (reserved_stock >= 0),
		low_stock_threshold INTEGER DEFAULT 10,
		track_inventory BOOLEAN DEFAULT TRUE,
		status VARCHAR(20) DEFAULT 'inactive' CHECK (status IN ('active', 'inactive', 'out_of_stock', 'discontinued')),
		visibility VARCHAR(20) DEFAULT 'visible' CHECK (visibility IN ('visible', 'hidden', 'catalog_only')),
		weight DECIMAL(8,3),
		length DECIMAL(8,2),
		width DECIMAL(8,2),
		height DECIMAL(8,2),
		images TEXT[],
		attributes JSONB DEFAULT '{}',
		meta_title VARCHAR(255),
		meta_description TEXT,
		meta_keywords TEXT,
		featured BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Inventory movements table
	CREATE TABLE IF NOT EXISTS inventory_movements (
		id VARCHAR(36) PRIMARY KEY,
		product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
		variant_id VARCHAR(36),
		movement_type VARCHAR(20) NOT NULL CHECK (movement_type IN ('in', 'out', 'adjustment', 'reserved', 'released')),
		quantity INTEGER NOT NULL,
		reference_type VARCHAR(50),
		reference_id VARCHAR(36),
		reason TEXT,
		created_by VARCHAR(36),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);
	CREATE INDEX IF NOT EXISTS idx_categories_path ON categories(path);
	CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
	CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
	CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
	CREATE INDEX IF NOT EXISTS idx_inventory_movements_product_id ON inventory_movements(product_id);

	-- Create triggers
	CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`
	
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	
	return nil
}

func cleanupTestDatabase(db *sql.DB) {
	// Clean up test data
	db.Exec("DELETE FROM inventory_movements")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM categories")
}

func setupTestRouter() *mux.Router {
	// Initialize repositories
	productRepo := repository.NewProductRepository(testDB)
	categoryRepo := repository.NewCategoryRepository(testDB)

	// Initialize services
	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService, categoryService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// Create router
	router := mux.NewRouter()

	// Product routes
	productRoutes := router.PathPrefix("/products").Subrouter()
	productRoutes.HandleFunc("", productHandler.ListProducts).Methods("GET")
	productRoutes.HandleFunc("", productHandler.CreateProduct).Methods("POST")
	productRoutes.HandleFunc("/search", productHandler.SearchProducts).Methods("GET")
	productRoutes.HandleFunc("/{id}", productHandler.GetProduct).Methods("GET")
	productRoutes.HandleFunc("/{id}", productHandler.UpdateProduct).Methods("PUT")
	productRoutes.HandleFunc("/{id}", productHandler.DeleteProduct).Methods("DELETE")
	productRoutes.HandleFunc("/{id}/reserve-stock", productHandler.ReserveStock).Methods("POST")

	// Category routes
	categoryRoutes := router.PathPrefix("/categories").Subrouter()
	categoryRoutes.HandleFunc("", categoryHandler.ListCategories).Methods("GET")
	categoryRoutes.HandleFunc("", categoryHandler.CreateCategory).Methods("POST")
	categoryRoutes.HandleFunc("/{id}", categoryHandler.GetCategory).Methods("GET")
	categoryRoutes.HandleFunc("/{id}", categoryHandler.UpdateCategory).Methods("PUT")
	categoryRoutes.HandleFunc("/{id}", categoryHandler.DeleteCategory).Methods("DELETE")

	return router
}

func TestProductIntegration_CreateAndGetProduct(t *testing.T) {
	router := setupTestRouter()
	
	// Clean up before test
	cleanupTestDatabase(testDB)
	
	// Create a product
	createReq := service.CreateProductRequest{
		SKU:         "TEST-001",
		Name:        "Test Product",
		Description: "A test product for integration testing",
		Price:       decimal.NewFromFloat(99.99),
		Currency:    "USD",
		Stock:       10,
		Status:      "active",
	}
	
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var createdProduct models.Product
	if err := json.Unmarshal(w.Body.Bytes(), &createdProduct); err != nil {
		t.Fatalf("Failed to unmarshal created product: %v", err)
	}
	
	// Get the product
	req = httptest.NewRequest("GET", "/products/"+createdProduct.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var retrievedProduct models.Product
	if err := json.Unmarshal(w.Body.Bytes(), &retrievedProduct); err != nil {
		t.Fatalf("Failed to unmarshal retrieved product: %v", err)
	}
	
	// Verify product data
	if retrievedProduct.SKU != createReq.SKU {
		t.Errorf("Expected SKU %s, got %s", createReq.SKU, retrievedProduct.SKU)
	}
	
	if retrievedProduct.Name != createReq.Name {
		t.Errorf("Expected name %s, got %s", createReq.Name, retrievedProduct.Name)
	}
	
	if !retrievedProduct.Price.Equal(createReq.Price) {
		t.Errorf("Expected price %s, got %s", createReq.Price.String(), retrievedProduct.Price.String())
	}
}

func TestProductIntegration_UpdateProduct(t *testing.T) {
	router := setupTestRouter()
	
	// Clean up before test
	cleanupTestDatabase(testDB)
	
	// Create a product first
	createReq := service.CreateProductRequest{
		SKU:         "TEST-002",
		Name:        "Test Product 2",
		Description: "Another test product",
		Price:       decimal.NewFromFloat(149.99),
		Currency:    "USD",
		Stock:       5,
		Status:      "active",
	}
	
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var createdProduct models.Product
	json.Unmarshal(w.Body.Bytes(), &createdProduct)
	
	// Update the product
	newName := "Updated Test Product 2"
	newPrice := decimal.NewFromFloat(199.99)
	updateReq := service.UpdateProductRequest{
		Name:  &newName,
		Price: &newPrice,
	}
	
	reqBody, _ = json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", "/products/"+createdProduct.ID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var updatedProduct models.Product
	if err := json.Unmarshal(w.Body.Bytes(), &updatedProduct); err != nil {
		t.Fatalf("Failed to unmarshal updated product: %v", err)
	}
	
	// Verify updates
	if updatedProduct.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, updatedProduct.Name)
	}
	
	if !updatedProduct.Price.Equal(newPrice) {
		t.Errorf("Expected price %s, got %s", newPrice.String(), updatedProduct.Price.String())
	}
}

func TestProductIntegration_ListProducts(t *testing.T) {
	router := setupTestRouter()
	
	// Clean up before test
	cleanupTestDatabase(testDB)
	
	// Create multiple products
	products := []service.CreateProductRequest{
		{
			SKU:         "TEST-LIST-001",
			Name:        "List Test Product 1",
			Description: "First product for list testing",
			Price:       decimal.NewFromFloat(99.99),
			Status:      "active",
		},
		{
			SKU:         "TEST-LIST-002",
			Name:        "List Test Product 2",
			Description: "Second product for list testing",
			Price:       decimal.NewFromFloat(149.99),
			Status:      "active",
		},
	}
	
	for _, product := range products {
		reqBody, _ := json.Marshal(product)
		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create product %s: status %d", product.SKU, w.Code)
		}
	}
	
	// List products
	req := httptest.NewRequest("GET", "/products?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var response service.ListProductsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal list response: %v", err)
	}
	
	if len(response.Products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(response.Products))
	}
	
	if response.Total != 2 {
		t.Errorf("Expected total 2, got %d", response.Total)
	}
}

func TestCategoryIntegration_CreateAndGetCategory(t *testing.T) {
	router := setupTestRouter()
	
	// Clean up before test
	cleanupTestDatabase(testDB)
	
	// Create a category
	createReq := service.CreateCategoryRequest{
		Name:        "Electronics",
		Description: "Electronic products and gadgets",
		IsActive:    true,
	}
	
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var createdCategory models.Category
	if err := json.Unmarshal(w.Body.Bytes(), &createdCategory); err != nil {
		t.Fatalf("Failed to unmarshal created category: %v", err)
	}
	
	// Get the category
	req = httptest.NewRequest("GET", "/categories/"+createdCategory.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var retrievedCategory models.Category
	if err := json.Unmarshal(w.Body.Bytes(), &retrievedCategory); err != nil {
		t.Fatalf("Failed to unmarshal retrieved category: %v", err)
	}
	
	// Verify category data
	if retrievedCategory.Name != createReq.Name {
		t.Errorf("Expected name %s, got %s", createReq.Name, retrievedCategory.Name)
	}
	
	if retrievedCategory.Description != createReq.Description {
		t.Errorf("Expected description %s, got %s", createReq.Description, retrievedCategory.Description)
	}
	
	if retrievedCategory.IsActive != createReq.IsActive {
		t.Errorf("Expected is_active %t, got %t", createReq.IsActive, retrievedCategory.IsActive)
	}
}

func TestProductIntegration_StockReservation(t *testing.T) {
	router := setupTestRouter()
	
	// Clean up before test
	cleanupTestDatabase(testDB)
	
	// Create a product with stock
	createReq := service.CreateProductRequest{
		SKU:         "TEST-STOCK-001",
		Name:        "Stock Test Product",
		Description: "Product for stock testing",
		Price:       decimal.NewFromFloat(99.99),
		Stock:       10,
		Status:      "active",
	}
	
	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var createdProduct models.Product
	json.Unmarshal(w.Body.Bytes(), &createdProduct)
	
	// Reserve stock
	reserveReq := service.StockReservationRequest{
		Quantity: 5,
	}
	
	reqBody, _ = json.Marshal(reserveReq)
	req = httptest.NewRequest("POST", "/products/"+createdProduct.ID+"/reserve-stock", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "success" {
		t.Errorf("Expected status success, got %s", response["status"])
	}
}