package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/product-service/internal/handlers"
	"github.com/shopsphere/product-service/internal/repository"
	"github.com/shopsphere/product-service/internal/service"
	"github.com/shopsphere/shared/utils"
)

func main() {
	ctx := context.Background()
	
	// Initialize logger
	utils.Logger.Info(ctx, "Starting Product Service...")

	// Initialize database connection
	dbConfig := utils.NewDatabaseConfig()
	dbConfig.DBName = "product_service"
	
	db, err := dbConfig.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// Initialize services
	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService, categoryService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// Create router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "product-service"}`))
	}).Methods("GET")

	// Product routes
	productRoutes := router.PathPrefix("/products").Subrouter()
	productRoutes.HandleFunc("", productHandler.ListProducts).Methods("GET")
	productRoutes.HandleFunc("", productHandler.CreateProduct).Methods("POST")
	productRoutes.HandleFunc("/search", productHandler.SearchProducts).Methods("GET")
	productRoutes.HandleFunc("/bulk-stock-update", productHandler.BulkUpdateStock).Methods("POST")
	productRoutes.HandleFunc("/sku/{sku}", productHandler.GetProductBySKU).Methods("GET")
	productRoutes.HandleFunc("/{id}", productHandler.GetProduct).Methods("GET")
	productRoutes.HandleFunc("/{id}", productHandler.UpdateProduct).Methods("PUT")
	productRoutes.HandleFunc("/{id}", productHandler.DeleteProduct).Methods("DELETE")
	productRoutes.HandleFunc("/{id}/reserve-stock", productHandler.ReserveStock).Methods("POST")
	productRoutes.HandleFunc("/{id}/release-stock", productHandler.ReleaseStock).Methods("POST")
	productRoutes.HandleFunc("/{id}/stock", productHandler.UpdateStock).Methods("PUT")

	// Category routes
	categoryRoutes := router.PathPrefix("/categories").Subrouter()
	categoryRoutes.HandleFunc("", categoryHandler.ListCategories).Methods("GET")
	categoryRoutes.HandleFunc("", categoryHandler.CreateCategory).Methods("POST")
	categoryRoutes.HandleFunc("/root", categoryHandler.GetRootCategories).Methods("GET")
	categoryRoutes.HandleFunc("/tree", categoryHandler.GetCategoryTree).Methods("GET")
	categoryRoutes.HandleFunc("/{id}", categoryHandler.GetCategory).Methods("GET")
	categoryRoutes.HandleFunc("/{id}", categoryHandler.UpdateCategory).Methods("PUT")
	categoryRoutes.HandleFunc("/{id}", categoryHandler.DeleteCategory).Methods("DELETE")
	categoryRoutes.HandleFunc("/{id}/children", categoryHandler.GetCategoryChildren).Methods("GET")
	categoryRoutes.HandleFunc("/{id}/path", categoryHandler.GetCategoryPath).Methods("GET")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	utils.Logger.Info(ctx, "Product Service listening on port", map[string]interface{}{"port": port})
	log.Fatal(http.ListenAndServe(":"+port, router))
}