package domain

import (
	"time"
)

type UserUpdate struct {
	Id string

	Username        *string
	PhoneNumber     *string
	Email           *string
	ProfileImageUrl *string
	Bio             *string
	DisplayName     *string
	EmailVerified   *bool
	PhoneVerified   *bool
	BirthDate       *time.Time
	BioStatus       *string
	AccountBadge    *AccountBadgeType
	FriendsCount    *uint32
	FollowersCount  *uint32
	FollowingCount  *uint32
	PostsCount      *uint32
}

type PrivacySettingsUpdate struct {
	UserId string

	AvatarVisibility    *Visibility
	PhoneVisibility     *Visibility
	EmailVisibility     *Visibility
	LastSeenVisibility  *Visibility
	ReadReceiptsEnabled *bool
	FindByUsername      *bool
}
