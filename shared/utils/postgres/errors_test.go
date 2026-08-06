package utils_postgres_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	utils_postgres "github.com/TheAmgadX/moltaqa-backend/shared/utils/postgres"
)

// pgErr is a convenience constructor that builds a *pgconn.PgError with a
// given Postgres error code so the tests stay readable.
func pgErr(code string) *pgconn.PgError {
	return &pgconn.PgError{Code: code}
}

func TestMapDBErrorToServiceError_NilInput(t *testing.T) {
	if got := utils_postgres.MapDBErrorToServiceError(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %v", got)
	}
}

func TestMapDBErrorToServiceError_ErrNoRows(t *testing.T) {
	got := utils_postgres.MapDBErrorToServiceError(pgx.ErrNoRows)
	if !errors.Is(got, utils_postgres.ErrNotFound) {
		t.Fatalf("pgx.ErrNoRows → want ErrNotFound, got %v", got)
	}
}

func TestMapDBErrorToServiceError_PostgresCodes(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr error
	}{
		// ── Already exists ───────────────────────────────────────────────────
		{name: "unique violation → ErrAlreadyExists", code: "23505", wantErr: utils_postgres.ErrAlreadyExists},

		// ── Conflict ─────────────────────────────────────────────────────────
		{name: "foreign key violation → ErrConflict", code: "23503", wantErr: utils_postgres.ErrConflict},
		{name: "restrict violation → ErrConflict", code: "23001", wantErr: utils_postgres.ErrConflict},
		{name: "exclusion violation → ErrConflict", code: "23P01", wantErr: utils_postgres.ErrConflict},
		{name: "serialization failure → ErrConflict", code: "40001", wantErr: utils_postgres.ErrConflict},
		{name: "deadlock detected → ErrConflict", code: "40P01", wantErr: utils_postgres.ErrConflict},
		{name: "lock not available → ErrConflict", code: "55P03", wantErr: utils_postgres.ErrConflict},

		// ── Invalid input ─────────────────────────────────────────────────────
		{name: "not null violation → ErrInvalidInput", code: "23502", wantErr: utils_postgres.ErrInvalidInput},
		{name: "check violation → ErrInvalidInput", code: "23514", wantErr: utils_postgres.ErrInvalidInput},
		{name: "invalid text representation → ErrInvalidInput", code: "22P02", wantErr: utils_postgres.ErrInvalidInput},
		{name: "numeric value out of range → ErrInvalidInput", code: "22003", wantErr: utils_postgres.ErrInvalidInput},
		{name: "string data right truncation → ErrInvalidInput", code: "22001", wantErr: utils_postgres.ErrInvalidInput},
		{name: "invalid datetime format → ErrInvalidInput", code: "22007", wantErr: utils_postgres.ErrInvalidInput},
		{name: "division by zero → ErrInvalidInput", code: "22012", wantErr: utils_postgres.ErrInvalidInput},

		// ── Permission denied ─────────────────────────────────────────────────
		{name: "insufficient privilege → ErrPermissionDenied", code: "42501", wantErr: utils_postgres.ErrPermissionDenied},
		{name: "invalid authorization spec → ErrPermissionDenied", code: "28000", wantErr: utils_postgres.ErrPermissionDenied},
		{name: "invalid password → ErrPermissionDenied", code: "28P01", wantErr: utils_postgres.ErrPermissionDenied},

		// ── Timeout ───────────────────────────────────────────────────────────
		{name: "query canceled → ErrTimeout", code: "57014", wantErr: utils_postgres.ErrTimeout},

		// ── Unavailable ───────────────────────────────────────────────────────
		{name: "too many connections → ErrUnavailable", code: "53300", wantErr: utils_postgres.ErrUnavailable},
		{name: "connection failure → ErrUnavailable", code: "08006", wantErr: utils_postgres.ErrUnavailable},
		{name: "connection refused → ErrUnavailable", code: "08001", wantErr: utils_postgres.ErrUnavailable},
		{name: "connection timeout → ErrUnavailable", code: "08004", wantErr: utils_postgres.ErrUnavailable},
		{name: "connection does not exist → ErrUnavailable", code: "08003", wantErr: utils_postgres.ErrUnavailable},
		{name: "disk full → ErrUnavailable", code: "53100", wantErr: utils_postgres.ErrUnavailable},
		{name: "out of memory → ErrUnavailable", code: "53200", wantErr: utils_postgres.ErrUnavailable},
		{name: "configuration limit → ErrUnavailable", code: "53400", wantErr: utils_postgres.ErrUnavailable},
		{name: "admin shutdown → ErrUnavailable", code: "57P01", wantErr: utils_postgres.ErrUnavailable},
		{name: "crash shutdown → ErrUnavailable", code: "57P02", wantErr: utils_postgres.ErrUnavailable},
		{name: "cannot connect now → ErrUnavailable", code: "57P03", wantErr: utils_postgres.ErrUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := utils_postgres.MapDBErrorToServiceError(pgErr(tc.code))
			if !errors.Is(got, tc.wantErr) {
				t.Fatalf("code %q: want %v, got %v", tc.code, tc.wantErr, got)
			}
		})
	}
}

// Programming errors should be passed through unchanged (not masked as domain
// errors) so they surface loudly in logs.
func TestMapDBErrorToServiceError_ProgrammingErrorsPassThrough(t *testing.T) {
	programmingCodes := []string{
		"42601", // syntax error
		"42703", // undefined column
		"42P01", // undefined table
		"42701", // duplicate column
		"42P07", // duplicate table
		"25P02", // in failed SQL transaction
	}

	for _, code := range programmingCodes {
		t.Run("code "+code+" passes through", func(t *testing.T) {
			original := pgErr(code)
			got := utils_postgres.MapDBErrorToServiceError(original)
			// Must be the same error instance — not wrapped or replaced.
			if got != original {
				t.Fatalf("code %q should pass through unchanged, got %v", code, got)
			}
		})
	}
}

// An entirely unknown Postgres code or a plain Go error should pass through.
func TestMapDBErrorToServiceError_UnknownError(t *testing.T) {
	plain := errors.New("some unknown error")
	got := utils_postgres.MapDBErrorToServiceError(plain)
	if got != plain {
		t.Fatalf("unknown error should pass through, got %v", got)
	}
}

func TestMapDBErrorToServiceError_UnknownPostgresCode(t *testing.T) {
	unknown := pgErr("99999")
	got := utils_postgres.MapDBErrorToServiceError(unknown)
	// Unknown pg codes fall to the default → returned as-is.
	if got != unknown {
		t.Fatalf("unknown pg code should pass through, got %v", got)
	}
}
