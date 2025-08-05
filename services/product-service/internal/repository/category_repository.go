package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

type categoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create creates a new category
func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	if category.ID == "" {
		category.ID = uuid.New().String()
	}
	
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	
	// Calculate path and level
	if category.ParentID != nil {
		parent, err := r.GetByID(ctx, *category.ParentID)
		if err != nil {
			return err
		}
		category.Path = parent.Path + "/" + category.ID
		category.Level = parent.Level + 1
	} else {
		category.Path = "/" + category.ID
		category.Level = 0
	}
	
	query := `
		INSERT INTO categories (id, name, description, parent_id, path, level, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description, category.ParentID,
		category.Path, category.Level, category.IsActive,
		category.CreatedAt, category.UpdatedAt,
	)
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return utils.NewConflictError("category with this name already exists")
		}
		return utils.NewInternalError("failed to create category", err)
	}
	
	return nil
}

// GetByID retrieves a category by ID
func (r *categoryRepository) GetByID(ctx context.Context, id string) (*models.Category, error) {
	query := `
		SELECT id, name, description, parent_id, path, level, is_active, created_at, updated_at
		FROM categories 
		WHERE id = $1`
	
	category := &models.Category{}
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.ParentID,
		&category.Path, &category.Level, &category.IsActive,
		&category.CreatedAt, &category.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("category")
		}
		return nil, utils.NewInternalError("failed to get category", err)
	}
	
	return category, nil
}

// Update updates a category
func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	category.UpdatedAt = time.Now()
	
	// If parent changed, recalculate path and level
	if category.ParentID != nil {
		parent, err := r.GetByID(ctx, *category.ParentID)
		if err != nil {
			return err
		}
		category.Path = parent.Path + "/" + category.ID
		category.Level = parent.Level + 1
	} else {
		category.Path = "/" + category.ID
		category.Level = 0
	}
	
	query := `
		UPDATE categories SET 
			name = $2, description = $3, parent_id = $4, path = $5, 
			level = $6, is_active = $7, updated_at = $8
		WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description, category.ParentID,
		category.Path, category.Level, category.IsActive, category.UpdatedAt,
	)
	
	if err != nil {
		return utils.NewInternalError("failed to update category", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.NewInternalError("failed to get rows affected", err)
	}
	
	if rowsAffected == 0 {
		return utils.NewNotFoundError("category")
	}
	
	// Update paths of all children if path changed
	if err := r.updateChildrenPaths(ctx, category); err != nil {
		return err
	}
	
	return nil
}

// Delete deletes a category
func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	// Check if category has children
	children, err := r.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	
	if len(children) > 0 {
		return utils.NewConflictError("cannot delete category with children")
	}
	
	// Check if category has products
	var productCount int
	countQuery := `SELECT COUNT(*) FROM products WHERE category_id = $1`
	err = r.db.QueryRowContext(ctx, countQuery, id).Scan(&productCount)
	if err != nil {
		return utils.NewInternalError("failed to count products in category", err)
	}
	
	if productCount > 0 {
		return utils.NewConflictError("cannot delete category with products")
	}
	
	query := `DELETE FROM categories WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return utils.NewInternalError("failed to delete category", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.NewInternalError("failed to get rows affected", err)
	}
	
	if rowsAffected == 0 {
		return utils.NewNotFoundError("category")
	}
	
	return nil
}

// List retrieves categories with filtering
func (r *categoryRepository) List(ctx context.Context, filter CategoryFilter) ([]*models.Category, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	if filter.ParentID != nil {
		if *filter.ParentID == "" {
			conditions = append(conditions, "parent_id IS NULL")
		} else {
			conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
			args = append(args, *filter.ParentID)
			argIndex++
		}
	}
	
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}
	
	if filter.Level != nil {
		conditions = append(conditions, fmt.Sprintf("level = $%d", argIndex))
		args = append(args, *filter.Level)
		argIndex++
	}
	
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	// Build query
	query := fmt.Sprintf(`
		SELECT id, name, description, parent_id, path, level, is_active, created_at, updated_at
		FROM categories %s
		ORDER BY level, name
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)
	
	args = append(args, filter.Limit, filter.Offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, utils.NewInternalError("failed to list categories", err)
	}
	defer rows.Close()
	
	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.ParentID,
			&category.Path, &category.Level, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt,
		)
		
		if err != nil {
			return nil, utils.NewInternalError("failed to scan category", err)
		}
		
		categories = append(categories, category)
	}
	
	if err := rows.Err(); err != nil {
		return nil, utils.NewInternalError("failed to iterate categories", err)
	}
	
	return categories, nil
}

