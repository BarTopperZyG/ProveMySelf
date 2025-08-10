package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// SecurityHeaders middleware adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		
		// Require HTTPS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", 
			"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
		
		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Feature Policy / Permissions Policy
		w.Header().Set("Permissions-Policy", 
			"camera=(), microphone=(), geolocation=(), payment=()")

		next.ServeHTTP(w, r)
	})
}

// RateLimiter represents a simple rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimit middleware implements rate limiting per IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		now := time.Now()

		// Clean old requests
		rl.cleanOldRequests(ip, now)

		// Check rate limit
		if len(rl.requests[ip]) >= rl.limit {
			log.Warn().
				Str("ip", ip).
				Int("requests", len(rl.requests[ip])).
				Int("limit", rl.limit).
				Msg("rate limit exceeded")

			SendJSONError(w, http.StatusTooManyRequests, "rate_limited", 
				"Rate limit exceeded. Please try again later.")
			return
		}

		// Record request
		rl.requests[ip] = append(rl.requests[ip], now)

		next.ServeHTTP(w, r)
	})
}

// cleanOldRequests removes requests outside the time window
func (rl *RateLimiter) cleanOldRequests(ip string, now time.Time) {
	requests := rl.requests[ip]
	cutoff := now.Add(-rl.window)
	
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	
	rl.requests[ip] = validRequests
}

// getClientIP extracts the real client IP from request headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	return ip
}

// AuthContextKey represents keys used in authentication context
type AuthContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey AuthContextKey = "user_id"
	// UserRoleKey is the context key for user role
	UserRoleKey AuthContextKey = "user_role"
	// UserEmailKey is the context key for user email
	UserEmailKey AuthContextKey = "user_email"
)

// User represents an authenticated user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// AuthenticateJWT middleware validates JWT tokens (skeleton implementation)
func AuthenticateJWT(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				SendJSONError(w, http.StatusUnauthorized, "missing_token", "Authorization header required")
				return
			}

			// Check Bearer prefix
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				SendJSONError(w, http.StatusUnauthorized, "invalid_token_format", "Token must be prefixed with 'Bearer '")
				return
			}

			token := strings.TrimPrefix(authHeader, bearerPrefix)
			
			// TODO: Implement actual JWT validation
			// For now, this is a skeleton that accepts any non-empty token in development
			if token == "" {
				SendJSONError(w, http.StatusUnauthorized, "empty_token", "Token cannot be empty")
				return
			}

			// Mock user for development (replace with actual JWT parsing)
			user := &User{
				ID:    "dev-user-123",
				Email: "dev@example.com",
				Role:  "admin",
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
			ctx = context.WithValue(ctx, UserEmailKey, user.Email)
			ctx = context.WithValue(ctx, UserRoleKey, user.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware allows both authenticated and anonymous requests
func OptionalAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			
			// If no auth header, continue without authentication
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			// If auth header present, validate it
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				next.ServeHTTP(w, r) // Continue without auth for malformed headers
				return
			}

			token := strings.TrimPrefix(authHeader, bearerPrefix)
			if token == "" {
				next.ServeHTTP(w, r) // Continue without auth for empty tokens
				return
			}

			// TODO: Implement actual JWT validation
			// Mock user for development
			user := &User{
				ID:    "dev-user-123",
				Email: "dev@example.com",
				Role:  "admin",
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
			ctx = context.WithValue(ctx, UserEmailKey, user.Email)
			ctx = context.WithValue(ctx, UserRoleKey, user.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware requires a specific user role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if userRole == "" {
				SendJSONError(w, http.StatusUnauthorized, "authentication_required", "Authentication required")
				return
			}

			if userRole != role && userRole != "admin" { // Admin can access everything
				SendJSONError(w, http.StatusForbidden, "insufficient_permissions", 
					"Insufficient permissions for this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions to extract user information from context

// GetUserID returns the user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserEmail returns the user email from context
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

// GetUserRole returns the user role from context
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

// IsAuthenticated returns true if the request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != ""
}

// CORS middleware with configurable options
func CORS(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range origins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}