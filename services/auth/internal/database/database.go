package database

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teammachinist/tutuplapak/services/auth/internal/logger"
)

//go:embed migrations/001_create_users_auth_table.sql
var migrationSQL string

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
		logger.InfoCtx(ctx, "Executing auth core database migration")
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
			logger.InfoCtx(ctx, "Executing auth core dummy data seeding")
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
            AND table_name = 'users_auth'
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
	err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users_auth").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check seed data: %w", err)
	}

	return count == 0, nil // Need seeding if no data exists
}

func (db *DB) runMigrations(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
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
