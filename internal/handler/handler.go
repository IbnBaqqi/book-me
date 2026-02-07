package handler

import (
	"github.com/IbnBaqqi/book-me/external/google"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/IbnBaqqi/book-me/internal/service"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"log/slog"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	db               *database.Queries
	session          *sessions.CookieStore
	oauthConfig      *oauth2.Config
	auth             *auth.Service
	email            *email.Service
	calendar         *google.CalendarService
	
	reservation		 *service.ReservationService
	userService      *service.UserService
	logger           *slog.Logger
	// authService        *services.AuthService
}

// New creates a new Handler with all dependencies injected
func New(
	db *database.Queries,
	sessionStore *sessions.CookieStore,
	oauthConfig *oauth2.Config,
	authService *auth.Service,
	emailService *email.Service,
	calendarService *google.CalendarService,
	userService     *service.UserService,
) *Handler {
	return &Handler{
		db:               db,
		session:          sessionStore,
		oauthConfig:      oauthConfig,
		auth:             authService,
		email:            emailService,
		calendar:         calendarService,
		reservation: 	  service.NewReservationService(db, emailService, calendarService),
		userService:      userService,
	}
}
