package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog/log"
)

// Database wraps a SQL database connection
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(databaseURL string) (*Database, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("database connection established")

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB instance
func (d *Database) DB() *sql.DB {
	return d.db
}

// HealthCheck checks if the database is accessible
func (d *Database) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Migrate runs database migrations
func (d *Database) Migrate(ctx context.Context) error {
	log.Info().Msg("running database migrations")

	// Create projects table
	createProjectsTable := `
		CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(200) NOT NULL CHECK (char_length(title) > 0),
			description TEXT,
			tags JSONB DEFAULT '[]'::jsonb,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			published_at TIMESTAMP WITH TIME ZONE
		);
	`

	if _, err := d.db.ExecContext(ctx, createProjectsTable); err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}

	// Create index on created_at for sorting
	createProjectsIndex := `
		CREATE INDEX IF NOT EXISTS idx_projects_created_at 
		ON projects (created_at DESC);
	`

	if _, err := d.db.ExecContext(ctx, createProjectsIndex); err != nil {
		return fmt.Errorf("failed to create projects index: %w", err)
	}

	// Create updated_at trigger function
	createUpdatedAtFunction := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`

	if _, err := d.db.ExecContext(ctx, createUpdatedAtFunction); err != nil {
		return fmt.Errorf("failed to create updated_at function: %w", err)
	}

	// Create trigger for projects
	createProjectsUpdatedAtTrigger := `
		DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
		CREATE TRIGGER update_projects_updated_at
			BEFORE UPDATE ON projects
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`

	if _, err := d.db.ExecContext(ctx, createProjectsUpdatedAtTrigger); err != nil {
		return fmt.Errorf("failed to create projects updated_at trigger: %w", err)
	}

	// Create items table
	createItemsTable := `
		CREATE TABLE IF NOT EXISTS items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL CHECK (type IN ('title', 'media', 'choice', 'multi_choice', 'text_entry', 'ordering', 'hotspot')),
			title VARCHAR(500) NOT NULL CHECK (char_length(title) > 0),
			content JSONB DEFAULT '{}'::jsonb,
			position INTEGER NOT NULL CHECK (position >= 0),
			required BOOLEAN DEFAULT false,
			points INTEGER CHECK (points IS NULL OR (points >= 0 AND points <= 1000)),
			explanation TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, position)
		);
	`

	if _, err := d.db.ExecContext(ctx, createItemsTable); err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}

	// Create indexes for items
	createItemsIndexes := `
		CREATE INDEX IF NOT EXISTS idx_items_project_position 
		ON items (project_id, position ASC);
		
		CREATE INDEX IF NOT EXISTS idx_items_created_at 
		ON items (created_at DESC);
	`

	if _, err := d.db.ExecContext(ctx, createItemsIndexes); err != nil {
		return fmt.Errorf("failed to create items indexes: %w", err)
	}

	// Create trigger for items
	createItemsUpdatedAtTrigger := `
		DROP TRIGGER IF EXISTS update_items_updated_at ON items;
		CREATE TRIGGER update_items_updated_at
			BEFORE UPDATE ON items
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`

	if _, err := d.db.ExecContext(ctx, createItemsUpdatedAtTrigger); err != nil {
		return fmt.Errorf("failed to create items updated_at trigger: %w", err)
	}

	log.Info().Msg("database migrations completed successfully")
	return nil
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Error().Err(rbErr).Msg("failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}