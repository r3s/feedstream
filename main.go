package main

import (
	"log"
	"net/http"
	"rss-reader/api"
	"rss-reader/config"
	"rss-reader/db"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()
	db.InitDB(cfg)

	app := &api.App{Router: mux.NewRouter(), Config: cfg}
	app.RegisterRoutes()

	log.Println("Server started on port", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, app.Router))
}
