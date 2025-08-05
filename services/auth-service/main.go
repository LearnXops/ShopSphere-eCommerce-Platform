package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	
	"github.com/shopsphere/auth-service/internal/auth"
	"github.com/shopsphere/auth-service/internal/handlers"
	"github.com/shopsphere/auth-service/internal/jwt"
	"github.com/shopsphere/auth-service/internal/middleware"
	"github.com/shopsphere/auth-service/internal/repository"
	"github.com/shopsphere/auth-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

func main() {
	ctx := context.Background()
	
	// Initialize logger
	utils.Logger.SetServiceName("auth-service")
	utils.Logger.Info(ctx, "Starting Auth Service...")

	// Get configuration from environment
	config := getConfig()

	// Initialize database connection
	db, err := initDatabase(config.DatabaseURL)
	if err != nil {
		utils.Logger.Fatal(ctx, "Failed to initialize database", err)
	}
	defer db.Close()

	// Initialize services
	userRepo := repository.NewUserRepository(db)
	passwordService := auth.NewPasswordService()
	jwtService := jwt.NewJWTService(
		config.JWTAccessSecret,
		config.JWTRefreshSecret,
		config.JWTIssuer,
		config.AccessTokenTTL,
		config.RefreshTokenTTL,
	)
	authService := service.NewAuthService(userRepo, jwtService, passwordService)

	// Initialize handlers and middleware
	authHandlers := handlers.NewAuthHandlers(authService)
	rbacMiddleware := middleware.NewRBACMiddleware(authService)

	// Create router
	router := mux.NewRouter()

	// Add logging middleware
	router.Use(utils.LogMiddleware("auth-service"))

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "auth-service"}`))
	}).Methods("GET")

	// Authentication endpoints
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", authHandlers.Login).Methods("POST")
	authRouter.HandleFunc("/refresh", authHandlers.RefreshToken).Methods("POST")
	authRouter.HandleFunc("/logout", authHandlers.Logout).Methods("POST")
	authRouter.HandleFunc("/validate", authHandlers.ValidateToken).Methods("POST")
	
	// Protected endpoints
	protectedRouter := router.PathPrefix("/auth").Subrouter()
	protectedRouter.Use(rbacMiddleware.Authenticate)
	protectedRouter.Use(rbacMiddleware.RequireActiveUser)
	protectedRouter.HandleFunc("/me", authHandlers.Me).Methods("GET")

	// Start cleanup routine for expired sessions
	go startCleanupRoutine(ctx, authService)

	utils.Logger.Info(ctx, "Auth Service listening on port "+config.Port, map[string]interface{}{
		"port": config.Port,
	})
	
	log.Fatal(http.ListenAndServe(":"+config.Port, router))
}

// Config holds application configuration
type Config struct {
	Port             string
	DatabaseURL      string
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTIssuer        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

// getConfig loads configuration from environment variables
func getConfig() *Config {
	config := &Config{
		Port:             getEnv("PORT", "8001"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://user:password@localhost/shopsphere_auth?sslmode=disable"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-super-secret-access-key-change-in-production"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-in-production"),
		JWTIssuer:        getEnv("JWT_ISSUER", "shopsphere-auth"),
		AccessTokenTTL:   15 * time.Minute,  // 15 minutes
		RefreshTokenTTL:  7 * 24 * time.Hour, // 7 days
	}

	// Parse TTL from environment if provided
	if accessTTL := os.Getenv("JWT_ACCESS_TTL"); accessTTL != "" {
		if duration, err := time.ParseDuration(accessTTL); err == nil {
			config.AccessTokenTTL = duration
		}
	}

	if refreshTTL := os.Getenv("JWT_REFRESH_TTL"); refreshTTL != "" {
		if duration, err := time.ParseDuration(refreshTTL); err == nil {
			config.RefreshTokenTTL = duration
		}
	}

	return config
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// initDatabase initializes database connection
func initDatabase(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// startCleanupRoutine starts a routine to clean up expired sessions
func startCleanupRoutine(ctx context.Context, authService *service.AuthService) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := authService.CleanupExpiredSessions(ctx); err != nil {
				utils.Logger.Error(ctx, "Failed to cleanup expired sessions", err)
			} else {
				utils.Logger.Info(ctx, "Successfully cleaned up expired sessions")
			}
		case <-ctx.Done():
			return
		}
	}
}