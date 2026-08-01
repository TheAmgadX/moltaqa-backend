package domain

import "time"

type Session struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func NewSession(id, userId string, createdAt, expiresAt time.Time) Session {
	return Session{
		Id:        id,
		UserId:    userId,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		RevokedAt: nil,
	}
}
