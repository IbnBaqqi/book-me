package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/external/google"
	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
)

type reservationDTO struct {
	ID        int64     `json:"Id"`
	RoomID    int64     `json:"roomId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	CreatedBy UserDto   `json:"createdBy"`
}

type UserDto struct {
	ID   int64  `json:"Id"`
	Name string `json:"name"`
}

func (cfg *apiConfig) handlerCreateReservation(w http.ResponseWriter, r *http.Request) {

	type createReservationRequest struct {
		RoomID    int64     `json:"roomId"`
		StartTime time.Time `json:"startTime"`
		EndTime   time.Time `json:"endTime"`
	}

	// Get authenticated user from context
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	req := createReservationRequest{}

	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Fetch room
	room, err := cfg.db.GetRoomByID(r.Context(), req.RoomID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Room not found", err)
		return
	}

	start := req.StartTime
	end := req.EndTime

	// Validate time
	if start.Before(time.Now()) {
		respondWithError(w, http.StatusBadRequest, "You can't book past times", err)
		return
	}

	if !end.After(start) {
		respondWithError(w, http.StatusBadRequest, "End time must be after start time", err)
		return
	}

	// Overlap check
	overlap, err := cfg.db.ExistsOverlappingReservation(r.Context(), database.ExistsOverlappingReservationParams{
		RoomID:    req.RoomID,
		StartTime: end,
		EndTime:   start,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal error", err)
		return
	}

	if overlap {
		respondWithError(w, http.StatusBadRequest, "This time slot is already booked", err)
		return
	}

	// Max duration rule (students only)
	duration := end.Sub(start)
	maxMinutes := 240

	if duration.Minutes() > float64(maxMinutes) && currentUser.Role == "STUDENT" {
		respondWithError(w, http.StatusBadRequest, "reservation exceeds maximum allowed duration of 4 hours", err)
		return
	}

	// 5. Persist reservation
	reservation, err := cfg.db.CreateReservation(r.Context(), database.CreateReservationParams{
		UserID:    int64(currentUser.ID),
		RoomID:    room.ID,
		StartTime: start,
		EndTime:   end,
		Status:    "RESERVED",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal error", err)
		return
	}

	// Create Google Calendar event asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build reservation object for calendar service
		calendarReservation := &google.Reservation{
			StartTime: reservation.StartTime,
			EndTime:   reservation.EndTime,
			CreatedBy: currentUser.Name,
			Room:      room.Name,
		}

		eventID, err := cfg.CalendarService.CreateGoogleEvent(ctx, calendarReservation)
		if err != nil {
			log.Printf("Failed to create Google Calendar event: %v", err)
			return
		}

		// Update reservation with event ID
		if eventID != "" {
			updateErr := cfg.db.UpdateGoogleCalID(ctx, database.UpdateGoogleCalIDParams{
				ID:         reservation.ID,
				GcalEventID: sql.NullString{String: eventID, Valid: eventID != ""},
			})
			if updateErr != nil {
				log.Printf("Failed to update reservation with calendar event ID: %v", updateErr)
			}
		}
	}()

	// TODO use redis instead
	dbUser, err := cfg.db.GetUser(r.Context(), int64(currentUser.ID))

	// Send confirmation email (async)
	cfg.EmailService.SendConfirmation(
		r.Context(),
		dbUser.Email,
		room.Name,
		req.StartTime.Format("Monday, January 2, 2006 at 3:04 PM"),
		req.EndTime.Format("Monday, January 2, 2006 at 3:04 PM"),
	)

	respondWithJSON(w, http.StatusCreated, reservationDTO{
		ID:        reservation.ID,
		RoomID:    reservation.RoomID,
		StartTime: reservation.StartTime,
		EndTime:   reservation.EndTime,
		CreatedBy: UserDto{
			ID:   int64(currentUser.ID),
			Name: currentUser.Name,
		},
	})
}
