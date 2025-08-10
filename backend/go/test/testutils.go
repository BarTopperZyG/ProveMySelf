package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgreSQLContainer wraps a PostgreSQL test container
type PostgreSQLContainer struct {
	testcontainers.Container
	ConnectionString string
}

// StartPostgreSQLContainer starts a PostgreSQL container for testing
func StartPostgreSQLContainer(ctx context.Context) (*PostgreSQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	connectionString := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", hostIP, mappedPort.Port())

	return &PostgreSQLContainer{
		Container:        container,
		ConnectionString: connectionString,
	}, nil
}

// TestFixture provides common test utilities
type TestFixture struct {
	T   *testing.T
	Ctx context.Context
}

// NewTestFixture creates a new test fixture
func NewTestFixture(t *testing.T) *TestFixture {
	return &TestFixture{
		T:   t,
		Ctx: context.Background(),
	}
}

// ProjectBuilder helps build test projects
type ProjectBuilder struct {
	title       string
	description *string
	tags        []string
}

// NewProjectBuilder creates a new project builder
func NewProjectBuilder() *ProjectBuilder {
	return &ProjectBuilder{
		title: "Test Project",
	}
}

// WithTitle sets the project title
func (b *ProjectBuilder) WithTitle(title string) *ProjectBuilder {
	b.title = title
	return b
}

// WithDescription sets the project description
func (b *ProjectBuilder) WithDescription(description string) *ProjectBuilder {
	b.description = &description
	return b
}

// WithTags sets the project tags
func (b *ProjectBuilder) WithTags(tags ...string) *ProjectBuilder {
	b.tags = tags
	return b
}

// Build returns the configured project data
func (b *ProjectBuilder) Build() (string, *string, []string) {
	return b.title, b.description, b.tags
}

// AssertTimeWithinRange asserts that a timestamp is within an acceptable range
func AssertTimeWithinRange(t *testing.T, expected, actual time.Time, delta time.Duration) {
	t.Helper()
	
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	
	if diff > delta {
		t.Errorf("Time difference too large: expected %v, actual %v, delta %v, diff %v", 
			expected, actual, delta, diff)
	}
}

// CleanupFunc is a function type for cleanup operations
type CleanupFunc func()

// NoopCleanup is a no-operation cleanup function
func NoopCleanup() {}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}