package search

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopsphere/shared/models"
	"github.com/shopspring/decimal"
)

// BenchmarkElasticsearchClient_IndexProduct benchmarks product indexing
func BenchmarkElasticsearchClient_IndexProduct(b *testing.B) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Create a sample product
	product := &models.Product{
		ID:          "benchmark-product",
		SKU:         "BENCH-001",
		Name:        "Benchmark Product",
		Description: "A product for benchmarking indexing performance",
		CategoryID:  "benchmark-category",
		Price:       decimal.NewFromFloat(99.99),
		Currency:    "USD",
		Stock:       100,
		Status:      models.ProductActive,
		Images:      []string{"image1.jpg", "image2.jpg"},
		Attributes: models.ProductAttributes{
			Brand:  "BenchmarkBrand",
			Color:  "Blue",
			Size:   "L",
			Weight: 2.5,
			Custom: map[string]interface{}{
				"material": "cotton",
				"origin":   "USA",
			},
		},
		Featured:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		product.ID = fmt.Sprintf("benchmark-product-%d", i)
		err := client.IndexProduct(ctx, product)
		if err != nil {
			b.Fatalf("Failed to index product: %v", err)
		}
	}
}

// BenchmarkElasticsearchClient_BulkIndexProducts benchmarks bulk indexing
func BenchmarkElasticsearchClient_BulkIndexProducts(b *testing.B) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Create sample products for bulk indexing
	createProducts := func(count int) []*models.Product {
		products := make([]*models.Product, count)
		for i := 0; i < count; i++ {
			products[i] = &models.Product{
				ID:          fmt.Sprintf("bulk-product-%d", i),
				SKU:         fmt.Sprintf("BULK-%03d", i),
				Name:        fmt.Sprintf("Bulk Product %d", i),
				Description: fmt.Sprintf("Bulk product number %d for benchmarking", i),
				CategoryID:  "bulk-category",
				Price:       decimal.NewFromFloat(float64(10 + i)),
				Currency:    "USD",
				Stock:       50,
				Status:      models.ProductActive,
				Attributes: models.ProductAttributes{
					Brand:  "BulkBrand",
					Custom: make(map[string]interface{}),
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}
		return products
	}
	
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_products", 10},
		{"50_products", 50},
		{"100_products", 100},
		{"500_products", 500},
	}
	
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			products := createProducts(bm.count)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				err := client.BulkIndexProducts(ctx, products)
				if err != nil {
					b.Fatalf("Failed to bulk index products: %v", err)
				}
			}
		})
	}
}

