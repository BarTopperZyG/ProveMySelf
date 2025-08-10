package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/provemyself/backend/internal/core"
	"github.com/provemyself/backend/internal/types"
)

// ItemHandler handles item-related HTTP requests
type ItemHandler struct {
	service  *core.ItemService
	validate *validator.Validate
}

// NewItemHandler creates a new item handler
func NewItemHandler(service *core.ItemService, validate *validator.Validate) *ItemHandler {
	return &ItemHandler{
		service:  service,
		validate: validate,
	}
}

// CreateItem handles POST /api/v1/projects/{projectId}/items
// @Summary Create item
// @Description Create a new quiz item in a project
// @Tags Items
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param request body types.CreateItemRequest true "Item creation request"
// @Success 201 {object} types.ItemResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 422 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/items [post]
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	var req types.CreateItemRequest
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

	// Validate content structure based on item type
	if err := h.validateItemContent(req.Type, req.Content); err != nil {
		h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_content", err.Error())
		return
	}

	item, err := h.service.Create(ctx, projectID, req.Type, req.Title, req.Content, req.Position, req.Required, req.Points, req.Explanation)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to create item")

		switch {
		case errors.Is(err, core.ErrProjectNotFound):
			h.sendJSONError(w, http.StatusNotFound, "project_not_found", "Project not found")
		case errors.Is(err, core.ErrItemTitleTooShort):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_short", "Item title is too short")
		case errors.Is(err, core.ErrItemTitleTooLong):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_long", "Item title is too long")
		case errors.Is(err, core.ErrItemInvalidType):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_type", "Invalid item type")
		case errors.Is(err, core.ErrItemInvalidPosition):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_position", "Invalid position")
		case errors.Is(err, core.ErrItemInvalidContent):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_content", "Invalid content for item type")
		default:
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to create item")
		}
		return
	}

	response := types.ItemResponse{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Type:        item.Type,
		Title:       item.Title,
		Content:     item.Content,
		Position:    item.Position,
		Required:    item.Required,
		Points:      item.Points,
		Explanation: item.Explanation,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	h.sendJSONResponse(w, http.StatusCreated, response)
}

