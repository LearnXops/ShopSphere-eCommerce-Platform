package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopsphere/product-service/internal/handlers"
	"github.com/shopsphere/product-service/internal/repository"
	"github.com/shopsphere/product-service/internal/search"
	"github.com/shopsphere/product-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestElasticsearchIntegration tests the complete Elasticsearch integration
func TestElasticsearchIntegration(t *testing.T) {
	// Skip if running in CI without Elasticsearch
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Initialize test database
	dbConfig := utils.NewDatabaseConfig()
	dbConfig.DBName = "product_service_test"
	
	db, err := dbConfig.Connect()
	require.NoError(t, err)
	defer db.Close()

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// Initialize mock Elasticsearch client for testing
	searchService := search.NewMockElasticsearchClient()
	analyticsService := search.NewAnalyticsService(db)

	// Initialize services
	productService := service.NewProductService(productRepo, categoryRepo, searchService, analyticsService)
	categoryService := service.NewCategoryService(categoryRepo)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService, categoryService)

	// Create test router
	router := mux.NewRouter()
	productRoutes := router.PathPrefix("/products").Subrouter()
	productRoutes.HandleFunc("/search/advanced", productHandler.AdvancedSearch).Methods("POST")
	productRoutes.HandleFunc("/search/suggestions", productHandler.GetSearchSuggestions).Methods("GET")
	productRoutes.HandleFunc("/search/reindex", productHandler.BulkIndexProducts).Methods("POST")

	// Test data
	testProducts := []*models.Product{
		{
			ID:          "test-product-1",
			SKU:         "TEST-001",
			Name:        "Red Cotton T-Shirt",
			Description: "Comfortable red cotton t-shirt for everyday wear",
			CategoryID:  "clothing",
			Price:       decimal.NewFromFloat(29.99),
			Currency:    "USD",
			Stock:       50,
			Status:      models.ProductActive,
			Images:      []string{"red-tshirt-1.jpg", "red-tshirt-2.jpg"},
			Attributes: models.ProductAttributes{
				Brand:  "TestBrand",
				Color:  "Red",
				Size:   "M",
				Weight: 0.2,
				Custom: map[string]interface{}{
					"material": "100% Cotton",
					"care":     "Machine wash cold",
				},
			},
			Featured:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "test-product-2",
			SKU:         "TEST-002",
			Name:        "Blue Denim Jeans",
			Description: "Classic blue denim jeans with comfortable fit",
			CategoryID:  "clothing",
			Price:       decimal.NewFromFloat(79.99),
			Currency:    "USD",
			Stock:       30,
			Status:      models.ProductActive,
			Images:      []string{"blue-jeans-1.jpg"},
			Attributes: models.ProductAttributes{
				Brand:  "TestBrand",
				Color:  "Blue",
				Size:   "32x34",
				Weight: 0.8,
				Custom: map[string]interface{}{
					"material": "98% Cotton, 2% Elastane",
					"fit":      "Regular",
				},
			},
			Featured:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "test-product-3",
			SKU:         "TEST-003",
			Name:        "Green Baseball Cap",
			Description: "Adjustable green baseball cap with logo",
			CategoryID:  "accessories",
			Price:       decimal.NewFromFloat(24.99),
			Currency:    "USD",
			Stock:       100,
			Status:      models.ProductActive,
			Images:      []string{"green-cap-1.jpg"},
			Attributes: models.ProductAttributes{
				Brand:  "SportsBrand",
				Color:  "Green",
				Size:   "One Size",
				Weight: 0.1,
				Custom: map[string]interface{}{
					"material":   "Cotton Twill",
					"adjustable": true,
				},
			},
			Featured:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Index test products
	err = searchService.BulkIndexProducts(ctx, testProducts)
	require.NoError(t, err)

	t.Run("Advanced Search", func(t *testing.T) {
		// Test advanced search with query
		searchReq := service.AdvancedSearchRequest{
			Query: "cotton",
			Filters: map[string]interface{}{
				"category_id": "clothing",
			},
			Facets: []string{"brand", "color"},
			From:   0,
			Size:   10,
		}

		reqBody, _ := json.Marshal(searchReq)
		req := httptest.NewRequest("POST", "/products/search/advanced", strings.NewReader(string(reqBody)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.AdvancedSearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Greater(t, response.Total, int64(0))
		assert.NotEmpty(t, response.Products)
		
		// Verify that returned products match the search criteria
		for _, product := range response.Products {
			assert.Equal(t, "clothing", product.CategoryID)
			assert.Contains(t, strings.ToLower(product.Name+" "+product.Description), "cotton")
		}
	})

	t.Run("Search Suggestions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/products/search/suggestions?q=red&size=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.SearchSuggestionsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Suggestions)
		
		// Verify suggestions contain the query term
		for _, suggestion := range response.Suggestions {
			assert.Contains(t, strings.ToLower(suggestion), "red")
		}
	})

	t.Run("Bulk Reindex", func(t *testing.T) {
		reindexReq := map[string]interface{}{
			"product_ids": []string{"test-product-1", "test-product-2"},
		}

		reqBody, _ := json.Marshal(reindexReq)
		req := httptest.NewRequest("POST", "/products/search/reindex", strings.NewReader(string(reqBody)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Products indexed successfully", response["message"])
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("Search with Filters", func(t *testing.T) {
		// Test search with multiple filters
		searchReq := service.AdvancedSearchRequest{
			Query: "",
			Filters: map[string]interface{}{
				"brand":  "TestBrand",
				"status": "active",
			},
			Sort: []service.SortField{
				{Field: "price", Order: "asc"},
			},
			From: 0,
			Size: 10,
		}

		reqBody, _ := json.Marshal(searchReq)
		req := httptest.NewRequest("POST", "/products/search/advanced", strings.NewReader(string(reqBody)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.AdvancedSearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify all returned products match the filters
		for _, product := range response.Products {
			assert.Equal(t, "TestBrand", product.Attributes.Brand)
			assert.Equal(t, models.ProductActive, product.Status)
		}
	})

	t.Run("Search with Pagination", func(t *testing.T) {
		// Test pagination
		searchReq := service.AdvancedSearchRequest{
			Query: "",
			From:  0,
			Size:  2,
		}

		reqBody, _ := json.Marshal(searchReq)
		req := httptest.NewRequest("POST", "/products/search/advanced", strings.NewReader(string(reqBody)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.AdvancedSearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.LessOrEqual(t, len(response.Products), 2)
		assert.Equal(t, 0, response.From)
		assert.Equal(t, 2, response.Size)
	})

	t.Run("Empty Search Results", func(t *testing.T) {
		// Test search with no results
		searchReq := service.AdvancedSearchRequest{
			Query: "nonexistent-product-xyz",
			From:  0,
			Size:  10,
		}

		reqBody, _ := json.Marshal(searchReq)
		req := httptest.NewRequest("POST", "/products/search/advanced", strings.NewReader(string(reqBody)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.AdvancedSearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Empty(t, response.Products)
		assert.Equal(t, int64(0), response.Total)
	})

	t.Run("Invalid Search Request", func(t *testing.T) {
		// Test invalid JSON
		req := httptest.NewRequest("POST", "/products/search/advanced", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Search Suggestions Validation", func(t *testing.T) {
		// Test missing query parameter
		req := httptest.NewRequest("GET", "/products/search/suggestions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// BenchmarkElasticsearchIntegration benchmarks the search functionality
func BenchmarkElasticsearchIntegration(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	ctx := context.Background()
	searchService := search.NewMockElasticsearchClient()

	// Create test products
	products := make([]*models.Product, 1000)
	for i := 0; i < 1000; i++ {
		products[i] = &models.Product{
			ID:          fmt.Sprintf("bench-product-%d", i),
			SKU:         fmt.Sprintf("BENCH-%04d", i),
			Name:        fmt.Sprintf("Benchmark Product %d", i),
			Description: fmt.Sprintf("Product for benchmarking number %d", i),
			CategoryID:  fmt.Sprintf("category-%d", i%10),
			Price:       decimal.NewFromFloat(float64(10 + (i % 100))),
			Currency:    "USD",
			Stock:       i % 50,
			Status:      models.ProductActive,
			Attributes: models.ProductAttributes{
				Brand:  fmt.Sprintf("Brand%d", i%5),
				Color:  []string{"Red", "Blue", "Green", "Yellow", "Black"}[i%5],
				Size:   []string{"XS", "S", "M", "L", "XL"}[i%5],
				Custom: make(map[string]interface{}),
			},
			Featured:  i%10 == 0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Index products
	err := searchService.BulkIndexProducts(ctx, products)
	if err != nil {
		b.Fatalf("Failed to index products: %v", err)
	}

	b.ResetTimer()

	b.Run("SimpleSearch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := search.SearchRequest{
				Query: "Product",
				From:  0,
				Size:  20,
			}
			_, err := searchService.SearchProducts(ctx, req)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	})

	b.Run("FilteredSearch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := search.SearchRequest{
				Query: "Product",
				Filters: map[string]interface{}{
					"category_id": "category-1",
					"brand":       "Brand1",
				},
				From: 0,
				Size: 20,
			}
			_, err := searchService.SearchProducts(ctx, req)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	})

	b.Run("Suggestions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := searchService.GetSearchSuggestions(ctx, "Product", 10)
			if err != nil {
				b.Fatalf("Suggestions failed: %v", err)
			}
		}
	})
}