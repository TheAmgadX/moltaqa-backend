package domain

type ContactLookupType int

const (
	ContactLookupTypeEmail ContactLookupType = iota
	ContactLookupTypePhone
)

type ContactLookup struct {
	Type  ContactLookupType
	Value string
}

// ContactRequest represents a request to register a contact for a user.
//
// Contact may be either an email or a phone number.
type ContactRequest struct {
	UserId        string
	ContactLookup ContactLookup
}

func (c ContactRequest) TypeString() string {
	switch c.ContactLookup.Type {
	case ContactLookupTypeEmail:
		return "email"
	case ContactLookupTypePhone:
		return "phone_number"
	default:
		return ""
	}
}

func (c ContactRequest) Value() string {
	return c.ContactLookup.Value
}
