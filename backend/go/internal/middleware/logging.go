package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	// UserIDKey is the context key for user ID (when available)
	UserIDKey contextKey = "user_id"
	// TraceIDKey is the context key for distributed tracing
	TraceIDKey contextKey = "trace_id"
)

// LoggingMiddleware provides enhanced request logging
type LoggingMiddleware struct {
	logger zerolog.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: log.Logger,
	}
}

// RequestLogger logs HTTP requests with detailed context
func (l *LoggingMiddleware) RequestLogger(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger: l.logger})(next)
}

// RequestID adds a unique request ID to each request
func (l *LoggingMiddleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get request ID from header first (for distributed systems)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a new UUID if not provided
			requestID = uuid.New().String()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserContext adds user information to the request context (for authenticated requests)
func (l *LoggingMiddleware) UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This would be populated by authentication middleware
		// For now, we'll skip if no auth headers present
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// StructuredLogger implements chi's LogFormatter interface with structured logging
type StructuredLogger struct {
	logger zerolog.Logger
}

// NewLogEntry creates a new log entry for a request
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{logger: l.logger}
	
	// Extract request context values
	requestID := GetRequestID(r.Context())
	userID := GetUserID(r.Context())

	// Create log event with request context
	logEvent := l.logger.Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Str("remote_addr", getRemoteAddr(r)).
		Str("user_agent", r.UserAgent()).
		Str("request_id", requestID).
		Str("proto", r.Proto)

	// Add user ID if available
	if userID != "" {
		logEvent = logEvent.Str("user_id", userID)
	}

	// Add trace ID if available (for distributed tracing)
	if traceID := GetTraceID(r.Context()); traceID != "" {
		logEvent = logEvent.Str("trace_id", traceID)
	}

	// Log request headers for debugging (excluding sensitive ones)
	if l.logger.GetLevel() <= zerolog.DebugLevel {
		headers := make(map[string]string)
		for name, values := range r.Header {
			if !isSensitiveHeader(name) {
				headers[name] = strings.Join(values, ", ")
			}
		}
		if len(headers) > 0 {
			logEvent = logEvent.Interface("headers", headers)
		}
	}

	logEvent.Msg("request started")
	entry.logger = l.logger
	
	return entry
}

// StructuredLoggerEntry represents a log entry for a single request
type StructuredLoggerEntry struct {
	logger    zerolog.Logger
	startTime time.Time
}

// Write logs the response for a request
func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	logEvent := l.logger.Info().
		Int("status", status).
		Int("bytes", bytes).
		Dur("elapsed", elapsed)

	// Add response headers for debugging
	if l.logger.GetLevel() <= zerolog.DebugLevel {
		responseHeaders := make(map[string]string)
		for name, values := range header {
			if !isSensitiveHeader(name) {
				responseHeaders[name] = strings.Join(values, ", ")
			}
		}
		if len(responseHeaders) > 0 {
			logEvent = logEvent.Interface("response_headers", responseHeaders)
		}
	}

	// Add extra context if provided
	if extra != nil {
		logEvent = logEvent.Interface("extra", extra)
	}

	// Determine log level based on status code
	switch {
	case status >= 500:
		logEvent = l.logger.Error().
			Int("status", status).
			Int("bytes", bytes).
			Dur("elapsed", elapsed)
	case status >= 400:
		logEvent = l.logger.Warn().
			Int("status", status).
			Int("bytes", bytes).
			Dur("elapsed", elapsed)
	}

	logEvent.Msg("request completed")
}

// Panic logs panic information
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.logger.Error().
		Interface("panic", v).
		Bytes("stack", stack).
		Msg("request panic")
}

// PanicRecovery provides panic recovery with detailed logging
func (l *LoggingMiddleware) PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with full context
				requestID := GetRequestID(r.Context())
				userID := GetUserID(r.Context())

				logEvent := l.logger.Error().
					Interface("panic", err).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Str("remote_addr", getRemoteAddr(r)).
					Str("request_id", requestID).
					Bytes("stack", debug.Stack())

				if userID != "" {
					logEvent = logEvent.Str("user_id", userID)
				}

				logEvent.Msg("panic recovered")

				// Return 500 error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Context helper functions

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetTraceID retrieves the trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// Helper functions

// getRemoteAddr returns the real client IP address
func getRemoteAddr(r *http.Request) string {
	// Check for X-Forwarded-For header (load balancer/proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(name string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
		"x-access-token",
	}
	
	lowerName := strings.ToLower(name)
	for _, sensitive := range sensitiveHeaders {
		if lowerName == sensitive {
			return true
		}
	}
	
	return false
}