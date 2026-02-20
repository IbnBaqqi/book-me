package handler

import (
	"encoding/json"
	"net/http"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/dto"
	"github.com/IbnBaqqi/book-me/internal/service"
	appvalidator "github.com/IbnBaqqi/book-me/internal/validator"
)

const maxRequestBodySize int64 = 1 * 1024 * 1024 // 1MB

// CreateReservation is handler to handles creation of a new reservation
//
// POST /reservations
func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {

	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	req := dto.CreateReservationRequest{}
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the request
	if err := appvalidator.Validate(req); err != nil {
		handleError(w, err)
		return
	}

	// Call service
	reservation, err := h.reservation.CreateReservation(r.Context(), service.CreateReservationInput{
		UserID:    currentUser.ID,
		UserName:  currentUser.Name,
		UserRole:  currentUser.Role,
		RoomID:    req.RoomID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})

	if err != nil {
		handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, dto.ReservationDto{
		ID:        reservation.ID,
		RoomID:    reservation.RoomID,
		StartTime: reservation.StartTime,
		EndTime:   reservation.EndTime,
		CreatedBy: dto.UserDto{
			ID:   currentUser.ID,
			Name: currentUser.Name,
		},
	})
}

// GetReservations Handler handles fetching reservations and group them
//
// GET /reservations
func (h *Handler) GetReservations(w http.ResponseWriter, r *http.Request) {

	// Validate & parse query parameters
	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		handleError(w, err)
		return
	}

	// Get authenticated user from context
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Build service input
	input := service.GetReservationsInput{
		StartDate: startDate,
		EndDate:   endDate,
		UserID:    currentUser.ID,
		UserRole:  currentUser.Role,
	}

	// Call service
	reserved, err := h.reservation.GetReservations(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, reserved)
}

// CancelReservation handler handles cancelling a reservation.
//
// DELETE /reservations/{id}
func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {

	// Extract & validate ID from path parameter
	id, err := parseReservationID(r)
	if err != nil {
		handleError(w, err)
		return
	}

	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input := service.CancelReservationInput{
		ID:       id,
		UserID:   currentUser.ID,
		UserRole: currentUser.Role,
	}

	// Call service
	err = h.reservation.CancelReservation(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
