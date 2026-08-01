package domain

import "time"

type RefreshToken struct {
	UserId           string
	SessionId        string
	RefreshTokenHash string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	LastUsedAt       time.Time
}

func NewRefreshToken(userId, sessionId, refreshTokenHash string, createdAt, expiresAt time.Time) RefreshToken {
	return RefreshToken{
		UserId:           userId,
		SessionId:        sessionId,
		RefreshTokenHash: refreshTokenHash,
		CreatedAt:        createdAt,
		ExpiresAt:        expiresAt,
		LastUsedAt:       createdAt,
	}
}
