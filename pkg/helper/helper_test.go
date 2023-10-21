package helper

import (
	"errors"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetIntEnv(t *testing.T) {
	_ = os.Setenv("TEST_INT", "123")

	// Test
	got := GetIntEnvWithDefault("TEST_INT", 0)
	if got != 123 {
		t.Errorf("GetIntEnvWithDefault() = %d; want 123", got)
	}
}

func TestGetStringEnv(t *testing.T) {
	_ = os.Setenv("TEST_STRING", "123")

	// Test
	got := GetStringEnvWithDefault("TEST_STRING", "")
	if got != "123" {
		t.Errorf("GetStringEnvWithDefault() = %s; want 123", got)
	}
}

func TestIsEnvSet(t *testing.T) {
	_ = os.Setenv("TEST_STRING", "123")

	// Test
	got := IsEnvSet("TEST_STRING")
	if !got {
		t.Errorf("IsEnvSet() = %t; want true", got)
	}

	_ = os.Unsetenv("TEST_STRING")

	// Test
	got = IsEnvSet("TEST_STRING")
	if got {
		t.Errorf("IsEnvSet() = %t; want false", got)
	}
}

func TestHandleError(t *testing.T) {
	err := errors.New("test error 1")
	errMsg := HandleError(err, false, "test error: %v", err)
	require.Contains(t, errMsg, "test error: test error 1")
}
