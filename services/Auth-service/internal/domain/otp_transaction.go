package domain

type OTPTransaction struct {
	OTPHash  string
	Email    string
	Phone    string
	Action   AuthAction
	Attempts int
}

func NewOTPTransaction(otpHash, email, phone string, action AuthAction) OTPTransaction {
	return OTPTransaction{
		OTPHash:  otpHash,
		Email:    email,
		Phone:    phone,
		Action:   action,
		Attempts: 0,
	}
}
