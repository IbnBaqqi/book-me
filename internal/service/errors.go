package service

import (
	"fmt"
	"net/http"
)

// ServiceError represents a business logic error with a status code
//
//nolint:revive // intentional naming for clarity
type ServiceError struct {
	Err        error
	Message    string
	StatusCode int
}

// Error implements the error interface
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

// Predefined errors - Service errors
var (
	ErrGetUserFailed = &ServiceError{
		Message:    "failed to get User",
		StatusCode: http.StatusInternalServerError,
	}
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
	ErrTimeSlotTaken = &ServiceError{
		Message:    "this time slot is already booked",
		StatusCode: http.StatusConflict,
	}
	ErrExceedsMaxDuration = &ServiceError{
		Message:    "reservation exceeds maximum allowed duration",
		StatusCode: http.StatusBadRequest,
	}
	ErrReservationFetchFailed = &ServiceError{
		Message:    "failed to fetch reservations",
		StatusCode: http.StatusInternalServerError,
	}
	ErrUnauthorizedCancellation = &ServiceError{
		Message:    "unauthorized to cancel this reservation",
		StatusCode: http.StatusForbidden,
	}
)
