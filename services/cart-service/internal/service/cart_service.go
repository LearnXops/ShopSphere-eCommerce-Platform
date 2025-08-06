package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopsphere/cart-service/internal/repository"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopspring/decimal"
)

// CartService defines the interface for cart business logic
type CartService interface {
	GetCart(ctx context.Context, userID, sessionID string) (*models.Cart, error)
	AddItem(ctx context.Context, userID, sessionID, productID, sku, name string, price decimal.Decimal, quantity int) (*models.Cart, error)
	UpdateItem(ctx context.Context, userID, sessionID, productID string, quantity int) (*models.Cart, error)
	RemoveItem(ctx context.Context, userID, sessionID, productID string) (*models.Cart, error)
	ClearCart(ctx context.Context, userID, sessionID string) error
	MigrateGuestCart(ctx context.Context, sessionID, userID string) (*models.Cart, error)
	ValidateCart(ctx context.Context, cart *models.Cart) (*CartValidationResult, error)
	ExtendCartExpiry(ctx context.Context, userID, sessionID string, duration time.Duration) (*models.Cart, error)
	CleanupExpiredCarts(ctx context.Context) error
}

// CartValidationResult represents the result of cart validation
type CartValidationResult struct {
	IsValid          bool                    `json:"is_valid"`
	InvalidItems     []CartValidationItem    `json:"invalid_items,omitempty"`
	PriceChanges     []CartPriceChange       `json:"price_changes,omitempty"`
	UnavailableItems []CartValidationItem    `json:"unavailable_items,omitempty"`
	TotalAmount      decimal.Decimal         `json:"total_amount"`
}

// CartValidationItem represents an invalid cart item
type CartValidationItem struct {
	ProductID string `json:"product_id"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
}

// CartPriceChange represents a price change in cart items
type CartPriceChange struct {
	ProductID string          `json:"product_id"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	OldPrice  decimal.Decimal `json:"old_price"`
	NewPrice  decimal.Decimal `json:"new_price"`
}

// cartService implements CartService
type cartService struct {
	cartRepo        repository.CartRepository
	productService  ProductService // Interface to product service for validation
}

// ProductService interface for product validation
type ProductService interface {
	GetProduct(ctx context.Context, productID string) (*ProductInfo, error)
	ValidateStock(ctx context.Context, productID string, quantity int) (bool, error)
}

