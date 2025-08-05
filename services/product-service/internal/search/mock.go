package search

import (
	"context"
	"strings"

	"github.com/shopsphere/shared/models"
)

// MockElasticsearchClient is a mock implementation for testing
type MockElasticsearchClient struct {
	products    map[string]*models.Product
	searchCalls []SearchRequest
}

// NewMockElasticsearchClient creates a new mock Elasticsearch client
func NewMockElasticsearchClient() *MockElasticsearchClient {
	return &MockElasticsearchClient{
		products:    make(map[string]*models.Product),
		searchCalls: []SearchRequest{},
	}
}

func (m *MockElasticsearchClient) IndexProduct(ctx context.Context, product *models.Product) error {
	m.products[product.ID] = product
	return nil
}

func (m *MockElasticsearchClient) BulkIndexProducts(ctx context.Context, products []*models.Product) error {
	for _, product := range products {
		m.products[product.ID] = product
	}
	return nil
}

func (m *MockElasticsearchClient) DeleteProduct(ctx context.Context, productID string) error {
	delete(m.products, productID)
	return nil
}

func (m *MockElasticsearchClient) SearchProducts(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	m.searchCalls = append(m.searchCalls, req)
	
	var matchedProducts []*models.Product
	for _, product := range m.products {
		if m.matchesQuery(product, req) {
			matchedProducts = append(matchedProducts, product)
		}
	}
	
	// Apply pagination
	from := req.From
	size := req.Size
	if from >= len(matchedProducts) {
		matchedProducts = []*models.Product{}
	} else {
		end := from + size
		if end > len(matchedProducts) {
			end = len(matchedProducts)
		}
		matchedProducts = matchedProducts[from:end]
	}
	
	return &SearchResponse{
		Products: matchedProducts,
		Total:    int64(len(m.products)),
		From:     req.From,
		Size:     req.Size,
		Facets:   make(map[string][]FacetValue),
	}, nil
}

func (m *MockElasticsearchClient) GetSearchSuggestions(ctx context.Context, query string, size int) ([]string, error) {
	var suggestions []string
	for _, product := range m.products {
		if len(suggestions) >= size {
			break
		}
		if contains(product.Name, query) {
			suggestions = append(suggestions, product.Name)
		}
	}
	return suggestions, nil
}

func (m *MockElasticsearchClient) matchesQuery(product *models.Product, req SearchRequest) bool {
	if req.Query != "" {
		if !contains(product.Name, req.Query) && !contains(product.Description, req.Query) {
			return false
		}
	}
	
	// Apply filters
	for field, value := range req.Filters {
		switch field {
		case "category_id":
			if product.CategoryID != value.(string) {
				return false
			}
		case "status":
			if string(product.Status) != value.(string) {
				return false
			}
		case "brand":
			if product.Attributes.Brand != value.(string) {
				return false
			}
		}
	}
	
	return true
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}