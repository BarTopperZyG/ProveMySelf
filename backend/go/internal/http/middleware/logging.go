package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ContextKey represents keys used in request context
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// LoggerKey is the context key for the request logger
	LoggerKey ContextKey = "logger"
)

// RequestID middleware adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StructuredLogger middleware provides structured logging for HTTP requests
func StructuredLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context
			requestID := GetRequestID(r.Context())

			// Create request-specific logger
			reqLogger := logger.With().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Logger()

			// Add logger to context
			ctx := context.WithValue(r.Context(), LoggerKey, &reqLogger)

			// Wrap response writer to capture status code and size
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Log request start
			reqLogger.Info().
				Str("query", r.URL.RawQuery).
				Int64("content_length", r.ContentLength).
				Msg("request started")

			// Process request
			next.ServeHTTP(ww, r.WithContext(ctx))

			// Calculate duration
			duration := time.Since(start)

			// Log request completion
			reqLogger.Info().
				Int("status", ww.Status()).
				Int("bytes_written", ww.BytesWritten()).
				Dur("duration", duration).
				Msg("request completed")

			// Log slow requests as warnings
			if duration > 5*time.Second {
				reqLogger.Warn().
					Dur("duration", duration).
					Msg("slow request detected")
			}

			// Log client errors as warnings
			if ww.Status() >= 400 && ww.Status() < 500 {
				reqLogger.Warn().
					Int("status", ww.Status()).
					Msg("client error")
			}

			// Log server errors as errors
			if ww.Status() >= 500 {
				reqLogger.Error().
					Int("status", ww.Status()).
					Msg("server error")
			}
		})
	}
}

// GetRequestID returns the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetLogger returns the request logger from context
func GetLogger(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*zerolog.Logger); ok {
		return logger
	}
	return &log.Logger
}

// LogError logs an error with context information
func LogError(ctx context.Context, err error, message string) {
	logger := GetLogger(ctx)
	logger.Error().
		Err(err).
		Str("request_id", GetRequestID(ctx)).
		Msg(message)
}

// LogWarn logs a warning with context information
func LogWarn(ctx context.Context, message string) {
	logger := GetLogger(ctx)
	logger.Warn().
		Str("request_id", GetRequestID(ctx)).
		Msg(message)
}

// LogInfo logs an info message with context information
func LogInfo(ctx context.Context, message string) {
	logger := GetLogger(ctx)
	logger.Info().
		Str("request_id", GetRequestID(ctx)).
		Msg(message)
}

// LogAudit logs an audit event
func LogAudit(ctx context.Context, action, resource string, details map[string]interface{}) {
	logger := GetLogger(ctx)
	event := logger.Info().
		Str("request_id", GetRequestID(ctx)).
		Str("audit_action", action).
		Str("audit_resource", resource)

	// Add details
	for key, value := range details {
		event = event.Interface(key, value)
	}

	event.Msg("audit event")
}

// Security middleware logs security-related events
func SecurityLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log suspicious patterns
			if len(r.URL.Path) > 1000 {
				logger.Warn().
					Str("request_id", GetRequestID(r.Context())).
					Str("remote_addr", r.RemoteAddr).
					Str("path", r.URL.Path).
					Msg("unusually long path detected")
			}

			// Log potential injection attempts
			query := r.URL.RawQuery
			suspiciousPatterns := []string{"<script", "javascript:", "SELECT ", "UNION ", "DROP ", "INSERT "}
			for _, pattern := range suspiciousPatterns {
				if contains(query, pattern) || contains(r.URL.Path, pattern) {
					logger.Warn().
						Str("request_id", GetRequestID(r.Context())).
						Str("remote_addr", r.RemoteAddr).
						Str("path", r.URL.Path).
						Str("query", query).
						Str("pattern", pattern).
						Msg("potential injection attempt detected")
					break
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	// Simple case-insensitive contains check
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// indexOf finds the index of a substring
func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}