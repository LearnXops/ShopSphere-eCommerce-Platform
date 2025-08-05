package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductStatus represents the status of a product
type ProductStatus string

const (
	ProductActive      ProductStatus = "active"
	ProductInactive    ProductStatus = "inactive"
	ProductOutOfStock  ProductStatus = "out_of_stock"
	ProductDiscontinued ProductStatus = "discontinued"
)

// Product represents a product in the catalog
type Product struct {
	ID          string            `json:"id" db:"id"`
	SKU         string            `json:"sku" db:"sku"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	CategoryID  string            `json:"category_id" db:"category_id"`
	Price       decimal.Decimal   `json:"price" db:"price"`
	Currency    string            `json:"currency" db:"currency"`
	Stock       int               `json:"stock" db:"stock"`
	Status      ProductStatus     `json:"status" db:"status"`
	Images      []string          `json:"images" db:"images"`
	Attributes  ProductAttributes `json:"attributes" db:"attributes"`
	Featured    bool              `json:"featured" db:"featured"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// ProductAttributes represents additional product attributes
type ProductAttributes struct {
	Brand      string                 `json:"brand"`
	Color      string                 `json:"color"`
	Size       string                 `json:"size"`
	Weight     float64                `json:"weight"`
	Dimensions Dimensions             `json:"dimensions"`
	Custom     map[string]interface{} `json:"custom"`
}

// Dimensions represents product dimensions
type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"` // cm, in, etc.
}

// NewProduct creates a new product with default values
func NewProduct(sku, name, description, categoryID string, price decimal.Decimal) *Product {
	return &Product{
		ID:          uuid.New().String(),
		SKU:         sku,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		Price:       price,
		Currency:    "USD",
		Stock:       0,
		Status:      ProductInactive,
		Images:      []string{},
		Attributes:  ProductAttributes{Custom: make(map[string]interface{})},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewCategory creates a new category with default values
func NewCategory(name, description string, parentID *string) *Category {
	return &Category{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		ParentID:    parentID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Category represents a product category
type Category struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ParentID    *string   `json:"parent_id" db:"parent_id"`
	Path        string    `json:"path" db:"path"`
	Level       int       `json:"level" db:"level"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}