package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/utils"
)

func main() {
	utils.Logger.Println("Starting Shipping Service...")
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "shipping-service"}`))
	}).Methods("GET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8007"
	}
	utils.Logger.Printf("Shipping Service listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}