// Package core contains the business logic and domain models for quiz items.
// This file defines the Item entity, service, and repository interface following
// the same Clean Architecture principles as the Project domain.
package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/provemyself/backend/internal/types"
)

// Domain errors for item operations.
var (
	// ErrItemNotFound is returned when an item with the given ID doesn't exist.
	ErrItemNotFound = errors.New("item not found")
	
	// ErrItemTitleTooShort is returned when an item title is shorter than the minimum length.
	ErrItemTitleTooShort = errors.New("item title too short")
	
	// ErrItemTitleTooLong is returned when an item title exceeds the maximum length.
	ErrItemTitleTooLong = errors.New("item title too long")
	
	// ErrItemInvalidType is returned when an item type is not supported.
	ErrItemInvalidType = errors.New("invalid item type")
	
	// ErrItemInvalidPosition is returned when an item position is invalid.
	ErrItemInvalidPosition = errors.New("invalid item position")
	
	// ErrItemInvalidContent is returned when item content doesn't match the item type.
	ErrItemInvalidContent = errors.New("invalid content for item type")
)

// Item represents a quiz item/question entity in the ProveMySelf platform.
// Each item belongs to a project and represents a single quiz element such as
// a question, media block, or instructional content.
//
// Business Rules:
// - Title must be between 1 and 500 characters
// - Type must be one of the supported ItemType values
// - Position must be >= 0 and unique within a project
// - Points can be null (no scoring) or 0-1000
// - Content structure depends on the item type
type Item struct {
	// ID is the unique identifier for the item (UUID format).
	ID string
	
	// ProjectID is the ID of the project this item belongs to.
	ProjectID string
	
	// Type specifies what kind of quiz element this is.
	Type types.ItemType
	
	// Title is the display text/question text for the item.
	Title string
	
	// Content contains type-specific data (choices, media URLs, etc.).
	// The structure depends on the Type field.
	Content json.RawMessage
	
	// Position determines the order of items within a project.
	// Must be unique within a project, 0-based indexing.
	Position int
	
	// Required indicates whether this item must be answered.
	Required bool
	
	// Points specifies the scoring weight for this item.
	// Nil means no scoring, otherwise 0-1000 points.
	Points *int
	
	// Explanation provides feedback or explanation text.
	// Shown after the item is answered or in review mode.
	Explanation *string
	
	// CreatedAt is the timestamp when the item was first created.
	CreatedAt time.Time
	
	// UpdatedAt is the timestamp when the item was last modified.
	UpdatedAt time.Time
}

// ItemStore defines the contract for item data persistence.
type ItemStore interface {
	// Create persists a new item with the given parameters.
	Create(ctx context.Context, projectID string, itemType types.ItemType, title string, content json.RawMessage, position int, required bool, points *int, explanation *string) (*Item, error)
	
	// GetByID retrieves an item by its unique identifier.
	GetByID(ctx context.Context, id string) (*Item, error)
	
	// ListByProject retrieves all items for a specific project, ordered by position.
	ListByProject(ctx context.Context, projectID string) ([]*Item, error)
	
	// Update modifies an existing item with new values.
	Update(ctx context.Context, id string, itemType types.ItemType, title string, content json.RawMessage, position int, required bool, points *int, explanation *string) (*Item, error)
	
	// Delete permanently removes an item from storage.
	Delete(ctx context.Context, id string) error
	
	// UpdatePositions updates the position field for multiple items atomically.
	// Used for reordering items within a project.
	UpdatePositions(ctx context.Context, updates []PositionUpdate) error
}

// PositionUpdate represents a position change for an item.
type PositionUpdate struct {
	ItemID   string
	Position int
}

// ItemService provides business logic for quiz item operations.
type ItemService struct {
	itemStore   ItemStore
	projectStore ProjectStore
}

// NewItemService creates a new item service.
func NewItemService(itemStore ItemStore, projectStore ProjectStore) *ItemService {
	return &ItemService{
		itemStore:   itemStore,
		projectStore: projectStore,
	}
}

// Create validates and creates a new quiz item.
func (s *ItemService) Create(ctx context.Context, projectID string, itemType types.ItemType, title string, content interface{}, position int, required bool, points *int, explanation *string) (*Item, error) {
	// Validate business rules
	if err := s.validateTitle(title); err != nil {
		return nil, err
	}
	
	if err := s.validateType(itemType); err != nil {
		return nil, err
	}
	
	if err := s.validatePosition(position); err != nil {
		return nil, err
	}
	
	// Ensure project exists
	_, err := s.projectStore.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project exists: %w", err)
	}
	
	// Serialize content
	contentBytes, err := s.serializeContent(itemType, content)
	if err != nil {
		return nil, err
	}
	
	// Create the item
	item, err := s.itemStore.Create(ctx, projectID, itemType, title, contentBytes, position, required, points, explanation)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}
	
	return item, nil
}

// GetByID retrieves an item by ID.
func (s *ItemService) GetByID(ctx context.Context, id string) (*Item, error) {
	item, err := s.itemStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// ListByProject retrieves all items for a project, ordered by position.
func (s *ItemService) ListByProject(ctx context.Context, projectID string) ([]*Item, error) {
	// Ensure project exists
	_, err := s.projectStore.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project exists: %w", err)
	}
	
	items, err := s.itemStore.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	
	return items, nil
}

