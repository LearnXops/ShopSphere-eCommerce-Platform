package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/utils"
)

func main() {
	// Initialize logger
	utils.Logger.Println("Starting Cart Service...")

	// Create router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "cart-service"}`))
	}).Methods("GET")

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	utils.Logger.Printf("Cart Service listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}