package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/provemyself/backend/internal/core"
	"github.com/provemyself/backend/internal/types"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	service  *core.ProjectService
	validate *validator.Validate
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service *core.ProjectService, validate *validator.Validate) *ProjectHandler {
	return &ProjectHandler{
		service:  service,
		validate: validate,
	}
}

// ListProjects handles GET /api/v1/projects
// @Summary List projects
// @Description Retrieve a list of quiz projects for the authenticated user
// @Tags Projects
// @Param limit query int false "Maximum number of projects to return" minimum(1) maximum(100) default(20)
// @Param offset query int false "Number of projects to skip" minimum(0) default(0)
// @Produce json
// @Success 200 {object} types.ProjectListResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects [get]
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Parse query parameters
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get projects from service
	projects, total, err := h.service.List(ctx, limit, offset)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list projects")
		h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to list projects")
		return
	}

	// Convert to response format
	projectResponses := make([]types.ProjectResponse, len(projects))
	for i, project := range projects {
		projectResponses[i] = types.ProjectResponse{
			ID:          project.ID,
			Title:       project.Title,
			Description: project.Description,
			Tags:        project.Tags,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
			PublishedAt: project.PublishedAt,
		}
	}

	response := types.ProjectListResponse{
		Projects: projectResponses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// CreateProject handles POST /api/v1/projects
// @Summary Create project
// @Description Create a new quiz project
// @Tags Projects
// @Accept json
// @Produce json
// @Param request body types.CreateProjectRequest true "Project creation request"
// @Success 201 {object} types.ProjectResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 422 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req types.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode request")
		h.sendJSONError(w, http.StatusBadRequest, "invalid_request_body", "Invalid request body")
		return
	}

	if err := h.validate.StructCtx(ctx, req); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("validation failed")
		h.sendJSONError(w, http.StatusBadRequest, "validation_failed", "Validation failed", err.Error())
		return
	}

	project, err := h.service.Create(ctx, req.Title, req.Description, req.Tags)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create project")
		
		switch {
		case errors.Is(err, core.ErrProjectTitleTooShort):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_short", "Project title is too short")
		case errors.Is(err, core.ErrProjectTitleTooLong):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_long", "Project title is too long")
		default:
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to create project")
		}
		return
	}

	response := types.ProjectResponse{
		ID:          project.ID,
		Title:       project.Title,
		Description: project.Description,
		Tags:        project.Tags,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		PublishedAt: project.PublishedAt,
	}

	h.sendJSONResponse(w, http.StatusCreated, response)
}

// GetProject handles GET /api/v1/projects/{projectId}
// @Summary Get project
// @Description Retrieve a specific project by ID
// @Tags Projects
// @Param projectId path string true "Project ID" format(uuid)
// @Produce json
// @Success 200 {object} types.ProjectResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	project, err := h.service.GetByID(ctx, projectID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to get project")
		
		if errors.Is(err, core.ErrProjectNotFound) {
			h.sendJSONError(w, http.StatusNotFound, "project_not_found", "Project not found")
		} else {
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to get project")
		}
		return
	}

	response := types.ProjectResponse{
		ID:          project.ID,
		Title:       project.Title,
		Description: project.Description,
		Tags:        project.Tags,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		PublishedAt: project.PublishedAt,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// UpdateProject handles PUT /api/v1/projects/{projectId}
// @Summary Update project
// @Description Update an existing project
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param request body types.UpdateProjectRequest true "Project update request"
// @Success 200 {object} types.ProjectResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 422 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId} [put]
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	var req types.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode request")
		h.sendJSONError(w, http.StatusBadRequest, "invalid_request_body", "Invalid request body")
		return
	}

	if err := h.validate.StructCtx(ctx, req); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("validation failed")
		h.sendJSONError(w, http.StatusBadRequest, "validation_failed", "Validation failed", err.Error())
		return
	}

	project, err := h.service.Update(ctx, projectID, req.Title, req.Description, req.Tags)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to update project")
		
		switch {
		case errors.Is(err, core.ErrProjectNotFound):
			h.sendJSONError(w, http.StatusNotFound, "project_not_found", "Project not found")
		case errors.Is(err, core.ErrProjectTitleTooShort):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_short", "Project title is too short")
		case errors.Is(err, core.ErrProjectTitleTooLong):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_long", "Project title is too long")
		default:
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to update project")
		}
		return
	}

	response := types.ProjectResponse{
		ID:          project.ID,
		Title:       project.Title,
		Description: project.Description,
		Tags:        project.Tags,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		PublishedAt: project.PublishedAt,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// DeleteProject handles DELETE /api/v1/projects/{projectId}
// @Summary Delete project
// @Description Delete a project by ID
// @Tags Projects
// @Param projectId path string true "Project ID" format(uuid)
// @Success 204 "Project deleted successfully"
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId} [delete]
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	err := h.service.Delete(ctx, projectID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to delete project")
		
		if errors.Is(err, core.ErrProjectNotFound) {
			h.sendJSONError(w, http.StatusNotFound, "project_not_found", "Project not found")
		} else {
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to delete project")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PublishProject handles POST /api/v1/projects/{projectId}/publish
// @Summary Publish project
// @Description Mark a project as published
// @Tags Projects
// @Param projectId path string true "Project ID" format(uuid)
// @Produce json
// @Success 200 {object} types.ProjectResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 409 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/publish [post]
func (h *ProjectHandler) PublishProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	project, err := h.service.Publish(ctx, projectID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to publish project")
		
		if errors.Is(err, core.ErrProjectNotFound) {
			h.sendJSONError(w, http.StatusNotFound, "project_not_found", "Project not found")
		} else {
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to publish project")
		}
		return
	}

	response := types.ProjectResponse{
		ID:          project.ID,
		Title:       project.Title,
		Description: project.Description,
		Tags:        project.Tags,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		PublishedAt: project.PublishedAt,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// Helper methods for consistent JSON responses

func (h *ProjectHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode JSON response")
	}
}

func (h *ProjectHandler) sendJSONError(w http.ResponseWriter, statusCode int, code, message string, details ...string) {
	var detailsPtr *string
	if len(details) > 0 {
		detailsPtr = &details[0]
	}

	errorResponse := types.ErrorResponse{
		Error: types.ErrorDetail{
			Code:    code,
			Message: message,
			Details: detailsPtr,
		},
	}

	h.sendJSONResponse(w, statusCode, errorResponse)
}