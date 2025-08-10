package types

import "time"

// ItemType represents the type of quiz item/question
type ItemType string

const (
	// ItemTypeTitle represents a title/heading block
	ItemTypeTitle ItemType = "title"
	// ItemTypeMedia represents a media block (image, video, audio)
	ItemTypeMedia ItemType = "media"
	// ItemTypeChoice represents a single-choice question
	ItemTypeChoice ItemType = "choice"
	// ItemTypeMultiChoice represents a multiple-choice question
	ItemTypeMultiChoice ItemType = "multi_choice"
	// ItemTypeTextEntry represents a text input question
	ItemTypeTextEntry ItemType = "text_entry"
	// ItemTypeOrdering represents a drag-and-drop ordering question
	ItemTypeOrdering ItemType = "ordering"
	// ItemTypeHotspot represents a hotspot/click-area question
	ItemTypeHotspot ItemType = "hotspot"
)

// CreateItemRequest represents a request to create a new quiz item
type CreateItemRequest struct {
	Type        ItemType    `json:"type" validate:"required,oneof=title media choice multi_choice text_entry ordering hotspot"`
	Title       string      `json:"title" validate:"required,min=1,max=500"`
	Content     interface{} `json:"content,omitempty"`
	Position    int         `json:"position" validate:"min=0"`
	Required    bool        `json:"required"`
	Points      *int        `json:"points,omitempty" validate:"omitempty,min=0,max=1000"`
	Explanation *string     `json:"explanation,omitempty" validate:"omitempty,max=1000"`
}

// UpdateItemRequest represents a request to update an existing quiz item
type UpdateItemRequest struct {
	Type        ItemType    `json:"type" validate:"required,oneof=title media choice multi_choice text_entry ordering hotspot"`
	Title       string      `json:"title" validate:"required,min=1,max=500"`
	Content     interface{} `json:"content,omitempty"`
	Position    int         `json:"position" validate:"min=0"`
	Required    bool        `json:"required"`
	Points      *int        `json:"points,omitempty" validate:"omitempty,min=0,max=1000"`
	Explanation *string     `json:"explanation,omitempty" validate:"omitempty,max=1000"`
}

// ItemResponse represents a quiz item in API responses
type ItemResponse struct {
	ID          string      `json:"id"`
	ProjectID   string      `json:"project_id"`
	Type        ItemType    `json:"type"`
	Title       string      `json:"title"`
	Content     interface{} `json:"content,omitempty"`
	Position    int         `json:"position"`
	Required    bool        `json:"required"`
	Points      *int        `json:"points,omitempty"`
	Explanation *string     `json:"explanation,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ItemListResponse represents a list of quiz items
type ItemListResponse struct {
	Items     []ItemResponse `json:"items"`
	Total     int            `json:"total"`
	ProjectID string         `json:"project_id"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

// PositionUpdateRequest represents a request to update item positions
type PositionUpdateRequest struct {
	ItemID   string `json:"item_id" validate:"required,uuid"`
	Position int    `json:"position" validate:"required,min=0"`
}

// Choice represents an option for choice-type questions
type Choice struct {
	ID      string `json:"id" validate:"required"`
	Text    string `json:"text" validate:"required,min=1,max=500"`
	Correct bool   `json:"correct"`
}

// ChoiceContent represents the content structure for choice/multi-choice questions
type ChoiceContent struct {
	Choices []Choice `json:"choices" validate:"required,min=1,max=10,dive"`
}

// MediaContent represents the content structure for media items
type MediaContent struct {
	URL         string  `json:"url" validate:"required,url"`
	MediaType   string  `json:"media_type" validate:"required,oneof=image video audio"`
	AltText     *string `json:"alt_text,omitempty" validate:"omitempty,max=200"`
	Caption     *string `json:"caption,omitempty" validate:"omitempty,max=500"`
	Autoplay    bool    `json:"autoplay"`
	ShowControls bool   `json:"show_controls"`
}

// TextEntryContent represents the content structure for text entry questions
type TextEntryContent struct {
	MaxLength    *int    `json:"max_length,omitempty" validate:"omitempty,min=1,max=10000"`
	Placeholder  *string `json:"placeholder,omitempty" validate:"omitempty,max=100"`
	Multiline    bool    `json:"multiline"`
	CorrectAnswer *string `json:"correct_answer,omitempty" validate:"omitempty,max=10000"`
}

// OrderingContent represents the content structure for ordering questions
type OrderingContent struct {
	Items []OrderingItem `json:"items" validate:"required,min=2,max=10,dive"`
}

// OrderingItem represents an item in ordering questions
type OrderingItem struct {
	ID           string `json:"id" validate:"required"`
	Text         string `json:"text" validate:"required,min=1,max=500"`
	CorrectOrder int    `json:"correct_order" validate:"required,min=1"`
}

// HotspotContent represents the content structure for hotspot questions
type HotspotContent struct {
	ImageURL  string        `json:"image_url" validate:"required,url"`
	AltText   *string       `json:"alt_text,omitempty" validate:"omitempty,max=200"`
	Hotspots  []Hotspot     `json:"hotspots" validate:"required,min=1,max=20,dive"`
}

// Hotspot represents a clickable area on an image
type Hotspot struct {
	ID      string     `json:"id" validate:"required"`
	Shape   string     `json:"shape" validate:"required,oneof=rectangle circle polygon"`
	Coords  []float64  `json:"coords" validate:"required,min=2"`
	Correct bool       `json:"correct"`
	Feedback *string   `json:"feedback,omitempty" validate:"omitempty,max=200"`
}