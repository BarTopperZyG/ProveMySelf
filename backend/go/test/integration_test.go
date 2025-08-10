//go:build integration

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/provemyself/backend/internal/config"
	"github.com/provemyself/backend/internal/core"
	"github.com/provemyself/backend/internal/http/handlers"
	"github.com/provemyself/backend/internal/types"
)

// IntegrationTestSuite contains integration tests for the API
type IntegrationTestSuite struct {
	suite.Suite
	server *httptest.Server
	client *http.Client
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Initialize services
	projectService := core.NewProjectService()
	validate := validator.New()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	projectHandler := handlers.NewProjectHandler(projectService, validate)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", healthHandler.GetHealth)

		// Projects
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.ListProjects)
			r.Post("/", projectHandler.CreateProject)
			r.Get("/{projectId}", projectHandler.GetProject)
		})
	})

	// Create test server
	suite.server = httptest.NewServer(r)
	suite.client = suite.server.Client()
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.server.Close()
}

func (suite *IntegrationTestSuite) TestHealthEndpoint() {
	// Act
	resp, err := suite.client.Get(suite.server.URL + "/api/v1/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "application/json", resp.Header.Get("Content-Type"))

	var healthResponse types.HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", healthResponse.Status)
	assert.Equal(suite.T(), "0.1.0", healthResponse.Version)
	assert.NotZero(suite.T(), healthResponse.Timestamp)
}

func (suite *IntegrationTestSuite) TestProjectCRUDFlow() {
	// 1. Create a project
	createReq := types.CreateProjectRequest{
		Title:       "Integration Test Quiz",
		Description: stringPtr("A quiz created during integration testing"),
		Tags:        []string{"integration", "test"},
	}

	body, err := json.Marshal(createReq)
	require.NoError(suite.T(), err)

	resp, err := suite.client.Post(
		suite.server.URL+"/api/v1/projects",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdProject types.ProjectResponse
	err = json.NewDecoder(resp.Body).Decode(&createdProject)
	require.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), createdProject.ID)
	assert.Equal(suite.T(), "Integration Test Quiz", createdProject.Title)
	assert.Equal(suite.T(), "A quiz created during integration testing", *createdProject.Description)
	assert.Equal(suite.T(), []string{"integration", "test"}, createdProject.Tags)

	// 2. Get the project by ID
	resp, err = suite.client.Get(suite.server.URL + "/api/v1/projects/" + createdProject.ID)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var retrievedProject types.ProjectResponse
	err = json.NewDecoder(resp.Body).Decode(&retrievedProject)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), createdProject.ID, retrievedProject.ID)
	assert.Equal(suite.T(), createdProject.Title, retrievedProject.Title)

	// 3. List projects and verify it's included
	resp, err = suite.client.Get(suite.server.URL + "/api/v1/projects")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var projectList types.ProjectListResponse
	err = json.NewDecoder(resp.Body).Decode(&projectList)
	require.NoError(suite.T(), err)

	assert.GreaterOrEqual(suite.T(), projectList.Total, 1)
	assert.GreaterOrEqual(suite.T(), len(projectList.Projects), 1)

	// Check that our project is in the list
	found := false
	for _, project := range projectList.Projects {
		if project.ID == createdProject.ID {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Created project should be in the project list")
}

func (suite *IntegrationTestSuite) TestProjectValidation() {
	// Test with invalid data
	createReq := types.CreateProjectRequest{
		Title: "", // Empty title should fail validation
	}

	body, err := json.Marshal(createReq)
	require.NoError(suite.T(), err)

	resp, err := suite.client.Post(
		suite.server.URL+"/api/v1/projects",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var errorResponse types.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "validation_failed", errorResponse.Error.Code)
}

func (suite *IntegrationTestSuite) TestProjectNotFound() {
	// Try to get a non-existent project
	resp, err := suite.client.Get(suite.server.URL + "/api/v1/projects/nonexistent-id")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	var errorResponse types.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "project_not_found", errorResponse.Error.Code)
}

func (suite *IntegrationTestSuite) TestPagination() {
	// Create multiple projects
	for i := 1; i <= 5; i++ {
		createReq := types.CreateProjectRequest{
			Title: "Pagination Test Quiz " + string(rune('0'+i)),
		}

		body, err := json.Marshal(createReq)
		require.NoError(suite.T(), err)

		resp, err := suite.client.Post(
			suite.server.URL+"/api/v1/projects",
			"application/json",
			bytes.NewReader(body),
		)
		require.NoError(suite.T(), err)
		resp.Body.Close()

		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	}

	// Test pagination
	resp, err := suite.client.Get(suite.server.URL + "/api/v1/projects?limit=2&offset=0")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var projectList types.ProjectListResponse
	err = json.NewDecoder(resp.Body).Decode(&projectList)
	require.NoError(suite.T(), err)

	assert.LessOrEqual(suite.T(), len(projectList.Projects), 2)
	assert.Equal(suite.T(), 2, projectList.Limit)
	assert.Equal(suite.T(), 0, projectList.Offset)
	assert.GreaterOrEqual(suite.T(), projectList.Total, 5)
}

func (suite *IntegrationTestSuite) TestCORSHeaders() {
	req, err := http.NewRequest("OPTIONS", suite.server.URL+"/api/v1/health", nil)
	require.NoError(suite.T(), err)
	
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Contains(suite.T(), resp.Header.Get("Access-Control-Allow-Origins"), "http://localhost:3000")
}

// Run the integration test suite
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationTestSuite))
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}