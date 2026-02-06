package validator

// ValidationError represents input validation error with fields details
type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a validation error with field details
func NewValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{
		Message: "validation failed",
		Fields:  fields,
	}
}