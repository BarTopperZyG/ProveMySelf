package types

import "time"

// CreateProjectRequest represents a request to create a new project
type CreateProjectRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	Tags        []string `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
}

// UpdateProjectRequest represents a request to update an existing project
type UpdateProjectRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	Tags        []string `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// ProjectListResponse represents a paginated list of projects
type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
}