// Package validator provides request validation utilities.
package validator

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	validate         *validator.Validate // validator instance for struct validation
	schoolOpenHour   = 6                 // school opening hour (6 AM)
	schoolCloseHour  = 20                // school closing hour (8 PM)
	maxDateRangeDays = 60                // maximum allowed date range in days
)

func init() {
	validate = validator.New()

	// Register custom validators
	// I ignored errors because In practice,
	// RegisterValidation only fails if you pass an empty tag or nil function
	// so this will essentially never fail
	// but the linter is happy because the error is handled.
	_ = validate.RegisterValidation("futureTime", validateFutureTime)
	_ = validate.RegisterValidation("schoolHours", validateSchoolHours)
	_ = validate.RegisterValidation("maxDateRange", validateMaxDateRange)
	_ = validate.RegisterValidation("utc", validateUTC)
}

var helsinkiTZ, _ = time.LoadLocation("Europe/Helsinki")

// Validate validates a struct and returns ValidationError if validation fails
func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	// Convert validator errors to a map of field errors
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		fields := FormatValidationErrors(validationErrs)
		return NewValidationError(fields)
	}
	return err
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// validateFutureTime validate time must be in the future
func validateFutureTime(fl validator.FieldLevel) bool {
	t, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return t.After(time.Now())
}

func validateUTC(fl validator.FieldLevel) bool {
    t, ok := fl.Field().Interface().(time.Time)
    if !ok {
        return false
    }
    _, offset := t.Zone()
    return offset == 0
}

// validateSchoolHours checks if time is within school operating hours
func validateSchoolHours(fl validator.FieldLevel) bool {
	t, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}

	hour := t.In(helsinkiTZ).Hour()
	return hour >= schoolOpenHour && hour < schoolCloseHour
}

// validateMaxDateRange ensures date range doesn't exceed a maximum (e.g 60 days)
func validateMaxDateRange(fl validator.FieldLevel) bool {
	endDate, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}

	startDateField := fl.Parent().FieldByName("StartDate")
	if !startDateField.IsValid() {
		return false
	}

	startDate, ok := startDateField.Interface().(time.Time)
	if !ok {
		return false
	}

	diff := endDate.Sub(startDate)
	return diff.Hours() <= float64(maxDateRangeDays*24)
}

// FormatValidationErrors formats validator errors into user-friendly messages
func FormatValidationErrors(err error) map[string]string {
	errs := make(map[string]string)

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, fieldErr := range validationErrs {
			errs[fieldErr.Field()] = formatFieldError(fieldErr)
		}
	}

	return errs
}

func formatFieldError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "gt":
		return fmt.Sprintf("Must be greater than %s", err.Param())
	case "futureTime":
		return "Time must be in the future"
	case "gtfield":
		return fmt.Sprintf("Must be after %s", err.Param())
	case "gtefield":
		return fmt.Sprintf("Must be after or equal to %s", err.Param())
	case "datetime":
		return "Invalid date/time format"
	case "schoolHours":
		return "Time must be between 6:00 AM and 8:00 PM"
	case "maxDateRange":
		return "Date range cannot exceed 60 days"
	case "utc":
        return "Time must be in UTC format (e.g. 2026-02-23T06:00:00Z)"
	default:
		return fmt.Sprintf("Validation failed on '%s'", err.Tag())
	}
}
