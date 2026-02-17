package validator

import (
	"errors"
	"testing"
	"time"
)

// Test structs
type reservationRequest struct {
	RoomID    int64     `validate:"required,gt=0"`
	StartTime time.Time `validate:"required,futureTime,schoolHours"`
	EndTime   time.Time `validate:"required,gtfield=StartTime,schoolHours"`
}

// TestValidate_Success tests successful validation scenarios
func TestValidate_Success(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour)
	
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name: "valid reservation",
			input: reservationRequest{
				RoomID:    1,
				StartTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
				EndTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "reservation at school start time (6 AM)",
			input: reservationRequest{
				RoomID:    5,
				StartTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 6, 0, 0, 0, time.UTC),
				EndTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "reservation just before school end (7:59 PM)",
			input: reservationRequest{
				RoomID:    3,
				StartTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 18, 0, 0, 0, time.UTC),
				EndTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 19, 59, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if err != nil {
				t.Errorf("Validate() expected no error, got: %v", err)
			}
		})
	}
}

// TestValidate_RequiredFields tests required field validation
func TestValidate_RequiredFields(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedField string
	}{
		{
			name: "missing room ID",
			input: reservationRequest{
				StartTime: time.Now().Add(24 * time.Hour),
				EndTime:   time.Now().Add(25 * time.Hour),
			},
			expectedField: "RoomID",
		},
		{
			name: "missing start time",
			input: reservationRequest{
				RoomID:  1,
				EndTime: time.Now().Add(25 * time.Hour),
			},
			expectedField: "StartTime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected ValidationError, got: %T", err)
			}

			if _, exists := valErr.Fields[tt.expectedField]; !exists {
				t.Errorf("expected error for field %s, got fields: %v", tt.expectedField, valErr.Fields)
			}

			if valErr.Fields[tt.expectedField] != "This field is required" {
				t.Errorf("expected 'This field is required', got: %s", valErr.Fields[tt.expectedField])
			}
		})
	}
}

// TestValidate_FutureTime tests futureTime custom validator
func TestValidate_FutureTime(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		wantErr   bool
		errField  string
	}{
		{
			name:      "past start time",
			startTime: time.Now().Add(-1 * time.Hour),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, time.UTC),
			wantErr:   true,
			errField:  "StartTime",
		},
		{
			name:      "past end time (also violates afterField)",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
			endTime:   time.Now().Add(-1 * time.Hour),
			wantErr:   true,
			errField:  "EndTime",
		},
		{
			name:      "both in future and during school hours",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := reservationRequest{
				RoomID:    1,
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}

			err := Validate(req)

			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got: %v", err)
				}

				// Check that the error field exists (could be multiple errors)
				if _, exists := valErr.Fields[tt.errField]; !exists {
					t.Errorf("expected error for field %s, got fields: %v", tt.errField, valErr.Fields)
				}
			} else if err != nil {
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					t.Errorf("expected no error, got validation errors: %v", valErr.Fields)
				} else {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestValidate_AfterField tests afterField custom validator
func TestValidate_GtField(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		wantErr   bool
	}{
		{
			name:      "end time before start time",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
			wantErr:   true,
		},
		{
			name:      "end time equals start time",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
			wantErr:   true,
		},
		{
			name:      "end time after start time",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := reservationRequest{
				RoomID:    1,
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}

			err := Validate(req)

			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got: %v", err)
				}

				if _, exists := valErr.Fields["EndTime"]; !exists {
					t.Errorf("expected error for EndTime, got fields: %v", valErr.Fields)
				}

				expectedMsg := "Must be after StartTime"
				if valErr.Fields["EndTime"] != expectedMsg {
					t.Errorf("expected '%s', got: %s", expectedMsg, valErr.Fields["EndTime"])
				}
			} else if err != nil {
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					t.Errorf("expected no error, got validation errors: %v", valErr.Fields)
				} else {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestValidate_SchoolHours tests schoolHours custom validator
func TestValidate_SchoolHours(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour)
	
	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		wantErr   bool
		errField  string
	}{
		{
			name:      "start time too early (5:59 AM)",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 5, 59, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, time.UTC),
			wantErr:   true,
			errField:  "StartTime",
		},
		{
			name:      "end time too late (8 PM)",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 18, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 20, 0, 0, 0, time.UTC),
			wantErr:   true,
			errField:  "EndTime",
		},
		{
			name:      "both times during school hours",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "start at 6 AM (boundary)",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 6, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "end at 7:59 PM (boundary)",
			startTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 18, 0, 0, 0, time.UTC),
			endTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 19, 59, 0, 0, time.UTC),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := reservationRequest{
				RoomID:    1,
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}

			err := Validate(req)

			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got: %v", err)
				}

				if _, exists := valErr.Fields[tt.errField]; !exists {
					t.Errorf("expected error for field %s, got fields: %v", tt.errField, valErr.Fields)
				}
			} else if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// TestValidate_MaxDateRange tests maxDateRange custom validator
