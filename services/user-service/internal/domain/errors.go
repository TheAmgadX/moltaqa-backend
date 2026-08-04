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

// Helper function to return dynamic validation errors easily
func NewValidationError(format string, args ...any) error {
	return &ValidationError{
		Message: fmt.Sprintf(format, args...),
	}
}

// Domain Errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrInvalidUserInput = errors.New("invalid user input")

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
		return ErrUserNotFound

	case errors.Is(err, utils_postgres.ErrAlreadyExists):
		return ErrUserAlreadyExists

	case errors.Is(err, utils_postgres.ErrInvalidInput):
		return ErrInvalidUserInput

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
