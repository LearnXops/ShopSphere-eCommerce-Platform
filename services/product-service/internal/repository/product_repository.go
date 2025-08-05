package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

type productRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	if product.ID == "" {
		product.ID = uuid.New().String()
	}
	
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	
	attributesJSON, err := json.Marshal(product.Attributes)
	if err != nil {
		return utils.NewInternalError("failed to marshal product attributes", err)
	}
	
	query := `
		INSERT INTO products (
			id, sku, name, description, category_id, price, currency, stock, 
			status, weight, length, width, height, images, attributes, 
			featured, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)`
	
	_, err = r.db.ExecContext(ctx, query,
		product.ID, product.SKU, product.Name, product.Description, product.CategoryID,
		product.Price, product.Currency, product.Stock, product.Status,
		product.Attributes.Weight, product.Attributes.Dimensions.Length,
		product.Attributes.Dimensions.Width, product.Attributes.Dimensions.Height,
		pq.Array(product.Images), attributesJSON, false,
		product.CreatedAt, product.UpdatedAt,
	)
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return utils.NewConflictError("product with this SKU already exists")
		}
		return utils.NewInternalError("failed to create product", err)
	}
	
	return nil
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	query := `
		SELECT id, sku, name, description, category_id, price, currency, stock, 
			   reserved_stock, status, weight, length, width, height, images, 
			   attributes, featured, created_at, updated_at
		FROM products 
		WHERE id = $1`
	
	product := &models.Product{}
	var attributesJSON []byte
	var weight, length, width, height sql.NullFloat64
	var reservedStock int
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.SKU, &product.Name, &product.Description,
		&product.CategoryID, &product.Price, &product.Currency, &product.Stock,
		&reservedStock, &product.Status, &weight, &length, &width, &height,
		pq.Array(&product.Images), &attributesJSON, &product.Featured,
		&product.CreatedAt, &product.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("product")
		}
		return nil, utils.NewInternalError("failed to get product", err)
	}
	
	// Parse attributes
	if err := json.Unmarshal(attributesJSON, &product.Attributes); err != nil {
		return nil, utils.NewInternalError("failed to unmarshal product attributes", err)
	}
	
	// Set dimensions
	if weight.Valid {
		product.Attributes.Weight = weight.Float64
	}
	if length.Valid {
		product.Attributes.Dimensions.Length = length.Float64
	}
	if width.Valid {
		product.Attributes.Dimensions.Width = width.Float64
	}
	if height.Valid {
		product.Attributes.Dimensions.Height = height.Float64
	}
	
	return product, nil
}

// GetBySKU retrieves a product by SKU
func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	query := `
		SELECT id, sku, name, description, category_id, price, currency, stock, 
			   reserved_stock, status, weight, length, width, height, images, 
			   attributes, featured, created_at, updated_at
		FROM products 
		WHERE sku = $1`
	
	product := &models.Product{}
	var attributesJSON []byte
	var weight, length, width, height sql.NullFloat64
	var reservedStock int
	
	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&product.ID, &product.SKU, &product.Name, &product.Description,
		&product.CategoryID, &product.Price, &product.Currency, &product.Stock,
		&reservedStock, &product.Status, &weight, &length, &width, &height,
		pq.Array(&product.Images), &attributesJSON, &product.Featured,
		&product.CreatedAt, &product.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("product")
		}
		return nil, utils.NewInternalError("failed to get product", err)
	}
	
	// Parse attributes
	if err := json.Unmarshal(attributesJSON, &product.Attributes); err != nil {
		return nil, utils.NewInternalError("failed to unmarshal product attributes", err)
	}
	
	// Set dimensions
	if weight.Valid {
		product.Attributes.Weight = weight.Float64
	}
	if length.Valid {
		product.Attributes.Dimensions.Length = length.Float64
	}
	if width.Valid {
		product.Attributes.Dimensions.Width = width.Float64
	}
	if height.Valid {
		product.Attributes.Dimensions.Height = height.Float64
	}
	
	return product, nil
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()
	
	attributesJSON, err := json.Marshal(product.Attributes)
	if err != nil {
		return utils.NewInternalError("failed to marshal product attributes", err)
	}
	
	query := `
		UPDATE products SET 
			sku = $2, name = $3, description = $4, category_id = $5, price = $6, 
			currency = $7, stock = $8, status = $9, weight = $10, length = $11, 
			width = $12, height = $13, images = $14, attributes = $15, 
			featured = $16, updated_at = $17
		WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		product.ID, product.SKU, product.Name, product.Description, product.CategoryID,
		product.Price, product.Currency, product.Stock, product.Status,
		product.Attributes.Weight, product.Attributes.Dimensions.Length,
		product.Attributes.Dimensions.Width, product.Attributes.Dimensions.Height,
		pq.Array(product.Images), attributesJSON, product.Featured,
		product.UpdatedAt,
	)
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return utils.NewConflictError("product with this SKU already exists")
		}
		return utils.NewInternalError("failed to update product", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.NewInternalError("failed to get rows affected", err)
	}
	
	if rowsAffected == 0 {
		return utils.NewNotFoundError("product")
	}
	
	return nil
}

// Delete deletes a product
func (r *productRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return utils.NewInternalError("failed to delete product", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.NewInternalError("failed to get rows affected", err)
	}
	
	if rowsAffected == 0 {
		return utils.NewNotFoundError("product")
	}
	
	return nil
}

// List retrieves products with filtering and pagination
func (r *productRepository) List(ctx context.Context, filter ProductFilter) ([]*models.Product, int, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	if filter.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, filter.CategoryID)
		argIndex++
	}
	
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}
	
	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}
	
	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}
	
	if filter.SearchTerm != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.SearchTerm+"%")
		argIndex++
	}
	
	if filter.Featured != nil {
		conditions = append(conditions, fmt.Sprintf("featured = $%d", argIndex))
		args = append(args, *filter.Featured)
		argIndex++
	}
	
	if filter.InStock != nil && *filter.InStock {
		conditions = append(conditions, "stock > 0")
	}
	
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, utils.NewInternalError("failed to count products", err)
	}
	
	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", filter.SortBy, direction)
	}
	
	// Build main query
	query := fmt.Sprintf(`
		SELECT id, sku, name, description, category_id, price, currency, stock, 
			   reserved_stock, status, weight, length, width, height, images, 
			   attributes, featured, created_at, updated_at
		FROM products %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIndex, argIndex+1)
	
	args = append(args, filter.Limit, filter.Offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, utils.NewInternalError("failed to list products", err)
	}
	defer rows.Close()
	
	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var attributesJSON []byte
		var weight, length, width, height sql.NullFloat64
		var reservedStock int
		
		err := rows.Scan(
			&product.ID, &product.SKU, &product.Name, &product.Description,
			&product.CategoryID, &product.Price, &product.Currency, &product.Stock,
			&reservedStock, &product.Status, &weight, &length, &width, &height,
			pq.Array(&product.Images), &attributesJSON, &product.Featured,
			&product.CreatedAt, &product.UpdatedAt,
		)
		
		if err != nil {
			return nil, 0, utils.NewInternalError("failed to scan product", err)
		}
		
		// Parse attributes
		if err := json.Unmarshal(attributesJSON, &product.Attributes); err != nil {
			return nil, 0, utils.NewInternalError("failed to unmarshal product attributes", err)
		}
		
		// Set dimensions
		if weight.Valid {
			product.Attributes.Weight = weight.Float64
		}
		if length.Valid {
			product.Attributes.Dimensions.Length = length.Float64
		}
		if width.Valid {
			product.Attributes.Dimensions.Width = width.Float64
		}
		if height.Valid {
			product.Attributes.Dimensions.Height = height.Float64
		}
		
		products = append(products, product)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, utils.NewInternalError("failed to iterate products", err)
	}
	
	return products, total, nil
}

