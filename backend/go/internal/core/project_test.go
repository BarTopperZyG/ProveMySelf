package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectService_Create(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description *string
		tags        []string
		wantErr     error
		validate    func(t *testing.T, project *Project)
	}{
		{
			name:        "successful project creation",
			title:       "Test Quiz",
			description: stringPtr("A test quiz"),
			tags:        []string{"test", "quiz"},
			wantErr:     nil,
			validate: func(t *testing.T, project *Project) {
				assert.NotEmpty(t, project.ID)
				assert.Equal(t, "Test Quiz", project.Title)
				assert.Equal(t, "A test quiz", *project.Description)
				assert.Equal(t, []string{"test", "quiz"}, project.Tags)
				assert.WithinDuration(t, time.Now(), project.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), project.UpdatedAt, time.Second)
				assert.Nil(t, project.PublishedAt)
			},
		},
		{
			name:        "project creation with minimal data",
			title:       "Simple Quiz",
			description: nil,
			tags:        nil,
			wantErr:     nil,
			validate: func(t *testing.T, project *Project) {
				assert.NotEmpty(t, project.ID)
				assert.Equal(t, "Simple Quiz", project.Title)
				assert.Nil(t, project.Description)
				assert.Nil(t, project.Tags)
			},
		},
		{
			name:        "title too short",
			title:       "",
			description: nil,
			tags:        nil,
			wantErr:     ErrProjectTitleTooShort,
			validate:    nil,
		},
		{
			name:        "title too long",
			title:       string(make([]byte, 201)), // 201 characters
			description: nil,
			tags:        nil,
			wantErr:     ErrProjectTitleTooLong,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := NewProjectService()
			ctx := context.Background()

			// Act
			project, err := service.Create(ctx, tt.title, tt.description, tt.tags)

			// Assert
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, project)
				if tt.validate != nil {
					tt.validate(t, project)
				}
			}
		})
	}
}

func TestProjectService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(s *ProjectService) string // Returns the project ID to search for
		wantErr error
		validate func(t *testing.T, project *Project)
	}{
		{
			name: "successful project retrieval",
			setup: func(s *ProjectService) string {
				project, err := s.Create(context.Background(), "Test Quiz", stringPtr("A test"), nil)
				require.NoError(t, err)
				return project.ID
			},
			wantErr: nil,
			validate: func(t *testing.T, project *Project) {
				assert.Equal(t, "Test Quiz", project.Title)
				assert.Equal(t, "A test", *project.Description)
			},
		},
		{
			name: "project not found",
			setup: func(s *ProjectService) string {
				return "nonexistent-id"
			},
			wantErr: ErrProjectNotFound,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := NewProjectService()
			ctx := context.Background()
			projectID := tt.setup(service)

			// Act
			project, err := service.GetByID(ctx, projectID)

			// Assert
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, project)
				if tt.validate != nil {
					tt.validate(t, project)
				}
			}
		})
	}
}

func TestProjectService_List(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(s *ProjectService) // Create test data
		limit    int
		offset   int
		validate func(t *testing.T, projects []*Project, total int)
	}{
		{
			name: "list empty projects",
			setup: func(s *ProjectService) {
				// No projects created
			},
			limit:  20,
			offset: 0,
			validate: func(t *testing.T, projects []*Project, total int) {
				assert.Empty(t, projects)
				assert.Equal(t, 0, total)
			},
		},
		{
			name: "list all projects",
			setup: func(s *ProjectService) {
				_, err := s.Create(context.Background(), "Quiz 1", nil, nil)
				require.NoError(t, err)
				_, err = s.Create(context.Background(), "Quiz 2", nil, nil)
				require.NoError(t, err)
				_, err = s.Create(context.Background(), "Quiz 3", nil, nil)
				require.NoError(t, err)
			},
			limit:  20,
			offset: 0,
			validate: func(t *testing.T, projects []*Project, total int) {
				assert.Len(t, projects, 3)
				assert.Equal(t, 3, total)
				
				// Check that all projects are present (order doesn't matter for in-memory storage)
				titles := make([]string, len(projects))
				for i, p := range projects {
					titles[i] = p.Title
				}
				assert.Contains(t, titles, "Quiz 1")
				assert.Contains(t, titles, "Quiz 2")
				assert.Contains(t, titles, "Quiz 3")
			},
		},
		{
			name: "list with pagination - first page",
			setup: func(s *ProjectService) {
				for i := 1; i <= 5; i++ {
					_, err := s.Create(context.Background(), "Quiz "+string(rune('0'+i)), nil, nil)
					require.NoError(t, err)
				}
			},
			limit:  2,
			offset: 0,
			validate: func(t *testing.T, projects []*Project, total int) {
				assert.Len(t, projects, 2)
				assert.Equal(t, 5, total)
			},
		},
		{
			name: "list with pagination - second page",
			setup: func(s *ProjectService) {
				for i := 1; i <= 5; i++ {
					_, err := s.Create(context.Background(), "Quiz "+string(rune('0'+i)), nil, nil)
					require.NoError(t, err)
				}
			},
			limit:  2,
			offset: 2,
			validate: func(t *testing.T, projects []*Project, total int) {
				assert.Len(t, projects, 2)
				assert.Equal(t, 5, total)
			},
		},
		{
			name: "list with offset beyond total",
			setup: func(s *ProjectService) {
				_, err := s.Create(context.Background(), "Quiz 1", nil, nil)
				require.NoError(t, err)
			},
			limit:  10,
			offset: 10,
			validate: func(t *testing.T, projects []*Project, total int) {
				assert.Empty(t, projects)
				assert.Equal(t, 1, total)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service := NewProjectService()
			tt.setup(service)
			ctx := context.Background()

			// Act
			projects, total, err := service.List(ctx, tt.limit, tt.offset)

			// Assert
			assert.NoError(t, err)
			tt.validate(t, projects, total)
		})
	}
}

func TestProjectService_Create_UniqueIDs(t *testing.T) {
	// Arrange
	service := NewProjectService()
	ctx := context.Background()

	// Act - create multiple projects
	project1, err1 := service.Create(ctx, "Quiz 1", nil, nil)
	project2, err2 := service.Create(ctx, "Quiz 2", nil, nil)
	project3, err3 := service.Create(ctx, "Quiz 3", nil, nil)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	
	// All IDs should be unique
	assert.NotEqual(t, project1.ID, project2.ID)
	assert.NotEqual(t, project1.ID, project3.ID)
	assert.NotEqual(t, project2.ID, project3.ID)
	
	// All IDs should be valid UUIDs (not empty)
	assert.NotEmpty(t, project1.ID)
	assert.NotEmpty(t, project2.ID)
	assert.NotEmpty(t, project3.ID)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}