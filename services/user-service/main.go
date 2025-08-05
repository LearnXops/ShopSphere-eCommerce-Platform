package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/utils"
	"github.com/shopsphere/user-service/internal/handlers"
	"github.com/shopsphere/user-service/internal/repository"
	"github.com/shopsphere/user-service/internal/service"
)

func main() {
	ctx := context.Background()
	
	// Initialize logger
	utils.Logger.SetServiceName("user-service")
	utils.Logger.Info(ctx, "Starting User Service...")

	// Initialize database connection
	dbConfig := utils.NewDatabaseConfig()
	dbConfig.DBName = "user_service"
	
	db, err := dbConfig.Connect()
	if err != nil {
		utils.Logger.Fatal(ctx, "Failed to connect to database", err)
	}
	defer db.Close()

	// Initialize repository, service, and handler layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Create router
	router := mux.NewRouter()

	// Add logging middleware
	router.Use(utils.LogMiddleware("user-service"))

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "user-service"}`))
	}).Methods("GET")

	// User management endpoints
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Public endpoints
	api.HandleFunc("/users/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/users/verify-email", userHandler.VerifyEmail).Methods("POST")
	api.HandleFunc("/users/password-reset/request", userHandler.RequestPasswordReset).Methods("POST")
	api.HandleFunc("/users/password-reset/confirm", userHandler.ResetPassword).Methods("POST")
	
	// User endpoints (require authentication in production)
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}/password", userHandler.ChangePassword).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	
	// Admin endpoints (require admin role in production)
	api.HandleFunc("/users", userHandler.ListUsers).Methods("GET")
	api.HandleFunc("/users/search", userHandler.SearchUsers).Methods("GET")
	api.HandleFunc("/users/{id}/status", userHandler.UpdateUserStatus).Methods("PUT")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	utils.Logger.Info(ctx, "User Service listening on port", map[string]interface{}{
		"port": port,
	})
	log.Fatal(http.ListenAndServe(":"+port, router))
}