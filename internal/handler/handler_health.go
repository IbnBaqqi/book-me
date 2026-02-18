package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string            `json:"status"`
	Checks   map[string]string `json:"checks"`
}

// Health returns the server and dependencies health status.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	// Check database connectivity
	if err := h.checkDatabase(ctx); err != nil {
		checks["database"] = "unhealthy"
		allHealthy = false
		slog.Error("health check: database unhealthy", "error", err)
	} else {
		checks["database"] = "healthy"
	}

	// Check Google Calendar API
	// (calendar is non-critical so it doesn't fail health check)
	if err := h.checkCalendar(ctx); err != nil {
		checks["calendar"] = "degraded"
		slog.Warn("health check: calendar degraded", "error", err)
	} else {
		checks["calendar"] = "healthy"
	}

	// Email service check
	if h.email == nil {
		checks["email"] = "unhealthy: service not initialized"
		allHealthy = false
	} else {
		checks["email"] = "healthy"
	}

	// Determine overall status
	status := "healthy"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status: status,
		Checks: checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode health response", "error", err)
	}
}

// checkDatabase verifies database connectivity
func (h *Handler) checkDatabase(ctx context.Context) error {
	if h.db == nil {
		return nil
	}
	return h.db.PingContext(ctx)
}

// checkCalendar verifies Google Calendar API is reachable
func (h *Handler) checkCalendar(ctx context.Context) error {
	if h.calendar == nil {
		return fmt.Errorf("database connection not initialized")
	}

	return h.calendar.HealthCheck(ctx)
}
