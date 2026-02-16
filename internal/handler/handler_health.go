package handler

import (
	"log/slog"
	"net/http"
)

// Health returns the server health status.
func (h *Handler) Health (w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
    	slog.Error("failed to write response", "error", err)
	}
}