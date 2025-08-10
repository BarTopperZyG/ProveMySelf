package types

import "time"

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string              `json:"status"`
	Timestamp time.Time           `json:"timestamp"`
	Version   string              `json:"version"`
	Services  *HealthServices     `json:"services,omitempty"`
}

// HealthServices represents the status of dependent services
type HealthServices struct {
	Database string `json:"database,omitempty"`
	Storage  string `json:"storage,omitempty"`
}