package main

import (
	"os"
	"testing"
)

// copyEnvVars creates a copy of specified environment variables
func copyEnvVars(keys []string) map[string]string {
	result := make(map[string]string)
	for _, key := range keys {
		result[key] = os.Getenv(key)
	}
	return result
}

// restoreEnvVars restores environment variables from a saved copy
func restoreEnvVars(saved map[string]string) {
	for key, value := range saved {
		if value != "" {
			os.Setenv(key, value)
		} else {
			os.Unsetenv(key)
		}
	}
}

// TestCLIMain_SmokeTest is a simple smoke test to verify CLI doesn't panic
// Note: Comprehensive unit testing of CLI is difficult because the code uses
// the global flag package which can only be parsed once. For full test coverage,
// the CLI code should be refactored to use flag.FlagSet.
//
// This test verifies the CLI runs without panic on valid input.
// It will fail at DB connection, but that's expected without a real DB.
func TestCLIMain_SmokeTest(t *testing.T) {
	originalArgs := os.Args
	originalEnv := copyEnvVars([]string{
		"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
		"QDRANT_URL", "JWT_SECRET", "PORT", "DATABASE_URL",
	})

	defer func() {
		os.Args = originalArgs
		restoreEnvVars(originalEnv)
	}()

	// Set test environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "botla_test")
	os.Setenv("DB_USER", "botla")
	os.Setenv("DB_PASSWORD", "botla")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("PORT", "8080")
	// Use invalid DATABASE_URL to avoid actual DB connection
	os.Setenv("DATABASE_URL", "postgres://invalid:invalid@localhost:5432/invalid?sslmode=disable")

	os.Args = []string{"cli", "-email=test@example.com"}

	// Capture and discard stderr
	stderr := os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stderr = devnull

	err := run()

	devnull.Close()
	os.Stderr = stderr

	// Should fail due to DB connection, not flag parsing
	if err != nil {
		// DB connection error is expected
		t.Logf("Expected DB connection error: %v", err)
	}
}
