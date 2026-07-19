package kafka

type Topic string

func (t Topic) String() string {
	return string(t)
}

const (
	// User Topics
	UserCreated Topic = "user.created"
)

func Topics() []string {
	return []string{UserCreated.String()}
}
