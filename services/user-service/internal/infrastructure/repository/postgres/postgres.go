package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPostgresRepository struct {
	db *pgxpool.Pool
}

func NewUserPostgresRepository(db *pgxpool.Pool) *UserPostgresRepository {
	return &UserPostgresRepository{db: db}
}