// ListItems handles GET /api/v1/projects/{projectId}/items
// @Summary List items
// @Description Retrieve all items for a project with optional filtering and search
// @Tags Items
// @Param projectId path string true "Project ID" format(uuid)
// @Param type query string false "Filter by item type"
// @Param search query string false "Search in item titles and content"
// @Param required query bool false "Filter by required status"
// @Param limit query int false "Maximum number of items to return" minimum(1) maximum(100) default(50)
// @Param offset query int false "Number of items to skip" minimum(0) default(0)
// @Produce json
// @Success 200 {object} types.ItemListResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/items [get]
func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	// Parse query parameters
	itemType := r.URL.Query().Get("type")
	search := r.URL.Query().Get("search")
	requiredStr := r.URL.Query().Get("required")
	
	limit := 50
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

	// Parse required filter
	var required *bool
	if requiredStr != "" {
		if parsed, err := strconv.ParseBool(requiredStr); err == nil {
			required = &parsed
		}
	}

	// Validate item type if provided
	if itemType != "" {
		if !h.isValidItemType(itemType) {
			h.sendJSONError(w, http.StatusBadRequest, "invalid_type_filter", "Invalid item type filter")
			return
		}
	}

	items, err := h.service.ListByProject(ctx, projectID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to list items")

		if errors.Is(err, core.ErrProjectNotFound) {
			h.sendJSONError(w, http.StatusNotFound, "project_not_found", "Project not found")
		} else {
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to list items")
		}
		return
	}

	// Apply filters
	filteredItems := h.filterItems(items, itemType, search, required)
	
	// Apply pagination
	total := len(filteredItems)
	start := offset
	end := start + limit
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}
	
	paginatedItems := filteredItems[start:end]

	// Convert to response format
	itemResponses := make([]types.ItemResponse, len(paginatedItems))
	for i, item := range paginatedItems {
		itemResponses[i] = types.ItemResponse{
			ID:          item.ID,
			ProjectID:   item.ProjectID,
			Type:        item.Type,
			Title:       item.Title,
			Content:     item.Content,
			Position:    item.Position,
			Required:    item.Required,
			Points:      item.Points,
			Explanation: item.Explanation,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	response := types.ItemListResponse{
		Items:     itemResponses,
		Total:     total,
		ProjectID: projectID,
		Limit:     limit,
		Offset:    offset,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// GetItem handles GET /api/v1/projects/{projectId}/items/{itemId}
// @Summary Get item
// @Description Retrieve a specific item by ID
// @Tags Items
// @Param projectId path string true "Project ID" format(uuid)
// @Param itemId path string true "Item ID" format(uuid)
// @Produce json
// @Success 200 {object} types.ItemResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/items/{itemId} [get]
func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_item_id", "Item ID is required")
		return
	}

	item, err := h.service.GetByID(ctx, itemID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("item_id", itemID).Msg("failed to get item")

		if errors.Is(err, core.ErrItemNotFound) {
			h.sendJSONError(w, http.StatusNotFound, "item_not_found", "Item not found")
		} else {
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to get item")
		}
		return
	}

	response := types.ItemResponse{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Type:        item.Type,
		Title:       item.Title,
		Content:     item.Content,
		Position:    item.Position,
		Required:    item.Required,
		Points:      item.Points,
		Explanation: item.Explanation,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// UpdateItem handles PUT /api/v1/projects/{projectId}/items/{itemId}
// @Summary Update item
// @Description Update an existing item
// @Tags Items
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param itemId path string true "Item ID" format(uuid)
// @Param request body types.UpdateItemRequest true "Item update request"
// @Success 200 {object} types.ItemResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 422 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/items/{itemId} [put]
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_item_id", "Item ID is required")
		return
	}

	var req types.UpdateItemRequest
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

	// Validate content structure based on item type
	if err := h.validateItemContent(req.Type, req.Content); err != nil {
		h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_content", err.Error())
		return
	}

	item, err := h.service.Update(ctx, itemID, req.Type, req.Title, req.Content, req.Position, req.Required, req.Points, req.Explanation)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("item_id", itemID).Msg("failed to update item")

		switch {
		case errors.Is(err, core.ErrItemNotFound):
			h.sendJSONError(w, http.StatusNotFound, "item_not_found", "Item not found")
		case errors.Is(err, core.ErrItemTitleTooShort):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_short", "Item title is too short")
		case errors.Is(err, core.ErrItemTitleTooLong):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "title_too_long", "Item title is too long")
		case errors.Is(err, core.ErrItemInvalidType):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_type", "Invalid item type")
		case errors.Is(err, core.ErrItemInvalidPosition):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_position", "Invalid position")
		case errors.Is(err, core.ErrItemInvalidContent):
			h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_content", "Invalid content for item type")
		default:
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to update item")
		}
		return
	}

	response := types.ItemResponse{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Type:        item.Type,
		Title:       item.Title,
		Content:     item.Content,
		Position:    item.Position,
		Required:    item.Required,
		Points:      item.Points,
		Explanation: item.Explanation,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// DeleteItem handles DELETE /api/v1/projects/{projectId}/items/{itemId}
// @Summary Delete item
// @Description Delete an item by ID
// @Tags Items
// @Param projectId path string true "Project ID" format(uuid)
// @Param itemId path string true "Item ID" format(uuid)
// @Success 204 "Item deleted successfully"
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/items/{itemId} [delete]
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_item_id", "Item ID is required")
		return
	}

	err := h.service.Delete(ctx, itemID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("item_id", itemID).Msg("failed to delete item")

		if errors.Is(err, core.ErrItemNotFound) {
			h.sendJSONError(w, http.StatusNotFound, "item_not_found", "Item not found")
		} else {
			h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to delete item")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateItemPositions handles PUT /api/v1/projects/{projectId}/items/positions
// @Summary Update item positions
// @Description Update the positions of multiple items for reordering
// @Tags Items
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param request body []types.PositionUpdateRequest true "Array of position updates"
// @Success 200 {object} types.ItemListResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 422 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/items/positions [put]
func (h *ItemHandler) UpdateItemPositions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	var req []types.PositionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode position update request")
		h.sendJSONError(w, http.StatusBadRequest, "invalid_request_body", "Invalid request body")
		return
	}

	if len(req) == 0 {
		h.sendJSONError(w, http.StatusBadRequest, "empty_updates", "At least one position update is required")
		return
	}

	// Validate each position update
	for _, update := range req {
		if err := h.validate.StructCtx(ctx, update); err != nil {
			h.sendJSONError(w, http.StatusBadRequest, "validation_failed", "Invalid position update", err.Error())
			return
		}
	}

	// Convert to core types
	updates := make([]core.PositionUpdate, len(req))
	for i, update := range req {
		updates[i] = core.PositionUpdate{
			ItemID:   update.ItemID,
			Position: update.Position,
		}
	}

	// Update positions
	if err := h.service.UpdatePositions(ctx, updates); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to update item positions")
		h.sendJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to update item positions")
		return
	}

	// Return updated item list
	h.ListItems(w, r)
}

