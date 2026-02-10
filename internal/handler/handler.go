package handler

import (
	"github.com/IbnBaqqi/book-me/internal/google"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/IbnBaqqi/book-me/internal/service"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	session          *sessions.CookieStore
	oauthConfig      *oauth2.Config
	auth             *auth.Service
	email            *email.Service
	calendar         *google.CalendarService
	
	reservation		 *service.ReservationService
	userService      *service.UserService
	// authService        *services.AuthService
}

// New creates a new Handler with all dependencies injected
func New(
	sessionStore         *sessions.CookieStore,
	oauthConfig          *oauth2.Config,
	authService          *auth.Service,
	emailService         *email.Service,
	calendarService      *google.CalendarService,
	reservationService   *service.ReservationService,
	userService          *service.UserService,
) *Handler {
	return &Handler{
		session:          sessionStore,
		oauthConfig:      oauthConfig,
		auth:             authService,
		email:            emailService,
		calendar:         calendarService,
		reservation: 	  reservationService,
		userService:      userService,
	}
}
