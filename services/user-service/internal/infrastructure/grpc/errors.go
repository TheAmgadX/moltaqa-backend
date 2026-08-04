package grpc

import (
	"errors"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
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
	// domain errors:
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domain.ErrInvalidUserInput):
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
