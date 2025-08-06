package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/shopsphere/cart-service/internal/repository"
	"github.com/shopsphere/cart-service/internal/service"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopspring/decimal"
)

func TestMain(m *testing.M) {
	// Skip integration tests if Redis is not available
	if err := utils.CheckRedisConnection(); err != nil {
		os.Exit(0) // Skip tests
	}
	
	os.Exit(m.Run())
}

func setupRedisTest(t *testing.T) (repository.CartRepository, func()) {
	config := utils.NewRedisConfig()
	config.DB = 15 // Use test database
	
	client, err := config.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	
	repo := repository.NewCartRepository(client)
	
	cleanup := func() {
		// Clean up test data
		ctx := context.Background()
		client.FlushDB(ctx)
		client.Close()
	}
	
	return repo, cleanup
}

func TestCartRepository_Integration(t *testing.T) {
	repo, cleanup := setupRedisTest(t)
	defer cleanup()
	
	ctx := context.Background()
	
	t.Run("SaveAndGetCart", func(t *testing.T) {
		cart := models.NewCart("user1", "")
		cart.AddItem("prod1", "SKU1", "Product 1", decimal.NewFromFloat(19.99), 2)
		
		// Save cart
		err := repo.SaveCart(ctx, cart)
		if err != nil {
			t.Fatalf("Failed to save cart: %v", err)
		}
		
		// Get cart
		retrievedCart, err := repo.GetCart(ctx, "user1", "")
		if err != nil {
			t.Fatalf("Failed to get cart: %v", err)
		}
		
		if retrievedCart == nil {
			t.Fatal("Expected cart to be found")
		}
		
		if retrievedCart.ID != cart.ID {
			t.Errorf("Expected cart ID %s, got %s", cart.ID, retrievedCart.ID)
		}
		
		if len(retrievedCart.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(retrievedCart.Items))
		}
		
		if retrievedCart.Items[0].ProductID != "prod1" {
			t.Errorf("Expected ProductID 'prod1', got %s", retrievedCart.Items[0].ProductID)
		}
	})
	
	t.Run("GetCartByID", func(t *testing.T) {
		cart := models.NewCart("user2", "")
		cart.AddItem("prod2", "SKU2", "Product 2", decimal.NewFromFloat(29.99), 1)
		
		err := repo.SaveCart(ctx, cart)
		if err != nil {
			t.Fatalf("Failed to save cart: %v", err)
		}
		
		retrievedCart, err := repo.GetCartByID(ctx, cart.ID)
		if err != nil {
			t.Fatalf("Failed to get cart by ID: %v", err)
		}
		
		if retrievedCart == nil {
			t.Fatal("Expected cart to be found")
		}
		
		if retrievedCart.ID != cart.ID {
			t.Errorf("Expected cart ID %s, got %s", cart.ID, retrievedCart.ID)
		}
	})
	
	t.Run("DeleteCart", func(t *testing.T) {
		cart := models.NewCart("user3", "")
		err := repo.SaveCart(ctx, cart)
		if err != nil {
			t.Fatalf("Failed to save cart: %v", err)
		}
		
		// Delete cart
		err = repo.DeleteCart(ctx, cart.ID)
		if err != nil {
			t.Fatalf("Failed to delete cart: %v", err)
		}
		
		// Verify cart is deleted
		retrievedCart, err := repo.GetCartByID(ctx, cart.ID)
		if err != nil {
			t.Fatalf("Failed to get cart: %v", err)
		}
		
		if retrievedCart != nil {
			t.Error("Expected cart to be deleted")
		}
	})
	
	t.Run("ExpiredCartHandling", func(t *testing.T) {
		cart := models.NewCart("user4", "")
		cart.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
		
		err := repo.SaveCart(ctx, cart)
		if err != nil {
			t.Fatalf("Failed to save cart: %v", err)
		}
		
		// Try to get expired cart
		retrievedCart, err := repo.GetCart(ctx, "user4", "")
		if err != nil {
			t.Fatalf("Failed to get cart: %v", err)
		}
		
		if retrievedCart != nil {
			t.Error("Expected expired cart to be nil")
		}
	})
	
	t.Run("MigrateGuestCartToUser", func(t *testing.T) {
		// Create guest cart
		guestCart := models.NewCart("", "session123")
		guestCart.AddItem("prod5", "SKU5", "Product 5", decimal.NewFromFloat(39.99), 1)
		
		err := repo.SaveCart(ctx, guestCart)
		if err != nil {
			t.Fatalf("Failed to save guest cart: %v", err)
		}
		
		// Migrate to user
		err = repo.MigrateGuestCartToUser(ctx, "session123", "user5")
		if err != nil {
			t.Fatalf("Failed to migrate cart: %v", err)
		}
		
		// Verify user cart exists
		userCart, err := repo.GetCart(ctx, "user5", "")
		if err != nil {
			t.Fatalf("Failed to get user cart: %v", err)
		}
		
		if userCart == nil {
			t.Fatal("Expected user cart to exist")
		}
		
		if len(userCart.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(userCart.Items))
		}
		
		// Verify guest cart is gone
		guestCart, err = repo.GetCart(ctx, "", "session123")
		if err != nil {
			t.Fatalf("Failed to get guest cart: %v", err)
		}
		
		if guestCart != nil && len(guestCart.Items) > 0 {
			t.Error("Expected guest cart to be empty or nil")
		}
	})
}