// BenchmarkElasticsearchClient_SearchProducts benchmarks search operations
func BenchmarkElasticsearchClient_SearchProducts(b *testing.B) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Index sample products for searching
	products := make([]*models.Product, 1000)
	for i := 0; i < 1000; i++ {
		products[i] = &models.Product{
			ID:          fmt.Sprintf("search-product-%d", i),
			SKU:         fmt.Sprintf("SEARCH-%04d", i),
			Name:        fmt.Sprintf("Search Product %d", i),
			Description: fmt.Sprintf("Product for search benchmarking number %d", i),
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
	
	// Bulk index all products
	err := client.BulkIndexProducts(ctx, products)
	if err != nil {
		b.Fatalf("Failed to index products for benchmark: %v", err)
	}
	
	searchRequests := []struct {
		name    string
		request SearchRequest
	}{
		{
			name: "simple_query",
			request: SearchRequest{
				Query: "Product",
				From:  0,
				Size:  20,
			},
		},
		{
			name: "filtered_search",
			request: SearchRequest{
				Query: "Product",
				Filters: map[string]interface{}{
					"category_id": "category-1",
				},
				From: 0,
				Size: 20,
			},
		},
		{
			name: "complex_filtered_search",
			request: SearchRequest{
				Query: "Product",
				Filters: map[string]interface{}{
					"category_id": "category-1",
					"brand":       "Brand1",
					"status":      "active",
				},
				From: 0,
				Size: 20,
			},
		},
		{
			name: "large_result_set",
			request: SearchRequest{
				Query: "",
				From:  0,
				Size:  100,
			},
		},
		{
			name: "paginated_search",
			request: SearchRequest{
				Query: "Product",
				From:  50,
				Size:  20,
			},
		},
	}
	
	for _, sr := range searchRequests {
		b.Run(sr.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_, err := client.SearchProducts(ctx, sr.request)
				if err != nil {
					b.Fatalf("Search failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkElasticsearchClient_GetSearchSuggestions benchmarks suggestion queries
func BenchmarkElasticsearchClient_GetSearchSuggestions(b *testing.B) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Index products with various names for suggestions
	productNames := []string{
		"Red Shirt", "Red Shoes", "Red Hat", "Red Jacket",
		"Blue Jeans", "Blue Shirt", "Blue Shoes", "Blue Hat",
		"Green Pants", "Green Shirt", "Green Shoes", "Green Hat",
		"Black Dress", "Black Shirt", "Black Shoes", "Black Hat",
		"White Shirt", "White Shoes", "White Hat", "White Jacket",
	}
	
	for i, name := range productNames {
		product := &models.Product{
			ID:         fmt.Sprintf("suggestion-product-%d", i),
			Name:       name,
			Attributes: models.ProductAttributes{Custom: make(map[string]interface{})},
		}
		client.products[product.ID] = product
	}
	
	queries := []string{"Red", "Blue", "Shirt", "Shoes", "Hat"}
	
	for _, query := range queries {
		b.Run(fmt.Sprintf("query_%s", query), func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_, err := client.GetSearchSuggestions(ctx, query, 10)
				if err != nil {
					b.Fatalf("Suggestion query failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkElasticsearchClient_DeleteProduct benchmarks product deletion
func BenchmarkElasticsearchClient_DeleteProduct(b *testing.B) {
	client := NewMockElasticsearchClient()
	ctx := context.Background()
	
	// Pre-populate with products to delete
	for i := 0; i < b.N; i++ {
		product := &models.Product{
			ID:         fmt.Sprintf("delete-product-%d", i),
			Name:       fmt.Sprintf("Delete Product %d", i),
			Attributes: models.ProductAttributes{Custom: make(map[string]interface{})},
		}
		client.products[product.ID] = product
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		productID := fmt.Sprintf("delete-product-%d", i)
		err := client.DeleteProduct(ctx, productID)
		if err != nil {
			b.Fatalf("Failed to delete product: %v", err)
		}
	}
}

// BenchmarkSearchRequest_Creation benchmarks search request creation
func BenchmarkSearchRequest_Creation(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := SearchRequest{
			Query: fmt.Sprintf("query-%d", i),
			Filters: map[string]interface{}{
				"category_id": fmt.Sprintf("category-%d", i%10),
				"brand":       fmt.Sprintf("brand-%d", i%5),
				"status":      "active",
			},
			Sort: []SortField{
				{Field: "price", Order: "asc"},
				{Field: "created_at", Order: "desc"},
			},
			From:   i % 100,
			Size:   20,
			Facets: []string{"brand", "color", "size"},
		}
		
		// Use the request to avoid compiler optimization
		_ = req.Query
	}
}

// BenchmarkProductDocument_Conversion benchmarks product to document conversion
func BenchmarkProductDocument_Conversion(b *testing.B) {
	client := &ElasticsearchClient{}
	
	product := &models.Product{
		ID:          "benchmark-product",
		SKU:         "BENCH-001",
		Name:        "Benchmark Product",
		Description: "A product for benchmarking document conversion",
		CategoryID:  "benchmark-category",
		Price:       decimal.NewFromFloat(99.99),
		Currency:    "USD",
		Stock:       100,
		Status:      models.ProductActive,
		Images:      []string{"image1.jpg", "image2.jpg", "image3.jpg"},
		Attributes: models.ProductAttributes{
			Brand:  "BenchmarkBrand",
			Color:  "Blue",
			Size:   "L",
			Weight: 2.5,
			Custom: map[string]interface{}{
				"material":    "cotton",
				"origin":      "USA",
				"washable":    true,
				"temperature": 30,
			},
		},
		Featured:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		doc := client.productToDocument(product)
		// Use the document to avoid compiler optimization
		_ = doc.ID
	}
}