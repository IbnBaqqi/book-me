package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/IbnBaqqi/book-me/internal/validator"
)

// Query parameter validation structs
type dateRangeQuery struct {
	StartDate time.Time `validate:"required"`
	EndDate   time.Time `validate:"required,gtfield=StartDate"`
}

type reservationIDParam struct {
	ID int64 `validate:"required,gt=0"`
}

// parseDateRange extracts and validates start/end dates from query params
func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	startDateStr := r.URL.Query().Get("start")
	endDateStr := r.URL.Query().Get("end")

	// Check if parameters exist
	if startDateStr == "" || endDateStr == "" {
		return time.Time{}, time.Time{}, &validator.ValidationError{
			Message: "Missing required query parameters",
			Fields: map[string]string{
				"start": "Start date is required",
				"end":   "End date is required",
			},
		}
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, &validator.ValidationError{
			Message: "Invalid date format",
			Fields: map[string]string{
				"start": "Invalid start date format, expected YYYY-MM-DD",
			},
		}
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, &validator.ValidationError{
			Message: "Invalid date format",
			Fields: map[string]string{
				"end": "Invalid end date format, expected YYYY-MM-DD",
			},
		}
	}

	// Validate date range using validator
	dateRange := dateRangeQuery{
		StartDate: startDate,
		EndDate:   endDate,
	}

	if err := validator.Validate(dateRange); err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

// parseReservationID extracts and validates reservation ID from path
func parseReservationID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")

	if idStr == "" {
		return 0, &validator.ValidationError{
			Message: "Missing path parameter",
			Fields: map[string]string{
				"id": "Reservation ID is required",
			},
		}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, &validator.ValidationError{
			Message: "Invalid path parameter",
			Fields: map[string]string{
				"id": "Reservation ID must be a valid number",
			},
		}
	}

	// Validate ID is positive
	idParam := reservationIDParam{ID: id}
	if err := validator.Validate(idParam); err != nil {
		return 0, err
	}

	return id, nil
}
