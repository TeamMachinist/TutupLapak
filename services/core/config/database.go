package config

import (
	"context"
	"fmt"
	"log"
	"tutuplapak-core/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewDatabase(cfg *Config) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := db.New(pool)

	log.Println("Database connected successfully with sqlc")

	return &Database{
		Pool:    pool,
		Queries: queries,
	}, nil
}

func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		log.Println("Database connection closed")
	}
}
