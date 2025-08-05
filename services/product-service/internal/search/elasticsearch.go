package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/shopsphere/shared/models"
	"github.com/shopspring/decimal"
)

const (
	ProductIndex = "products"
)

// ElasticsearchClient wraps the Elasticsearch client with search functionality
type ElasticsearchClient struct {
	client *elasticsearch.Client
}

// NewElasticsearchClient creates a new Elasticsearch client
func NewElasticsearchClient(addresses []string) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	esClient := &ElasticsearchClient{client: client}
	
	// Initialize indices
	if err := esClient.initializeIndices(); err != nil {
		return nil, fmt.Errorf("failed to initialize indices: %w", err)
	}

	return esClient, nil
}

// ProductDocument represents a product document in Elasticsearch
type ProductDocument struct {
	ID          string                 `json:"id"`
	SKU         string                 `json:"sku"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CategoryID  string                 `json:"category_id"`
	Price       float64                `json:"price"`
	Currency    string                 `json:"currency"`
	Stock       int                    `json:"stock"`
	Status      string                 `json:"status"`
	Images      []string               `json:"images"`
	Brand       string                 `json:"brand"`
	Color       string                 `json:"color"`
	Size        string                 `json:"size"`
	Weight      float64                `json:"weight"`
	Featured    bool                   `json:"featured"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Custom      map[string]interface{} `json:"custom"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string            `json:"query"`
	Filters    map[string]interface{} `json:"filters"`
	Sort       []SortField       `json:"sort"`
	From       int               `json:"from"`
	Size       int               `json:"size"`
	Facets     []string          `json:"facets"`
}

// SortField represents a sort field
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc or desc
}

// SearchResponse represents a search response
type SearchResponse struct {
	Products []*models.Product          `json:"products"`
	Total    int64                      `json:"total"`
	Facets   map[string][]FacetValue    `json:"facets"`
	From     int                        `json:"from"`
	Size     int                        `json:"size"`
}

// FacetValue represents a facet value with count
type FacetValue struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// initializeIndices creates the product index with proper mappings
func (es *ElasticsearchClient) initializeIndices() error {
	ctx := context.Background()

	// Check if index exists
	req := esapi.IndicesExistsRequest{
		Index: []string{ProductIndex},
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	// If index exists, return
	if res.StatusCode == 200 {
		return nil
	}

	// Create index with mappings
	mapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"sku": {"type": "keyword"},
				"name": {
					"type": "text",
					"analyzer": "standard",
					"fields": {
						"keyword": {"type": "keyword"},
						"suggest": {
							"type": "completion",
							"analyzer": "simple"
						}
					}
				},
				"description": {
					"type": "text",
					"analyzer": "standard"
				},
				"category_id": {"type": "keyword"},
				"price": {"type": "double"},
				"currency": {"type": "keyword"},
				"stock": {"type": "integer"},
				"status": {"type": "keyword"},
				"images": {"type": "keyword"},
				"brand": {
					"type": "text",
					"fields": {"keyword": {"type": "keyword"}}
				},
				"color": {
					"type": "text",
					"fields": {"keyword": {"type": "keyword"}}
				},
				"size": {
					"type": "text",
					"fields": {"keyword": {"type": "keyword"}}
				},
				"weight": {"type": "double"},
				"featured": {"type": "boolean"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"},
				"custom": {"type": "object", "dynamic": true}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"product_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "stop", "snowball"]
					}
				}
			}
		}
	}`

	createReq := esapi.IndicesCreateRequest{
		Index: ProductIndex,
		Body:  strings.NewReader(mapping),
	}

	createRes, err := createReq.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("failed to create index: %s", createRes.String())
	}

	log.Printf("Created Elasticsearch index: %s", ProductIndex)
	return nil
}

// IndexProduct indexes a product document
func (es *ElasticsearchClient) IndexProduct(ctx context.Context, product *models.Product) error {
	doc := es.productToDocument(product)
	
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal product document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      ProductIndex,
		DocumentID: product.ID,
		Body:       bytes.NewReader(docJSON),
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to index product: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index product: %s", res.String())
	}

	return nil
}

// BulkIndexProducts indexes multiple products in bulk
func (es *ElasticsearchClient) BulkIndexProducts(ctx context.Context, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, product := range products {
		doc := es.productToDocument(product)
		
		// Index action
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": ProductIndex,
				"_id":    product.ID,
			},
		}
		
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')
		
		// Document
		docJSON, _ := json.Marshal(doc)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Index:   ProductIndex,
		Body:    &buf,
		Refresh: "wait_for",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to bulk index products: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to bulk index products: %s", res.String())
	}

	return nil
}

// DeleteProduct removes a product from the index
func (es *ElasticsearchClient) DeleteProduct(ctx context.Context, productID string) error {
	req := esapi.DeleteRequest{
		Index:      ProductIndex,
		DocumentID: productID,
		Refresh:    "wait_for",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete product: %s", res.String())
	}

	return nil
}

