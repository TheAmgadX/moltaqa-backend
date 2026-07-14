package domain

import (
	"time"
)

type UserUpdate struct {
	Id string

	Username        *string
	ProfileImageUrl *string
	Bio             *string
	DisplayName     *string
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