func TestCartService_Integration(t *testing.T) {
	repo, cleanup := setupRedisTest(t)
	defer cleanup()
	
	ctx := context.Background()
	cartService := service.NewCartService(repo, nil)
	
	t.Run("CompleteCartWorkflow", func(t *testing.T) {
		userID := "integration_user"
		price1 := decimal.NewFromFloat(19.99)
		price2 := decimal.NewFromFloat(29.99)
		
		// Add first item
		cart, err := cartService.AddItem(ctx, userID, "", "prod1", "SKU1", "Product 1", price1, 2)
		if err != nil {
			t.Fatalf("Failed to add first item: %v", err)
		}
		
		if len(cart.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(cart.Items))
		}
		
		// Add second item
		cart, err = cartService.AddItem(ctx, userID, "", "prod2", "SKU2", "Product 2", price2, 1)
		if err != nil {
			t.Fatalf("Failed to add second item: %v", err)
		}
		
		if len(cart.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(cart.Items))
		}
		
		// Verify subtotal
		expectedSubtotal := price1.Mul(decimal.NewFromInt(2)).Add(price2)
		if !cart.Subtotal.Equal(expectedSubtotal) {
			t.Errorf("Expected subtotal %s, got %s", expectedSubtotal.String(), cart.Subtotal.String())
		}
		
		// Update item quantity
		cart, err = cartService.UpdateItem(ctx, userID, "", "prod1", 3)
		if err != nil {
			t.Fatalf("Failed to update item: %v", err)
		}
		
		// Find the updated item
		var updatedItem *models.CartItem
		for _, item := range cart.Items {
			if item.ProductID == "prod1" {
				updatedItem = &item
				break
			}
		}
		
		if updatedItem == nil {
			t.Fatal("Expected to find updated item")
		}
		
		if updatedItem.Quantity != 3 {
			t.Errorf("Expected quantity 3, got %d", updatedItem.Quantity)
		}
		
		// Remove item
		cart, err = cartService.RemoveItem(ctx, userID, "", "prod2")
		if err != nil {
			t.Fatalf("Failed to remove item: %v", err)
		}
		
		if len(cart.Items) != 1 {
			t.Errorf("Expected 1 item after removal, got %d", len(cart.Items))
		}
		
		// Clear cart
		err = cartService.ClearCart(ctx, userID, "")
		if err != nil {
			t.Fatalf("Failed to clear cart: %v", err)
		}
		
		// Verify cart is empty
		cart, err = cartService.GetCart(ctx, userID, "")
		if err != nil {
			t.Fatalf("Failed to get cart: %v", err)
		}
		
		if len(cart.Items) != 0 {
			t.Errorf("Expected 0 items after clear, got %d", len(cart.Items))
		}
	})
	
	t.Run("CartExpiry", func(t *testing.T) {
		userID := "expiry_user"
		
		// Create cart
		cart, err := cartService.GetCart(ctx, userID, "")
		if err != nil {
			t.Fatalf("Failed to get cart: %v", err)
		}
		
		originalExpiry := cart.ExpiresAt
		
		// Extend expiry
		cart, err = cartService.ExtendCartExpiry(ctx, userID, "", 2*time.Hour)
		if err != nil {
			t.Fatalf("Failed to extend cart expiry: %v", err)
		}
		
		if !cart.ExpiresAt.After(originalExpiry) {
			t.Error("Expected expiry to be extended")
		}
		
		// Verify the extension persisted
		retrievedCart, err := cartService.GetCart(ctx, userID, "")
		if err != nil {
			t.Fatalf("Failed to get cart: %v", err)
		}
		
		if !retrievedCart.ExpiresAt.Equal(cart.ExpiresAt) {
			t.Error("Expected expiry extension to persist")
		}
	})
}
