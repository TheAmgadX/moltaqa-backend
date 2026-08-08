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

	// Auth notification topics
	AuthSendEmail Topic = "auth.send_email"
	AuthSendSMS   Topic = "auth.send_sms"
)

func Topics() []string {
	return []string{
		UserRegistered.String(), ContactRegistered.String(),
		UserUpdated.String(), UserDeleted.String(), UserRestored.String(),
		AuthSendEmail.String(),
		AuthSendSMS.String(),
	}
}
