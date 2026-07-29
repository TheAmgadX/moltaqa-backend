package utils_postgres

import (
	"errors"

	sharederrors "github.com/TheAmgadX/moltaqa-backend/shared/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation      = "23505"
	pgForeignKeyViolation  = "23503"
	pgNotNullViolation     = "23502"
	pgCheckViolation       = "23514"
	pgRestrictViolation    = "23501"
	pgSerializationFailure = "40001"
	pgTooManyConnections   = "53300"
	pgQueryCanceled        = "57014"
	pgConnectionFailure    = "08006"
	pgConnectionRefused    = "08001"
	pgConnectionTimeout    = "08004"
)

// MapDBErrorToServiceError maps PostgreSQL errors to shared domain errors.
func MapDBErrorToServiceError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return sharederrors.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return sharederrors.ErrAlreadyExists
		case pgForeignKeyViolation, pgRestrictViolation:
			return sharederrors.ErrConflict
		case pgNotNullViolation, pgCheckViolation:
			return sharederrors.ErrInvalidInput
		case pgSerializationFailure:
			return sharederrors.ErrConflict
		case pgTooManyConnections, pgQueryCanceled, pgConnectionFailure, pgConnectionRefused, pgConnectionTimeout:
			return sharederrors.ErrUnavailable
		default:
			return err
		}
	}

	return err
}
