package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/teammachinist/tutuplapak/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*database.Queries
	Pool *pgxpool.Pool
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

	queries := database.New(pool)

	return &DB{
		Queries: queries,
		Pool:    pool,
	}, nil
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
