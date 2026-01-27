package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/IbnBaqqi/book-me/external/google"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
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

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system env vars")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET must be set")
	}

	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		log.Fatal("CLIENT_ID must be set")
	}

	clientSecret := os.Getenv("SECRET")
	if clientSecret == "" {
		log.Fatal("SECRET must be set")
	}

	redirectURI := os.Getenv("REDIRECT_URI")
	if redirectURI == "" {
		log.Fatal("REDIRECT_URI must be set")
	}

	redirectTokenURI := os.Getenv("REDIRECT_TOKEN_URI")
	if redirectTokenURI == "" {
		log.Fatal("REDIRECT_TOKEN_URI must be set")
	}

	user42InfoURL := os.Getenv("USER_INFO_URL")
	if user42InfoURL == "" {
		log.Fatal("USER_INFO_URL must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// Google Calendar configuration
	googleCalendarScope := os.Getenv("GOOGLE_CALENDAR_SCOPE")
	if googleCalendarScope == "" {
		log.Fatal("GOOGLE_CALENDAR_SCOPE must be set")
	}

	googlePrivateKey := os.Getenv("GOOGLE_PRIVATE_KEY")
	if googlePrivateKey == "" {
		log.Fatal("GOOGLE_PRIVATE_KEY must be set")
	}

	googleServiceAccountEmail := os.Getenv("GOOGLE_SERVICE_ACCOUNT_EMAIL")
	if googleServiceAccountEmail == "" {
		log.Fatal("GOOGLE_SERVICE_ACCOUNT_EMAIL must be set")
	}

	googleTokenURI := os.Getenv("GOOGLE_TOKEN_URI")
	if googleTokenURI == "" {
		log.Fatal("GOOGLE_TOKEN_URI must be set")
	}

	googleTokenExpiration, err := strconv.ParseInt(os.Getenv("GOOGLE_TOKEN_EXPIRATION"), 10, 64)
	if err != nil {
		googleTokenExpiration = 3600 // Default 1 hour
	}

	googleCalendarURI := os.Getenv("GOOGLE_CALENDAR_URI")
	if googleCalendarURI == "" {
		log.Fatal("GOOGLE_CALENDAR_URI must be set")
	}

	googleCalendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if googleCalendarID == "" {
		googleCalendarID = "primary" // Default
	}

	// Initialize Google Calendar service
	authService := google.NewAuthService(
		googleCalendarScope,
		googlePrivateKey,
		googleServiceAccountEmail,
		googleTokenURI,
		googleTokenExpiration,
	)

	tokenManager := google.NewTokenManager(authService)
	calendarSvc := google.NewCalendarService(tokenManager, googleCalendarURI, googleCalendarID)

	// Email configuration
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		smtpPort = 587 // Default
	}

	emailCfg := email.Config{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     os.Getenv("FROM_NAME"),
		UseTLS:       os.Getenv("SMTP_USE_TLS") == "true",
	}

	emailSvc, err := email.NewService(emailCfg)
	if err != nil {
		log.Fatalf("failed to initialize email service: %v", err)
	}

	// open a connection to the database
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", dbURL)
	}

	// Using SQLC generated database package to create a new *database.Queries,
	// and storing in apiConfig struct so that handlers can access it:
	dbQueries := database.New(dbConn)

	// Configure OAuth2 for 42 Intra
	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"public"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  os.Getenv("OAUTH_AUTH_URI"),
			TokenURL: os.Getenv("OAUTH_TOKEN_URI"),
		},
	}

	apiCfg := &apiConfig{
		db:               dbQueries,
		sessionStore:     sessions.NewCookieStore([]byte(sessionSecret)),
		oauthConfig:      oauthCfg,
		auth:             auth.NewService(jwtSecret),
		EmailService:     emailSvc,
		CalendarService:  calendarSvc,
		redirectTokenURI: redirectTokenURI,
		user42InfoURL:    user42InfoURL,
		jwtSecret:        jwtSecret,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", healthHandler)

	mux.HandleFunc("GET /api/oauth/login", apiCfg.loginHandler)
	mux.HandleFunc("GET /oauth/callback", apiCfg.handlerCallback)

	mux.Handle(
		"POST /reservation",
		apiCfg.auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(apiCfg.handlerCreateReservation))))
	mux.Handle(
		"GET /reservation",
		apiCfg.auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(apiCfg.handlerFetchReservations))))
	mux.Handle(
		"DELETE /reservation/{id}",
		apiCfg.auth.Authenticate(
			auth.RequireAuth(
				http.HandlerFunc(apiCfg.handlerCancelReservation))))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
