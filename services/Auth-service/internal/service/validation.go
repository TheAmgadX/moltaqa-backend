package service

import (
	"regexp"

	"github.com/TheAmgadX/moltaqa-backend/services/auth-service/internal/domain"
	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
)

var validate = validator.New()

var otpRegex = regexp.MustCompile(`^\d{6}$`)

func validateEmail(email string) error {
	if err := validate.Var(email, "required,email"); err != nil {
		return domain.NewValidationError("invalid email: %s", email)
	}

	return nil
}

func validatePhone(phone string) error {
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return domain.NewValidationError("invalid phone number: %s", phone)
	}

	if !phonenumbers.IsValidNumber(num) {
		return domain.NewValidationError("invalid phone number: %s", phone)
	}

	return nil
}

func validateRecipient(otpTx domain.OTPTransaction) error {
	if otpTx.Email == "" && otpTx.Phone == "" {
		return domain.NewValidationError("email or phone number is required")
	}

	if otpTx.Email != "" && otpTx.Phone != "" {
		return domain.NewValidationError("only one recipient is allowed")
	}

	if otpTx.Email != "" {
		return validateEmail(otpTx.Email)
	}

	return validatePhone(otpTx.Phone)
}

func validateOTPRequest(otpTx domain.OTPTransaction) error {
	if err := validateRecipient(otpTx); err != nil {
		return err
	}

	if !otpRegex.MatchString(otpTx.OTPHash) {
		return domain.NewValidationError("invalid otp")
	}

	return nil
}

func validateAuthAction(action domain.AuthAction) error {
	switch action {
	case domain.ActionLogin, domain.ActionRestore, domain.ActionDelete:
		return nil
	default:
		return domain.ErrInvalidAction
	}
}

func validateRefreshToken(refreshToken string) error {
	if refreshToken == "" {
		return domain.ErrRefreshTokenInvalid
	}

	return nil
}
