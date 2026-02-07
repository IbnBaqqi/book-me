// internal/handler/parser_test.go
package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IbnBaqqi/book-me/internal/validator"
)

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		name          string
		startParam    string
		endParam      string
		wantErr       bool
		expectedStart time.Time
		expectedEnd   time.Time
		errorField    string
		errorMessage  string
	}{
		{
			name:          "valid date range",
			startParam:    "2026-02-01",
			endParam:      "2026-02-28",
			wantErr:       false,
			expectedStart: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "valid single day range",
			startParam:    "2026-02-15",
			endParam:      "2026-02-15",
			wantErr:       false,
			expectedStart: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "missing start parameter",
			startParam:   "",
			endParam:     "2026-02-28",
			wantErr:      true,
			errorField:   "start",
			errorMessage: "Start date is required",
		},
		{
			name:         "missing end parameter",
			startParam:   "2026-02-01",
			endParam:     "",
			wantErr:      true,
			errorField:   "end",
			errorMessage: "End date is required",
		},
		{
			name:         "missing both parameters",
			startParam:   "",
			endParam:     "",
			wantErr:      true,
			errorField:   "start",
			errorMessage: "Start date is required",
		},
		{
			name:         "invalid start date format",
			startParam:   "2026/02/01",
			endParam:     "2026-02-28",
			wantErr:      true,
			errorField:   "start",
			errorMessage: "Invalid start date format, expected YYYY-MM-DD",
		},
		{
			name:         "invalid end date format",
			startParam:   "2026-02-01",
			endParam:     "28-02-2026",
			wantErr:      true,
			errorField:   "end",
			errorMessage: "Invalid end date format, expected YYYY-MM-DD",
		},
		{
			name:         "malformed start date",
			startParam:   "not-a-date",
			endParam:     "2026-02-28",
			wantErr:      true,
			errorField:   "start",
			errorMessage: "Invalid start date format, expected YYYY-MM-DD",
		},
		{
			name:         "end date before start date",
			startParam:   "2026-02-28",
			endParam:     "2026-02-01",
			wantErr:      true,
			errorField:   "EndDate",
			errorMessage: "Must be after StartDate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/reservations", nil)
			q := req.URL.Query()
			if tt.startParam != "" {
				q.Add("start", tt.startParam)
			}
			if tt.endParam != "" {
				q.Add("end", tt.endParam)
			}
			req.URL.RawQuery = q.Encode()

			// Call parseDateRange
			startDate, endDate, err := parseDateRange(req)

			if tt.wantErr {
				// Expect error
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				var valErr *validator.ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got: %T", err)
				}

				if tt.errorField != "" {
					if _, exists := valErr.Fields[tt.errorField]; !exists {
						t.Errorf("expected error for field '%s', got fields: %v", tt.errorField, valErr.Fields)
					}

					if tt.errorMessage != "" && valErr.Fields[tt.errorField] != tt.errorMessage {
						t.Errorf("expected error message '%s', got: '%s'", tt.errorMessage, valErr.Fields[tt.errorField])
					}
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}

				if !startDate.Equal(tt.expectedStart) {
					t.Errorf("expected start date %v, got %v", tt.expectedStart, startDate)
				}

				if !endDate.Equal(tt.expectedEnd) {
					t.Errorf("expected end date %v, got %v", tt.expectedEnd, endDate)
				}
			}
		})
	}
}

func TestParseReservationID(t *testing.T) {
	tests := []struct {
		name         string
		pathValue    string
		setupRequest func(*http.Request)
		wantErr      bool
		expectedID   int64
		errorField   string
		errorMessage string
	}{
		{
			name:       "valid positive ID",
			pathValue:  "123",
			wantErr:    false,
			expectedID: 123,
		},
		{
			name:       "valid large ID",
			pathValue:  "999999",
			wantErr:    false,
			expectedID: 999999,
		},
		{
			name:         "missing ID parameter",
			pathValue:    "",
			wantErr:      true,
			errorField:   "id",
			errorMessage: "Reservation ID is required",
		},
		{
			name:         "invalid ID - not a number",
			pathValue:    "abc",
			wantErr:      true,
			errorField:   "id",
			errorMessage: "Reservation ID must be a valid number",
		},
		{
			name:         "invalid ID - negative number",
			pathValue:    "-5",
			wantErr:      true,
			errorField:   "ID",
			errorMessage: "Must be greater than 0",
		},
		{
			name:         "invalid ID - zero",
			pathValue:    "0",
			wantErr:      true,
			errorField:   "ID",
			errorMessage: "Must be greater than 0",
		},
		{
			name:         "invalid ID - decimal number",
			pathValue:    "123.45",
			wantErr:      true,
			errorField:   "id",
			errorMessage: "Reservation ID must be a valid number",
		},
		{
			name:         "invalid ID - special characters",
			pathValue:    "12@34",
			wantErr:      true,
			errorField:   "id",
			errorMessage: "Reservation ID must be a valid number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/reservations/"+tt.pathValue, nil)
			
			// Set path value using SetPathValue
			req.SetPathValue("id", tt.pathValue)

			// Call parseReservationID
			id, err := parseReservationID(req)

			if tt.wantErr {
				// Expect error
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				var valErr *validator.ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got: %T", err)
				}

				if tt.errorField != "" {
					if _, exists := valErr.Fields[tt.errorField]; !exists {
						t.Errorf("expected error for field '%s', got fields: %v", tt.errorField, valErr.Fields)
					}

					if tt.errorMessage != "" && valErr.Fields[tt.errorField] != tt.errorMessage {
						t.Errorf("expected error message '%s', got: '%s'", tt.errorMessage, valErr.Fields[tt.errorField])
					}
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}

				if id != tt.expectedID {
					t.Errorf("expected ID %d, got %d", tt.expectedID, id)
				}
			}
		})
	}
}