package validator


import (
    "fmt"
    "time"

    "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator.New()
    
    // Register custom validators
    validate.RegisterValidation("futureTime", validateFutureTime)
    validate.RegisterValidation("afterField", validateAfterField)
    validate.RegisterValidation("maxDurationStudent", validateMaxDurationStudent)
}

// Validate validates a struct and returns ValidationError if validation fails
func Validate(s interface{}) error {
    err := validate.Struct(s)
    if err == nil {
        return nil
    }

    // Convert validator errors to a map of field errors
    validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

    fields := FormatValidationErrors(validationErrs)
	return NewValidationError(fields)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
    return validate.Var(field, tag)
}

// Custom validator: time must be in the future
func validateFutureTime(fl validator.FieldLevel) bool {
    t, ok := fl.Field().Interface().(time.Time)
    if !ok {
        return false
    }
    return t.After(time.Now())
}

// Custom validator: field must be after another field
func validateAfterField(fl validator.FieldLevel) bool {
    endTime, ok := fl.Field().Interface().(time.Time)
    if !ok {
        return false
    }
    
    // Get the start time field
    startTimeField := fl.Parent().FieldByName(fl.Param())
    if !startTimeField.IsValid() {
        return false
    }
    
    startTime, ok := startTimeField.Interface().(time.Time)
    if !ok {
        return false
    }
    
    return endTime.After(startTime)
}

// Custom validator for student max duration (context-aware)
func validateMaxDurationStudent(fl validator.FieldLevel) bool {
    // This is just a placeholder - actual validation happens in service layer
    // because we need user role context
    return true
}

// FormatValidationErrors formats validator errors into user-friendly messages
func FormatValidationErrors(err error) map[string]string {
    errors := make(map[string]string)
    
    if validationErrs, ok := err.(validator.ValidationErrors); ok {
        for _, fieldErr := range validationErrs {
            errors[fieldErr.Field()] = formatFieldError(fieldErr)
        }
    }
    
    return errors
}

func formatFieldError(err validator.FieldError) string {
    switch err.Tag() {
    case "required":
        return "This field is required"
    case "gt":
        return fmt.Sprintf("Must be greater than %s", err.Param())
    case "futureTime":
        return "Time must be in the future"
    case "afterField":
        return fmt.Sprintf("Must be after %s", err.Param())
    case "datetime":
        return "Invalid date/time format"
    default:
        return fmt.Sprintf("Validation failed on '%s'", err.Tag())
    }
}