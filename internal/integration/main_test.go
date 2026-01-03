package integration

import (
	"os"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestMain(m *testing.M) {
	// Global setup: ensure clean state from potential previous aborted runs
	// We run this once before any tests start.
	// This cleans up any 'botla_it_*' schemas left over from previous runs.
	if err := fixtures.CleanupAllIntegrationSchemas(); err != nil {
		// Just log error but proceed, individual tests might still work or fail with more specific errors
		// fmt.Fprintf(os.Stderr, "Warning: failed to clean up schemas before tests: %v\n", err)
	}

	exitCode := m.Run()

	// Optional: We could cleanup after, but keeping schemas on failure can be useful for debugging.
	// fixtures.CleanupAllIntegrationSchemas()

	os.Exit(exitCode)
}
