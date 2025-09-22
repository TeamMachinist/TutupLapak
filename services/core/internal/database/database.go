package database

import (
	"context"
	"embed"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Embed directory as core services has multiple sql files

//go:embed migrations/*.sql
var migrationFS embed.FS

//go:embed seeds/seeds_data.sql
var seedSQL string

type DB struct {
	Queries *Queries
	Pool    *pgxpool.Pool
}

func NewDatabase(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Performance tuning for high RPS (58,900 total load test)
	config.MaxConns = 30                       // Maximum connections
	config.MinConns = 5                        // Keep warm connections
	config.MaxConnLifetime = 1 * time.Hour     // Recycle connections
	config.MaxConnIdleTime = 5 * time.Minute   // Close idle connections
	config.HealthCheckPeriod = 1 * time.Minute // Regular health checks

	// Connection timeout
	config.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		Queries: New(pool),
		Pool:    pool,
	}

	if err := db.initializeDatabase(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}

	return db, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

// Health check method
func (db *DB) HealthCheck(ctx context.Context) error {
	// Basic ping
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test query
	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("query test failed: %w", err)
	}

	return nil
}

// Get connection pool stats
func (db *DB) GetStats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// initializeDatabase handles migrations and seeding intelligently
func (db *DB) initializeDatabase(ctx context.Context) error {
	// Check if migration needed
	migrationNeeded, err := db.isMigrationNeeded(ctx)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if migrationNeeded {
		if err := db.runMigrations(ctx); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Check if seeding needed (development only)
	if os.Getenv("ENV") == "development" {
		seedingNeeded, err := db.isSeedingNeeded(ctx)
		if err != nil {
			return fmt.Errorf("failed to check seeding status: %w", err)
		}

		if seedingNeeded {
			if err := db.seedData(ctx); err != nil {
				return fmt.Errorf("seeding failed: %w", err)
			}
		}
	}

	return nil
}

func (db *DB) isMigrationNeeded(ctx context.Context) (bool, error) {
	// Check if main table exists
	var exists bool
	err := db.Pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'users'
        )
    `).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return !exists, nil // Need migration if table doesn't exist
}

func (db *DB) isSeedingNeeded(ctx context.Context) (bool, error) {
	// First check if table exists
	migrationNeeded, err := db.isMigrationNeeded(ctx)
	if err != nil {
		return false, err
	}

	if migrationNeeded {
		return false, nil // Can't seed if table doesn't exist
	}

	// Check if any seed data exists
	var count int
	err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check seed data: %w", err)
	}

	return count == 0, nil // Need seeding if no data exists
}

func (db *DB) runMigrations(ctx context.Context) error {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort files to ensure order (001_, 002_, etc.)
	var filenames []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)

	// Execute migrations in order
	for _, filename := range filenames {
		migrationSQL, err := migrationFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		_, err = db.Pool.Exec(ctx, string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		fmt.Printf("Applied migration: %s\n", filename)
	}

	return nil
}

func (db *DB) seedData(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx, seedSQL)
	if err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}
	return nil
}
