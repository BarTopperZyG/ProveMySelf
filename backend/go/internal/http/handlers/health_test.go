package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provemyself/backend/internal/types"
)

func TestHealthHandler_GetHealth(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		validateBody   func(t *testing.T, response types.HealthResponse)
	}{
		{
			name:           "successful health check",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, response types.HealthResponse) {
				assert.Equal(t, "healthy", response.Status)
				assert.Equal(t, "0.1.0", response.Version)
				assert.NotZero(t, response.Timestamp)
				
				require.NotNil(t, response.Services)
				assert.Equal(t, "healthy", response.Services.Database)
				assert.Equal(t, "healthy", response.Services.Storage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler := NewHealthHandler()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			rr := httptest.NewRecorder()

			// Act
			handler.GetHealth(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var response types.HealthResponse
			err := json.NewDecoder(rr.Body).Decode(&response)
			require.NoError(t, err)

			tt.validateBody(t, response)
		})
	}
}

func TestHealthHandler_GetHealth_ContentType(t *testing.T) {
	// Arrange
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	// Act
	handler.GetHealth(rr, req)

	// Assert
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

func TestHealthHandler_GetHealth_ResponseStructure(t *testing.T) {
	// Arrange
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	// Act
	handler.GetHealth(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Check required fields exist
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "version")
	assert.Contains(t, response, "services")

	services, ok := response["services"].(map[string]interface{})
	require.True(t, ok, "services should be an object")
	assert.Contains(t, services, "database")
	assert.Contains(t, services, "storage")
}