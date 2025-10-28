package main

import (
	"log"
	"net/http"
	"rss-reader/config"
	"rss-reader/internal/app"
	"time"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      application.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server started on port %s (environment: %s)", cfg.AppPort, cfg.Environment)
	if cfg.IsProduction() {
		log.Println("Running in production mode - ensure reverse proxy handles HTTPS")
	}
	
	log.Fatal(server.ListenAndServe())
}