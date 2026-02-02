package handler

import (
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/auth"
	"github.com/IbnBaqqi/book-me/internal/service"
)

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

// parseDateRange extracts and validates start/end dates from query params
func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	startDateStr := r.URL.Query().Get("start")
	endDateStr := r.URL.Query().Get("end")

	if startDateStr == "" || endDateStr == "" {
		return time.Time{}, time.Time{}, &service.ServiceError{
			Message:    "start and end date parameters are required",
			StatusCode: http.StatusBadRequest,
		}
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, &service.ServiceError{
			Message:    "invalid start date format, expected YYYY-MM-DD",
			StatusCode: http.StatusBadRequest,
		}
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, &service.ServiceError{
			Message:    "invalid end date format, expected YYYY-MM-DD",
			StatusCode: http.StatusBadRequest,
		}
	}

	// Validate date range
	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, &service.ServiceError{
			Message:    "end date must be after start date",
			StatusCode: http.StatusBadRequest,
		}
	}

	return startDate, endDate, nil
}
