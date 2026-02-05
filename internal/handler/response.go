package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/IbnBaqqi/book-me/internal/service"
)

// respondWithError sends a JSON error response with the specified HTTP status code and message.
// It also logs the provided error and message for server-side debugging.
func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	var serviceErr *service.ServiceError
	if err != nil && !errors.As(err, &serviceErr){
		log.Println(err)
	}

	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
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
