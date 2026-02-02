package service

import (
	"fmt"
	"net/http"
)

// ServiceError represents a business logic error with a code
// Commit message: Update error handling logic in internal/services/errors.go for improved clarity and maintainability.
type ServiceError struct {
	Err     error
	Message string
	StatusCode    int
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *ServiceError) Unwrap() error {
	return e.Err
}

var (
    ErrReservationNotFound = &ServiceError{
        Message:    "reservation not found",
        StatusCode: http.StatusNotFound,
    }
    ErrUnauthorized = &ServiceError{
        Message:    "you are not authorized to perform this action",
        StatusCode: http.StatusForbidden,
    }
    ErrRoomNotFound = &ServiceError{
        Message:    "room not found",
        StatusCode: http.StatusNotFound,
    }
    ErrInvalidTimeRange = &ServiceError{
        Message:    "invalid time range: end time must be after start time",
        StatusCode: http.StatusBadRequest,
    }
    ErrPastTime = &ServiceError{
        Message:    "cannot book past times",
        StatusCode: http.StatusBadRequest,
    }
    ErrTimeSlotTaken = &ServiceError{
        Message:    "this time slot is already booked",
        StatusCode: http.StatusConflict,
    }
    ErrExceedsMaxDuration = &ServiceError{
        Message:    "reservation exceeds maximum allowed duration",
        StatusCode: http.StatusBadRequest,
    }
)