package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/IbnBaqqi/book-me/internal/oauth"
	"github.com/IbnBaqqi/book-me/internal/service"
	appvalidator "github.com/IbnBaqqi/book-me/internal/validator"
)

type errorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// respondWithError sends a JSON error response with the specified HTTP status code and message.
func respondWithError(w http.ResponseWriter, code int, msg string) {

	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

// respondWithValidationError sends a JSON error response with BadRequest HTTP status code, message and details.
func respondWithValidationError(w http.ResponseWriter, code int, msg string, details map[string]string) {

	respondWithJSON(w, code, errorResponse{
		Error:   msg,
		Details: details,
	})
}

// respondWithJSON sends a JSON response with the specified HTTP status code and payload.
// It sets the Content-Type header to application/json and handles marshalling errors.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Error marshalling JSON", "error", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

// handleError handles all application errors
func handleError(w http.ResponseWriter, err error) {

	// Check for validation errors
	var validationErr *appvalidator.ValidationError
	if errors.As(err, &validationErr) {
		respondWithValidationError(w, http.StatusBadRequest, validationErr.Message, validationErr.Fields)
		return
	}

	// Check for service errors, log 5xx errors
	var serviceErr *service.ServiceError
	if errors.As(err, &serviceErr) {
		if serviceErr.StatusCode >= 500 {
			slog.Error("service error",
				"error", serviceErr,
			)
			respondWithError(w, serviceErr.StatusCode, "Internal server error")
			return
		}
		respondWithError(w, serviceErr.StatusCode, serviceErr.Message)
		return
	}

	// Check for oauth errors, log 5xx errors
	var oauthErr *oauth.OauthError
	if errors.As(err, &oauthErr){
		if oauthErr.StatusCode >= 500 {
			slog.Error("oauth error",
				"error", oauthErr,
			)
			respondWithError(w, oauthErr.StatusCode, "Internal server error")
			return
		}
		respondWithError(w, oauthErr.StatusCode, oauthErr.Message)
		return
	}

	// Unexpected errors
	slog.Error("unexpected error", "error", err)
	respondWithError(w, http.StatusInternalServerError, "Internal server error")
}
