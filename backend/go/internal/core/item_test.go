package core

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provemyself/backend/internal/types"
)

// mockItemStore implements ItemStore for testing
type mockItemStore struct {
	items       map[string]*Item
	projectItems map[string][]*Item
	lastError   error
}

func newMockItemStore() *mockItemStore {
	return &mockItemStore{
		items:       make(map[string]*Item),
		projectItems: make(map[string][]*Item),
	}
}

func (m *mockItemStore) Create(ctx context.Context, projectID string, itemType types.ItemType, title string, content json.RawMessage, position int, required bool, points *int, explanation *string) (*Item, error) {
	if m.lastError != nil {
		return nil, m.lastError
	}

	item := &Item{
		ID:          "test-item-id",
		ProjectID:   projectID,
		Type:        itemType,
		Title:       title,
		Content:     content,
		Position:    position,
		Required:    required,
		Points:      points,
		Explanation: explanation,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.items[item.ID] = item
	m.projectItems[projectID] = append(m.projectItems[projectID], item)
	return item, nil
}

func (m *mockItemStore) GetByID(ctx context.Context, id string) (*Item, error) {
	if m.lastError != nil {
		return nil, m.lastError
	}

	item, exists := m.items[id]
	if !exists {
		return nil, ErrItemNotFound
	}
	return item, nil
}

func (m *mockItemStore) ListByProject(ctx context.Context, projectID string) ([]*Item, error) {
	if m.lastError != nil {
		return nil, m.lastError
	}

	items, exists := m.projectItems[projectID]
	if !exists {
		return []*Item{}, nil
	}
	return items, nil
}

func (m *mockItemStore) Update(ctx context.Context, id string, itemType types.ItemType, title string, content json.RawMessage, position int, required bool, points *int, explanation *string) (*Item, error) {
	if m.lastError != nil {
		return nil, m.lastError
	}

	item, exists := m.items[id]
	if !exists {
		return nil, ErrItemNotFound
	}

	item.Type = itemType
	item.Title = title
	item.Content = content
	item.Position = position
	item.Required = required
	item.Points = points
	item.Explanation = explanation
	item.UpdatedAt = time.Now()

	return item, nil
}

func (m *mockItemStore) Delete(ctx context.Context, id string) error {
	if m.lastError != nil {
		return m.lastError
	}

	_, exists := m.items[id]
	if !exists {
		return ErrItemNotFound
	}

	delete(m.items, id)
	return nil
}

func (m *mockItemStore) UpdatePositions(ctx context.Context, updates []PositionUpdate) error {
	if m.lastError != nil {
		return m.lastError
	}

	for _, update := range updates {
		item, exists := m.items[update.ItemID]
		if exists {
			item.Position = update.Position
			item.UpdatedAt = time.Now()
		}
	}
	return nil
}

// mockProjectStore implements ProjectStore for testing
type mockProjectStore struct {
	projects  map[string]*Project
	lastError error
}

func newMockProjectStore() *mockProjectStore {
	return &mockProjectStore{
		projects: make(map[string]*Project),
	}
}

func (m *mockProjectStore) Create(ctx context.Context, title string, description *string, tags []string) (*Project, error) {
	project := &Project{
		ID:          "test-project-id",
		Title:       title,
		Description: description,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.projects[project.ID] = project
	return project, nil
}

func (m *mockProjectStore) GetByID(ctx context.Context, id string) (*Project, error) {
	if m.lastError != nil {
		return nil, m.lastError
	}

	project, exists := m.projects[id]
	if !exists {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

func (m *mockProjectStore) List(ctx context.Context, limit, offset int) ([]*Project, int, error) {
	return nil, 0, nil
}

func (m *mockProjectStore) Update(ctx context.Context, id string, title string, description *string, tags []string) (*Project, error) {
	return nil, nil
}

func (m *mockProjectStore) Delete(ctx context.Context, id string) error {
	return nil
}

func TestItemService_Create(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		itemType    types.ItemType
		title       string
		content     interface{}
		position    int
		required    bool
		points      *int
		explanation *string
		setupMocks  func(*mockItemStore, *mockProjectStore)
		wantErr     error
		validate    func(t *testing.T, item *Item)
	}{
		{
			name:      "successful choice item creation",
			projectID: "test-project-id",
			itemType:  types.ItemTypeChoice,
			title:     "Test Choice Question",
			content: types.ChoiceContent{
				Choices: []types.Choice{
					{ID: "choice1", Text: "Option A", Correct: true},
					{ID: "choice2", Text: "Option B", Correct: false},
				},
			},
			position: 0,
			required: true,
			points:   intPtr(10),
			setupMocks: func(itemStore *mockItemStore, projectStore *mockProjectStore) {
				projectStore.projects["test-project-id"] = &Project{ID: "test-project-id"}
			},
			wantErr: nil,
			validate: func(t *testing.T, item *Item) {
				assert.Equal(t, "test-project-id", item.ProjectID)
				assert.Equal(t, types.ItemTypeChoice, item.Type)
				assert.Equal(t, "Test Choice Question", item.Title)
				assert.Equal(t, 0, item.Position)
				assert.True(t, item.Required)
				assert.Equal(t, 10, *item.Points)
			},
		},
		{
			name:      "title too short",
			projectID: "test-project-id",
			itemType:  types.ItemTypeTitle,
			title:     "",
			position:  0,
			setupMocks: func(itemStore *mockItemStore, projectStore *mockProjectStore) {
				projectStore.projects["test-project-id"] = &Project{ID: "test-project-id"}
			},
			wantErr: ErrItemTitleTooShort,
		},
		{
			name:      "title too long",
			projectID: "test-project-id",
			itemType:  types.ItemTypeTitle,
			title:     string(make([]byte, 501)), // 501 characters
			position:  0,
			setupMocks: func(itemStore *mockItemStore, projectStore *mockProjectStore) {
				projectStore.projects["test-project-id"] = &Project{ID: "test-project-id"}
			},
			wantErr: ErrItemTitleTooLong,
		},
		{
			name:      "invalid item type",
			projectID: "test-project-id",
			itemType:  "invalid_type",
			title:     "Test Item",
			position:  0,
			setupMocks: func(itemStore *mockItemStore, projectStore *mockProjectStore) {
				projectStore.projects["test-project-id"] = &Project{ID: "test-project-id"}
			},
			wantErr: ErrItemInvalidType,
		},
		{
			name:      "invalid position",
			projectID: "test-project-id",
			itemType:  types.ItemTypeTitle,
			title:     "Test Item",
			position:  -1,
			setupMocks: func(itemStore *mockItemStore, projectStore *mockProjectStore) {
				projectStore.projects["test-project-id"] = &Project{ID: "test-project-id"}
			},
			wantErr: ErrItemInvalidPosition,
		},
		{
			name:      "project not found",
			projectID: "non-existent-project",
			itemType:  types.ItemTypeTitle,
			title:     "Test Item",
			position:  0,
			setupMocks: func(itemStore *mockItemStore, projectStore *mockProjectStore) {
				// Don't create the project
			},
			wantErr: ErrProjectNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemStore := newMockItemStore()
			projectStore := newMockProjectStore()

			if tt.setupMocks != nil {
				tt.setupMocks(itemStore, projectStore)
			}

			service := NewItemService(itemStore, projectStore)
			ctx := context.Background()

			item, err := service.Create(ctx, tt.projectID, tt.itemType, tt.title, tt.content, tt.position, tt.required, tt.points, tt.explanation)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, item)
			} else {
				require.NoError(t, err)
				require.NotNil(t, item)
				if tt.validate != nil {
					tt.validate(t, item)
				}
			}
		})
	}
}

