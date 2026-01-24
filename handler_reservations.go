package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/database"
)

func (cfg *apiConfig) handlerCreateReservation(w http.ResponseWriter, r *http.Request) {
	
	type createReservationRequest struct {
		RoomID    int64     `json:"roomId"`
		StartTime time.Time `json:"startTime"`
		EndTime   time.Time `json:"endTime"`
	}

	// Get authenticated user (OAuth/JWT)
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
	fmt.Println(req.RoomID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Room not found", err)
		return
	}

	start := req.StartTime
	end := req.EndTime

	// Validate time
	if start.Before(time.Now()) {
		respondWithError(w, http.StatusNotFound, "You can't book past times", err)
		return
	}

	if !end.After(start) {
		respondWithError(w, http.StatusNotFound, "End time must be after start time", err)
	}

	// Overlap check
	overlap, err := cfg.db.ExistsOverlappingReservation(r.Context(), database.ExistsOverlappingReservationParams{
		RoomID: req.RoomID,
		StartTime:  end,
		EndTime:    start,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal error", err)
	}

	if overlap {
		respondWithError(w, http.StatusBadRequest, "This time slot is already booked", err)
	}

	// Max duration rule (students only)
	duration := end.Sub(start)
	maxMinutes := 240

	if duration.Minutes() > float64(maxMinutes) && currentUser.Role == "STUDENT" {
		respondWithError(w, http.StatusBadRequest, "reservation exceeds maximum allowed duration of 4 hours", err)
	}

	// 5. Persist reservation
	reservation, err := cfg.db.CreateReservation(r.Context(), database.CreateReservationParams{
		UserID:		int64(currentUser.ID),
		RoomID:		room.ID,
		StartTime:	start,
		EndTime:	end,
		Status:		"RESERVED",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal error", err)
	}

	// handle google calender here
	// sending email

	respondWithJSON(w, http.StatusCreated, reservation) // change to return reservation dto without gcal
}