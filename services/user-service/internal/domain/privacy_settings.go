package domain

import "github.com/google/uuid"

type Visibility string

const (
	EVERYONE Visibility = "everyone"
	FRIENDS  Visibility = "friends"
	CONTACTS Visibility = "contacts"
	NOBODY   Visibility = "nobody"
)

type PrivacySettings struct {
	UserId              uuid.UUID
	AvatarVisibility    Visibility
	PhoneVisibility     Visibility
	EmailVisibility     Visibility
	LastSeenVisibility  Visibility
	ReadReceiptsEnabled bool
	FindByUsername      bool
}

func (p *Visibility) String() string {
	if *p == EVERYONE {
		return "EVERYONE"
	}
	if *p == FRIENDS {
		return "FRIENDS"
	}
	if *p == CONTACTS {
		return "CONTACTS"
	}
	return "NOBODY"
}
