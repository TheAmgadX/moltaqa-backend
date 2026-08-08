package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthPostgresRepository struct {
	db *pgxpool.Pool
}

func CreateDBConnection(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	return pool, nil
}

func NewAuthPostgresRepository(dsn string) (*AuthPostgresRepository, error) {
	db, err := CreateDBConnection(dsn)
	if err != nil {
		return nil, err
	}

	return &AuthPostgresRepository{db: db}, nil
}
