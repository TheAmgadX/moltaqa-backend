package grpc

import (
	"errors"

	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapServiceError(err error) error {
	if err == nil {
		return nil
	}

	var valErr *domain.ValidationError
	if errors.As(err, &valErr) {
		return status.Error(codes.InvalidArgument, valErr.Error())
	}

	switch {
	case errors.Is(err, domain.ErrOTPTransactionNotFound),
		errors.Is(err, domain.ErrRefreshTokenNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrAuthAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domain.ErrOTPMaxAttemptsExceeded):
		return status.Error(codes.ResourceExhausted, err.Error())

	case errors.Is(err, domain.ErrOTPResendCooldown):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, domain.ErrRefreshTokenExpired),
		errors.Is(err, domain.ErrRefreshTokenInvalid):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, domain.ErrOTPInvalid),
		errors.Is(err, domain.ErrInvalidAction),
		errors.Is(err, domain.ErrInvalidRecipient),
		errors.Is(err, domain.ErrInvalidAuthInput):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, domain.ErrRequestTimeout):
		return status.Error(codes.DeadlineExceeded, err.Error())

	case errors.Is(err, domain.ErrServiceUnavailable):
		return status.Error(codes.Unavailable, err.Error())

	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
