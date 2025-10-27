package api

import (
	"net/http"
	"rss-reader/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type App struct {
	Router *mux.Router
	Config *config.Config
	Store  *sessions.CookieStore
}

// RegisterRoutes registers the application routes
func (a *App) RegisterRoutes() {
	a.Store = sessions.NewCookieStore([]byte(a.Config.SessionSecret))

	a.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	a.Router.HandleFunc("/login", a.LoginHandler).Methods("GET", "POST")
	a.Router.HandleFunc("/logout", a.LogoutHandler).Methods("GET")

	// Authenticated routes
	authRouter := a.Router.PathPrefix("/").Subrouter()
	authRouter.Use(a.authMiddleware)
	authRouter.HandleFunc("/feeds", a.FeedsHandler).Methods("GET")
	authRouter.HandleFunc("/feeds/add", a.AddFeedHandler).Methods("GET", "POST")
	authRouter.HandleFunc("/feeds/refresh", a.RefreshFeedsHandler).Methods("GET")
	authRouter.HandleFunc("/feeds/manage", a.ManageFeedsHandler).Methods("GET")
	authRouter.HandleFunc("/feeds/edit/{id}", a.EditFeedHandler).Methods("GET", "POST")
	authRouter.HandleFunc("/feeds/delete/{id}", a.DeleteFeedHandler).Methods("POST")
	authRouter.HandleFunc("/feeds/debug", a.DebugHandler).Methods("GET")

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	a.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
}

func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.Store.Get(r, "session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		next.ServeHTTP(w, r)
	})
}
