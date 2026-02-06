package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/IbnBaqqi/book-me/internal/service"
	appvalidator "github.com/IbnBaqqi/book-me/internal/validator"
)

type errorResponse struct {
	Error string `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// respondWithError sends a JSON error response with the specified HTTP status code and message.
// It also logs the provided error and message for server-side debugging.
func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	var serviceErr *service.ServiceError
	if err != nil && !errors.As(err, &serviceErr){
		// log.Println(err) @TODO fix later cause invalid request body logs err
	}

	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}

	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

// respondWithValidationError sends a JSON error response with the specified HTTP status code and message and details.
// It also logs the provided error and message for server-side debugging.
func respondWithValidationError(w http.ResponseWriter, code int, msg string, details map[string]string) {

	respondWithJSON(w, code, errorResponse{
		Error: msg,
		Details: details,
	})
}

// respondWithJSON sends a JSON response with the specified HTTP status code and payload.
// It sets the Content-Type header to application/json and handles marshalling errors.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

// handleServiceError handles ServiceError and sends appropriate HTTP response
func handleServiceError(w http.ResponseWriter, err error) {
	var serviceErr *service.ServiceError
	if errors.As(err, &serviceErr) {
		respondWithError(w, serviceErr.StatusCode, serviceErr.Message, err)
		return
	}

	// Fallback for unexpected errors
	respondWithError(w, http.StatusInternalServerError, "internal server error", err)
}

// handleError handles all application errors (validation + service)
func handleValidationError(w http.ResponseWriter, err error) {
	// Check for validation errors (handler layer)
	var validationErr *appvalidator.ValidationError
	if errors.As(err, &validationErr) {
		respondWithValidationError(w, http.StatusBadRequest, validationErr.Message, validationErr.Fields)
		return
	}

	// // Check for service errors (service layer)
	// var serviceErr *service.ServiceError
	// if errors.As(err, &serviceErr) {
	// 	// Log errors based on severity
	// 	if serviceErr.StatusCode >= 500 {
	// 		// TODO: Add structured logging
	// 		// log.Error("service error", "error", serviceErr, "message", serviceErr.Message)
	// 	}
	// 	respondWithError(w, serviceErr.StatusCode, serviceErr.Message, nil)
	// 	return
	// }

	// Unexpected errors
	// log.Error("unexpected error", "error", err)
	respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
}