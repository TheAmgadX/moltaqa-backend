package repository

import (
	"context"
	"errors"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/auth-service/internal/domain"
	utils_postgres "github.com/TheAmgadX/moltaqa-backend/shared/utils/postgres"
	"github.com/jackc/pgx/v5"
)

func (r *AuthPostgresRepository) CreateRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error {
	if refreshToken == nil {
		return domain.ErrRefreshTokenInvalid
	}

	query := `
		INSERT INTO refresh_tokens (
			user_id,
			refresh_token_hash,
			created_at,
			expires_at,
			last_used_at,
			revoked_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		refreshToken.UserId,
		refreshToken.RefreshTokenHash,
		refreshToken.CreatedAt,
		refreshToken.ExpiresAt,
		refreshToken.LastUsedAt,
		refreshToken.RevokedAt,
	)

	return utils_postgres.MapDBErrorToServiceError(err)
}

func (r *AuthPostgresRepository) DeleteRefreshToken(ctx context.Context, refreshTokenHash string) error {
	if refreshTokenHash == "" {
		return domain.ErrRefreshTokenInvalid
	}

	query := `
		DELETE FROM refresh_tokens
		WHERE refresh_token_hash = $1
	`

	result, err := r.db.Exec(ctx, query, refreshTokenHash)
	if err != nil {
		return utils_postgres.MapDBErrorToServiceError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *AuthPostgresRepository) UpdateRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error {
	if refreshToken == nil {
		return domain.ErrRefreshTokenInvalid
	}

	if refreshToken.RefreshTokenHash == "" {
		return domain.ErrRefreshTokenInvalid
	}

	query := `
		UPDATE refresh_tokens
		SET
			last_used_at = $1,
			revoked_at = $2
		WHERE refresh_token_hash = $3
	`

	result, err := r.db.Exec(
		ctx,
		query,
		refreshToken.LastUsedAt,
		refreshToken.RevokedAt,
		refreshToken.RefreshTokenHash,
	)
	if err != nil {
		return utils_postgres.MapDBErrorToServiceError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *AuthPostgresRepository) GetRefreshToken(ctx context.Context, refreshTokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT
			user_id,
			refresh_token_hash,
			created_at,
			expires_at,
			last_used_at,
			revoked_at
		FROM refresh_tokens
		WHERE refresh_token_hash = $1
	`

	var (
		userID           string
		scannedTokenHash string
		createdAt        time.Time
		expiresAt        time.Time
		lastUsedAt       time.Time
		revokedAt        *time.Time
	)

	err := r.db.QueryRow(ctx, query, refreshTokenHash).Scan(
		&userID,
		&scannedTokenHash,
		&createdAt,
		&expiresAt,
		&lastUsedAt,
		&revokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRefreshTokenNotFound
		}

		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	return &domain.RefreshToken{
		UserId:           userID,
		RefreshTokenHash: scannedTokenHash,
		CreatedAt:        createdAt,
		ExpiresAt:        expiresAt,
		LastUsedAt:       lastUsedAt,
		RevokedAt:        revokedAt,
	}, nil
}
