package domain

import "errors"

var (
	ErrOTPTransactionNotFound = errors.New("otp transaction not found")
	ErrOTPInvalid             = errors.New("invalid otp")
	ErrOTPExpired             = errors.New("otp expired")
	ErrOTPMaxAttemptsExceeded = errors.New("otp max attempts exceeded")
	ErrOTPResendCooldown      = errors.New("otp resend cooldown active")

	ErrSessionNotFound = errors.New("session not found")
	ErrSessionRevoked  = errors.New("session revoked")
	ErrSessionExpired  = errors.New("session expired")

	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenInvalid  = errors.New("invalid refresh token")

	ErrInvalidAction    = errors.New("invalid auth action")
	ErrInvalidRecipient = errors.New("invalid email or phone")
)
