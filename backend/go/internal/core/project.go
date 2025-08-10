// Package core contains the business logic and domain models for the ProveMySelf API.
// It defines the core entities, interfaces, and business rules that are independent
// of external dependencies like databases, HTTP frameworks, or third-party services.
//
// The core package follows Clean Architecture principles:
// - Entities: Core business objects (Project)
// - Use Cases: Business logic implementation (ProjectService)
// - Repository Interfaces: Data access contracts (ProjectStore)
//
// This package should have no dependencies on external frameworks or infrastructure.
package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain errors for project operations.
// These errors represent business rule violations and are independent
// of infrastructure concerns.
var (
	// ErrProjectNotFound is returned when a project with the given ID doesn't exist.
	ErrProjectNotFound = errors.New("project not found")
	
	// ErrProjectTitleTooShort is returned when a project title is shorter than the minimum length.
	ErrProjectTitleTooShort = errors.New("project title too short")
	
	// ErrProjectTitleTooLong is returned when a project title exceeds the maximum length.
	ErrProjectTitleTooLong = errors.New("project title too long")
)

// Project represents a quiz project entity in the ProveMySelf platform.
// It contains all the metadata for a quiz project, including title, description,
// tags for categorization, and timestamps for lifecycle management.
//
// Business Rules:
// - Title must be between 1 and 200 characters
// - Description is optional and can be up to 1000 characters
// - Tags are optional, maximum 10 tags, each tag max 50 characters
// - Projects can be published at most once (PublishedAt is immutable once set)
// - CreatedAt and UpdatedAt are managed automatically
type Project struct {
	// ID is the unique identifier for the project (UUID format).
	ID string
	
	// Title is the human-readable name of the project.
	// Required field, must be 1-200 characters.
	Title string
	
	// Description provides detailed information about the project.
	// Optional field, maximum 1000 characters when present.
	Description *string
	
	// Tags are labels used for categorizing and filtering projects.
	// Optional field, maximum 10 tags, each tag maximum 50 characters.
	Tags []string
	
	// CreatedAt is the timestamp when the project was first created.
	// Set automatically and never changes.
	CreatedAt time.Time
	
	// UpdatedAt is the timestamp when the project was last modified.
	// Updated automatically on any change to the project.
	UpdatedAt time.Time
	
	// PublishedAt is the timestamp when the project was published.
	// Nil until the project is published, then immutable once set.
	PublishedAt *time.Time
}

// ProjectStore defines the contract for project data persistence.
// This interface abstracts the data layer, allowing different implementations
// (PostgreSQL, MongoDB, in-memory, etc.) without changing business logic.
//
// All methods should be safe for concurrent use and handle context cancellation.
type ProjectStore interface {
	// Create persists a new project with the given parameters.
	// Returns the created project with generated ID and timestamps.
	// Returns domain errors for validation failures.
	Create(ctx context.Context, title string, description *string, tags []string) (*Project, error)
	
	// GetByID retrieves a project by its unique identifier.
	// Returns ErrProjectNotFound if the project doesn't exist.
	GetByID(ctx context.Context, id string) (*Project, error)
	
	// List retrieves a paginated list of projects ordered by creation date (desc).
	// Returns the projects slice, total count, and any error.
	// Limit and offset are used for pagination.
	List(ctx context.Context, limit, offset int) ([]*Project, int, error)
	
	// Update modifies an existing project with new values.
	// Returns the updated project with new UpdatedAt timestamp.
	// Returns ErrProjectNotFound if the project doesn't exist.
	Update(ctx context.Context, id string, title string, description *string, tags []string) (*Project, error)
	
	// Delete permanently removes a project from storage.
	// Returns ErrProjectNotFound if the project doesn't exist.
	Delete(ctx context.Context, id string) error
	
	// Publish marks a project as published by setting PublishedAt timestamp.
	// Can only be called once per project (PublishedAt is immutable).
	// Returns ErrProjectNotFound if the project doesn't exist.
	Publish(ctx context.Context, id string) (*Project, error)
	
	// SearchByTitle finds projects by searching title and description fields.
	// Returns paginated results matching the search term (case-insensitive).
	SearchByTitle(ctx context.Context, searchTerm string, limit, offset int) ([]*Project, int, error)
}

// ProjectService implements the use cases for project management.
// It encapsulates the business logic and orchestrates operations between
// the domain entities and the data access layer.
//
// The service validates business rules, coordinates transactions, and
// ensures data consistency. It's the primary entry point for all
// project-related operations in the application.
type ProjectService struct {
	// store provides data persistence capabilities for projects.
	store ProjectStore
}

// NewProjectService creates a new project service
func NewProjectService(store ProjectStore) *ProjectService {
	return &ProjectService{
		store: store,
	}
}

// Create creates a new project
func (s *ProjectService) Create(ctx context.Context, title string, description *string, tags []string) (*Project, error) {
	if len(title) < 1 {
		return nil, ErrProjectTitleTooShort
	}
	if len(title) > 200 {
		return nil, ErrProjectTitleTooLong
	}

	// Validate tags
	if len(tags) > 10 {
		return nil, fmt.Errorf("too many tags: maximum 10 allowed, got %d", len(tags))
	}
	for _, tag := range tags {
		if len(tag) > 50 {
			return nil, fmt.Errorf("tag too long: maximum 50 characters, got %d", len(tag))
		}
	}

	return s.store.Create(ctx, title, description, tags)
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(ctx context.Context, id string) (*Project, error) {
	return s.store.GetByID(ctx, id)
}

// List retrieves projects with pagination
func (s *ProjectService) List(ctx context.Context, limit, offset int) ([]*Project, int, error) {
	return s.store.List(ctx, limit, offset)
}

// Update updates a project
func (s *ProjectService) Update(ctx context.Context, id string, title string, description *string, tags []string) (*Project, error) {
	if len(title) < 1 {
		return nil, ErrProjectTitleTooShort
	}
	if len(title) > 200 {
		return nil, ErrProjectTitleTooLong
	}

	// Validate tags
	if len(tags) > 10 {
		return nil, fmt.Errorf("too many tags: maximum 10 allowed, got %d", len(tags))
	}
	for _, tag := range tags {
		if len(tag) > 50 {
			return nil, fmt.Errorf("tag too long: maximum 50 characters, got %d", len(tag))
		}
	}

	return s.store.Update(ctx, id, title, description, tags)
}

// Delete deletes a project
func (s *ProjectService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// Publish publishes a project
func (s *ProjectService) Publish(ctx context.Context, id string) (*Project, error) {
	return s.store.Publish(ctx, id)
}

// SearchByTitle searches projects by title
func (s *ProjectService) SearchByTitle(ctx context.Context, searchTerm string, limit, offset int) ([]*Project, int, error) {
	return s.store.SearchByTitle(ctx, searchTerm, limit, offset)
}