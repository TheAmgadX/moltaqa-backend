package grpc

import (
	"errors"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapServiceError(err error) error {
	switch {
	// domain errors:
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domain.ErrInvalidUserId),
		errors.Is(err, domain.ErrInvalidUserInput),
		errors.Is(err, domain.ErrEmptyUserIdSlice),
		errors.Is(err, domain.ErrNothingToUpdate):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrNothingToUpdate):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
