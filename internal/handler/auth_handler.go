package handler

import (
	"html/template"
	"log"
	"net/http"
	"rss-reader/internal/middleware"
	"rss-reader/internal/service"

	"github.com/gorilla/csrf"
)

type AuthHandler struct {
	authService *service.AuthService
	authMw      *middleware.AuthMiddleware
}

func NewAuthHandler(authService *service.AuthService, authMw *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		authMw:      authMw,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.showLoginPage(w, r, nil)
		return
	}

	if r.Method == "POST" {
		h.handleLoginPost(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *AuthHandler) showLoginPage(w http.ResponseWriter, r *http.Request, data map[string]string) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		data = make(map[string]string)
	}
	
	templateData := map[string]interface{}{
		"Email":     data["Email"],
		"Message":   data["Message"],
		"Error":     data["Error"],
		"csrfField": csrf.TemplateField(r),
	}

	tmpl.Execute(w, templateData)
}

func (h *AuthHandler) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	otp := r.FormValue("otp")

	if otp == "" {
		h.handleSendOTP(w, r, email)
	} else {
		h.handleVerifyOTP(w, r, email, otp)
	}
}

func (h *AuthHandler) handleSendOTP(w http.ResponseWriter, r *http.Request, email string) {
	err := h.authService.SendOTP(email)
	if err != nil {
		log.Printf("Error sending OTP to %s: %v", email, err)
		h.showLoginPage(w, r, map[string]string{
			"Email": email,
			"Error": "Failed to send OTP. Please try again.",
		})
		return
	}

	h.showLoginPage(w, r, map[string]string{
		"Email":   email,
		"Message": "An OTP has been sent to your email.",
	})
}

func (h *AuthHandler) handleVerifyOTP(w http.ResponseWriter, r *http.Request, email, otp string) {
	user, err := h.authService.VerifyOTP(email, otp)
	if err != nil {
		log.Printf("OTP verification failed for %s: %v", email, err)
		h.showLoginPage(w, r, map[string]string{
			"Email": email,
			"Error": "Invalid or expired OTP. Please try again.",
		})
		return
	}

	if err := h.authMw.SetUserSession(w, r, user.ID); err != nil {
		log.Printf("Failed to set session for user %d: %v", user.ID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s logged in successfully", email)
	http.Redirect(w, r, "/feeds", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.authMw.ClearSession(w, r); err != nil {
		log.Printf("Error clearing session: %v", err)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}