// UpdateStock updates product stock
func (r *productRepository) UpdateStock(ctx context.Context, productID string, quantity int) error {
	return r.executeStockOperation(ctx, productID, quantity, "adjustment", "Manual stock update")
}

// ReserveStock reserves stock for a product
func (r *productRepository) ReserveStock(ctx context.Context, productID string, quantity int) error {
	return r.executeStockOperation(ctx, productID, quantity, "reserved", "Stock reserved for cart")
}

// ReleaseStock releases reserved stock
func (r *productRepository) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	return r.executeStockOperation(ctx, productID, quantity, "released", "Stock released from cart")
}

// GetAvailableStock returns available stock (total - reserved)
func (r *productRepository) GetAvailableStock(ctx context.Context, productID string) (int, error) {
	query := `SELECT stock - reserved_stock FROM products WHERE id = $1`
	
	var availableStock int
	err := r.db.QueryRowContext(ctx, query, productID).Scan(&availableStock)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, utils.NewNotFoundError("product")
		}
		return 0, utils.NewInternalError("failed to get available stock", err)
	}
	
	return availableStock, nil
}

// BulkUpdateStock updates stock for multiple products
func (r *productRepository) BulkUpdateStock(ctx context.Context, updates []StockUpdate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return utils.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()
	
	for _, update := range updates {
		err := r.executeStockOperationTx(ctx, tx, update.ProductID, update.Quantity, update.Type, update.Reason)
		if err != nil {
			return err
		}
	}
	
	if err := tx.Commit(); err != nil {
		return utils.NewInternalError("failed to commit transaction", err)
	}
	
	return nil
}

// executeStockOperation executes a stock operation and records the movement
func (r *productRepository) executeStockOperation(ctx context.Context, productID string, quantity int, movementType, reason string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return utils.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()
	
	err = r.executeStockOperationTx(ctx, tx, productID, quantity, movementType, reason)
	if err != nil {
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return utils.NewInternalError("failed to commit transaction", err)
	}
	
	return nil
}

// executeStockOperationTx executes a stock operation within a transaction
func (r *productRepository) executeStockOperationTx(ctx context.Context, tx *sql.Tx, productID string, quantity int, movementType, reason string) error {
	// Record inventory movement
	movementQuery := `
		INSERT INTO inventory_movements (id, product_id, movement_type, quantity, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := tx.ExecContext(ctx, movementQuery,
		uuid.New().String(), productID, movementType, quantity, reason, time.Now())
	
	if err != nil {
		return utils.NewInternalError("failed to record inventory movement", err)
	}
	
	return nil
}