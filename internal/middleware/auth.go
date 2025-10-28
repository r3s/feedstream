package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type AuthMiddleware struct {
	store *sessions.CookieStore
}

func NewAuthMiddleware(store *sessions.CookieStore) *AuthMiddleware {
	return &AuthMiddleware{
		store: store,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.store.Get(r, "session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) GetUserID(r *http.Request) (int, bool) {
	session, err := m.store.Get(r, "session")
	if err != nil {
		return 0, false
	}

	userID, ok := session.Values["user_id"].(int)
	return userID, ok
}

func (m *AuthMiddleware) SetUserSession(w http.ResponseWriter, r *http.Request, userID int) error {
	session, err := m.store.Get(r, "session")
	if err != nil {
		return err
	}

	session.Values["authenticated"] = true
	session.Values["user_id"] = userID

	return session.Save(r, w)
}

func (m *AuthMiddleware) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, "session")
	if err != nil {
		return err
	}

	session.Values["authenticated"] = false
	delete(session.Values, "user_id")

	return session.Save(r, w)
}

func (m *AuthMiddleware) GetSession(r *http.Request) (*sessions.Session, error) {
	return m.store.Get(r, "session")
}