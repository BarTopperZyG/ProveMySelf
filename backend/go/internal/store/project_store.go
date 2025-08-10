package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/provemyself/backend/internal/core"
)

// ProjectStore implements project data access using PostgreSQL
type ProjectStore struct {
	db *Database
}

// NewProjectStore creates a new project store
func NewProjectStore(db *Database) *ProjectStore {
	return &ProjectStore{db: db}
}

// Create creates a new project in the database
func (s *ProjectStore) Create(ctx context.Context, title string, description *string, tags []string) (*core.Project, error) {
	var project core.Project

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO projects (title, description, tags)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, tags, created_at, updated_at, published_at
	`

	row := s.db.DB().QueryRowContext(ctx, query, title, description, tagsJSON)

	var tagsRaw []byte
	err = row.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&tagsRaw,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.PublishedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23514": // check_violation
				if pqErr.Constraint == "projects_title_check" {
					return nil, core.ErrProjectTitleTooShort
				}
			case "23505": // unique_violation
				return nil, fmt.Errorf("project already exists: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Unmarshal tags
	if err := json.Unmarshal(tagsRaw, &project.Tags); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal project tags")
		project.Tags = []string{} // Fallback to empty slice
	}

	log.Info().
		Str("project_id", project.ID).
		Str("title", project.Title).
		Msg("project created successfully")

	return &project, nil
}

// GetByID retrieves a project by ID
func (s *ProjectStore) GetByID(ctx context.Context, id string) (*core.Project, error) {
	var project core.Project

	query := `
		SELECT id, title, description, tags, created_at, updated_at, published_at
		FROM projects
		WHERE id = $1
	`

	row := s.db.DB().QueryRowContext(ctx, query, id)

	var tagsRaw []byte
	err := row.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&tagsRaw,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.PublishedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Unmarshal tags
	if err := json.Unmarshal(tagsRaw, &project.Tags); err != nil {
		log.Warn().Err(err).Str("project_id", id).Msg("failed to unmarshal project tags")
		project.Tags = []string{} // Fallback to empty slice
	}

	return &project, nil
}

// List retrieves projects with pagination
func (s *ProjectStore) List(ctx context.Context, limit, offset int) ([]*core.Project, int, error) {
	// First, get the total count
	var total int
	countQuery := `SELECT COUNT(*) FROM projects`
	if err := s.db.DB().QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Get the projects
	query := `
		SELECT id, title, description, tags, created_at, updated_at, published_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.DB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*core.Project
	for rows.Next() {
		var project core.Project
		var tagsRaw []byte

		err := rows.Scan(
			&project.ID,
			&project.Title,
			&project.Description,
			&tagsRaw,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.PublishedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}

		// Unmarshal tags
		if err := json.Unmarshal(tagsRaw, &project.Tags); err != nil {
			log.Warn().Err(err).Str("project_id", project.ID).Msg("failed to unmarshal project tags")
			project.Tags = []string{} // Fallback to empty slice
		}

		projects = append(projects, &project)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate projects: %w", err)
	}

	return projects, total, nil
}

// Update updates a project
func (s *ProjectStore) Update(ctx context.Context, id string, title string, description *string, tags []string) (*core.Project, error) {
	// Convert tags to JSON
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE projects 
		SET title = $1, description = $2, tags = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, title, description, tags, created_at, updated_at, published_at
	`

	row := s.db.DB().QueryRowContext(ctx, query, title, description, tagsJSON, id)

	var project core.Project
	var tagsRaw []byte
	err = row.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&tagsRaw,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.PublishedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrProjectNotFound
		}
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23514" && pqErr.Constraint == "projects_title_check" {
				return nil, core.ErrProjectTitleTooShort
			}
		}
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Unmarshal tags
	if err := json.Unmarshal(tagsRaw, &project.Tags); err != nil {
		log.Warn().Err(err).Str("project_id", id).Msg("failed to unmarshal project tags")
		project.Tags = []string{} // Fallback to empty slice
	}

	log.Info().
		Str("project_id", project.ID).
		Str("title", project.Title).
		Msg("project updated successfully")

	return &project, nil
}

// Delete deletes a project
func (s *ProjectStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := s.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return core.ErrProjectNotFound
	}

	log.Info().
		Str("project_id", id).
		Msg("project deleted successfully")

	return nil
}

// Publish marks a project as published
func (s *ProjectStore) Publish(ctx context.Context, id string) (*core.Project, error) {
	query := `
		UPDATE projects 
		SET published_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND published_at IS NULL
		RETURNING id, title, description, tags, created_at, updated_at, published_at
	`

	row := s.db.DB().QueryRowContext(ctx, query, id)

	var project core.Project
	var tagsRaw []byte
	err := row.Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&tagsRaw,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.PublishedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Check if project exists but is already published
			var exists bool
			checkQuery := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`
			if checkErr := s.db.DB().QueryRowContext(ctx, checkQuery, id).Scan(&exists); checkErr != nil {
				return nil, fmt.Errorf("failed to check project existence: %w", checkErr)
			}
			if !exists {
				return nil, core.ErrProjectNotFound
			}
			return nil, fmt.Errorf("project is already published")
		}
		return nil, fmt.Errorf("failed to publish project: %w", err)
	}

	// Unmarshal tags
	if err := json.Unmarshal(tagsRaw, &project.Tags); err != nil {
		log.Warn().Err(err).Str("project_id", id).Msg("failed to unmarshal project tags")
		project.Tags = []string{} // Fallback to empty slice
	}

	log.Info().
		Str("project_id", project.ID).
		Msg("project published successfully")

	return &project, nil
}

// SearchByTitle searches projects by title
func (s *ProjectStore) SearchByTitle(ctx context.Context, searchTerm string, limit, offset int) ([]*core.Project, int, error) {
	searchPattern := "%" + searchTerm + "%"

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM projects 
		WHERE title ILIKE $1 OR description ILIKE $1
	`
	if err := s.db.DB().QueryRowContext(ctx, countQuery, searchPattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get projects
	query := `
		SELECT id, title, description, tags, created_at, updated_at, published_at
		FROM projects
		WHERE title ILIKE $1 OR description ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.DB().QueryContext(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search projects: %w", err)
	}
	defer rows.Close()

	var projects []*core.Project
	for rows.Next() {
		var project core.Project
		var tagsRaw []byte

		err := rows.Scan(
			&project.ID,
			&project.Title,
			&project.Description,
			&tagsRaw,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.PublishedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}

		// Unmarshal tags
		if err := json.Unmarshal(tagsRaw, &project.Tags); err != nil {
			log.Warn().Err(err).Str("project_id", project.ID).Msg("failed to unmarshal project tags")
			project.Tags = []string{}
		}

		projects = append(projects, &project)
	}

	return projects, total, nil
}