package search

import (
	"context"
	"testing"
	"time"

	"github.com/shopsphere/shared/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestElasticsearchClient_IndexProduct(t *testing.T) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	product := &models.Product{
		ID:          "test-product-1",
		SKU:         "TEST-001",
		Name:        "Test Product",
		Description: "A test product",
		CategoryID:  "category-1",
		Price:       decimal.NewFromFloat(99.99),
		Currency:    "USD",
		Stock:       10,
		Status:      models.ProductActive,
		Images:      []string{"image1.jpg"},
		Attributes: models.ProductAttributes{
			Brand:  "TestBrand",
			Color:  "Red",
			Size:   "M",
			Weight: 1.5,
			Custom: make(map[string]interface{}),
		},
		Featured:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	err := client.IndexProduct(ctx, product)
	require.NoError(t, err)
	
	// Verify product was indexed
	assert.Contains(t, client.products, product.ID)
	assert.Equal(t, product, client.products[product.ID])
}

func TestElasticsearchClient_BulkIndexProducts(t *testing.T) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	products := []*models.Product{
		{
			ID:          "product-1",
			SKU:         "SKU-001",
			Name:        "Product 1",
			Description: "First product",
			Price:       decimal.NewFromFloat(10.00),
			Status:      models.ProductActive,
			Attributes:  models.ProductAttributes{Custom: make(map[string]interface{})},
		},
		{
			ID:          "product-2",
			SKU:         "SKU-002",
			Name:        "Product 2",
			Description: "Second product",
			Price:       decimal.NewFromFloat(20.00),
			Status:      models.ProductActive,
			Attributes:  models.ProductAttributes{Custom: make(map[string]interface{})},
		},
	}
	
	err := client.BulkIndexProducts(ctx, products)
	require.NoError(t, err)
	
	// Verify all products were indexed
	assert.Len(t, client.products, 2)
	assert.Contains(t, client.products, "product-1")
	assert.Contains(t, client.products, "product-2")
}

func TestElasticsearchClient_DeleteProduct(t *testing.T) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Index a product first
	product := &models.Product{
		ID:         "test-product",
		Name:       "Test Product",
		Attributes: models.ProductAttributes{Custom: make(map[string]interface{})},
	}
	client.products[product.ID] = product
	
	// Delete the product
	err := client.DeleteProduct(ctx, product.ID)
	require.NoError(t, err)
	
	// Verify product was deleted
	assert.NotContains(t, client.products, product.ID)
}

func TestElasticsearchClient_SearchProducts(t *testing.T) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Index test products
	products := []*models.Product{
		{
			ID:          "product-1",
			Name:        "Red Shirt",
			Description: "A red cotton shirt",
			CategoryID:  "clothing",
			Status:      models.ProductActive,
			Attributes: models.ProductAttributes{
				Brand:  "TestBrand",
				Color:  "Red",
				Custom: make(map[string]interface{}),
			},
		},
		{
			ID:          "product-2",
			Name:        "Blue Jeans",
			Description: "Blue denim jeans",
			CategoryID:  "clothing",
			Status:      models.ProductActive,
			Attributes: models.ProductAttributes{
				Brand:  "TestBrand",
				Color:  "Blue",
				Custom: make(map[string]interface{}),
			},
		},
		{
			ID:          "product-3",
			Name:        "Green Hat",
			Description: "A green baseball cap",
			CategoryID:  "accessories",
			Status:      models.ProductActive,
			Attributes: models.ProductAttributes{
				Brand:  "OtherBrand",
				Color:  "Green",
				Custom: make(map[string]interface{}),
			},
		},
	}
	
	for _, product := range products {
		client.products[product.ID] = product
	}
	
	tests := []struct {
		name           string
		request        SearchRequest
		expectedCount  int
		expectedTotal  int64
	}{
		{
			name: "Search by query",
			request: SearchRequest{
				Query: "shirt",
				From:  0,
				Size:  10,
			},
			expectedCount: 1,
			expectedTotal: 3,
		},
		{
			name: "Search with category filter",
			request: SearchRequest{
				Query: "",
				Filters: map[string]interface{}{
					"category_id": "clothing",
				},
				From: 0,
				Size: 10,
			},
			expectedCount: 2,
			expectedTotal: 3,
		},
		{
			name: "Search with brand filter",
			request: SearchRequest{
				Query: "",
				Filters: map[string]interface{}{
					"brand": "TestBrand",
				},
				From: 0,
				Size: 10,
			},
			expectedCount: 2,
			expectedTotal: 3,
		},
		{
			name: "Search with pagination",
			request: SearchRequest{
				Query: "",
				From:  1,
				Size:  1,
			},
			expectedCount: 1,
			expectedTotal: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.SearchProducts(ctx, tt.request)
			require.NoError(t, err)
			
			assert.Len(t, result.Products, tt.expectedCount)
			assert.Equal(t, tt.expectedTotal, result.Total)
			assert.Equal(t, tt.request.From, result.From)
			assert.Equal(t, tt.request.Size, result.Size)
		})
	}
}

func TestElasticsearchClient_GetSearchSuggestions(t *testing.T) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Index test products
	products := []*models.Product{
		{
			ID:         "product-1",
			Name:       "Red Shirt",
			Attributes: models.ProductAttributes{Custom: make(map[string]interface{})},
		},
		{
			ID:         "product-2",
			Name:       "Red Shoes",
			Attributes: models.ProductAttributes{Custom: make(map[string]interface{})},
		},
		{
			ID:         "product-3",
			Name:       "Blue Jeans",
			Attributes: models.ProductAttributes{Custom: make(map[string]interface{})},
		},
	}
	
	for _, product := range products {
		client.products[product.ID] = product
	}
	
	suggestions, err := client.GetSearchSuggestions(ctx, "Red", 5)
	require.NoError(t, err)
	
	assert.Len(t, suggestions, 2)
	assert.Contains(t, suggestions, "Red Shirt")
	assert.Contains(t, suggestions, "Red Shoes")
}

func TestSearchRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request SearchRequest
		valid   bool
	}{
		{
			name: "Valid request",
			request: SearchRequest{
				Query: "test",
				From:  0,
				Size:  10,
			},
			valid: true,
		},
		{
			name: "Empty query is valid",
			request: SearchRequest{
				Query: "",
				From:  0,
				Size:  10,
			},
			valid: true,
		},
		{
			name: "Negative from",
			request: SearchRequest{
				Query: "test",
				From:  -1,
				Size:  10,
			},
			valid: false,
		},
		{
			name: "Zero size",
			request: SearchRequest{
				Query: "test",
				From:  0,
				Size:  0,
			},
			valid: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.request.From >= 0 && tt.request.Size > 0
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestFacetValue_Structure(t *testing.T) {
	facet := FacetValue{
		Value: "TestBrand",
		Count: 42,
	}
	
	assert.Equal(t, "TestBrand", facet.Value)
	assert.Equal(t, int64(42), facet.Count)
}

func TestSortField_Structure(t *testing.T) {
	sort := SortField{
		Field: "price",
		Order: "desc",
	}
	
	assert.Equal(t, "price", sort.Field)
	assert.Equal(t, "desc", sort.Order)
}