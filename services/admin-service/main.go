package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/shopsphere/admin-service/internal/handlers"
	"github.com/shopsphere/admin-service/internal/repository"
	"github.com/shopsphere/admin-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

func main() {
	ctx := context.Background()
	
	// Initialize logger
	logger := utils.NewStructuredLogger(os.Stdout, "info", "admin-service")
	logger.Info(ctx, "Starting Admin Service...")

	// Initialize database connection
	dbConfig := utils.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "shopsphere"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := dbConfig.Connect()
	if err != nil {
		logger.Error(ctx, "Failed to connect to database", err)
		log.Fatal("Database connection failed")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Error(ctx, "Failed to ping database", err)
		log.Fatal("Database connection failed")
	}

	// Initialize repository
	adminRepo := repository.NewPostgresAdminRepository(db, logger)

	// Initialize service
	adminService := service.NewAdminService(adminRepo, logger)

	// Initialize handlers
	adminHandler := handlers.NewAdminHandler(adminService, logger)

	// Setup routes
	router := mux.NewRouter()

	// Admin User Management
	router.HandleFunc("/admin/users", adminHandler.CreateAdminUser).Methods("POST")
	router.HandleFunc("/admin/users/{id}", adminHandler.GetAdminUser).Methods("GET")
	router.HandleFunc("/admin/users/{id}", adminHandler.UpdateAdminUser).Methods("PUT")
	router.HandleFunc("/admin/users/{id}", adminHandler.DeleteAdminUser).Methods("DELETE")
	router.HandleFunc("/admin/users", adminHandler.ListAdminUsers).Methods("GET")

	// Activity Logs
	router.HandleFunc("/admin/activity-logs", adminHandler.GetActivityLogs).Methods("GET")

	// Dashboard and Analytics
	router.HandleFunc("/admin/dashboard/metrics", adminHandler.GetDashboardMetrics).Methods("GET")
	router.HandleFunc("/admin/analytics/revenue", adminHandler.GetRevenueData).Methods("GET")
	router.HandleFunc("/admin/analytics/orders/distribution", adminHandler.GetOrderStatusDistribution).Methods("GET")
	router.HandleFunc("/admin/analytics/products/top", adminHandler.GetTopProducts).Methods("GET")
	router.HandleFunc("/admin/analytics/users/growth", adminHandler.GetUserGrowthData).Methods("GET")

	// System Alerts
	router.HandleFunc("/admin/alerts", adminHandler.CreateSystemAlert).Methods("POST")
	router.HandleFunc("/admin/alerts/{id}", adminHandler.GetSystemAlert).Methods("GET")
	router.HandleFunc("/admin/alerts/{id}/resolve", adminHandler.ResolveSystemAlert).Methods("POST")
	router.HandleFunc("/admin/alerts/{id}", adminHandler.DeleteSystemAlert).Methods("DELETE")
	router.HandleFunc("/admin/alerts", adminHandler.ListSystemAlerts).Methods("GET")

	// Dashboard Configs
	router.HandleFunc("/admin/dashboard/configs", adminHandler.CreateDashboardConfig).Methods("POST")
	router.HandleFunc("/admin/dashboard/configs/{name}", adminHandler.GetDashboardConfig).Methods("GET")
	router.HandleFunc("/admin/dashboard/configs/{id}", adminHandler.UpdateDashboardConfig).Methods("PUT")
	router.HandleFunc("/admin/dashboard/configs/{id}", adminHandler.DeleteDashboardConfig).Methods("DELETE")
	router.HandleFunc("/admin/dashboard/configs", adminHandler.ListDashboardConfigs).Methods("GET")

	// Bulk Operations
	router.HandleFunc("/admin/bulk-operations", adminHandler.CreateBulkOperation).Methods("POST")
	router.HandleFunc("/admin/bulk-operations/{id}", adminHandler.GetBulkOperation).Methods("GET")
	router.HandleFunc("/admin/bulk-operations", adminHandler.ListBulkOperations).Methods("GET")

	// System Metrics
	router.HandleFunc("/admin/metrics/update", adminHandler.UpdateSystemMetrics).Methods("POST")

	// Health Check
	router.HandleFunc("/health", adminHandler.HealthCheck).Methods("GET")

	// Add CORS middleware
	router.Use(corsMiddleware)

	// Start server
	port := getEnv("PORT", "8010")
	logger.Info(ctx, "Admin Service listening", map[string]interface{}{"port": port})
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-User-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}