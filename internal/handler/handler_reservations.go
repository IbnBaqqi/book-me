package handler

import (
	"encoding/json"
	"net/http"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/dto"
	"github.com/IbnBaqqi/book-me/internal/service"
	appvalidator "github.com/IbnBaqqi/book-me/internal/validator"
)

// Handler to create a new reservation
func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {

	// Get authenticated user from context
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB

	// Decode with strict validation
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	req := dto.CreateReservationRequest{}
	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate the request
	if err := appvalidator.Validate(req); err != nil {
		handleValidationError(w, err)
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
		handleServiceError(w, err)
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

// Handler to Fetch reservations and group them
func (h *Handler) GetReservations(w http.ResponseWriter, r *http.Request) {

	// Validate & parse query parameters
	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		handleValidationError(w, err)
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
		handleServiceError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, reserved)
}

// Handler to cancel reservation
func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {

	// Extract & validate ID from path parameter
	id, err := parseReservationID(r)
	if err != nil {
		handleValidationError(w, err)
		return
	}

	// Get authenticated user from context
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// Build service input
	input := service.CancelReservationInput{
		ID:       id,
		UserID:   currentUser.ID,
		UserRole: currentUser.Role,
	}

	// Call service
	err = h.reservation.CancelReservation(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
