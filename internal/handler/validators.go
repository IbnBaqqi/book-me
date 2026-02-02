package handler

import (
	"net/http"
	"time"

	"github.com/IbnBaqqi/book-me/internal/service"
)

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