// Update validates and updates an existing item.
func (s *ItemService) Update(ctx context.Context, id string, itemType types.ItemType, title string, content interface{}, position int, required bool, points *int, explanation *string) (*Item, error) {
	// Validate business rules
	if err := s.validateTitle(title); err != nil {
		return nil, err
	}
	
	if err := s.validateType(itemType); err != nil {
		return nil, err
	}
	
	if err := s.validatePosition(position); err != nil {
		return nil, err
	}
	
	// Serialize content
	contentBytes, err := s.serializeContent(itemType, content)
	if err != nil {
		return nil, err
	}
	
	// Update the item
	item, err := s.itemStore.Update(ctx, id, itemType, title, contentBytes, position, required, points, explanation)
	if err != nil {
		return nil, err
	}
	
	return item, nil
}

// Delete removes an item.
func (s *ItemService) Delete(ctx context.Context, id string) error {
	return s.itemStore.Delete(ctx, id)
}

// validateTitle ensures the title meets business rules.
func (s *ItemService) validateTitle(title string) error {
	if len(title) < 1 {
		return ErrItemTitleTooShort
	}
	if len(title) > 500 {
		return ErrItemTitleTooLong
	}
	return nil
}

// validateType ensures the item type is supported.
func (s *ItemService) validateType(itemType types.ItemType) error {
	switch itemType {
	case types.ItemTypeTitle, types.ItemTypeMedia, types.ItemTypeChoice,
		types.ItemTypeMultiChoice, types.ItemTypeTextEntry,
		types.ItemTypeOrdering, types.ItemTypeHotspot:
		return nil
	default:
		return ErrItemInvalidType
	}
}

// validatePosition ensures the position is valid.
func (s *ItemService) validatePosition(position int) error {
	if position < 0 {
		return ErrItemInvalidPosition
	}
	return nil
}

// serializeContent converts content to JSON based on item type.
func (s *ItemService) serializeContent(itemType types.ItemType, content interface{}) (json.RawMessage, error) {
	if content == nil {
		return json.RawMessage("{}"), nil
	}
	
	// Validate content structure based on type
	switch itemType {
	case types.ItemTypeChoice, types.ItemTypeMultiChoice:
		if _, ok := content.(types.ChoiceContent); !ok {
			// Try to unmarshal if it's already JSON
			if contentBytes, ok := content.([]byte); ok {
				var choiceContent types.ChoiceContent
				if err := json.Unmarshal(contentBytes, &choiceContent); err != nil {
					return nil, fmt.Errorf("%w: invalid choice content structure", ErrItemInvalidContent)
				}
			} else {
				return nil, fmt.Errorf("%w: expected ChoiceContent for choice items", ErrItemInvalidContent)
			}
		}
	case types.ItemTypeMedia:
		if _, ok := content.(types.MediaContent); !ok {
			if contentBytes, ok := content.([]byte); ok {
				var mediaContent types.MediaContent
				if err := json.Unmarshal(contentBytes, &mediaContent); err != nil {
					return nil, fmt.Errorf("%w: invalid media content structure", ErrItemInvalidContent)
				}
			} else {
				return nil, fmt.Errorf("%w: expected MediaContent for media items", ErrItemInvalidContent)
			}
		}
	case types.ItemTypeTextEntry:
		if _, ok := content.(types.TextEntryContent); !ok {
			if contentBytes, ok := content.([]byte); ok {
				var textContent types.TextEntryContent
				if err := json.Unmarshal(contentBytes, &textContent); err != nil {
					return nil, fmt.Errorf("%w: invalid text entry content structure", ErrItemInvalidContent)
				}
			} else {
				return nil, fmt.Errorf("%w: expected TextEntryContent for text entry items", ErrItemInvalidContent)
			}
		}
	case types.ItemTypeOrdering:
		if _, ok := content.(types.OrderingContent); !ok {
			if contentBytes, ok := content.([]byte); ok {
				var orderingContent types.OrderingContent
				if err := json.Unmarshal(contentBytes, &orderingContent); err != nil {
					return nil, fmt.Errorf("%w: invalid ordering content structure", ErrItemInvalidContent)
				}
			} else {
				return nil, fmt.Errorf("%w: expected OrderingContent for ordering items", ErrItemInvalidContent)
			}
		}
	case types.ItemTypeHotspot:
		if _, ok := content.(types.HotspotContent); !ok {
			if contentBytes, ok := content.([]byte); ok {
				var hotspotContent types.HotspotContent
				if err := json.Unmarshal(contentBytes, &hotspotContent); err != nil {
					return nil, fmt.Errorf("%w: invalid hotspot content structure", ErrItemInvalidContent)
				}
			} else {
				return nil, fmt.Errorf("%w: expected HotspotContent for hotspot items", ErrItemInvalidContent)
			}
		}
	}
	
	// Serialize to JSON
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize content: %w", err)
	}
	
	return json.RawMessage(contentBytes), nil
}