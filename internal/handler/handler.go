package handler

import (
	"github.com/IbnBaqqi/book-me/external/google"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/IbnBaqqi/book-me/internal/service"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	db               *database.Queries
	session          *sessions.CookieStore
	oauthConfig      *oauth2.Config
	auth             *auth.Service
	email            *email.Service
	calendar         *google.CalendarService
	
	// Services (business logic)
	reservation		 *service.ReservationService
	// authService        *services.AuthService
	// userService        *services.UserService

	redirectTokenURI string
	user42InfoURL    string
	jwtSecret        string
}

// New creates a new Handler with all dependencies injected
func New(
	db *database.Queries,
	sessionStore *sessions.CookieStore,
	oauthConfig *oauth2.Config,
	authService *auth.Service,
	emailService *email.Service,
	calendarService *google.CalendarService,
	redirectTokenURI string,
	user42InfoURL string,
	jwtSecret string,
) *Handler {
	return &Handler{
		db:               db,
		session:          sessionStore,
		oauthConfig:      oauthConfig,
		auth:             authService,
		email:            emailService,
		calendar:         calendarService,
		reservation: 	  service.NewReservationService(db, emailService, calendarService),
		redirectTokenURI: redirectTokenURI,
		user42InfoURL:    user42InfoURL,
		jwtSecret:        jwtSecret,
	}
}
