package service

import (
	"context"

	"github.com/shopsphere/product-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// CategoryService handles category business logic
type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*models.Category, error) {
	// Validate request
	if err := s.validateCreateCategoryRequest(req); err != nil {
		return nil, err
	}
	
	// Validate parent exists if provided
	if req.ParentID != nil && *req.ParentID != "" {
		_, err := s.categoryRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, utils.NewValidationError("invalid parent_id")
		}
	}
	
	// Create category
	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    req.IsActive,
	}
	
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}
	
	return category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	if id == "" {
		return nil, utils.NewValidationError("category ID is required")
	}
	
	return s.categoryRepo.GetByID(ctx, id)
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(ctx context.Context, id string, req UpdateCategoryRequest) (*models.Category, error) {
	if id == "" {
		return nil, utils.NewValidationError("category ID is required")
	}
	
	// Get existing category
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Validate request
	if err := s.validateUpdateCategoryRequest(req); err != nil {
		return nil, err
	}
	
	// Validate parent exists if provided and prevent circular reference
	if req.ParentID != nil {
		if *req.ParentID != "" {
			parent, err := s.categoryRepo.GetByID(ctx, *req.ParentID)
			if err != nil {
				return nil, utils.NewValidationError("invalid parent_id")
			}
			
			// Check for circular reference
			if s.wouldCreateCircularReference(ctx, id, *req.ParentID) {
				return nil, utils.NewValidationError("cannot set parent: would create circular reference")
			}
			
			// Check depth limit (e.g., max 5 levels)
			if parent.Level >= 4 {
				return nil, utils.NewValidationError("maximum category depth exceeded")
			}
		}
		category.ParentID = req.ParentID
	}
	
	// Update fields
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}
	
	return category, nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	if id == "" {
		return utils.NewValidationError("category ID is required")
	}
	
	return s.categoryRepo.Delete(ctx, id)
}

// ListCategories retrieves categories with filtering
func (s *CategoryService) ListCategories(ctx context.Context, req ListCategoriesRequest) (*ListCategoriesResponse, error) {
	// Validate request
	if err := s.validateListCategoriesRequest(req); err != nil {
		return nil, err
	}
	
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 200 {
		req.Limit = 200
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	
	// Build filter
	filter := repository.CategoryFilter{
		ParentID: req.ParentID,
		IsActive: req.IsActive,
		Level:    req.Level,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}
	
	categories, err := s.categoryRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return &ListCategoriesResponse{
		Categories: categories,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}, nil
}

// GetCategoryChildren retrieves all direct children of a category
func (s *CategoryService) GetCategoryChildren(ctx context.Context, parentID string) ([]*models.Category, error) {
	if parentID == "" {
		return nil, utils.NewValidationError("parent ID is required")
	}
	
	// Verify parent exists
	_, err := s.categoryRepo.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	
	return s.categoryRepo.GetChildren(ctx, parentID)
}

// GetCategoryPath retrieves the full path from root to the specified category
func (s *CategoryService) GetCategoryPath(ctx context.Context, categoryID string) ([]*models.Category, error) {
	if categoryID == "" {
		return nil, utils.NewValidationError("category ID is required")
	}
	
	return s.categoryRepo.GetPath(ctx, categoryID)
}

// GetRootCategories retrieves all root categories (categories without parent)
func (s *CategoryService) GetRootCategories(ctx context.Context) ([]*models.Category, error) {
	emptyParent := ""
	filter := repository.CategoryFilter{
		ParentID: &emptyParent,
		IsActive: &[]bool{true}[0], // Only active categories
		Limit:    100,
		Offset:   0,
	}
	
	return s.categoryRepo.List(ctx, filter)
}

// GetCategoryTree retrieves the complete category tree structure
func (s *CategoryService) GetCategoryTree(ctx context.Context) ([]*CategoryTreeNode, error) {
	// Get all categories
	filter := repository.CategoryFilter{
		IsActive: &[]bool{true}[0],
		Limit:    1000, // Large limit to get all categories
		Offset:   0,
	}
	
	categories, err := s.categoryRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// Build tree structure
	return s.buildCategoryTree(categories), nil
}

// wouldCreateCircularReference checks if setting a parent would create a circular reference
func (s *CategoryService) wouldCreateCircularReference(ctx context.Context, categoryID, parentID string) bool {
	// Get the path of the potential parent
	path, err := s.categoryRepo.GetPath(ctx, parentID)
	if err != nil {
		return false // If we can't get the path, assume no circular reference
	}
	
	// Check if the category is already in the parent's path
	for _, cat := range path {
		if cat.ID == categoryID {
			return true
		}
	}
	
	return false
}

// buildCategoryTree builds a hierarchical tree structure from flat category list
func (s *CategoryService) buildCategoryTree(categories []*models.Category) []*CategoryTreeNode {
	// Create a map for quick lookup
	categoryMap := make(map[string]*models.Category)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}
	
	// Create tree nodes
	nodeMap := make(map[string]*CategoryTreeNode)
	var rootNodes []*CategoryTreeNode
	
	for _, cat := range categories {
		node := &CategoryTreeNode{
			Category: cat,
			Children: []*CategoryTreeNode{},
		}
		nodeMap[cat.ID] = node
		
		if cat.ParentID == nil {
			rootNodes = append(rootNodes, node)
		}
	}
	
	// Build parent-child relationships
	for _, cat := range categories {
		if cat.ParentID != nil {
			if parentNode, exists := nodeMap[*cat.ParentID]; exists {
				if childNode, exists := nodeMap[cat.ID]; exists {
					parentNode.Children = append(parentNode.Children, childNode)
				}
			}
		}
	}
	
	return rootNodes
}

// validateCreateCategoryRequest validates create category request
func (s *CategoryService) validateCreateCategoryRequest(req CreateCategoryRequest) error {
	v := utils.NewValidator()
	
	v.Required("name", req.Name).MaxLength("name", req.Name, 255)
	v.MaxLength("description", req.Description, 1000)
	
	if v.HasErrors() {
		return utils.NewValidationError(v.Errors().Error())
	}
	
	return nil
}

// validateUpdateCategoryRequest validates update category request
func (s *CategoryService) validateUpdateCategoryRequest(req UpdateCategoryRequest) error {
	v := utils.NewValidator()
	
	if req.Name != nil {
		v.Required("name", *req.Name).MaxLength("name", *req.Name, 255)
	}
	
	if req.Description != nil {
		v.MaxLength("description", *req.Description, 1000)
	}
	
	if v.HasErrors() {
		return utils.NewValidationError(v.Errors().Error())
	}
	
	return nil
}

// validateListCategoriesRequest validates list categories request
func (s *CategoryService) validateListCategoriesRequest(req ListCategoriesRequest) error {
	if req.Limit < 0 {
		return utils.NewValidationError("limit must be non-negative")
	}
	
	if req.Offset < 0 {
		return utils.NewValidationError("offset must be non-negative")
	}
	
	return nil
}