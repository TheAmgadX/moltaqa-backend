package domain

type LookupType int

const (
	LookUpId LookupType = iota
	LookupUsername
	LookupEmail
	LookupPhone
)

type Lookup struct {
	Type  LookupType
	Value string
}

func (l Lookup) TypeString() string {
	switch l.Type {
	case LookUpId:
		return "id"
	case LookupUsername:
		return "username"
	case LookupEmail:
		return "email"
	case LookupPhone:
		return "phone"
	default:
		return ""
	}
}
