package utils_postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrConflict         = errors.New("conflict")
	ErrUnavailable      = errors.New("unavailable")
	ErrPermissionDenied = errors.New("permission denied") // new
	ErrTimeout          = errors.New("timeout")           // new
)

// Postgres error codes: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	// --- Class 22 — Data Exception ---
	pgStringDataRightTruncation = "22001"
	pgNumericValueOutOfRange    = "22003"
	pgInvalidDatetimeFormat     = "22007"
	pgDivisionByZero            = "22012"
	pgInvalidTextRepresentation = "22P02" // e.g. malformed UUID/int input

	// --- Class 23 — Integrity Constraint Violation ---
	pgIntegrityConstraintViolation = "23000"
	pgRestrictViolation            = "23001"
	pgNotNullViolation             = "23502"
	pgForeignKeyViolation          = "23503"
	pgUniqueViolation              = "23505"
	pgCheckViolation               = "23514"
	pgExclusionViolation           = "23P01"

	// --- Class 25 — Invalid Transaction State ---
	pgInFailedSQLTransaction = "25P02" // "current transaction is aborted"

	// --- Class 28 — Invalid Authorization Specification ---
	pgInvalidAuthorizationSpec = "28000"
	pgInvalidPassword          = "28P01"

	// --- Class 40 — Transaction Rollback ---
	pgTransactionRollback        = "40000"
	pgSerializationFailure       = "40001"
	pgDeadlockDetected           = "40P01"
	pgStatementCompletionUnknown = "40003"

	// --- Class 42 — Syntax Error or Access Rule Violation ---
	pgSyntaxError           = "42601"
	pgInsufficientPrivilege = "42501"
	pgUndefinedColumn       = "42703"
	pgUndefinedTable        = "42P01"
	pgDuplicateColumn       = "42701"
	pgDuplicateTable        = "42P07"

	// --- Class 53 — Insufficient Resources ---
	pgInsufficientResources = "53000"
	pgDiskFull              = "53100"
	pgOutOfMemory           = "53200"
	pgTooManyConnections    = "53300"
	pgConfigurationLimit    = "53400"

	// --- Class 55 — Object Not In Prerequisite State ---
	pgObjectNotInPrerequisiteState = "55000"
	pgLockNotAvailable             = "55P03"

	// --- Class 57 — Operator Intervention ---
	pgQueryCanceled    = "57014"
	pgAdminShutdown    = "57P01"
	pgCrashShutdown    = "57P02"
	pgCannotConnectNow = "57P03"

	// --- Class 58 — System Error ---
	pgConnectionFailure      = "08006"
	pgConnectionRefused      = "08001"
	pgConnectionTimeout      = "08004"
	pgConnectionDoesNotExist = "08003"
)

// MapDBErrorToServiceError maps PostgreSQL errors to shared domain errors.
func MapDBErrorToServiceError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {

		// Not found
		// (no separate PG code beyond pgx.ErrNoRows in practice)

		// Already exists
		case pgUniqueViolation:
			return ErrAlreadyExists

		// Conflict — data integrity or concurrency conflicts
		case pgForeignKeyViolation,
			pgRestrictViolation,
			pgExclusionViolation,
			pgSerializationFailure,
			pgDeadlockDetected,
			pgLockNotAvailable:
			return ErrConflict

		// Invalid input — client sent bad data
		case pgNotNullViolation,
			pgCheckViolation,
			pgInvalidTextRepresentation,
			pgNumericValueOutOfRange,
			pgStringDataRightTruncation,
			pgInvalidDatetimeFormat,
			pgDivisionByZero:
			return ErrInvalidInput

		// Permission denied
		case pgInsufficientPrivilege,
			pgInvalidAuthorizationSpec,
			pgInvalidPassword:
			return ErrPermissionDenied

		// Timeout — distinct from general unavailability
		case pgQueryCanceled:
			return ErrTimeout

		// Unavailable — connection/resource/infra problems
		case pgTooManyConnections,
			pgConnectionFailure,
			pgConnectionRefused,
			pgConnectionTimeout,
			pgConnectionDoesNotExist,
			pgDiskFull,
			pgOutOfMemory,
			pgConfigurationLimit,
			pgAdminShutdown,
			pgCrashShutdown,
			pgCannotConnectNow:
			return ErrUnavailable

		// Programming errors (bugs in your SQL, not runtime conditions) —
		// don't expose these as domain errors, let them bubble as-is so
		// they show up loudly in logs/monitoring instead of being masked.
		case pgSyntaxError,
			pgUndefinedColumn,
			pgUndefinedTable,
			pgDuplicateColumn,
			pgDuplicateTable,
			pgInFailedSQLTransaction:
			return err

		default:
			return err
		}
	}

	return err
}
