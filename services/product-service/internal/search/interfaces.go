package search

import (
	"context"
	"time"

	"github.com/shopsphere/shared/models"
)

// SearchService defines the interface for product search operations
type SearchService interface {
	// IndexProduct indexes a single product
	IndexProduct(ctx context.Context, product *models.Product) error
	
	// BulkIndexProducts indexes multiple products in bulk
	BulkIndexProducts(ctx context.Context, products []*models.Product) error
	
	// DeleteProduct removes a product from the search index
	DeleteProduct(ctx context.Context, productID string) error
	
	// SearchProducts performs advanced product search
	SearchProducts(ctx context.Context, req SearchRequest) (*SearchResponse, error)
	
	// GetSearchSuggestions returns search suggestions
	GetSearchSuggestions(ctx context.Context, query string, size int) ([]string, error)
}

// SearchAnalytics defines the interface for search analytics
type SearchAnalytics interface {
	// RecordSearch records a search query for analytics
	RecordSearch(ctx context.Context, query string, userID string, resultsCount int) error
	
	// GetPopularSearches returns popular search terms
	GetPopularSearches(ctx context.Context, limit int) ([]SearchTerm, error)
	
	// GetSearchMetrics returns search performance metrics
	GetSearchMetrics(ctx context.Context, from, to time.Time) (*SearchMetrics, error)
}

// SearchTerm represents a search term with frequency
type SearchTerm struct {
	Term      string `json:"term"`
	Frequency int    `json:"frequency"`
}

// SearchMetrics represents search performance metrics
type SearchMetrics struct {
	TotalSearches     int64   `json:"total_searches"`
	AverageResults    float64 `json:"average_results"`
	ZeroResultsRate   float64 `json:"zero_results_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	PopularTerms      []SearchTerm `json:"popular_terms"`
}