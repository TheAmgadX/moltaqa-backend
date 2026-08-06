package domain

import (
	"errors"
	"fmt"

	utils_postgres "github.com/TheAmgadX/moltaqa-backend/shared/utils/postgres"
)

// ValidationError represents a field-level dynamic validation error.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError returns dynamic validation errors.
func NewValidationError(format string, args ...any) error {
	return &ValidationError{
		Message: fmt.Sprintf(format, args...),
	}
}

var (
	ErrOTPTransactionNotFound = errors.New("otp transaction not found")
	ErrOTPInvalid             = errors.New("invalid otp")
	ErrOTPExpired             = errors.New("otp expired")
	ErrOTPMaxAttemptsExceeded = errors.New("otp max attempts exceeded")
	ErrOTPResendCooldown      = errors.New("otp resend cooldown active")

	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenInvalid  = errors.New("invalid refresh token")

	ErrInvalidAction    = errors.New("invalid auth action")
	ErrInvalidRecipient = errors.New("invalid email or phone")
	ErrUserNotFound     = errors.New("user not found")

	ErrAuthAlreadyExists  = errors.New("auth resource already exists")
	ErrInvalidAuthInput   = errors.New("invalid auth input")
	ErrConflict           = errors.New("state conflict")
	ErrServiceUnavailable = errors.New("database service unavailable")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrRequestTimeout     = errors.New("database request timeout")
	ErrInternal           = errors.New("internal server error")
)

// MapPostgresErrorToDomain translates generic shared DB errors to domain-specific errors
func MapPostgresErrorToDomain(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, utils_postgres.ErrNotFound):
		return ErrRefreshTokenNotFound

	case errors.Is(err, utils_postgres.ErrAlreadyExists):
		return ErrAuthAlreadyExists

	case errors.Is(err, utils_postgres.ErrInvalidInput):
		return ErrInvalidAuthInput

	case errors.Is(err, utils_postgres.ErrConflict):
		return ErrConflict

	case errors.Is(err, utils_postgres.ErrPermissionDenied):
		return ErrPermissionDenied

	case errors.Is(err, utils_postgres.ErrTimeout):
		return ErrRequestTimeout

	case errors.Is(err, utils_postgres.ErrUnavailable):
		return ErrServiceUnavailable

	default:
		// SQL syntax errors, unknown errors, or unmapped errors fall through here
		return ErrInternal
	}
}
