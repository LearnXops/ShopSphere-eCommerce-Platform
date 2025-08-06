package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/shopsphere/cart-service/internal/handlers"
	"github.com/shopsphere/cart-service/internal/repository"
	"github.com/shopsphere/cart-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

func main() {
	ctx := context.Background()
	
	// Initialize logger
	utils.Logger.Info(ctx, "Starting Cart Service...", nil)

	// Initialize Redis connection
	redisConfig := utils.NewRedisConfig()
	redisClient, err := redisConfig.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	utils.Logger.Info(ctx, "Connected to Redis successfully", map[string]interface{}{
		"host": redisConfig.Host,
		"port": redisConfig.Port,
		"db":   redisConfig.DB,
	})

	// Initialize repositories
	cartRepo := repository.NewCartRepository(redisClient)

	// Initialize services
	// Note: ProductService is nil for now - can be integrated later for validation
	cartService := service.NewCartService(cartRepo, nil)

	// Initialize handlers
	cartHandler := handlers.NewCartHandler(cartService)

	// Create router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "cart-service"}`))
	}).Methods("GET")

	// Cart routes
	cartRoutes := router.PathPrefix("/cart").Subrouter()
	cartRoutes.HandleFunc("", cartHandler.GetCart).Methods("GET")
	cartRoutes.HandleFunc("/items", cartHandler.AddItem).Methods("POST")
	cartRoutes.HandleFunc("/items/{productId}", cartHandler.UpdateItem).Methods("PUT")
	cartRoutes.HandleFunc("/items/{productId}", cartHandler.RemoveItem).Methods("DELETE")
	cartRoutes.HandleFunc("/clear", cartHandler.ClearCart).Methods("POST")
	cartRoutes.HandleFunc("/migrate", cartHandler.MigrateGuestCart).Methods("POST")
	cartRoutes.HandleFunc("/validate", cartHandler.ValidateCart).Methods("GET")
	cartRoutes.HandleFunc("/extend-expiry", cartHandler.ExtendExpiry).Methods("POST")
	cartRoutes.HandleFunc("/summary", cartHandler.GetCartSummary).Methods("GET")

	// Admin routes
	adminRoutes := router.PathPrefix("/admin").Subrouter()
	adminRoutes.HandleFunc("/cleanup-expired", cartHandler.CleanupExpiredCarts).Methods("POST")

	// Start cleanup routine for expired carts
	go startCleanupRoutine(ctx, cartService)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	utils.Logger.Info(ctx, "Cart Service listening on port", map[string]interface{}{"port": port})
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// startCleanupRoutine starts a background routine to cleanup expired carts
func startCleanupRoutine(ctx context.Context, cartService service.CartService) {
	ticker := time.NewTicker(1 * time.Hour) // Run cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := cartService.CleanupExpiredCarts(ctx); err != nil {
				utils.Logger.Error(ctx, "Failed to cleanup expired carts", err, nil)
			} else {
				utils.Logger.Info(ctx, "Successfully cleaned up expired carts", nil)
			}
		case <-ctx.Done():
			return
		}
	}
}