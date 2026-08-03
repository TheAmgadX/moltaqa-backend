package kafka

type Topic string

func (t Topic) String() string {
	return string(t)
}

const (
	// User Topics
	UserCreated Topic = "user.created"

	// Auth notification topics
	AuthSendEmail Topic = "auth.send_email"
	AuthSendSMS   Topic = "auth.send_sms"
)

func Topics() []string {
	return []string{
		UserCreated.String(),
		AuthSendEmail.String(),
		AuthSendSMS.String(),
	}
}
