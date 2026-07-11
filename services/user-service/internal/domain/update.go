package domain

import "time"

type UserUpdate struct {
	Id string

	Username        *string
	DisplayName     *string
	Email           *string
	Phone           *string
	ProfileImageURL *string
	Bio             *string
	BioStatus       *string
	BirthDate       *time.Time
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
