package domain

import "time"

type RefreshToken struct {
	UserId           string
	RefreshTokenHash string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	LastUsedAt       time.Time
	RevokedAt        *time.Time
}

func NewRefreshToken(userId, refreshTokenHash string, createdAt, expiresAt time.Time) RefreshToken {
	return RefreshToken{
		UserId:           userId,
		RefreshTokenHash: refreshTokenHash,
		CreatedAt:        createdAt,
		ExpiresAt:        expiresAt,
		LastUsedAt:       createdAt,
		RevokedAt:        nil,
	}
}