// BulkCreateItems handles POST /api/v1/projects/{projectId}/items/bulk
// @Summary Bulk create items
// @Description Create multiple items at once
// @Tags Items
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID" format(uuid)
// @Param request body []types.CreateItemRequest true "Array of items to create"
// @Success 201 {object} types.ItemListResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 422 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /projects/{projectId}/items/bulk [post]
func (h *ItemHandler) BulkCreateItems(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		h.sendJSONError(w, http.StatusBadRequest, "missing_project_id", "Project ID is required")
		return
	}

	var req []types.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode bulk create request")
		h.sendJSONError(w, http.StatusBadRequest, "invalid_request_body", "Invalid request body")
		return
	}

	if len(req) == 0 {
		h.sendJSONError(w, http.StatusBadRequest, "empty_items", "At least one item is required")
		return
	}

	if len(req) > 100 {
		h.sendJSONError(w, http.StatusBadRequest, "too_many_items", "Maximum 100 items can be created at once")
		return
	}

	// Validate each item
	for i, itemReq := range req {
		if err := h.validate.StructCtx(ctx, itemReq); err != nil {
			h.sendJSONError(w, http.StatusBadRequest, "validation_failed", 
				fmt.Sprintf("Item %d validation failed: %s", i+1, err.Error()))
			return
		}

		if err := h.validateItemContent(itemReq.Type, itemReq.Content); err != nil {
			h.sendJSONError(w, http.StatusUnprocessableEntity, "invalid_content", 
				fmt.Sprintf("Item %d: %s", i+1, err.Error()))
			return
		}
	}

	// Create items
	createdItems := make([]*core.Item, 0, len(req))
	for _, itemReq := range req {
		item, err := h.service.Create(ctx, projectID, itemReq.Type, itemReq.Title, itemReq.Content, 
			itemReq.Position, itemReq.Required, itemReq.Points, itemReq.Explanation)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("project_id", projectID).Msg("failed to create item in bulk operation")
			h.sendJSONError(w, http.StatusInternalServerError, "bulk_create_failed", 
				"Failed to create some items in bulk operation")
			return
		}
		createdItems = append(createdItems, item)
	}

	// Convert to response format
	itemResponses := make([]types.ItemResponse, len(createdItems))
	for i, item := range createdItems {
		itemResponses[i] = types.ItemResponse{
			ID:          item.ID,
			ProjectID:   item.ProjectID,
			Type:        item.Type,
			Title:       item.Title,
			Content:     item.Content,
			Position:    item.Position,
			Required:    item.Required,
			Points:      item.Points,
			Explanation: item.Explanation,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	response := types.ItemListResponse{
		Items:     itemResponses,
		Total:     len(itemResponses),
		ProjectID: projectID,
	}

	h.sendJSONResponse(w, http.StatusCreated, response)
}

// validateItemContent validates that the content structure matches the item type
func (h *ItemHandler) validateItemContent(itemType types.ItemType, content interface{}) error {
	if content == nil {
		return nil // Content is optional for some item types
	}

	switch itemType {
	case types.ItemTypeChoice, types.ItemTypeMultiChoice:
		return h.validateChoiceContent(content)
	case types.ItemTypeMedia:
		return h.validateMediaContent(content)
	case types.ItemTypeTextEntry:
		return h.validateTextEntryContent(content)
	case types.ItemTypeOrdering:
		return h.validateOrderingContent(content)
	case types.ItemTypeHotspot:
		return h.validateHotspotContent(content)
	case types.ItemTypeTitle:
		// Title items don't need content validation
		return nil
	default:
		return fmt.Errorf("unsupported item type: %s", itemType)
	}
}

// validateChoiceContent validates choice/multi-choice question content
func (h *ItemHandler) validateChoiceContent(content interface{}) error {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("invalid content format: %w", err)
	}

	var choiceContent types.ChoiceContent
	if err := json.Unmarshal(contentBytes, &choiceContent); err != nil {
		return fmt.Errorf("invalid choice content structure: %w", err)
	}

	if err := h.validate.Struct(choiceContent); err != nil {
		return fmt.Errorf("choice content validation failed: %w", err)
	}

	// Check that at least one choice is marked as correct
	hasCorrect := false
	for _, choice := range choiceContent.Choices {
		if choice.Correct {
			hasCorrect = true
			break
		}
	}
	if !hasCorrect {
		return fmt.Errorf("at least one choice must be marked as correct")
	}

	return nil
}

