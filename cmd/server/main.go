package main

import (
	"log"
	"net/http"

	"github.com/AlexTLDR/pizzeria/internal/app"
)

func main() {
	// Initialize the application (encapsulates .env loading, DB, OAuth setup)
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// Set up the HTTP handlers
	handler := application.SetupHandlers()

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
