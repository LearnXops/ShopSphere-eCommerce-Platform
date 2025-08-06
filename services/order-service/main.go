package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/order-service/internal/handlers"
	"github.com/shopsphere/order-service/internal/repository"
	"github.com/shopsphere/order-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

func main() {
	ctx := context.Background()

	// Initialize logger
	utils.Logger.Info(ctx, "Starting Order Service...", nil, nil)

	// Initialize database connection
	dbConfig := utils.NewDatabaseConfig()
	dbConfig.DBName = "order_service"
	
	db, err := dbConfig.Connect()
	if err != nil {
		utils.Logger.Fatal(ctx, "Failed to connect to database", err)
	}
	defer db.Close()

	// Initialize repository
	orderRepo := repository.NewPostgresOrderRepository(db)

	// Initialize service (with nil product and inventory services for now)
	orderService := service.NewOrderService(orderRepo, nil, nil)

	// Initialize handlers
	orderHandler := handlers.NewOrderHandler(orderService)

	// Create router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Order routes
	api.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	api.HandleFunc("/orders/{id}", orderHandler.GetOrder).Methods("GET")
	api.HandleFunc("/orders/{id}", orderHandler.UpdateOrder).Methods("PUT")
	api.HandleFunc("/orders/{id}/status", orderHandler.UpdateOrderStatus).Methods("PATCH")
	api.HandleFunc("/orders/{id}/cancel", orderHandler.CancelOrder).Methods("POST")
	api.HandleFunc("/orders/{id}/history", orderHandler.GetOrderStatusHistory).Methods("GET")
	api.HandleFunc("/orders/{id}/summary", orderHandler.GetOrderSummary).Methods("GET")
	api.HandleFunc("/orders/user/{userId}", orderHandler.GetUserOrders).Methods("GET")
	api.HandleFunc("/orders/search", orderHandler.SearchOrders).Methods("GET")
	api.HandleFunc("/orders/validate", orderHandler.ValidateOrder).Methods("POST")

	// Health check endpoint
	router.HandleFunc("/health", orderHandler.HealthCheck).Methods("GET")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	utils.Logger.Info(ctx, "Order Service listening", nil, map[string]interface{}{"port": port})
	log.Fatal(http.ListenAndServe(":"+port, router))
}