package kafka

type Topic string

func (t Topic) String() string {
	return string(t)
}

const (
	// User Topics
	UserRegistered    Topic = "user.registered"
	ContactRegistered Topic = "contact.registered"
	UserUpdated       Topic = "user.updated"
	UserDeleted       Topic = "user.deleted"
	UserRestored      Topic = "user.restored"
)

func Topics() []string {
	return []string{UserRegistered.String(), ContactRegistered.String(),
		UserUpdated.String(), UserDeleted.String(), UserRestored.String(),
	}
}
