package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/utils"
)

func main() {
	utils.Logger.Println("Starting Search Service...")
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "search-service"}`))
	}).Methods("GET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8011"
	}
	utils.Logger.Printf("Search Service listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}