func TestValidate_MaxDateRange(t *testing.T) {
	type dateRangeStruct struct {
		StartDate time.Time `validate:"required"`
		EndDate   time.Time `validate:"required,maxDateRange"`
	}

	baseDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		wantErr   bool
	}{
		{
			name:      "valid range - 1 day",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 1),
			wantErr:   false,
		},
		{
			name:      "valid range - 30 days",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 30),
			wantErr:   false,
		},
		{
			name:      "valid range - exactly 60 days",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 60),
			wantErr:   false,
		},
		{
			name:      "valid range - 59 days 23 hours",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 59).Add(23 * time.Hour),
			wantErr:   false,
		},
		{
			name:      "invalid range - 61 days",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 61),
			wantErr:   true,
		},
		{
			name:      "invalid range - 90 days",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 90),
			wantErr:   true,
		},
		{
			name:      "invalid range - 180 days",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 180),
			wantErr:   true,
		},
		{
			name:      "invalid range - 1 year",
			startDate: baseDate,
			endDate:   baseDate.AddDate(1, 0, 0),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dateRange := dateRangeStruct{
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
			}

			err := Validate(dateRange)

			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got: %v", err)
				}

				if _, exists := valErr.Fields["EndDate"]; !exists {
					t.Errorf("expected error for EndDate, got fields: %v", valErr.Fields)
				}

				expectedMsg := "Date range cannot exceed 60 days"
				if valErr.Fields["EndDate"] != expectedMsg {
					t.Errorf("expected '%s', got: %s", expectedMsg, valErr.Fields["EndDate"])
				}
			} else if err != nil {
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					t.Errorf("expected no error, got validation errors: %v", valErr.Fields)
				} else {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestValidate_DateRangeWithMultipleValidations tests combining maxDateRange with other validators
func TestValidate_DateRangeWithMultipleValidations(t *testing.T) {
	type dateRangeStruct struct {
		StartDate time.Time `validate:"required"`
		EndDate   time.Time `validate:"required,gtfield=StartDate,maxDateRange"`
	}

	baseDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		startDate   time.Time
		endDate     time.Time
		wantErr     bool
		expectedErr string
	}{
		{
			name:      "valid - all constraints pass",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 30),
			wantErr:   false,
		},
		{
			name:        "invalid - end before start",
			startDate:   baseDate,
			endDate:     baseDate.AddDate(0, 0, -1),
			wantErr:     true,
			expectedErr: "Must be after StartDate",
		},
		{
			name:        "invalid - exceeds max range",
			startDate:   baseDate,
			endDate:     baseDate.AddDate(0, 0, 90),
			wantErr:     true,
			expectedErr: "Date range cannot exceed 60 days",
		},
		{
			name:      "edge case - exactly at limit",
			startDate: baseDate,
			endDate:   baseDate.AddDate(0, 0, 60),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dateRange := dateRangeStruct{
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
			}

			err := Validate(dateRange)

			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected ValidationError, got: %v", err)
				}

				if _, exists := valErr.Fields["EndDate"]; !exists {
					t.Errorf("expected error for EndDate, got fields: %v", valErr.Fields)
				}

				if tt.expectedErr != "" && valErr.Fields["EndDate"] != tt.expectedErr {
					t.Errorf("expected '%s', got: %s", tt.expectedErr, valErr.Fields["EndDate"])
				}
			} else if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// TestValidate_MultipleErrors tests multiple validation errors
func TestValidate_MultipleErrors(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)
	
	req := reservationRequest{
		RoomID:    0, // Invalid: must be > 0
		StartTime: yesterday, // Invalid: past time
		EndTime:   yesterday.Add(-1 * time.Hour), // Invalid: past time & before start
	}

	err := Validate(req)

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}

	// Should have multiple field errors
	if len(valErr.Fields) < 2 {
		t.Errorf("expected at least 2 field errors, got: %d (%v)", len(valErr.Fields), valErr.Fields)
	}

	// Check that RoomID error exists
	if _, exists := valErr.Fields["RoomID"]; !exists {
		t.Error("expected error for RoomID")
	}

	// Check error message
	if valErr.Message != "validation failed" {
		t.Errorf("expected 'validation failed', got: %s", valErr.Message)
	}
}

// TestValidateVar tests ValidateVar function
func TestValidateVar(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		tag     string
		wantErr bool
	}{
		{
			name:    "valid email",
			value:   "test@example.com",
			tag:     "email",
			wantErr: false,
		},
		{
			name:    "invalid email",
			value:   "not-an-email",
			tag:     "email",
			wantErr: true,
		},
		{
			name:    "required field present",
			value:   "something",
			tag:     "required",
			wantErr: false,
		},
		{
			name:    "required field missing",
			value:   "",
			tag:     "required",
			wantErr: true,
		},
		{
			name:    "number greater than zero",
			value:   10,
			tag:     "gt=0",
			wantErr: false,
		},
		{
			name:    "number not greater than zero",
			value:   0,
			tag:     "gt=0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.value, tt.tag)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFormatValidationErrors tests error message formatting
func TestFormatValidationErrors(t *testing.T) {
	// This is more of an integration test
	req := reservationRequest{
		RoomID: 0,
	}

	err := Validate(req)

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}

	// Check that error messages are user-friendly
	for field, msg := range valErr.Fields {
		if msg == "" {
			t.Errorf("field %s has empty error message", field)
		}
		// Messages shouldn't contain technical jargon
		if msg == "validation failed on 'gt'" {
			t.Errorf("field %s has unfriendly error message: %s", field, msg)
		}
	}
}

// TestValidationError_Error tests ValidationError.Error() method
func TestValidationError_Error(t *testing.T) {
	valErr := &ValidationError{
		Message: "validation failed",
		Fields: map[string]string{
			"Email": "Invalid email format",
		},
	}

	errMsg := valErr.Error()
	if errMsg != "validation failed" {
		t.Errorf("expected 'validation failed', got: %s", errMsg)
	}
}