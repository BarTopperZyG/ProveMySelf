package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/provemyself/backend/internal/types"
)

// ErrorHandler middleware handles panics and errors
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("error", err).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("panic recovered")

				SendJSONError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SendJSONError sends a standardized JSON error response
func SendJSONError(w http.ResponseWriter, statusCode int, code, message string, details ...string) {
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
		// Fallback to plain text error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SendJSONResponse sends a standardized JSON success response
func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode JSON response")
		SendJSONError(w, http.StatusInternalServerError, "encoding_error", "Failed to encode response")
	}
}