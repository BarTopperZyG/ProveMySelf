package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/provemyself/backend/internal/core"
	"github.com/provemyself/backend/internal/types"
)

// ItemStore implements item data access using PostgreSQL
type ItemStore struct {
	db *Database
}

// NewItemStore creates a new item store
func NewItemStore(db *Database) *ItemStore {
	return &ItemStore{db: db}
}

// Create creates a new item in the database
func (s *ItemStore) Create(ctx context.Context, projectID string, itemType types.ItemType, title string, content json.RawMessage, position int, required bool, points *int, explanation *string) (*core.Item, error) {
	var item core.Item

	query := `
		INSERT INTO items (project_id, type, title, content, position, required, points, explanation)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, project_id, type, title, content, position, required, points, explanation, created_at, updated_at
	`

	row := s.db.DB().QueryRowContext(ctx, query, projectID, string(itemType), title, content, position, required, points, explanation)

	var contentRaw []byte
	var typeStr string
	err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&typeStr,
		&item.Title,
		&contentRaw,
		&item.Position,
		&item.Required,
		&item.Points,
		&item.Explanation,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("failed to create item: no rows returned")
		}
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	item.Type = types.ItemType(typeStr)
	item.Content = json.RawMessage(contentRaw)

	return &item, nil
}

// GetByID retrieves an item by its ID
func (s *ItemStore) GetByID(ctx context.Context, id string) (*core.Item, error) {
	var item core.Item

	query := `
		SELECT id, project_id, type, title, content, position, required, points, explanation, created_at, updated_at
		FROM items
		WHERE id = $1
	`

	row := s.db.DB().QueryRowContext(ctx, query, id)

	var contentRaw []byte
	var typeStr string
	err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&typeStr,
		&item.Title,
		&contentRaw,
		&item.Position,
		&item.Required,
		&item.Points,
		&item.Explanation,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to get item by ID: %w", err)
	}

	item.Type = types.ItemType(typeStr)
	item.Content = json.RawMessage(contentRaw)

	return &item, nil
}

// ListByProject retrieves all items for a project, ordered by position
func (s *ItemStore) ListByProject(ctx context.Context, projectID string) ([]*core.Item, error) {
	query := `
		SELECT id, project_id, type, title, content, position, required, points, explanation, created_at, updated_at
		FROM items
		WHERE project_id = $1
		ORDER BY position ASC
	`

	rows, err := s.db.DB().QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*core.Item
	for rows.Next() {
		var item core.Item
		var contentRaw []byte
		var typeStr string

		err := rows.Scan(
			&item.ID,
			&item.ProjectID,
			&typeStr,
			&item.Title,
			&contentRaw,
			&item.Position,
			&item.Required,
			&item.Points,
			&item.Explanation,
			&item.CreatedAt,
			&item.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan item row: %w", err)
		}

		item.Type = types.ItemType(typeStr)
		item.Content = json.RawMessage(contentRaw)
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return items, nil
}

// Update updates an existing item
func (s *ItemStore) Update(ctx context.Context, id string, itemType types.ItemType, title string, content json.RawMessage, position int, required bool, points *int, explanation *string) (*core.Item, error) {
	var item core.Item

	query := `
		UPDATE items
		SET type = $2, title = $3, content = $4, position = $5, required = $6, points = $7, explanation = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, type, title, content, position, required, points, explanation, created_at, updated_at
	`

	row := s.db.DB().QueryRowContext(ctx, query, id, string(itemType), title, content, position, required, points, explanation)

	var contentRaw []byte
	var typeStr string
	err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&typeStr,
		&item.Title,
		&contentRaw,
		&item.Position,
		&item.Required,
		&item.Points,
		&item.Explanation,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, core.ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	item.Type = types.ItemType(typeStr)
	item.Content = json.RawMessage(contentRaw)

	return &item, nil
}

// Delete removes an item from the database
func (s *ItemStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM items WHERE id = $1`

	result, err := s.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return core.ErrItemNotFound
	}

	return nil
}

// UpdatePositions updates the position field for multiple items atomically
func (s *ItemStore) UpdatePositions(ctx context.Context, updates []core.PositionUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := s.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Ctx(ctx).Error().Err(rollbackErr).Msg("failed to rollback transaction")
			}
		}
	}()

	// Update each position
	query := `UPDATE items SET position = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	for _, update := range updates {
		_, err = tx.ExecContext(ctx, query, update.ItemID, update.Position)
		if err != nil {
			return fmt.Errorf("failed to update position for item %s: %w", update.ItemID, err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}