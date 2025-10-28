package app

import (
	"log"
	"net/http"
	"rss-reader/config"
	"rss-reader/internal/database"
	"rss-reader/internal/handler"
	"rss-reader/internal/middleware"
	"rss-reader/internal/repository"
	"rss-reader/internal/service"
	"rss-reader/pkg/datetime"
	"rss-reader/pkg/email"
	"rss-reader/pkg/security"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type Application struct {
	Router       *mux.Router
	Config       *config.Config
	DBManager    *database.Manager
	AuthHandler  *handler.AuthHandler
	FeedHandler  *handler.FeedHandler
	AuthMiddleware *middleware.AuthMiddleware
}

func New(cfg *config.Config) (*Application, error) {
	dbConfig := database.Config{
		ConnectionString: cfg.DatabaseURL,
		Host:             cfg.DBHost,
		Port:             cfg.DBPort,
		User:             cfg.DBUser,
		Password:         cfg.DBPassword,
		DBName:           cfg.DBName,
	}

	dbManager, err := database.NewManager(dbConfig)
	if err != nil {
		return nil, err
	}

	db := dbManager.GetDB()
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	feedRepo := repository.NewFeedRepository(db)
	feedItemRepo := repository.NewFeedItemRepository(db)
	otpGenerator := security.NewOTPGenerator()
	dateFormatter := datetime.NewFormatter()
	emailService, err := email.NewResendService(cfg.ResendAPIKey, cfg.EmailFrom)
	if err != nil {
		log.Printf("Warning: Email service initialization failed: %v", err)
		log.Println("Authentication will not work without email service")
	}
	authService := service.NewAuthService(userRepo, otpRepo, emailService, otpGenerator)
	feedService := service.NewFeedService(feedRepo, feedItemRepo, dateFormatter)
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	authMiddleware := middleware.NewAuthMiddleware(sessionStore)
	authHandler := handler.NewAuthHandler(authService, authMiddleware)
	feedHandler := handler.NewFeedHandler(feedService, authMiddleware)
	router := mux.NewRouter()

	app := &Application{
		Router:         router,
		Config:         cfg,
		DBManager:      dbManager,
		AuthHandler:    authHandler,
		FeedHandler:    feedHandler,
		AuthMiddleware: authMiddleware,
	}

	app.setupRoutes()

	return app, nil
}

func (a *Application) setupRoutes() {
	a.Router.HandleFunc("/", a.redirectToLogin).Methods("GET")
	a.Router.HandleFunc("/login", a.AuthHandler.Login).Methods("GET", "POST")
	a.Router.HandleFunc("/logout", a.AuthHandler.Logout).Methods("GET")
	protected := a.Router.PathPrefix("/").Subrouter()
	protected.Use(a.AuthMiddleware.RequireAuth)

	protected.HandleFunc("/feeds", a.FeedHandler.ViewFeeds).Methods("GET")
	protected.HandleFunc("/feeds/add", a.FeedHandler.AddFeed).Methods("GET", "POST")
	protected.HandleFunc("/feeds/refresh", a.FeedHandler.RefreshFeeds).Methods("GET")
	protected.HandleFunc("/feeds/manage", a.FeedHandler.ManageFeeds).Methods("GET")
	protected.HandleFunc("/feeds/edit/{id}", a.FeedHandler.EditFeed).Methods("GET", "POST")
	protected.HandleFunc("/feeds/delete/{id}", a.FeedHandler.DeleteFeed).Methods("POST")
	protected.HandleFunc("/feeds/import", a.FeedHandler.ImportFeeds).Methods("POST")
	protected.HandleFunc("/feeds/export", a.FeedHandler.ExportFeeds).Methods("GET")
	protected.HandleFunc("/feeds/debug", a.FeedHandler.Debug).Methods("GET")
	a.Router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)
}

func (a *Application) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (a *Application) Close() error {
	if a.DBManager != nil {
		return a.DBManager.Close()
	}
	return nil
}