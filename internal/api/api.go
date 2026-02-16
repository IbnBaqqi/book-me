// Package api wires up dependencies and defines HTTP routes for the book-me server.
//
//nolint:revive // api is a clear and intentional package name
package api

import (
	"database/sql"
	"fmt"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/config"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/IbnBaqqi/book-me/internal/google"
	"github.com/IbnBaqqi/book-me/internal/oauth"
	"github.com/IbnBaqqi/book-me/internal/service"
	"golang.org/x/oauth2"
)

// API holds all dependencies for the API handlers
type API struct {
	DB               *database.Queries
	sqlDB            *sql.DB
	Oauth            *oauth.Service
	Auth             *auth.Service
	EmailService     *email.Service
	CalendarService  *google.CalendarService
	Reservation      *service.ReservationService
}

// New initializes all services and returns a pointer to API
func New(cfg *config.Config, db *database.DB) (*API, error) {

	// Using SQLC generated database package to create a new *database.Queries,
	dbQueries := database.New(db)

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

	// Initialize 42 oauth provider & service
	oauth42 := oauth.NewProvider42(dbQueries, oauthConfig, cfg.App.SessionSecret, cfg.App.RedirectTokenURI,cfg.App.User42InfoURL)
	oauthService := oauth.NewService(oauth42)

	// Initialize auth service for app (JWT)
	authService := auth.NewService(cfg.App.JWTSecret)
	
	// Initialize reservation service
	reservationService := service.NewReservationService(dbQueries, db, emailService, calendarService)

	return &API{
		DB:               dbQueries,
		sqlDB:            db.DB,
		Oauth:            oauthService,
		Auth:             authService,
		EmailService:     emailService,
		CalendarService:  calendarService,
		Reservation:      reservationService,
	}, nil
}
