// Package middleware provides HTTP middleware components for the ProveMySelf API.
// This package includes error handling, validation, logging, and health monitoring middleware.
package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/provemyself/backend/internal/types"
)

// ErrorHandler provides standardized error handling middleware
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler middleware
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// Recovery middleware with standardized error responses
func (e *ErrorHandler) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("panic", err).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Str("remote_addr", r.RemoteAddr).
					Msg("panic recovered")

				e.sendErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ValidationError handles validation errors with detailed messages
func (e *ErrorHandler) ValidationError(w http.ResponseWriter, err error) {
	log.Warn().
		Err(err).
		Msg("validation error")

	e.sendErrorResponse(w, http.StatusBadRequest, "validation_error", "Request validation failed", err.Error())
}

// NotFoundError handles 404 errors
func (e *ErrorHandler) NotFoundError(w http.ResponseWriter, resource string) {
	e.sendErrorResponse(w, http.StatusNotFound, "not_found", resource+" not found")
}

// UnauthorizedError handles 401 errors
func (e *ErrorHandler) UnauthorizedError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentication required"
	}
	e.sendErrorResponse(w, http.StatusUnauthorized, "unauthorized", message)
}

// ForbiddenError handles 403 errors
func (e *ErrorHandler) ForbiddenError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	e.sendErrorResponse(w, http.StatusForbidden, "forbidden", message)
}

// ConflictError handles 409 errors
func (e *ErrorHandler) ConflictError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	e.sendErrorResponse(w, http.StatusConflict, "conflict", message)
}

// InternalError handles 500 errors
func (e *ErrorHandler) InternalError(w http.ResponseWriter, err error) {
	log.Error().
		Err(err).
		Msg("internal server error")

	e.sendErrorResponse(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred")
}

// sendErrorResponse sends a standardized error response
func (e *ErrorHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details ...string) {
	var detailsPtr *string
	if len(details) > 0 {
		detailsPtr = &details[0]
	}

	errorResponse := types.ErrorResponse{
		Error: types.ErrorDetail{
			Code:    code,
			Message: message,
			Details: detailsPtr,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Error().Err(err).Msg("failed to encode error response")
		// Fallback to plain text response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}