// ProductInfo represents basic product information for validation
type ProductInfo struct {
	ID          string          `json:"id"`
	SKU         string          `json:"sku"`
	Name        string          `json:"name"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	IsAvailable bool            `json:"is_available"`
}

// NewCartService creates a new cart service
func NewCartService(cartRepo repository.CartRepository, productService ProductService) CartService {
	return &cartService{
		cartRepo:       cartRepo,
		productService: productService,
	}
}

// GetCart retrieves or creates a cart for the user/session
func (s *cartService) GetCart(ctx context.Context, userID, sessionID string) (*models.Cart, error) {
	cart, err := s.cartRepo.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if cart == nil {
		// Create new cart
		cart = models.NewCart(userID, sessionID)
		if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
			return nil, fmt.Errorf("failed to create new cart: %w", err)
		}
		
		utils.Logger.Info(ctx, "Created new cart", map[string]interface{}{
			"cart_id":    cart.ID,
			"user_id":    userID,
			"session_id": sessionID,
		})
	}

	return cart, nil
}

// AddItem adds an item to the cart
func (s *cartService) AddItem(ctx context.Context, userID, sessionID, productID, sku, name string, price decimal.Decimal, quantity int) (*models.Cart, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}

	// Validate product if product service is available
	if s.productService != nil {
		if available, err := s.productService.ValidateStock(ctx, productID, quantity); err != nil {
			utils.Logger.Error(ctx, "Failed to validate product stock", err, map[string]interface{}{
				"product_id": productID,
				"quantity":   quantity,
			})
		} else if !available {
			return nil, fmt.Errorf("insufficient stock for product %s", productID)
		}
	}

	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if item already exists in cart
	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID && item.SKU == sku {
			// Update existing item
			cart.Items[i].Quantity += quantity
			cart.Items[i].Total = item.Price.Mul(decimal.NewFromInt(int64(cart.Items[i].Quantity)))
			cart.Items[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		// Add new item
		cart.AddItem(productID, sku, name, price, quantity)
	}

	cart.UpdatedAt = time.Now()
	cart.CalculateSubtotal()

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	utils.Logger.Info(ctx, "Added item to cart", map[string]interface{}{
		"cart_id":    cart.ID,
		"product_id": productID,
		"quantity":   quantity,
		"user_id":    userID,
		"session_id": sessionID,
	})

	return cart, nil
}

// UpdateItem updates the quantity of an item in the cart
func (s *cartService) UpdateItem(ctx context.Context, userID, sessionID, productID string, quantity int) (*models.Cart, error) {
	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if quantity > 0 {
		// Validate stock if product service is available
		if s.productService != nil {
			if available, err := s.productService.ValidateStock(ctx, productID, quantity); err != nil {
				utils.Logger.Error(ctx, "Failed to validate product stock", err, map[string]interface{}{
					"product_id": productID,
					"quantity":   quantity,
				})
			} else if !available {
				return nil, fmt.Errorf("insufficient stock for product %s", productID)
			}
		}
	}

	if !cart.UpdateItem(productID, quantity) {
		return nil, fmt.Errorf("item not found in cart")
	}

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	utils.Logger.Info(ctx, "Updated cart item", map[string]interface{}{
		"cart_id":    cart.ID,
		"product_id": productID,
		"quantity":   quantity,
		"user_id":    userID,
		"session_id": sessionID,
	})

	return cart, nil
}

// RemoveItem removes an item from the cart
func (s *cartService) RemoveItem(ctx context.Context, userID, sessionID, productID string) (*models.Cart, error) {
	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if !cart.RemoveItem(productID) {
		return nil, fmt.Errorf("item not found in cart")
	}

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	utils.Logger.Info(ctx, "Removed item from cart", map[string]interface{}{
		"cart_id":    cart.ID,
		"product_id": productID,
		"user_id":    userID,
		"session_id": sessionID,
	})

	return cart, nil
}

// ClearCart removes all items from the cart
func (s *cartService) ClearCart(ctx context.Context, userID, sessionID string) error {
	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return err
	}

	cart.Items = []models.CartItem{}
	cart.Subtotal = decimal.Zero
	cart.UpdatedAt = time.Now()

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	utils.Logger.Info(ctx, "Cleared cart", map[string]interface{}{
		"cart_id":    cart.ID,
		"user_id":    userID,
		"session_id": sessionID,
	})

	return nil
}

// MigrateGuestCart migrates a guest cart to a user cart
func (s *cartService) MigrateGuestCart(ctx context.Context, sessionID, userID string) (*models.Cart, error) {
	if err := s.cartRepo.MigrateGuestCartToUser(ctx, sessionID, userID); err != nil {
		return nil, fmt.Errorf("failed to migrate guest cart: %w", err)
	}

	// Get the migrated cart
	cart, err := s.GetCart(ctx, userID, "")
	if err != nil {
		return nil, err
	}

	utils.Logger.Info(ctx, "Migrated guest cart to user", map[string]interface{}{
		"cart_id":    cart.ID,
		"user_id":    userID,
		"session_id": sessionID,
	})

	return cart, nil
}

// ValidateCart validates all items in the cart against current product data
func (s *cartService) ValidateCart(ctx context.Context, cart *models.Cart) (*CartValidationResult, error) {
	result := &CartValidationResult{
		IsValid:          true,
		InvalidItems:     []CartValidationItem{},
		PriceChanges:     []CartPriceChange{},
		UnavailableItems: []CartValidationItem{},
		TotalAmount:      decimal.Zero,
	}

	if s.productService == nil {
		// If no product service available, assume cart is valid
		result.TotalAmount = cart.Subtotal
		return result, nil
	}

	for _, item := range cart.Items {
		productInfo, err := s.productService.GetProduct(ctx, item.ProductID)
		if err != nil {
			result.IsValid = false
			result.InvalidItems = append(result.InvalidItems, CartValidationItem{
				ProductID: item.ProductID,
				SKU:       item.SKU,
				Name:      item.Name,
				Reason:    "Product not found",
			})
			continue
		}

		// Check availability
		if !productInfo.IsAvailable {
			result.IsValid = false
			result.UnavailableItems = append(result.UnavailableItems, CartValidationItem{
				ProductID: item.ProductID,
				SKU:       item.SKU,
				Name:      item.Name,
				Reason:    "Product no longer available",
			})
			continue
		}

		// Check stock
		if available, err := s.productService.ValidateStock(ctx, item.ProductID, item.Quantity); err != nil || !available {
			result.IsValid = false
			result.UnavailableItems = append(result.UnavailableItems, CartValidationItem{
				ProductID: item.ProductID,
				SKU:       item.SKU,
				Name:      item.Name,
				Reason:    "Insufficient stock",
			})
			continue
		}

		// Check price changes
		if !item.Price.Equal(productInfo.Price) {
			result.PriceChanges = append(result.PriceChanges, CartPriceChange{
				ProductID: item.ProductID,
				SKU:       item.SKU,
				Name:      item.Name,
				OldPrice:  item.Price,
				NewPrice:  productInfo.Price,
			})
			// Update total with new price
			result.TotalAmount = result.TotalAmount.Add(productInfo.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
		} else {
			result.TotalAmount = result.TotalAmount.Add(item.Total)
		}
	}

	return result, nil
}

// ExtendCartExpiry extends the expiry time of a cart
func (s *cartService) ExtendCartExpiry(ctx context.Context, userID, sessionID string, duration time.Duration) (*models.Cart, error) {
	cart, err := s.GetCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	cart.ExtendExpiry(duration)

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to extend cart expiry: %w", err)
	}

	utils.Logger.Info(ctx, "Extended cart expiry", map[string]interface{}{
		"cart_id":    cart.ID,
		"user_id":    userID,
		"session_id": sessionID,
		"expires_at": cart.ExpiresAt,
	})

	return cart, nil
}

// CleanupExpiredCarts removes all expired carts
func (s *cartService) CleanupExpiredCarts(ctx context.Context) error {
	if err := s.cartRepo.DeleteExpiredCarts(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired carts: %w", err)
	}

	utils.Logger.Info(ctx, "Cleaned up expired carts", nil)
	return nil
}