func TestItemService_GetByID(t *testing.T) {
	itemStore := newMockItemStore()
	projectStore := newMockProjectStore()
	service := NewItemService(itemStore, projectStore)

	// Setup test item
	testItem := &Item{
		ID:        "test-item-id",
		ProjectID: "test-project-id",
		Type:      types.ItemTypeChoice,
		Title:     "Test Item",
		Position:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	itemStore.items["test-item-id"] = testItem

	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		item, err := service.GetByID(ctx, "test-item-id")
		require.NoError(t, err)
		assert.Equal(t, testItem.ID, item.ID)
		assert.Equal(t, testItem.Title, item.Title)
	})

	t.Run("item not found", func(t *testing.T) {
		item, err := service.GetByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrItemNotFound)
		assert.Nil(t, item)
	})
}

func TestItemService_ListByProject(t *testing.T) {
	itemStore := newMockItemStore()
	projectStore := newMockProjectStore()
	service := NewItemService(itemStore, projectStore)

	// Setup test project
	projectStore.projects["test-project-id"] = &Project{ID: "test-project-id"}

	// Setup test items
	items := []*Item{
		{ID: "item1", ProjectID: "test-project-id", Position: 0},
		{ID: "item2", ProjectID: "test-project-id", Position: 1},
	}
	itemStore.projectItems["test-project-id"] = items

	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		result, err := service.ListByProject(ctx, "test-project-id")
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "item1", result[0].ID)
		assert.Equal(t, "item2", result[1].ID)
	})

	t.Run("project not found", func(t *testing.T) {
		result, err := service.ListByProject(ctx, "non-existent-project")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrProjectNotFound)
		assert.Nil(t, result)
	})
}

