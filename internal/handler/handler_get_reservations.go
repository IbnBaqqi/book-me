package handler

import (
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/service"
)

func (h *Handler) GetReservations(w http.ResponseWriter, r *http.Request) {

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start")
	endDateStr := r.URL.Query().Get("end")

	if startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "start and end parameters are required", nil)
		return
	}

	// Parse dates (ISO format: YYYY-MM-DD)
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid end date format", err)
		return
	}

	// Get authenticated user from context
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := service.GetReservationsInput{
		StartDate: startDate,
		EndDate:   endDate,
		UserID:    int64(currentUser.ID),
		UserRole:  currentUser.Role,
	}

	// Call service method
	reserved, err := h.reservation.GetReservations(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, reserved)
}
