package handler

import (
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
	"github.com/IbnBaqqi/book-me/internal/email"
	"github.com/IbnBaqqi/book-me/internal/google"
	"github.com/IbnBaqqi/book-me/internal/oauth"
	"github.com/IbnBaqqi/book-me/internal/service"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	db               *database.DB
	oauth            *oauth.Service
	auth             *auth.Service
	email            *email.Service
	calendar         *google.CalendarService
	reservation		 *service.ReservationService
}

// New creates a new Handler with all dependencies injected
func New(
	db                   *database.DB,
	oauthService         *oauth.Service,
	authService          *auth.Service,
	emailService         *email.Service,
	calendarService      *google.CalendarService,
	reservationService   *service.ReservationService,
) *Handler {
	return &Handler{
		db:               db,
		oauth:            oauthService,
		auth:             authService,
		email:            emailService,
		calendar:         calendarService,
		reservation: 	  reservationService,
	}
}
