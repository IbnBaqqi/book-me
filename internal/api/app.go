package api

import (
	"database/sql"
	"fmt"

	"github.com/IbnBaqqi/book-me/external/google"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/config"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
)

// API holds all dependencies for the API handlers
type API struct {
	DB               *database.Queries
	SessionStore     *sessions.CookieStore
	OAuthConfig      *oauth2.Config
	Auth             *auth.Service
	EmailService     *email.Service
	CalendarService  *google.CalendarService
	RedirectTokenURI string
	User42InfoURL    string
	JWTSecret        string
}

// New initializes all services and returns a pointer to API
func New(cfg *config.Config) (*API, error) {
	// Initialize database
	dbConn, err := sql.Open("postgres", cfg.App.DBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize Database
	dbQueries := database.New(dbConn)

	// Initialize Google Calendar service
	calendarService, err := google.NewCalendarService(
		cfg.Google.PrivateKey,
		cfg.Google.ServiceAccountEmail,
		cfg.Google.TokenURI,
		cfg.Google.CalendarScope,
		cfg.Google.CalendarID,
	)
	if err != nil {
        return nil, fmt.Errorf("failed to initialize calendar service: %w", err)
    }
	
	// Initialize email service
	emailCfg := email.Config{
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUsername: cfg.Email.SMTPUsername,
		SMTPPassword: cfg.Email.SMTPPassword,
		FromEmail:    cfg.Email.FromEmail,
		FromName:     cfg.Email.FromName,
		UseTLS:       cfg.Email.UseTLS,     
	}

	emailService, err := email.NewService(emailCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize email service: %w", err)
	}

	// Initialize OAuth2 config for 42 auth
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.App.ClientID,
		ClientSecret: cfg.App.ClientSecret,
		RedirectURL:  cfg.App.RedirectURI,
		Scopes:       []string{"public"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.App.OAuthAuthURI,
			TokenURL: cfg.App.OAuthTokenURI,
		},
	}

	// Initialize auth service for app (JWT)
	authService := auth.NewService(cfg.App.JWTSecret)

	return &API{
		DB:               dbQueries,
		SessionStore:     sessions.NewCookieStore([]byte(cfg.App.SessionSecret)),
		OAuthConfig:      oauthConfig,
		Auth:             authService,
		EmailService:     emailService,
		CalendarService:  calendarService,
		RedirectTokenURI: cfg.App.RedirectTokenURI,
		User42InfoURL:    cfg.App.User42InfoURL,
		JWTSecret:        cfg.App.JWTSecret,
	}, nil
}