// GetChildren retrieves all direct children of a category
func (r *categoryRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Category, error) {
	query := `
		SELECT id, name, description, parent_id, path, level, is_active, created_at, updated_at
		FROM categories 
		WHERE parent_id = $1
		ORDER BY name`
	
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, utils.NewInternalError("failed to get children categories", err)
	}
	defer rows.Close()
	
	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.ParentID,
			&category.Path, &category.Level, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt,
		)
		
		if err != nil {
			return nil, utils.NewInternalError("failed to scan category", err)
		}
		
		categories = append(categories, category)
	}
	
	if err := rows.Err(); err != nil {
		return nil, utils.NewInternalError("failed to iterate categories", err)
	}
	
	return categories, nil
}

// GetPath retrieves the full path of categories from root to the specified category
func (r *categoryRepository) GetPath(ctx context.Context, categoryID string) ([]*models.Category, error) {
	category, err := r.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	
	// Parse path to get all category IDs
	pathParts := strings.Split(strings.Trim(category.Path, "/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		return []*models.Category{category}, nil
	}
	
	// Build query to get all categories in path
	placeholders := make([]string, len(pathParts))
	args := make([]interface{}, len(pathParts))
	for i, part := range pathParts {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = part
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, description, parent_id, path, level, is_active, created_at, updated_at
		FROM categories 
		WHERE id IN (%s)
		ORDER BY level`,
		strings.Join(placeholders, ","))
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, utils.NewInternalError("failed to get category path", err)
	}
	defer rows.Close()
	
	var categories []*models.Category
	for rows.Next() {
		cat := &models.Category{}
		
		err := rows.Scan(
			&cat.ID, &cat.Name, &cat.Description, &cat.ParentID,
			&cat.Path, &cat.Level, &cat.IsActive,
			&cat.CreatedAt, &cat.UpdatedAt,
		)
		
		if err != nil {
			return nil, utils.NewInternalError("failed to scan category", err)
		}
		
		categories = append(categories, cat)
	}
	
	if err := rows.Err(); err != nil {
		return nil, utils.NewInternalError("failed to iterate categories", err)
	}
	
	return categories, nil
}

// updateChildrenPaths updates the paths of all children when a category's path changes
func (r *categoryRepository) updateChildrenPaths(ctx context.Context, category *models.Category) error {
	// Get all descendants
	query := `
		SELECT id, parent_id, level
		FROM categories 
		WHERE path LIKE $1 AND id != $2
		ORDER BY level`
	
	rows, err := r.db.QueryContext(ctx, query, category.Path+"/%", category.ID)
	if err != nil {
		return utils.NewInternalError("failed to get descendant categories", err)
	}
	defer rows.Close()
	
	type descendant struct {
		ID       string
		ParentID *string
		Level    int
	}
	
	var descendants []descendant
	for rows.Next() {
		var d descendant
		err := rows.Scan(&d.ID, &d.ParentID, &d.Level)
		if err != nil {
			return utils.NewInternalError("failed to scan descendant category", err)
		}
		descendants = append(descendants, d)
	}
	
	// Update paths level by level
	for _, desc := range descendants {
		var newPath string
		var newLevel int
		
		if desc.ParentID != nil {
			parent, err := r.GetByID(ctx, *desc.ParentID)
			if err != nil {
				return err
			}
			newPath = parent.Path + "/" + desc.ID
			newLevel = parent.Level + 1
		} else {
			newPath = "/" + desc.ID
			newLevel = 0
		}
		
		updateQuery := `UPDATE categories SET path = $1, level = $2 WHERE id = $3`
		_, err := r.db.ExecContext(ctx, updateQuery, newPath, newLevel, desc.ID)
		if err != nil {
			return utils.NewInternalError("failed to update descendant category path", err)
		}
	}
	
	return nil
}