package domain

type Event struct {
	Email  string
	Phone  string
	OTP    string
	Action AuthAction
}

func NewEvent(email, phone, otp string, action AuthAction) Event {
	return Event{
		Email:  email,
		Phone:  phone,
		OTP:    otp,
		Action: action,
	}
}
