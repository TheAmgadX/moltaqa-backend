package grpc_utils

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetGRPCErrorMessage(err error) string {
	if st, ok := status.FromError(err); ok {
		return st.Message()
	}
	return err.Error()
}

func MapGRPCErrCodesToHttpErrCodes(err error) int {
	status, ok := status.FromError(err)

	if !ok {
		return http.StatusInternalServerError
	}

	var httpStatus int

	switch status.Code() {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	default:
		httpStatus = http.StatusInternalServerError
	}

	return httpStatus
}