func TestItemService_Update(t *testing.T) {
	itemStore := newMockItemStore()
	projectStore := newMockProjectStore()
	service := NewItemService(itemStore, projectStore)

	// Setup test item
	testItem := &Item{
		ID:        "test-item-id",
		ProjectID: "test-project-id",
		Type:      types.ItemTypeChoice,
		Title:     "Original Title",
		Position:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	itemStore.items["test-item-id"] = testItem

	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		newContent := types.ChoiceContent{
			Choices: []types.Choice{
				{ID: "choice1", Text: "Updated Option A", Correct: true},
			},
		}

		item, err := service.Update(ctx, "test-item-id", types.ItemTypeChoice, "Updated Title", newContent, 1, true, intPtr(20), stringPtr("Updated explanation"))
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", item.Title)
		assert.Equal(t, 1, item.Position)
		assert.True(t, item.Required)
		assert.Equal(t, 20, *item.Points)
		assert.Equal(t, "Updated explanation", *item.Explanation)
	})

	t.Run("item not found", func(t *testing.T) {
		item, err := service.Update(ctx, "non-existent-id", types.ItemTypeChoice, "Title", nil, 0, false, nil, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrItemNotFound)
		assert.Nil(t, item)
	})
}

func TestItemService_Delete(t *testing.T) {
	itemStore := newMockItemStore()
	projectStore := newMockProjectStore()
	service := NewItemService(itemStore, projectStore)

	// Setup test item
	itemStore.items["test-item-id"] = &Item{ID: "test-item-id"}

	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		err := service.Delete(ctx, "test-item-id")
		require.NoError(t, err)
	})

	t.Run("item not found", func(t *testing.T) {
		err := service.Delete(ctx, "non-existent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrItemNotFound)
	})
}

func TestItemService_validateType(t *testing.T) {
	service := &ItemService{}

	validTypes := []types.ItemType{
		types.ItemTypeTitle,
		types.ItemTypeMedia,
		types.ItemTypeChoice,
		types.ItemTypeMultiChoice,
		types.ItemTypeTextEntry,
		types.ItemTypeOrdering,
		types.ItemTypeHotspot,
	}

	for _, itemType := range validTypes {
		t.Run(string(itemType), func(t *testing.T) {
			err := service.validateType(itemType)
			assert.NoError(t, err)
		})
	}

	t.Run("invalid type", func(t *testing.T) {
		err := service.validateType("invalid_type")
		assert.ErrorIs(t, err, ErrItemInvalidType)
	})
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}