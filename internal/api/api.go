package api

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IbnBaqqi/book-me/internal/google"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/config"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/IbnBaqqi/book-me/internal/service"
	"github.com/IbnBaqqi/book-me/internal/logger"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
)

// API holds all dependencies for the API handlers
type API struct {
	DB               *database.Queries
	dbConn           *sql.DB
	SessionStore     *sessions.CookieStore
	OAuthConfig      *oauth2.Config
	Auth             *auth.Service
	EmailService     *email.Service
	CalendarService  *google.CalendarService
	UserService      *service.UserService
	Reservation      *service.ReservationService
}

// New initializes all services and returns a pointer to API
func New(ctx context.Context, cfg *config.Config) (*API, error) {
	// Initialize database
	dbConn, err := sql.Open("postgres", cfg.App.DBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection, ping with context
	if err := dbConn.PingContext(ctx); err != nil {
		dbConn.Close()
		logger.Log.Error("failed to ping database", "error", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Using SQLC generated database package to create a new *database.Queries,
	dbQueries := database.New(dbConn)

	// Initialize Google Calendar service
	calendarService, err := google.NewCalendarService(
		cfg.Google.CredentialsFile,
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

	// Initialize user service
	userService := service.NewUserService(dbQueries, cfg.App.RedirectTokenURI, cfg.App.User42InfoURL)
	
	// Initialize reservation service
	reservationService := service.NewReservationService(dbQueries, emailService, calendarService)

	return &API{
		DB:               dbQueries,
		dbConn:           dbConn,
		SessionStore:     sessions.NewCookieStore([]byte(cfg.App.SessionSecret)),
		OAuthConfig:      oauthConfig,
		Auth:             authService,
		EmailService:     emailService,
		CalendarService:  calendarService,
		UserService:      userService,
		Reservation:      reservationService,
	}, nil
}

// Close: close database connection
func (a *API) Close() error {
	if a.dbConn != nil {
		if err := a.dbConn.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	return nil
}
