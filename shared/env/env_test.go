package env_test

import (
	"testing"

	"github.com/TheAmgadX/moltaqa-backend/shared/env"
)

// ── GetString ─────────────────────────────────────────────────────────────────

func TestGetString_SetEnv_ReturnsValue(t *testing.T) {
	t.Setenv("TEST_STR_KEY", "hello")
	got := env.GetString("TEST_STR_KEY", "fallback")
	if got != "hello" {
		t.Fatalf("want %q, got %q", "hello", got)
	}
}

func TestGetString_UnsetEnv_ReturnsFallback(t *testing.T) {
	got := env.GetString("TEST_STR_MISSING_XYZ", "fallback_value")
	if got != "fallback_value" {
		t.Fatalf("want fallback %q, got %q", "fallback_value", got)
	}
}

func TestGetString_EmptyValue_ReturnsEmptyString(t *testing.T) {
	t.Setenv("TEST_STR_EMPTY", "")
	got := env.GetString("TEST_STR_EMPTY", "fallback")
	// An explicitly set-but-empty env var should return "" not the fallback.
	if got != "" {
		t.Fatalf("want empty string for set-but-empty var, got %q", got)
	}
}

// ── GetInt ────────────────────────────────────────────────────────────────────

func TestGetInt_SetEnv_ReturnsIntValue(t *testing.T) {
	t.Setenv("TEST_INT_KEY", "42")
	got := env.GetInt("TEST_INT_KEY", 0)
	if got != 42 {
		t.Fatalf("want 42, got %d", got)
	}
}

func TestGetInt_UnsetEnv_ReturnsFallback(t *testing.T) {
	got := env.GetInt("TEST_INT_MISSING_XYZ", 99)
	if got != 99 {
		t.Fatalf("want fallback 99, got %d", got)
	}
}

func TestGetInt_InvalidValue_ReturnsFallback(t *testing.T) {
	t.Setenv("TEST_INT_INVALID", "not-a-number")
	got := env.GetInt("TEST_INT_INVALID", 7)
	if got != 7 {
		t.Fatalf("want fallback 7 for non-numeric value, got %d", got)
	}
}

func TestGetInt_ZeroValue_ReturnsZero(t *testing.T) {
	t.Setenv("TEST_INT_ZERO", "0")
	got := env.GetInt("TEST_INT_ZERO", 99)
	if got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestGetInt_NegativeValue_ReturnsNegative(t *testing.T) {
	t.Setenv("TEST_INT_NEGATIVE", "-5")
	got := env.GetInt("TEST_INT_NEGATIVE", 0)
	if got != -5 {
		t.Fatalf("want -5, got %d", got)
	}
}

// ── GetBool ───────────────────────────────────────────────────────────────────

func TestGetBool_True_ReturnsTrue(t *testing.T) {
	t.Setenv("TEST_BOOL_KEY", "true")
	got := env.GetBool("TEST_BOOL_KEY", false)
	if !got {
		t.Fatal("want true, got false")
	}
}

func TestGetBool_False_ReturnsFalse(t *testing.T) {
	t.Setenv("TEST_BOOL_FALSE", "false")
	got := env.GetBool("TEST_BOOL_FALSE", true)
	if got {
		t.Fatal("want false, got true")
	}
}

func TestGetBool_UnsetEnv_ReturnsFallback(t *testing.T) {
	got := env.GetBool("TEST_BOOL_MISSING_XYZ", true)
	if !got {
		t.Fatal("want fallback true for unset env var")
	}
}

func TestGetBool_InvalidValue_ReturnsFallback(t *testing.T) {
	t.Setenv("TEST_BOOL_INVALID", "not-a-bool")
	got := env.GetBool("TEST_BOOL_INVALID", true)
	if !got {
		t.Fatal("want fallback true for invalid bool string")
	}
}

func TestGetBool_NumericTrue_ReturnsTrue(t *testing.T) {
	// strconv.ParseBool accepts "1" as true.
	t.Setenv("TEST_BOOL_ONE", "1")
	got := env.GetBool("TEST_BOOL_ONE", false)
	if !got {
		t.Fatal("want true for \"1\", got false")
	}
}

func TestGetBool_NumericFalse_ReturnsFalse(t *testing.T) {
	// strconv.ParseBool accepts "0" as false.
	t.Setenv("TEST_BOOL_ZERO", "0")
	got := env.GetBool("TEST_BOOL_ZERO", true)
	if got {
		t.Fatal("want false for \"0\", got true")
	}
}
