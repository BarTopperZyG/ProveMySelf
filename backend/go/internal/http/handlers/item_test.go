package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/provemyself/backend/internal/core"
	"github.com/provemyself/backend/internal/types"
)

// MockItemService is a mock implementation of core.ItemService
type MockItemService struct {
	mock.Mock
}

func (m *MockItemService) Create(ctx context.Context, projectID string, itemType types.ItemType, title string, content interface{}, position int, required bool, points *int, explanation *string) (*core.Item, error) {
	args := m.Called(ctx, projectID, itemType, title, content, position, required, points, explanation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Item), args.Error(1)
}

func (m *MockItemService) GetByID(ctx context.Context, id string) (*core.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Item), args.Error(1)
}

func (m *MockItemService) ListByProject(ctx context.Context, projectID string) ([]*core.Item, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.Item), args.Error(1)
}

func (m *MockItemService) Update(ctx context.Context, id string, itemType types.ItemType, title string, content interface{}, position int, required bool, points *int, explanation *string) (*core.Item, error) {
	args := m.Called(ctx, id, itemType, title, content, position, required, points, explanation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Item), args.Error(1)
}

func (m *MockItemService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestItemHandler_CreateItem(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		requestBody    interface{}
		setupMock      func(*MockItemService)
		expectedStatus int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:      "successful item creation",
			projectID: "test-project-id",
			requestBody: types.CreateItemRequest{
				Type:     types.ItemTypeChoice,
				Title:    "Test Question",
				Position: 0,
				Required: true,
				Points:   intPtr(10),
			},
			setupMock: func(mockService *MockItemService) {
				mockService.On("Create", mock.Anything, "test-project-id", types.ItemTypeChoice, "Test Question", mock.Anything, 0, true, intPtr(10), (*string)(nil)).Return(&core.Item{
					ID:        "test-item-id",
					ProjectID: "test-project-id",
					Type:      types.ItemTypeChoice,
					Title:     "Test Question",
					Position:  0,
					Required:  true,
					Points:    intPtr(10),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var response types.ItemResponse
				require.NoError(t, json.Unmarshal(body, &response))
				assert.Equal(t, "test-item-id", response.ID)
				assert.Equal(t, "test-project-id", response.ProjectID)
				assert.Equal(t, types.ItemTypeChoice, response.Type)
				assert.Equal(t, "Test Question", response.Title)
				assert.Equal(t, 0, response.Position)
				assert.True(t, response.Required)
				assert.Equal(t, 10, *response.Points)
			},
		},
		{
			name:      "invalid request body",
			projectID: "test-project-id",
			requestBody: "invalid json",
			setupMock: func(mockService *MockItemService) {
				// No mock setup needed for this test
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "invalid_request_body", errorResponse.Error.Code)
			},
		},
		{
			name:      "validation failed - title too short",
			projectID: "test-project-id",
			requestBody: types.CreateItemRequest{
				Type:     types.ItemTypeChoice,
				Title:    "", // Invalid: empty title
				Position: 0,
			},
			setupMock: func(mockService *MockItemService) {
				// No mock setup needed for validation errors
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "validation_failed", errorResponse.Error.Code)
			},
		},
		{
			name:      "project not found",
			projectID: "non-existent-project",
			requestBody: types.CreateItemRequest{
				Type:     types.ItemTypeChoice,
				Title:    "Test Question",
				Position: 0,
			},
			setupMock: func(mockService *MockItemService) {
				mockService.On("Create", mock.Anything, "non-existent-project", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return((*core.Item)(nil), core.ErrProjectNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "project_not_found", errorResponse.Error.Code)
			},
		},
		{
			name:      "title too short error",
			projectID: "test-project-id",
			requestBody: types.CreateItemRequest{
				Type:     types.ItemTypeChoice,
				Title:    "Valid Title", // Valid for validation, but service will return error
				Position: 0,
			},
			setupMock: func(mockService *MockItemService) {
				mockService.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return((*core.Item)(nil), core.ErrItemTitleTooShort)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "title_too_short", errorResponse.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.setupMock(mockService)

			handler := NewItemHandler(mockService, validator.New())

			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/{projectId}/items", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Setup chi context with projectId parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("projectId", tt.projectID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.CreateItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_ListItems(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		setupMock      func(*MockItemService)
		expectedStatus int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:      "successful list",
			projectID: "test-project-id",
			setupMock: func(mockService *MockItemService) {
				items := []*core.Item{
					{
						ID:        "item1",
						ProjectID: "test-project-id",
						Type:      types.ItemTypeChoice,
						Title:     "Question 1",
						Position:  0,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        "item2",
						ProjectID: "test-project-id",
						Type:      types.ItemTypeTitle,
						Title:     "Title Block",
						Position:  1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				mockService.On("ListByProject", mock.Anything, "test-project-id").Return(items, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response types.ItemListResponse
				require.NoError(t, json.Unmarshal(body, &response))
				assert.Equal(t, 2, response.Total)
				assert.Len(t, response.Items, 2)
				assert.Equal(t, "item1", response.Items[0].ID)
				assert.Equal(t, "item2", response.Items[1].ID)
				assert.Equal(t, "test-project-id", response.ProjectID)
			},
		},
		{
			name:      "project not found",
			projectID: "non-existent-project",
			setupMock: func(mockService *MockItemService) {
				mockService.On("ListByProject", mock.Anything, "non-existent-project").Return(([]*core.Item)(nil), core.ErrProjectNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "project_not_found", errorResponse.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.setupMock(mockService)

			handler := NewItemHandler(mockService, validator.New())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/{projectId}/items", nil)
			
			// Setup chi context with projectId parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("projectId", tt.projectID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.ListItems(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_GetItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		setupMock      func(*MockItemService)
		expectedStatus int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:   "successful get",
			itemID: "test-item-id",
			setupMock: func(mockService *MockItemService) {
				item := &core.Item{
					ID:        "test-item-id",
					ProjectID: "test-project-id",
					Type:      types.ItemTypeChoice,
					Title:     "Test Question",
					Position:  0,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockService.On("GetByID", mock.Anything, "test-item-id").Return(item, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response types.ItemResponse
				require.NoError(t, json.Unmarshal(body, &response))
				assert.Equal(t, "test-item-id", response.ID)
				assert.Equal(t, "test-project-id", response.ProjectID)
				assert.Equal(t, types.ItemTypeChoice, response.Type)
			},
		},
		{
			name:   "item not found",
			itemID: "non-existent-item",
			setupMock: func(mockService *MockItemService) {
				mockService.On("GetByID", mock.Anything, "non-existent-item").Return((*core.Item)(nil), core.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "item_not_found", errorResponse.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.setupMock(mockService)

			handler := NewItemHandler(mockService, validator.New())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/{projectId}/items/{itemId}", nil)
			
			// Setup chi context with itemId parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("itemId", tt.itemID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_UpdateItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		requestBody    interface{}
		setupMock      func(*MockItemService)
		expectedStatus int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:   "successful update",
			itemID: "test-item-id",
			requestBody: types.UpdateItemRequest{
				Type:     types.ItemTypeChoice,
				Title:    "Updated Question",
				Position: 1,
				Required: false,
			},
			setupMock: func(mockService *MockItemService) {
				updatedItem := &core.Item{
					ID:        "test-item-id",
					ProjectID: "test-project-id",
					Type:      types.ItemTypeChoice,
					Title:     "Updated Question",
					Position:  1,
					Required:  false,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockService.On("Update", mock.Anything, "test-item-id", types.ItemTypeChoice, "Updated Question", mock.Anything, 1, false, (*int)(nil), (*string)(nil)).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response types.ItemResponse
				require.NoError(t, json.Unmarshal(body, &response))
				assert.Equal(t, "test-item-id", response.ID)
				assert.Equal(t, "Updated Question", response.Title)
				assert.Equal(t, 1, response.Position)
				assert.False(t, response.Required)
			},
		},
		{
			name:   "item not found",
			itemID: "non-existent-item",
			requestBody: types.UpdateItemRequest{
				Type:     types.ItemTypeChoice,
				Title:    "Updated Question",
				Position: 0,
			},
			setupMock: func(mockService *MockItemService) {
				mockService.On("Update", mock.Anything, "non-existent-item", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return((*core.Item)(nil), core.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "item_not_found", errorResponse.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.setupMock(mockService)

			handler := NewItemHandler(mockService, validator.New())

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/{projectId}/items/{itemId}", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Setup chi context with itemId parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("itemId", tt.itemID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.UpdateItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_DeleteItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		setupMock      func(*MockItemService)
		expectedStatus int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:   "successful delete",
			itemID: "test-item-id",
			setupMock: func(mockService *MockItemService) {
				mockService.On("Delete", mock.Anything, "test-item-id").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "item not found",
			itemID: "non-existent-item",
			setupMock: func(mockService *MockItemService) {
				mockService.On("Delete", mock.Anything, "non-existent-item").Return(core.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, body []byte) {
				var errorResponse types.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResponse))
				assert.Equal(t, "item_not_found", errorResponse.Error.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.setupMock(mockService)

			handler := NewItemHandler(mockService, validator.New())

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/{projectId}/items/{itemId}", nil)
			
			// Setup chi context with itemId parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("itemId", tt.itemID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.DeleteItem(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, rr.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}