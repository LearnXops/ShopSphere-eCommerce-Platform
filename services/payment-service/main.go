package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/payment-service/internal/gateway"
	"github.com/shopsphere/payment-service/internal/handlers"
	"github.com/shopsphere/payment-service/internal/repository"
	"github.com/shopsphere/payment-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

func main() {
	// Initialize logger
	logger := utils.NewStructuredLogger(os.Stdout, utils.LogLevelInfo, "payment-service")
	utils.Logger = logger

	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8006"
	}

	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeSecretKey == "" {
		log.Fatal("STRIPE_SECRET_KEY environment variable is required")
	}

	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if stripeWebhookSecret == "" {
		log.Fatal("STRIPE_WEBHOOK_SECRET environment variable is required")
	}

	// Initialize database connection
	dbConfig := &utils.DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
		User:     getEnvOrDefault("DB_USER", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("DB_NAME", "shopsphere_payments"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	db, err := dbConfig.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Printf("Connected to database: %s", dbConfig.DBName)

	// Initialize components
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentGateway := gateway.NewStripeGateway(stripeSecretKey, stripeWebhookSecret)
	paymentService := service.NewPaymentService(paymentRepo, paymentGateway)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	// Create router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		
		// Check database connection
		if err := db.PingContext(ctx); err != nil {
			logger.Error(ctx, "Database health check failed", err)
			utils.WriteErrorResponse(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection failed")
			return
		}
		
		utils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
			"status":  "healthy",
			"service": "payment-service",
			"version": "1.0.0",
		})
	}).Methods("GET")

	// Register payment routes
	paymentHandler.RegisterRoutes(router)

	// Add logging middleware
	router.Use(utils.LogMiddleware("payment-service"))

	log.Printf("Payment service starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health - Health check")
	log.Printf("  POST /payments - Create payment")
	log.Printf("  GET  /payments/{id} - Get payment")
	log.Printf("  POST /payments/{id}/process - Process payment")
	log.Printf("  POST /payments/{id}/cancel - Cancel payment")
	log.Printf("  POST /payments/{id}/retry - Retry payment")
	log.Printf("  POST /payment-methods - Create payment method")
	log.Printf("  POST /refunds - Create refund")
	log.Printf("  POST /webhooks/stripe - Stripe webhook")

	log.Fatal(http.ListenAndServe(":"+port, router))
}

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}