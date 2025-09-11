package main

import (
	"context"
	"fmt"
	"tutuplapak-files/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*db.Queries
	Pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := db.New(pool)

	return &DB{
		Queries: queries,
		Pool:    pool,
	}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}