package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/provemyself/backend/internal/core"
	"github.com/provemyself/backend/internal/types"
)

// MockProjectService is a mock implementation of core.ProjectService
type MockProjectService struct {
	mock.Mock
}

func (m *MockProjectService) Create(ctx context.Context, title string, description *string, tags []string) (*core.Project, error) {
	args := m.Called(ctx, title, description, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Project), args.Error(1)
}

func (m *MockProjectService) GetByID(ctx context.Context, id string) (*core.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Project), args.Error(1)
}

func (m *MockProjectService) List(ctx context.Context, limit, offset int) ([]*core.Project, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*core.Project), args.Int(1), args.Error(2)
}

func TestProjectHandler_CreateProject(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    types.CreateProjectRequest
		mockSetup      func(m *MockProjectService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "successful project creation",
			requestBody: types.CreateProjectRequest{
				Title:       "Test Quiz",
				Description: stringPtr("A test quiz"),
				Tags:        []string{"test", "quiz"},
			},
			mockSetup: func(m *MockProjectService) {
				m.On("Create", mock.Anything, "Test Quiz", stringPtr("A test quiz"), []string{"test", "quiz"}).
					Return(&core.Project{
						ID:          "test-id-123",
						Title:       "Test Quiz",
						Description: stringPtr("A test quiz"),
						Tags:        []string{"test", "quiz"},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body []byte) {
				var response types.ProjectResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Equal(t, "test-id-123", response.ID)
				assert.Equal(t, "Test Quiz", response.Title)
				assert.Equal(t, "A test quiz", *response.Description)
				assert.Equal(t, []string{"test", "quiz"}, response.Tags)
			},
		},
		{
			name: "validation error - empty title",
			requestBody: types.CreateProjectRequest{
				Title: "",
			},
			mockSetup: func(m *MockProjectService) {
				// No mock setup needed as validation should fail first
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response types.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Equal(t, "validation_failed", response.Error.Code)
				assert.Contains(t, response.Error.Message, "Validation failed")
			},
		},
		{
			name: "service error - title too short",
			requestBody: types.CreateProjectRequest{
				Title: "A",
			},
			mockSetup: func(m *MockProjectService) {
				m.On("Create", mock.Anything, "A", (*string)(nil), []string(nil)).
					Return(nil, core.ErrProjectTitleTooShort)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body []byte) {
				var response types.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Equal(t, "title_too_short", response.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockProjectService)
			tt.mockSetup(mockService)

			handler := NewProjectHandler(mockService, validator.New())

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Act
			handler.CreateProject(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.validateBody(t, rr.Body.Bytes())

			mockService.AssertExpectations(t)
		})
	}
}

func TestProjectHandler_GetProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		mockSetup      func(m *MockProjectService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name:      "successful project retrieval",
			projectID: "test-id-123",
			mockSetup: func(m *MockProjectService) {
				m.On("GetByID", mock.Anything, "test-id-123").
					Return(&core.Project{
						ID:    "test-id-123",
						Title: "Test Quiz",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response types.ProjectResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Equal(t, "test-id-123", response.ID)
				assert.Equal(t, "Test Quiz", response.Title)
			},
		},
		{
			name:      "project not found",
			projectID: "nonexistent",
			mockSetup: func(m *MockProjectService) {
				m.On("GetByID", mock.Anything, "nonexistent").
					Return(nil, core.ErrProjectNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body []byte) {
				var response types.ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Equal(t, "project_not_found", response.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockProjectService)
			tt.mockSetup(mockService)

			handler := NewProjectHandler(mockService, validator.New())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+tt.projectID, nil)
			rr := httptest.NewRecorder()

			// Set up Chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("projectId", tt.projectID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Act
			handler.GetProject(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.validateBody(t, rr.Body.Bytes())

			mockService.AssertExpectations(t)
		})
	}
}

func TestProjectHandler_ListProjects(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(m *MockProjectService)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name:        "successful project listing with defaults",
			queryParams: "",
			mockSetup: func(m *MockProjectService) {
				projects := []*core.Project{
					{ID: "1", Title: "Quiz 1"},
					{ID: "2", Title: "Quiz 2"},
				}
				m.On("List", mock.Anything, 20, 0).
					Return(projects, 2, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response types.ProjectListResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Len(t, response.Projects, 2)
				assert.Equal(t, 2, response.Total)
				assert.Equal(t, 20, response.Limit)
				assert.Equal(t, 0, response.Offset)
			},
		},
		{
			name:        "successful project listing with pagination",
			queryParams: "?limit=10&offset=5",
			mockSetup: func(m *MockProjectService) {
				projects := []*core.Project{
					{ID: "6", Title: "Quiz 6"},
				}
				m.On("List", mock.Anything, 10, 5).
					Return(projects, 50, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response types.ProjectListResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Len(t, response.Projects, 1)
				assert.Equal(t, 50, response.Total)
				assert.Equal(t, 10, response.Limit)
				assert.Equal(t, 5, response.Offset)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockProjectService)
			tt.mockSetup(mockService)

			handler := NewProjectHandler(mockService, validator.New())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/projects"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			// Act
			handler.ListProjects(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.validateBody(t, rr.Body.Bytes())

			mockService.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}