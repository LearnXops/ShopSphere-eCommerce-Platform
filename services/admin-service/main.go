package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/utils"
)

func main() {
	utils.Logger.Println("Starting Admin Service...")
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "admin-service"}`))
	}).Methods("GET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8010"
	}
	utils.Logger.Printf("Admin Service listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}