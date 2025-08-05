package repository

import (
	"context"

	"github.com/shopsphere/shared/models"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	// Product CRUD operations
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ProductFilter) ([]*models.Product, int, error)
	
	// Inventory operations
	UpdateStock(ctx context.Context, productID string, quantity int) error
	ReserveStock(ctx context.Context, productID string, quantity int) error
	ReleaseStock(ctx context.Context, productID string, quantity int) error
	GetAvailableStock(ctx context.Context, productID string) (int, error)
	
	// Bulk operations
	BulkUpdateStock(ctx context.Context, updates []StockUpdate) error
}

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id string) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter CategoryFilter) ([]*models.Category, error)
	GetChildren(ctx context.Context, parentID string) ([]*models.Category, error)
	GetPath(ctx context.Context, categoryID string) ([]*models.Category, error)
}

// ProductFilter represents filtering options for products
type ProductFilter struct {
	CategoryID   string
	Status       string
	MinPrice     *float64
	MaxPrice     *float64
	SearchTerm   string
	Featured     *bool
	InStock      *bool
	Limit        int
	Offset       int
	SortBy       string
	SortOrder    string
}

// CategoryFilter represents filtering options for categories
type CategoryFilter struct {
	ParentID  *string
	IsActive  *bool
	Level     *int
	Limit     int
	Offset    int
}

// StockUpdate represents a stock update operation
type StockUpdate struct {
	ProductID string
	Quantity  int
	Type      string // "in", "out", "adjustment"
	Reason    string
}