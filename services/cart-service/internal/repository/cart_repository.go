package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
)

// CartRepository defines the interface for cart operations
type CartRepository interface {
	GetCart(ctx context.Context, userID, sessionID string) (*models.Cart, error)
	SaveCart(ctx context.Context, cart *models.Cart) error
	DeleteCart(ctx context.Context, cartID string) error
	GetCartByID(ctx context.Context, cartID string) (*models.Cart, error)
	UpdateCartExpiry(ctx context.Context, cartID string, expiresAt time.Time) error
	GetExpiredCarts(ctx context.Context) ([]*models.Cart, error)
	DeleteExpiredCarts(ctx context.Context) error
	MigrateGuestCartToUser(ctx context.Context, sessionID, userID string) error
}

// RedisCartRepository implements CartRepository using Redis
type RedisCartRepository struct {
	client *redis.Client
}

// NewCartRepository creates a new cart repository
func NewCartRepository(client *redis.Client) CartRepository {
	return &RedisCartRepository{
		client: client,
	}
}

// GetCart retrieves a cart by user ID or session ID
func (r *RedisCartRepository) GetCart(ctx context.Context, userID, sessionID string) (*models.Cart, error) {
	var key string
	if userID != "" {
		key = fmt.Sprintf("cart:user:%s", userID)
	} else {
		key = fmt.Sprintf("cart:session:%s", sessionID)
	}

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cart not found
		}
		return nil, fmt.Errorf("failed to get cart from Redis: %w", err)
	}

	var cart models.Cart
	if err := json.Unmarshal([]byte(data), &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart data: %w", err)
	}

	// Check if cart is expired
	if time.Now().After(cart.ExpiresAt) {
		// Delete expired cart
		r.client.Del(ctx, key)
		return nil, nil
	}

	return &cart, nil
}

// SaveCart saves a cart to Redis
func (r *RedisCartRepository) SaveCart(ctx context.Context, cart *models.Cart) error {
	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart data: %w", err)
	}

	// Determine the key based on user ID or session ID
	var key string
	if cart.UserID != "" {
		key = fmt.Sprintf("cart:user:%s", cart.UserID)
	} else {
		key = fmt.Sprintf("cart:session:%s", cart.SessionID)
	}

	// Set with expiration
	expiration := time.Until(cart.ExpiresAt)
	if expiration <= 0 {
		expiration = 24 * time.Hour // Default 24 hours
	}

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to save cart to Redis: %w", err)
	}

	// Also save by cart ID for direct access
	cartIDKey := fmt.Sprintf("cart:id:%s", cart.ID)
	if err := r.client.Set(ctx, cartIDKey, data, expiration).Err(); err != nil {
		utils.Logger.Error(ctx, "Failed to save cart by ID", err, map[string]interface{}{
			"cart_id": cart.ID,
		})
	}

	return nil
}

// DeleteCart deletes a cart from Redis
func (r *RedisCartRepository) DeleteCart(ctx context.Context, cartID string) error {
	// First get the cart to determine the user/session key
	cart, err := r.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil // Cart doesn't exist
	}

	// Delete by user/session key
	var key string
	if cart.UserID != "" {
		key = fmt.Sprintf("cart:user:%s", cart.UserID)
	} else {
		key = fmt.Sprintf("cart:session:%s", cart.SessionID)
	}

	// Delete both keys
	cartIDKey := fmt.Sprintf("cart:id:%s", cartID)
	
	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, cartIDKey)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete cart from Redis: %w", err)
	}

	return nil
}

// GetCartByID retrieves a cart by its ID
func (r *RedisCartRepository) GetCartByID(ctx context.Context, cartID string) (*models.Cart, error) {
	key := fmt.Sprintf("cart:id:%s", cartID)
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cart not found
		}
		return nil, fmt.Errorf("failed to get cart by ID from Redis: %w", err)
	}

	var cart models.Cart
	if err := json.Unmarshal([]byte(data), &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart data: %w", err)
	}

	// Check if cart is expired
	if time.Now().After(cart.ExpiresAt) {
		// Delete expired cart
		r.DeleteCart(ctx, cartID)
		return nil, nil
	}

	return &cart, nil
}