// SearchProducts performs advanced product search
func (es *ElasticsearchClient) SearchProducts(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	query := es.buildSearchQuery(req)
	
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	searchReq := esapi.SearchRequest{
		Index: []string{ProductIndex},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := searchReq.Do(ctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return es.parseSearchResponse(searchResult, req)
}

// GetSearchSuggestions returns search suggestions
func (es *ElasticsearchClient) GetSearchSuggestions(ctx context.Context, query string, size int) ([]string, error) {
	if size <= 0 {
		size = 10
	}

	searchQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"product_suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "name.suggest",
					"size":  size,
				},
			},
		},
	}

	queryJSON, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggestion query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{ProductIndex},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute suggestion search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("suggestion search error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode suggestion response: %w", err)
	}

	return es.parseSuggestionResponse(result)
}

// productToDocument converts a product model to an Elasticsearch document
func (es *ElasticsearchClient) productToDocument(product *models.Product) *ProductDocument {
	price, _ := product.Price.Float64()
	
	return &ProductDocument{
		ID:          product.ID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		CategoryID:  product.CategoryID,
		Price:       price,
		Currency:    product.Currency,
		Stock:       product.Stock,
		Status:      string(product.Status),
		Images:      product.Images,
		Brand:       product.Attributes.Brand,
		Color:       product.Attributes.Color,
		Size:        product.Attributes.Size,
		Weight:      product.Attributes.Weight,
		Featured:    product.Featured,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
		Custom:      product.Attributes.Custom,
	}
}

// buildSearchQuery builds an Elasticsearch query from search request
func (es *ElasticsearchClient) buildSearchQuery(req SearchRequest) map[string]interface{} {
	query := map[string]interface{}{
		"from": req.From,
		"size": req.Size,
	}

	// Build the main query
	var boolQuery map[string]interface{}
	
	if req.Query != "" {
		// Multi-match query for text search
		boolQuery = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":  req.Query,
							"fields": []string{"name^3", "description^2", "brand", "sku"},
							"type":   "best_fields",
							"fuzziness": "AUTO",
						},
					},
				},
			},
		}
	} else {
		boolQuery = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"match_all": map[string]interface{}{},
					},
				},
			},
		}
	}

	// Add filters
	if len(req.Filters) > 0 {
		var filters []interface{}
		
		for field, value := range req.Filters {
			switch field {
			case "category_id", "status", "brand", "color", "size":
				filters = append(filters, map[string]interface{}{
					"term": map[string]interface{}{
						field: value,
					},
				})
			case "price_min":
				if priceMin, ok := value.(float64); ok {
					filters = append(filters, map[string]interface{}{
						"range": map[string]interface{}{
							"price": map[string]interface{}{
								"gte": priceMin,
							},
						},
					})
				}
			case "price_max":
				if priceMax, ok := value.(float64); ok {
					filters = append(filters, map[string]interface{}{
						"range": map[string]interface{}{
							"price": map[string]interface{}{
								"lte": priceMax,
							},
						},
					})
				}
			case "in_stock":
				if inStock, ok := value.(bool); ok && inStock {
					filters = append(filters, map[string]interface{}{
						"range": map[string]interface{}{
							"stock": map[string]interface{}{
								"gt": 0,
							},
						},
					})
				}
			case "featured":
				if featured, ok := value.(bool); ok {
					filters = append(filters, map[string]interface{}{
						"term": map[string]interface{}{
							"featured": featured,
						},
					})
				}
			}
		}
		
		if len(filters) > 0 {
			boolQuery["bool"].(map[string]interface{})["filter"] = filters
		}
	}

	query["query"] = boolQuery

	// Add sorting
	if len(req.Sort) > 0 {
		var sorts []interface{}
		for _, sort := range req.Sort {
			order := "asc"
			if sort.Order == "desc" {
				order = "desc"
			}
			sorts = append(sorts, map[string]interface{}{
				sort.Field: map[string]interface{}{
					"order": order,
				},
			})
		}
		query["sort"] = sorts
	} else {
		// Default sorting by relevance score and then by created_at
		query["sort"] = []interface{}{
			map[string]interface{}{
				"_score": map[string]interface{}{
					"order": "desc",
				},
			},
			map[string]interface{}{
				"created_at": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	}

	// Add aggregations for facets
	if len(req.Facets) > 0 {
		aggs := make(map[string]interface{})
		for _, facet := range req.Facets {
			switch facet {
			case "brand", "color", "size", "status":
				aggs[facet] = map[string]interface{}{
					"terms": map[string]interface{}{
						"field": facet + ".keyword",
						"size":  20,
					},
				}
			case "price_ranges":
				aggs["price_ranges"] = map[string]interface{}{
					"range": map[string]interface{}{
						"field": "price",
						"ranges": []interface{}{
							map[string]interface{}{"to": 25},
							map[string]interface{}{"from": 25, "to": 50},
							map[string]interface{}{"from": 50, "to": 100},
							map[string]interface{}{"from": 100, "to": 200},
							map[string]interface{}{"from": 200},
						},
					},
				}
			}
		}
		if len(aggs) > 0 {
			query["aggs"] = aggs
		}
	}

	return query
}

// parseSearchResponse parses Elasticsearch search response
func (es *ElasticsearchClient) parseSearchResponse(result map[string]interface{}, req SearchRequest) (*SearchResponse, error) {
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid search response format")
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid total format in search response")
	}

	totalValue, ok := total["value"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid total value in search response")
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid hits format in search response")
	}

	var products []*models.Product
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		product, err := es.documentToProduct(source)
		if err != nil {
			log.Printf("Failed to convert document to product: %v", err)
			continue
		}

		products = append(products, product)
	}

	response := &SearchResponse{
		Products: products,
		Total:    int64(totalValue),
		From:     req.From,
		Size:     req.Size,
		Facets:   make(map[string][]FacetValue),
	}

	// Parse aggregations/facets
	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		for facetName, aggResult := range aggs {
			if buckets := es.extractBuckets(aggResult); buckets != nil {
				response.Facets[facetName] = buckets
			}
		}
	}

	return response, nil
}

