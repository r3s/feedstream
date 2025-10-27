package main

import (
	"log"
	"net/http"
	"rss-reader/api"
	"rss-reader/config"
	"rss-reader/db"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()
	db.InitDB(cfg)

	// Debug timezone information
	now := time.Now()
	log.Printf("Server timezone: %s", now.Format("MST"))
	log.Printf("Server time: %s", now.Format("2006-01-02 15:04:05 MST"))
	log.Printf("UTC time: %s", now.UTC().Format("2006-01-02 15:04:05 UTC"))

	app := &api.App{Router: mux.NewRouter(), Config: cfg}
	app.RegisterRoutes()

	log.Println("Server started on port", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, app.Router))
}
