package internal

import (
	"context"
	"fmt"

	"github.com/teammachinist/tutuplapak/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*database.Queries
	Pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
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
