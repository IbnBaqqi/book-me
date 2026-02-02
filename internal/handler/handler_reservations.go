package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/service"
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

// Handler to create a new reservation
func (h *Handler) CreateReservation(w http.ResponseWriter, r *http.Request) {

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

	// TODO use redis instead
	dbUser, err := h.db.GetUser(r.Context(), int64(currentUser.ID))

	// Call service
	reservation, err := h.reservation.CreateReservation(r.Context(), service.CreateReservationInput{
		UserID:    int64(currentUser.ID),
		UserName:  currentUser.Name,
		UserEmail: dbUser.Email,
		UserRole:  currentUser.Role,
		RoomID:    req.RoomID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})

	if err != nil {
		handleServiceError(w, err)
		return
	}

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

// Handler to Fetch reservations and group them
func (h *Handler) GetReservations(w http.ResponseWriter, r *http.Request) {

	// Validate & parse query parameters
	startDate, endDate, err := parseDateRange(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
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
		UserID:    int64(currentUser.ID),
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

	// Build service input
	input := service.CancelReservationInput{
		ID: id,
		UserID: currentUser.ID,
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

// handleServiceError maps service errors to HTTP responses
func handleServiceError(w http.ResponseWriter, err error) {
	var serviceErr *service.ServiceError
	if errors.As(err, &serviceErr) {
		respondWithError(w, serviceErr.StatusCode, serviceErr.Message, err)
		return
	}

	// Fallback for unexpected errors
	respondWithError(w, http.StatusInternalServerError, "internal server error", err)
}
