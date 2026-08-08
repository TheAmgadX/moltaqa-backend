package repository

import (
	"context"

	"github.com/TheAmgadX/moltaqa-backend/services/auth-service/internal/domain"
)

// AuthRepository defines the persistence layer for auth-related data.
type AuthRepository interface {
	// Commands
	CreateRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error
	DeleteRefreshToken(ctx context.Context, refreshTokenHash string) error
	UpdateRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error

	// Queries
	GetRefreshToken(ctx context.Context, refreshTokenHash string) (*domain.RefreshToken, error)
}
