package validator

import "fmt"

// ValidationError represents input validation error with fields details
type ValidationError struct {
	Err     error
	Message string
	Fields  map[string]string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a validation error with field details
func NewValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{
		Message: "validation failed",
		Fields:  fields,
	}
}

// Wrap wraps an existing error with additional context
func (e *ValidationError) Wrap(err error) *ValidationError {
	return &ValidationError{
		Err:     err,
		Message: e.Message,
		Fields:  e.Fields,
	}
}
