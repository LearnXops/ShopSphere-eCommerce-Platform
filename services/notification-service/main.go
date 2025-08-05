package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/shopsphere/shared/utils"
)

func main() {
	utils.Logger.Println("Starting Notification Service...")
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "notification-service"}`))
	}).Methods("GET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8009"
	}
	utils.Logger.Printf("Notification Service listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}