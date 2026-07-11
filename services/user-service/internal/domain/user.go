package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountBadgeType int

const (
	UNVERIFIED AccountBadgeType = iota
	BLUE_BADGE
	GOLDEN_BADGE
	SILVER_BADGE
)

type User struct {
	Id              uuid.UUID
	Username        string
	PhoneNumber     string
	Email           string
	ProfileImageUrl string
	Bio             string
	DisplayName     string
	EmailVerified   bool
	PhoneVerified   bool
	BirthDate       time.Time
	BioStatus       string
	AccountBadge    AccountBadgeType
	FriendsCount    uint32
	FollowersCount  uint32
	FollowingCount  uint32
	PostsCount      uint32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       time.Time
}
