package api

import (
	"net/http"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/handler"
)

// SetupRoutes configures all HTTP routes and middleware
func SetupRoutes(cfg *API) *http.ServeMux {
	mux := http.NewServeMux()

	// Create handlers with injected dependencies
	h := handler.New(
		cfg.DB,
		cfg.SessionStore,
		cfg.OAuthConfig,
		cfg.Auth,
		cfg.EmailService,
		cfg.CalendarService,
		cfg.RedirectTokenURI,
		cfg.User42InfoURL,
		cfg.JWTSecret,
	)

	// Health check (public)
	mux.HandleFunc("GET /api/healthz", h.Health)

	// Authentication routes (public)
	mux.HandleFunc("GET /api/oauth/login", h.Login)
	mux.HandleFunc("GET /oauth/callback", h.Callback)

	// Reservation routes (authenticated)
	mux.Handle(
		"POST /reservation",
		cfg.Auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(h.CreateReservation))))

	mux.Handle(
		"GET /reservation",
		cfg.Auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(h.GetReservations))))

	mux.Handle(
		"DELETE /reservation/{id}",
		cfg.Auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(h.CancelReservation))))

	return mux
}
