package domain

import "fmt"

// Notification is the auth notification event consumed from Kafka.
type Notification struct {
	Email  string
	Phone  string
	OTP    string
	Action string
}

func (n Notification) ValidateEmail() error {
	if n.Email == "" {
		return fmt.Errorf("email recipient is required")
	}
	if n.OTP == "" {
		return fmt.Errorf("OTP is required")
	}
	return nil
}

func (n Notification) ValidateSMS() error {
	if n.Phone == "" {
		return fmt.Errorf("SMS recipient is required")
	}
	if n.OTP == "" {
		return fmt.Errorf("OTP is required")
	}
	return nil
}
