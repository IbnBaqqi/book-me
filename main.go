package main

import (
	"log"
	"net/http"

	"github.com/IbnBaqqi/book-me/external/google"
	"github.com/IbnBaqqi/book-me/internal/api"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/config"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
)

type apiConfig struct {
	db               *database.Queries
	sessionStore     *sessions.CookieStore
	oauthConfig      *oauth2.Config
	auth             *auth.Service
	EmailService     *email.Service
	CalendarService  *google.CalendarService
	redirectTokenURI string
	user42InfoURL    string
	jwtSecret        string
}

func main() {
	const port = "8080"

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	apiCfg, err := api.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize API config: %v", err)
	}

	// Setup routes
	mux := api.SetupRoutes(apiCfg)

	// mux := http.NewServeMux()

	// mux.HandleFunc("GET /api/healthz", healthHandler)

	// mux.HandleFunc("GET /api/oauth/login", apiCfg.loginHandler)
	// mux.HandleFunc("GET /oauth/callback", apiCfg.handlerCallback)

	// mux.Handle(
	// 	"POST /reservation",
	// 	apiCfg.auth.Authenticate(
	// 		auth.RequireAuth(
	// 			http.HandlerFunc(apiCfg.handlerCreateReservation))))
	// mux.Handle(
	// 	"GET /reservation",
	// 	apiCfg.auth.Authenticate(
	// 		auth.RequireAuth(
	// 			http.HandlerFunc(apiCfg.handlerFetchReservations))))
	// mux.Handle(
	// 	"DELETE /reservation/{id}",
	// 	apiCfg.auth.Authenticate(
	// 		auth.RequireAuth(
	// 			http.HandlerFunc(apiCfg.handlerCancelReservation))))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
