package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError holds a field-level error message.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors converts validator.ValidationErrors into a
// human-readable slice of ValidationError structs.
func FormatValidationErrors(err error) []ValidationError {
	var errs []ValidationError

	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); !ok {
		// Not a validation error — return generic message
		return []ValidationError{{Field: "request", Message: err.Error()}}
	}

	for _, e := range validationErrors {
		errs = append(errs, ValidationError{
			Field:   toSnakeCase(e.Field()),
			Message: fieldMessage(e),
		})
	}
	return errs
}

// fieldMessage returns a user-friendly error message for a validation tag.
func fieldMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", toSnakeCase(e.Field()))
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", toSnakeCase(e.Field()), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", toSnakeCase(e.Field()), e.Param())
	case "email":
		return "must be a valid email address"
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", toSnakeCase(e.Field()), e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", toSnakeCase(e.Field()), e.Param())
	default:
		return fmt.Sprintf("%s is invalid", toSnakeCase(e.Field()))
	}
}

// toSnakeCase converts a PascalCase field name to snake_case for JSON friendliness.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// errors package needed for As()
