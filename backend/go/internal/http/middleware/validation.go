package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// FormatValidationError formats validator errors into a user-friendly format
func FormatValidationError(err error) (string, string) {
	var validationErrors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationErr := range validatorErrors {
			fieldError := ValidationError{
				Field:   validationErr.Field(),
				Tag:     validationErr.Tag(),
				Value:   fmt.Sprintf("%v", validationErr.Value()),
				Message: getValidationErrorMessage(validationErr),
			}
			validationErrors = append(validationErrors, fieldError)
		}
	}

	if len(validationErrors) == 0 {
		return "validation_failed", "Validation failed"
	}

	// Create detailed error message
	var messages []string
	for _, validationErr := range validationErrors {
		messages = append(messages, validationErr.Message)
	}

	details := strings.Join(messages, "; ")
	return "validation_failed", details
}

// getValidationErrorMessage returns a user-friendly message for validation errors
func getValidationErrorMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters", field, fe.Param())
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field)
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("Field '%s' must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", field, fe.Param())
	case "dive":
		return fmt.Sprintf("Array field '%s' contains invalid items", field)
	case "gte":
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("Field '%s' must be less than or equal to %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("Field '%s' must be greater than %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("Field '%s' must be less than %s", field, fe.Param())
	default:
		return fmt.Sprintf("Field '%s' failed validation rule '%s'", field, fe.Tag())
	}
}

// ValidateJSON validates JSON request body
func ValidateJSON(v *validator.Validate, data interface{}) error {
	return v.Struct(data)
}

// DecodeAndValidateJSON decodes JSON and validates it
func DecodeAndValidateJSON(r *http.Request, v *validator.Validate, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	if err := ValidateJSON(v, dst); err != nil {
		return err
	}

	return nil
}

// RequestSizeLimit middleware limits the size of request bodies
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				log.Warn().
					Int64("content_length", r.ContentLength).
					Int64("max_bytes", maxBytes).
					Msg("request body too large")

				SendJSONError(w, http.StatusRequestEntityTooLarge, "request_too_large", 
					fmt.Sprintf("Request body too large. Maximum size is %d bytes", maxBytes))
				return
			}

			// Limit the request body reader
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeJSON middleware ensures request has JSON content type
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check content type for requests with bodies
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				log.Warn().
					Str("content_type", contentType).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("invalid content type")

				SendJSONError(w, http.StatusUnsupportedMediaType, "invalid_content_type", 
					"Content-Type must be application/json")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// ValidatorExtensions registers custom validation rules
func ValidatorExtensions(v *validator.Validate) {
	// Register custom validation for project tags
	v.RegisterValidation("project_tag", validateProjectTag)
}

// validateProjectTag validates project tag format
func validateProjectTag(fl validator.FieldLevel) bool {
	tag := fl.Field().String()
	
	// Tags must be 1-50 characters, alphanumeric + hyphen/underscore
	if len(tag) < 1 || len(tag) > 50 {
		return false
	}

	for _, char := range tag {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return false
		}
	}

	return true
}