// UpdateCartExpiry updates the expiration time of a cart
func (r *RedisCartRepository) UpdateCartExpiry(ctx context.Context, cartID string, expiresAt time.Time) error {
	cart, err := r.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}
	if cart == nil {
		return fmt.Errorf("cart not found")
	}

	cart.ExpiresAt = expiresAt
	cart.UpdatedAt = time.Now()

	return r.SaveCart(ctx, cart)
}

// GetExpiredCarts retrieves all expired carts (for cleanup)
func (r *RedisCartRepository) GetExpiredCarts(ctx context.Context) ([]*models.Cart, error) {
	// This is a simplified implementation
	// In production, you might want to use a separate index for expired carts
	var expiredCarts []*models.Cart
	
	// Scan for all cart keys
	iter := r.client.Scan(ctx, 0, "cart:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue // Skip this key
		}

		var cart models.Cart
		if err := json.Unmarshal([]byte(data), &cart); err != nil {
			continue // Skip invalid data
		}

		if time.Now().After(cart.ExpiresAt) {
			expiredCarts = append(expiredCarts, &cart)
		}
	}

	return expiredCarts, iter.Err()
}

// DeleteExpiredCarts removes all expired carts
func (r *RedisCartRepository) DeleteExpiredCarts(ctx context.Context) error {
	expiredCarts, err := r.GetExpiredCarts(ctx)
	if err != nil {
		return err
	}

	for _, cart := range expiredCarts {
		if err := r.DeleteCart(ctx, cart.ID); err != nil {
			utils.Logger.Error(ctx, "Failed to delete expired cart", err, map[string]interface{}{
				"cart_id": cart.ID,
			})
		}
	}

	return nil
}

// MigrateGuestCartToUser migrates a guest cart to a user cart
func (r *RedisCartRepository) MigrateGuestCartToUser(ctx context.Context, sessionID, userID string) error {
	// Get the guest cart
	guestCart, err := r.GetCart(ctx, "", sessionID)
	if err != nil {
		return err
	}
	if guestCart == nil {
		return nil // No guest cart to migrate
	}

	// Check if user already has a cart
	userCart, err := r.GetCart(ctx, userID, "")
	if err != nil {
		return err
	}

	if userCart != nil {
		// Merge guest cart items into user cart
		for _, guestItem := range guestCart.Items {
			found := false
			for i, userItem := range userCart.Items {
				if userItem.ProductID == guestItem.ProductID && userItem.SKU == guestItem.SKU {
					// Update quantity and total
					userCart.Items[i].Quantity += guestItem.Quantity
					userCart.Items[i].Total = userItem.Price.Mul(userCart.Items[i].Total.Div(userItem.Price).Add(guestItem.Total.Div(guestItem.Price)))
					userCart.Items[i].UpdatedAt = time.Now()
					found = true
					break
				}
			}
			if !found {
				// Add new item
				userCart.Items = append(userCart.Items, guestItem)
			}
		}
		userCart.UpdatedAt = time.Now()
		userCart.CalculateSubtotal()
	} else {
		// Convert guest cart to user cart
		guestCart.UserID = userID
		guestCart.UpdatedAt = time.Now()
		userCart = guestCart
	}

	// Save the updated user cart
	if err := r.SaveCart(ctx, userCart); err != nil {
		return err
	}

	// Delete the guest cart
	guestKey := fmt.Sprintf("cart:session:%s", sessionID)
	if err := r.client.Del(ctx, guestKey).Err(); err != nil {
		utils.Logger.Error(ctx, "Failed to delete guest cart after migration", err, map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		})
	}

	return nil
}
