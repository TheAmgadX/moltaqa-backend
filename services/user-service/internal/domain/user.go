package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountBadgeType string

const (
	UNVERIFIED   AccountBadgeType = "unverified"
	BLUE_BADGE   AccountBadgeType = "blue_badge"
	GOLDEN_BADGE AccountBadgeType = "golden_badge"
	SILVER_BADGE AccountBadgeType = "silver_badge"
)

type User struct {
	Id              uuid.UUID
	Username        string
	PhoneNumber     string
	Email           string
	ProfileImageUrl string
	Bio             string
	DisplayName     string
	EmailVerified   time.Time
	PhoneVerified   time.Time
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
