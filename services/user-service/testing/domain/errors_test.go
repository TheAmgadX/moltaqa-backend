package domain_test

import (
	"errors"
	"testing"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	utils_postgres "github.com/TheAmgadX/moltaqa-backend/shared/utils/postgres"
)

func TestMapPostgresErrorToDomain_NilInput(t *testing.T) {
	if got := domain.MapPostgresErrorToDomain(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %v", got)
	}
}

func TestMapPostgresErrorToDomain_Mappings(t *testing.T) {
	tests := []struct {
		name    string
		input   error
		wantErr error
	}{
		{
			name:    "ErrNotFound → ErrUserNotFound",
			input:   utils_postgres.ErrNotFound,
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:    "ErrAlreadyExists → ErrUserAlreadyExists",
			input:   utils_postgres.ErrAlreadyExists,
			wantErr: domain.ErrUserAlreadyExists,
		},
		{
			name:    "ErrInvalidInput → ErrInvalidUserInput",
			input:   utils_postgres.ErrInvalidInput,
			wantErr: domain.ErrInvalidUserInput,
		},
		{
			name:    "ErrConflict → ErrConflict",
			input:   utils_postgres.ErrConflict,
			wantErr: domain.ErrConflict,
		},
		{
			name:    "ErrPermissionDenied → ErrPermissionDenied",
			input:   utils_postgres.ErrPermissionDenied,
			wantErr: domain.ErrPermissionDenied,
		},
		{
			name:    "ErrTimeout → ErrRequestTimeout",
			input:   utils_postgres.ErrTimeout,
			wantErr: domain.ErrRequestTimeout,
		},
		{
			name:    "ErrUnavailable → ErrServiceUnavailable",
			input:   utils_postgres.ErrUnavailable,
			wantErr: domain.ErrServiceUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.MapPostgresErrorToDomain(tc.input)
			if !errors.Is(got, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, got)
			}
		})
	}
}

// Any error that doesn't match a known utils_postgres sentinel should be
// collapsed to ErrInternal — not leaked to callers.
func TestMapPostgresErrorToDomain_UnknownErrorBecomesInternal(t *testing.T) {
	unknown := errors.New("some raw db driver error")
	got := domain.MapPostgresErrorToDomain(unknown)
	if !errors.Is(got, domain.ErrInternal) {
		t.Fatalf("unknown error should map to ErrInternal, got %v", got)
	}
}

// Validate that ValidationError implements the error interface correctly and
// can be dynamically constructed.
func TestValidationError_Message(t *testing.T) {
	err := domain.NewValidationError("field %s is required", "email")
	if err == nil {
		t.Fatal("expected a non-nil ValidationError")
	}
	want := "field email is required"
	if err.Error() != want {
		t.Fatalf("want %q, got %q", want, err.Error())
	}
}
