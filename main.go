package main

import (
	"log"
	"net/http"
	"rss-reader/config"
	"rss-reader/internal/app"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize application with dependency injection
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// Start server
	log.Printf("Server started on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, application.Router))
}