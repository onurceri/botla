package middleware

import (
	"testing"
)

func TestNewRateLimiterFromEnvWithPrefix(t *testing.T) {
	// This test is for the old rate limiter API which has been replaced
	// with plan-based rate limiting. Integration tests cover the new behavior.
	t.Skip("Deprecated: Rate limiter now uses plan-based configuration")
}