// validateMediaContent validates media item content
func (h *ItemHandler) validateMediaContent(content interface{}) error {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("invalid content format: %w", err)
	}

	var mediaContent types.MediaContent
	if err := json.Unmarshal(contentBytes, &mediaContent); err != nil {
		return fmt.Errorf("invalid media content structure: %w", err)
	}

	return h.validate.Struct(mediaContent)
}

// validateTextEntryContent validates text entry question content
func (h *ItemHandler) validateTextEntryContent(content interface{}) error {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("invalid content format: %w", err)
	}

	var textContent types.TextEntryContent
	if err := json.Unmarshal(contentBytes, &textContent); err != nil {
		return fmt.Errorf("invalid text entry content structure: %w", err)
	}

	return h.validate.Struct(textContent)
}

// validateOrderingContent validates ordering question content
func (h *ItemHandler) validateOrderingContent(content interface{}) error {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("invalid content format: %w", err)
	}

	var orderingContent types.OrderingContent
	if err := json.Unmarshal(contentBytes, &orderingContent); err != nil {
		return fmt.Errorf("invalid ordering content structure: %w", err)
	}

	if err := h.validate.Struct(orderingContent); err != nil {
		return fmt.Errorf("ordering content validation failed: %w", err)
	}

	// Check that order numbers are sequential and start from 1
	orderMap := make(map[int]bool)
	for _, item := range orderingContent.Items {
		if item.CorrectOrder < 1 {
			return fmt.Errorf("order numbers must start from 1")
		}
		orderMap[item.CorrectOrder] = true
	}

	// Check for gaps in sequence
	for i := 1; i <= len(orderingContent.Items); i++ {
		if !orderMap[i] {
			return fmt.Errorf("order numbers must be sequential starting from 1")
		}
	}

	return nil
}

// validateHotspotContent validates hotspot question content
func (h *ItemHandler) validateHotspotContent(content interface{}) error {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("invalid content format: %w", err)
	}

	var hotspotContent types.HotspotContent
	if err := json.Unmarshal(contentBytes, &hotspotContent); err != nil {
		return fmt.Errorf("invalid hotspot content structure: %w", err)
	}

	if err := h.validate.Struct(hotspotContent); err != nil {
		return fmt.Errorf("hotspot content validation failed: %w", err)
	}

	// Check that at least one hotspot is marked as correct
	hasCorrect := false
	for _, hotspot := range hotspotContent.Hotspots {
		if hotspot.Correct {
			hasCorrect = true
			break
		}
	}
	if !hasCorrect {
		return fmt.Errorf("at least one hotspot must be marked as correct")
	}

	return nil
}

// isValidItemType checks if the given string is a valid item type
func (h *ItemHandler) isValidItemType(itemType string) bool {
	validTypes := []string{
		string(types.ItemTypeTitle),
		string(types.ItemTypeMedia),
		string(types.ItemTypeChoice),
		string(types.ItemTypeMultiChoice),
		string(types.ItemTypeTextEntry),
		string(types.ItemTypeOrdering),
		string(types.ItemTypeHotspot),
	}

	for _, validType := range validTypes {
		if itemType == validType {
			return true
		}
	}
	return false
}

// filterItems applies filters to the items list
func (h *ItemHandler) filterItems(items []*core.Item, itemType, search string, required *bool) []*core.Item {
	filtered := make([]*core.Item, 0, len(items))

	for _, item := range items {
		// Filter by type
		if itemType != "" && string(item.Type) != itemType {
			continue
		}

		// Filter by required status
		if required != nil && item.Required != *required {
			continue
		}

		// Filter by search term
		if search != "" {
			searchLower := strings.ToLower(search)
			titleLower := strings.ToLower(item.Title)
			
			// Search in title
			if strings.Contains(titleLower, searchLower) {
				filtered = append(filtered, item)
				continue
			}

			// Search in content (basic string search)
			contentStr := string(item.Content)
			if strings.Contains(strings.ToLower(contentStr), searchLower) {
				filtered = append(filtered, item)
				continue
			}

			// If no match found, skip this item
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered
}

// sendJSONResponse sends a JSON response
func (h *ItemHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Error().Err(err).Msg("failed to encode JSON response")
		}
	}
}

// sendJSONError sends a JSON error response
func (h *ItemHandler) sendJSONError(w http.ResponseWriter, statusCode int, code, message string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := types.ErrorResponse{
		Error: types.Error{
			Code:    code,
			Message: message,
		},
	}

	if len(details) > 0 && details[0] != "" {
		errorResponse.Error.Details = &details[0]
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Error().Err(err).Msg("failed to encode JSON error response")
	}
}