package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/IbnBaqqi/book-me/internal/auth"
)

var (
	ErrReservationNotFound      = errors.New("reservation not found")
	ErrUnauthorizedCancellation = errors.New("unauthorized to cancel this reservation")
)

func (cfg *apiConfig) handlerCancelReservation(w http.ResponseWriter, r *http.Request) {

	// Extract ID from path parameter
	idStr := r.PathValue("id")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "reservation ID is required", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid reservation ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user from context
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Call service method
	err = cfg.cancelReservation(r.Context(), id, currentUser)
	if err != nil {
		// Handle different error types
		switch err {
		case ErrReservationNotFound:
			respondWithError(w, http.StatusNotFound, "Reservation doesn't exist", err)
		case ErrUnauthorizedCancellation:
			respondWithError(w, http.StatusForbidden, "You didn't book this slot", err)
		default:
			respondWithError(w, http.StatusInternalServerError, "failed to cancel reservation", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Service method
func (cfg *apiConfig) cancelReservation(
	ctx context.Context,
	id int64,
	currentUser auth.User,
) error {

	// Find reservation by ID
	reservation, err := cfg.db.GetReservationByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReservationNotFound
		}
		return err
	}

	// Check authorization
	isStaff := currentUser.Role == "STAFF"
	isOwner := reservation.UserID == int64(currentUser.ID)

	if !isStaff && !isOwner {
		return ErrUnauthorizedCancellation
	}

	// Delete from database
	err = cfg.db.DeleteReservation(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
