package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// MetricsCollector collects HTTP metrics
type MetricsCollector struct {
	requestDurations map[string]time.Duration
	requestCounts    map[string]int
	errorCounts      map[string]int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestDurations: make(map[string]time.Duration),
		requestCounts:    make(map[string]int),
		errorCounts:      make(map[string]int),
	}
}

// Metrics middleware collects HTTP metrics
func (mc *MetricsCollector) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Process request
		next.ServeHTTP(ww, r)

		// Calculate metrics
		duration := time.Since(start)
		status := ww.Status()
		method := r.Method
		path := r.URL.Path

		// Create metric key
		key := method + " " + path

		// Update metrics
		mc.requestCounts[key]++
		mc.requestDurations[key] = duration

		// Count errors
		if status >= 400 {
			errorKey := key + " " + strconv.Itoa(status)
			mc.errorCounts[errorKey]++
		}

		// Log performance metrics periodically
		if mc.requestCounts[key]%100 == 0 {
			log.Info().
				Str("method", method).
				Str("path", path).
				Int("count", mc.requestCounts[key]).
				Dur("avg_duration", mc.requestDurations[key]).
				Msg("performance metrics")
		}
	})
}

// GetMetrics returns current metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"request_counts":    mc.requestCounts,
		"request_durations": mc.requestDurations,
		"error_counts":      mc.errorCounts,
	}
}

// HealthMetrics tracks health check metrics
type HealthMetrics struct {
	LastCheck   time.Time `json:"last_check"`
	CheckCount  int       `json:"check_count"`
	FailureCount int      `json:"failure_count"`
	Uptime      string    `json:"uptime"`
}

var (
	healthMetrics = &HealthMetrics{
		LastCheck: time.Now(),
	}
	startTime = time.Now()
)

// UpdateHealthMetrics updates health check metrics
func UpdateHealthMetrics(success bool) {
	healthMetrics.LastCheck = time.Now()
	healthMetrics.CheckCount++
	healthMetrics.Uptime = time.Since(startTime).String()
	
	if !success {
		healthMetrics.FailureCount++
	}
}

// GetHealthMetrics returns current health metrics
func GetHealthMetrics() *HealthMetrics {
	return healthMetrics
}

// PerformanceMonitor logs performance warnings
func PerformanceMonitor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Process request
		next.ServeHTTP(w, r)

		duration := time.Since(start)

		// Log performance warnings
		if duration > 1*time.Second {
			log.Warn().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("duration", duration).
				Msg("slow request")
		}

		if duration > 5*time.Second {
			log.Error().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("duration", duration).
				Msg("very slow request")
		}
	})
}