package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/notification-service/internal/handlers"
	"github.com/shopsphere/notification-service/internal/repository"
	"github.com/shopsphere/notification-service/internal/service"
	"github.com/shopsphere/notification-service/internal/gateway"
)

func main() {
	logger := utils.NewStructuredLogger(os.Stdout, utils.LogLevelInfo, "notification-service")

	// Database configuration
	dbConfig := &utils.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "shopsphere"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to database
	db, err := dbConfig.Connect()
	if err != nil {
		logger.Error(context.Background(), "Failed to connect to database", err)
		log.Fatal("Database connection failed")
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewPostgresNotificationRepository(db, logger)

	// Initialize notification providers using factory
	providerFactory := gateway.NewProviderFactory(logger)
	
	emailProvider, err := providerFactory.CreateEmailProvider()
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize email provider", err)
		log.Fatal("Email provider initialization failed")
	}
	
	smsProvider := providerFactory.CreateSMSProvider()
	pushProvider := providerFactory.CreatePushProvider()

	// Initialize notification gateway
	notificationGateway := gateway.NewNotificationGateway(emailProvider, smsProvider, pushProvider)

	// Initialize service
	notificationService := service.NewNotificationService(repo, notificationGateway, logger)

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(notificationService, logger)

	// Setup router
	router := mux.NewRouter()

	// Register routes
	notificationHandler.RegisterRoutes(router)

	// Add basic CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-User-Role")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	port := getEnv("PORT", "8080")
	logger.Info(context.Background(), "Notification service starting", map[string]interface{}{
		"port": port,
	})

	log.Fatal(http.ListenAndServe(":"+port, router))
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}