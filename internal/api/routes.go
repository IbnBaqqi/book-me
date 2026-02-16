package api

import (
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/handler"
	"github.com/IbnBaqqi/book-me/internal/middleware"
	"golang.org/x/time/rate"
)

// SetupRoutes configures all HTTP routes and middleware
func SetupRoutes(cfg *API) *http.ServeMux {
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
	// OAuth endpoints: 5 requests per minute (prevent brute force)
	oauthLimiter := middleware.NewRateLimiter(rate.Every(12*time.Second), 5)
	
	// API endpoints: 30 requests per minute (normal usage)
	apiLimiter := middleware.NewRateLimiter(rate.Every(2*time.Second), 30)

	// Health check (public, no rate limit)
	mux.HandleFunc("GET /health", h.Health)

	// Authentication routes (rate limited)
	mux.Handle("GET /oauth/login", oauthLimiter.Limit(http.HandlerFunc(h.Login)))
	mux.Handle("GET /oauth/callback", oauthLimiter.Limit(http.HandlerFunc(h.Callback)))

	// Reservation routes (authenticated + rate limited)
	mux.Handle(
		"POST /api/v1/reservations",
		apiLimiter.Limit(
			cfg.Auth.Authenticate(
				auth.RequireAuth(
					http.HandlerFunc(h.CreateReservation)))))

	mux.Handle(
		"GET /api/v1/reservations",
		apiLimiter.Limit(
			cfg.Auth.Authenticate(
				auth.RequireAuth(
					http.HandlerFunc(h.GetReservations)))))

	mux.Handle(
		"DELETE /api/v1/reservations/{id}",
		apiLimiter.Limit(
			cfg.Auth.Authenticate(
				auth.RequireAuth(
					http.HandlerFunc(h.CancelReservation)))))

	return mux
}
