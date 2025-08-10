package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

// HealthMiddleware provides health and metrics endpoints
type HealthMiddleware struct {
	startTime time.Time
}

// NewHealthMiddleware creates a new health middleware
func NewHealthMiddleware() *HealthMiddleware {
	return &HealthMiddleware{
		startTime: time.Now(),
	}
}

// SystemMetrics represents system health metrics
type SystemMetrics struct {
	Uptime          string         `json:"uptime"`
	UptimeSeconds   float64        `json:"uptime_seconds"`
	Timestamp       time.Time      `json:"timestamp"`
	Version         string         `json:"version"`
	GoVersion       string         `json:"go_version"`
	NumGoroutines   int            `json:"num_goroutines"`
	Memory          MemoryStats    `json:"memory"`
	GarbageCollector GCStats       `json:"garbage_collector"`
	System          SystemStats    `json:"system"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc        uint64  `json:"alloc_bytes"`
	AllocMB      float64 `json:"alloc_mb"`
	TotalAlloc   uint64  `json:"total_alloc_bytes"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	Sys          uint64  `json:"sys_bytes"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
}

// GCStats represents garbage collection statistics
type GCStats struct {
	LastGC       time.Time `json:"last_gc"`
	NextGC       uint64    `json:"next_gc_bytes"`
	PauseTotal   uint64    `json:"pause_total_ns"`
	NumGC        uint32    `json:"num_gc"`
	GCCPUPercent float64   `json:"gc_cpu_percent"`
}

// SystemStats represents system-level statistics
type SystemStats struct {
	NumCPU       int    `json:"num_cpu"`
	NumCgoCall   int64  `json:"num_cgo_call"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
}

// Metrics returns current system metrics
func (h *HealthMiddleware) Metrics(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)
	metrics := SystemMetrics{
		Uptime:        uptime.String(),
		UptimeSeconds: uptime.Seconds(),
		Timestamp:     time.Now(),
		Version:       "0.1.0", // This should come from build info
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		Memory: MemoryStats{
			Alloc:        m.Alloc,
			AllocMB:      float64(m.Alloc) / 1024 / 1024,
			TotalAlloc:   m.TotalAlloc,
			TotalAllocMB: float64(m.TotalAlloc) / 1024 / 1024,
			Sys:          m.Sys,
			SysMB:        float64(m.Sys) / 1024 / 1024,
			NumGC:        m.NumGC,
		},
		GarbageCollector: GCStats{
			LastGC:       time.Unix(0, int64(m.LastGC)),
			NextGC:       m.NextGC,
			PauseTotal:   m.PauseTotalNs,
			NumGC:        m.NumGC,
			GCCPUPercent: m.GCCPUFraction * 100,
		},
		System: SystemStats{
			NumCPU:     runtime.NumCPU(),
			NumCgoCall: runtime.NumCgoCall(),
			GOOS:       runtime.GOOS,
			GOARCH:     runtime.GOARCH,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Error().Err(err).Msg("failed to encode metrics response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// LivenessProbe provides a simple liveness probe endpoint
func (h *HealthMiddleware) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("failed to encode liveness response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ReadinessProbe provides a readiness probe that can include dependency checks
func (h *HealthMiddleware) ReadinessProbe(dependencies []HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		allHealthy := true
		checks := make(map[string]string)

		// Check all dependencies
		for _, dep := range dependencies {
			if err := dep.HealthCheck(ctx); err != nil {
				checks[dep.Name()] = "unhealthy"
				allHealthy = false
				log.Warn().
					Err(err).
					Str("dependency", dep.Name()).
					Msg("dependency health check failed")
			} else {
				checks[dep.Name()] = "healthy"
			}
		}

		status := "ready"
		statusCode := http.StatusOK
		
		if !allHealthy {
			status = "not_ready"
			statusCode = http.StatusServiceUnavailable
		}

		response := map[string]interface{}{
			"status":    status,
			"timestamp": time.Now(),
			"checks":    checks,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error().Err(err).Msg("failed to encode readiness response")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// HealthChecker interface for dependency health checks
type HealthChecker interface {
	Name() string
	HealthCheck(ctx context.Context) error
}

// DatabaseHealthChecker implements health checking for databases
type DatabaseHealthChecker struct {
	name string
	ping func(context.Context) error
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(name string, pingFunc func(context.Context) error) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name: name,
		ping: pingFunc,
	}
}

// Name returns the name of the health checker
func (d *DatabaseHealthChecker) Name() string {
	return d.name
}

// HealthCheck performs the health check
func (d *DatabaseHealthChecker) HealthCheck(ctx context.Context) error {
	return d.ping(ctx)
}

// StorageHealthChecker implements health checking for storage services
type StorageHealthChecker struct {
	name  string
	check func(context.Context) error
}

// NewStorageHealthChecker creates a new storage health checker
func NewStorageHealthChecker(name string, checkFunc func(context.Context) error) *StorageHealthChecker {
	return &StorageHealthChecker{
		name:  name,
		check: checkFunc,
	}
}

// Name returns the name of the health checker
func (s *StorageHealthChecker) Name() string {
	return s.name
}

// HealthCheck performs the health check
func (s *StorageHealthChecker) HealthCheck(ctx context.Context) error {
	return s.check(ctx)
}