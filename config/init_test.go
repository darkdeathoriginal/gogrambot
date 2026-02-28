package config

import (
	"os"
	"testing"
)

func TestGetenv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	if val := Getenv("TEST_KEY", "fallback"); val != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", val)
	}

	if val := Getenv("NON_EXISTENT_KEY", "fallback"); val != "fallback" {
		t.Errorf("Expected 'fallback', got '%s'", val)
	}
}
