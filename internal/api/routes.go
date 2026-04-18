//nolint:revive // api is a clear and intentional package name
package api

import (
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/handler"
	"github.com/IbnBaqqi/book-me/internal/middleware"
	"golang.org/x/time/rate"
)

// SetupRoutes configures all HTTP routes and middleware
func SetupRoutes(cfg *API) http.Handler {
	mux := http.NewServeMux()

	// Create handlers with injected dependencies
	h := handler.New(
		cfg.DB,
		cfg.Oauth,
		cfg.Auth,
		cfg.EmailService,
		cfg.CalendarService,
		cfg.Reservation,
	)

	// Create rate limiters
	oauthLimiter := middleware.NewRateLimiter(rate.Every(1*time.Second), 20, false)
	apiLimiter := middleware.NewRateLimiter(rate.Every(1*time.Second), 100, false)

	// Create auth middleware
	authenticate := middleware.Authenticate(cfg.Auth)

	// Health check
	mux.HandleFunc("GET /api/v1/health", h.Health)

	// Authentication routes
	mux.Handle("GET /auth/42/login", oauthLimiter.Limit(http.HandlerFunc(h.Login42)))
	mux.Handle("GET /auth/42/callback", oauthLimiter.Limit(http.HandlerFunc(h.Callback42)))
	mux.Handle("GET /auth/keycloak/login", oauthLimiter.Limit(http.HandlerFunc(h.LoginKeycloak)))
	mux.Handle("GET /auth/keycloak/callback", oauthLimiter.Limit(http.HandlerFunc(h.CallbackKeycloak)))

	// Reservation routes
	mux.Handle(
		"POST /api/v1/reservations",
		apiLimiter.Limit(
			authenticate(
				middleware.RequireAuth(
					http.HandlerFunc(h.CreateReservation)))))

	mux.Handle(
		"GET /api/v1/reservations",
		apiLimiter.Limit(
			authenticate(
				middleware.RequireAuth(
					http.HandlerFunc(h.GetReservations)))))

	mux.Handle(
		"DELETE /api/v1/reservations/{id}",
		apiLimiter.Limit(
			authenticate(
				middleware.RequireAuth(
					http.HandlerFunc(h.CancelReservation)))))

	return middleware.Cors(mux)
}
