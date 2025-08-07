package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/models"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/shipping-service/internal/gateway"
	"github.com/shopsphere/shipping-service/internal/handlers"
	"github.com/shopsphere/shipping-service/internal/repository"
	"github.com/shopsphere/shipping-service/internal/service"
)

func main() {
	// Initialize logger
	logger := utils.NewStructuredLogger(os.Stdout, utils.LogLevelInfo, "shipping-service")
	logger.Info(context.Background(), "Starting Shipping Service", nil)

	// Initialize database connection
	dbConfig := &utils.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "shopsphere"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := dbConfig.Connect()
	if err != nil {
		logger.Error(context.Background(), "Failed to connect to database", err)
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewPostgresShippingRepository(db)

	// Initialize carrier gateway factory
	gatewayFactory := gateway.NewCarrierGatewayFactory()

	// Register mock gateways for development
	gatewayFactory.RegisterGateway(models.CarrierFedEx, gateway.NewMockCarrierGateway(models.CarrierFedEx))
	gatewayFactory.RegisterGateway(models.CarrierUPS, gateway.NewMockCarrierGateway(models.CarrierUPS))
	gatewayFactory.RegisterGateway(models.CarrierUSPS, gateway.NewMockCarrierGateway(models.CarrierUSPS))

	// Initialize service
	shippingService := service.NewShippingService(repo, gatewayFactory, logger)

	// Initialize handlers
	handler := handlers.NewShippingHandler(shippingService, logger)

	// Setup router
	router := mux.NewRouter()

	// Add basic middleware (simplified for now)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Register routes
	handler.RegisterRoutes(router)

	// Start server
	port := getEnv("PORT", "8007")
	logger.Info(context.Background(), "Shipping Service listening", map[string]interface{}{
		"port": port,
	})

	log.Fatal(http.ListenAndServe(":"+port, router))
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt gets environment variable as int with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}