// parseSuggestionResponse parses Elasticsearch suggestion response
func (es *ElasticsearchClient) parseSuggestionResponse(result map[string]interface{}) ([]string, error) {
	suggest, ok := result["suggest"].(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	productSuggest, ok := suggest["product_suggest"].([]interface{})
	if !ok || len(productSuggest) == 0 {
		return []string{}, nil
	}

	firstSuggest, ok := productSuggest[0].(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	options, ok := firstSuggest["options"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var suggestions []string
	for _, option := range options {
		optionMap, ok := option.(map[string]interface{})
		if !ok {
			continue
		}

		text, ok := optionMap["text"].(string)
		if ok {
			suggestions = append(suggestions, text)
		}
	}

	return suggestions, nil
}

// documentToProduct converts an Elasticsearch document to a product model
func (es *ElasticsearchClient) documentToProduct(doc map[string]interface{}) (*models.Product, error) {
	product := &models.Product{}

	if id, ok := doc["id"].(string); ok {
		product.ID = id
	}
	if sku, ok := doc["sku"].(string); ok {
		product.SKU = sku
	}
	if name, ok := doc["name"].(string); ok {
		product.Name = name
	}
	if description, ok := doc["description"].(string); ok {
		product.Description = description
	}
	if categoryID, ok := doc["category_id"].(string); ok {
		product.CategoryID = categoryID
	}
	if price, ok := doc["price"].(float64); ok {
		product.Price = decimal.NewFromFloat(price)
	}
	if currency, ok := doc["currency"].(string); ok {
		product.Currency = currency
	}
	if stock, ok := doc["stock"].(float64); ok {
		product.Stock = int(stock)
	}
	if status, ok := doc["status"].(string); ok {
		product.Status = models.ProductStatus(status)
	}
	if featured, ok := doc["featured"].(bool); ok {
		product.Featured = featured
	}

	// Handle images array
	if images, ok := doc["images"].([]interface{}); ok {
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				product.Images = append(product.Images, imgStr)
			}
		}
	}

	// Handle attributes
	product.Attributes = models.ProductAttributes{
		Custom: make(map[string]interface{}),
	}
	
	if brand, ok := doc["brand"].(string); ok {
		product.Attributes.Brand = brand
	}
	if color, ok := doc["color"].(string); ok {
		product.Attributes.Color = color
	}
	if size, ok := doc["size"].(string); ok {
		product.Attributes.Size = size
	}
	if weight, ok := doc["weight"].(float64); ok {
		product.Attributes.Weight = weight
	}
	if custom, ok := doc["custom"].(map[string]interface{}); ok {
		product.Attributes.Custom = custom
	}

	// Handle timestamps
	if createdAt, ok := doc["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			product.CreatedAt = t
		}
	}
	if updatedAt, ok := doc["updated_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			product.UpdatedAt = t
		}
	}

	return product, nil
}

// extractBuckets extracts facet buckets from aggregation result
func (es *ElasticsearchClient) extractBuckets(aggResult interface{}) []FacetValue {
	aggMap, ok := aggResult.(map[string]interface{})
	if !ok {
		return nil
	}

	buckets, ok := aggMap["buckets"].([]interface{})
	if !ok {
		return nil
	}

	var facetValues []FacetValue
	for _, bucket := range buckets {
		bucketMap, ok := bucket.(map[string]interface{})
		if !ok {
			continue
		}

		key, keyOk := bucketMap["key"].(string)
		docCount, countOk := bucketMap["doc_count"].(float64)

		if keyOk && countOk {
			facetValues = append(facetValues, FacetValue{
				Value: key,
				Count: int64(docCount),
			})
		}
	}

	return facetValues
}