package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/provemyself/backend/internal/types"
)

// ValidationMiddleware provides enhanced request validation
type ValidationMiddleware struct {
	validator    *validator.Validate
	errorHandler *ErrorHandler
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	validator := validator.New()
	
	// Register custom tag names for better error messages
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &ValidationMiddleware{
		validator:    validator,
		errorHandler: NewErrorHandler(),
	}
}

// ValidateJSON validates JSON request body against a struct
func (v *ValidationMiddleware) ValidateJSON(target interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip validation for GET, DELETE methods
			if r.Method == http.MethodGet || r.Method == http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			// Parse and validate JSON
			if err := json.NewDecoder(r.Body).Decode(target); err != nil {
				log.Warn().
					Err(err).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Msg("failed to decode JSON request")

				v.errorHandler.sendErrorResponse(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format", err.Error())
				return
			}

			// Validate struct
			if err := v.validator.StructCtx(r.Context(), target); err != nil {
				validationErrors := v.formatValidationErrors(err)
				log.Warn().
					Interface("validation_errors", validationErrors).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Msg("request validation failed")

				v.sendValidationErrorResponse(w, validationErrors)
				return
			}

			// Add validated data to request context
			ctx := context.WithValue(r.Context(), "validated_data", target)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ValidateQueryParams validates query parameters
func (v *ValidationMiddleware) ValidateQueryParams(validators map[string]func(string) error) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			var errors []types.ValidationError

			for param, validateFunc := range validators {
				value := query.Get(param)
				if value != "" {
					if err := validateFunc(value); err != nil {
						errors = append(errors, types.ValidationError{
							Field:   param,
							Tag:     "custom",
							Message: err.Error(),
						})
					}
				}
			}

			if len(errors) > 0 {
				log.Warn().
					Interface("validation_errors", errors).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Msg("query parameter validation failed")

				v.sendValidationErrorResponse(w, errors)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// formatValidationErrors converts validator errors to a structured format
func (v *ValidationMiddleware) formatValidationErrors(err error) []types.ValidationError {
	var validationErrors []types.ValidationError

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validationErrs {
			validationErrors = append(validationErrors, types.ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Message: v.getValidationMessage(err),
			})
		}
	}

	return validationErrors
}

// getValidationMessage returns a human-readable error message for validation tags
func (v *ValidationMiddleware) getValidationMessage(err validator.FieldError) string {
	field := err.Field()
	
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, err.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	case "dive":
		return fmt.Sprintf("Invalid item in %s", field)
	default:
		return fmt.Sprintf("%s failed validation (%s)", field, err.Tag())
	}
}

// sendValidationErrorResponse sends a structured validation error response
func (v *ValidationMiddleware) sendValidationErrorResponse(w http.ResponseWriter, errors []types.ValidationError) {
	response := types.ValidationErrorResponse{
		Error: types.ValidationErrorDetail{
			Code:    "validation_failed",
			Message: "Request validation failed",
			Errors:  errors,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("failed to encode validation error response")
		v.errorHandler.InternalError(w, err)
	}
}