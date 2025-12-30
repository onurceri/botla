package processing

import (
	"errors"
	"fmt"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	pkgErrors "github.com/onurceri/botla-co/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsRetryableError(t *testing.T) {
	// Legacy string-based tests (for backwards compatibility during transition)
	legacyTests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"connection refused", errors.New("connection refused"), true},
		{"timeout waiting for response", errors.New("timeout waiting for response"), true},
		{"rate limit exceeded", errors.New("rate limit exceeded"), true},
		{"status 429", errors.New("status 429"), true},
		{"invalid URL", errors.New("invalid URL"), false},
		{"parse error", errors.New("parse error"), false},
		{"unauthorized", errors.New("unauthorized"), false},
		{"context deadline exceeded", errors.New("context deadline exceeded"), true},
	}

	for _, tt := range legacyTests {
		t.Run("legacy_"+tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.retryable, result, "isRetryableError(%v)", tt.err)
		})
	}
}

func TestIsRetryableError_SentinelErrors(t *testing.T) {
	// Sentinel error tests - these are the NEW preferred way
	sentinelTests := []struct {
		name      string
		err       error
		retryable bool
	}{
		// Direct sentinel errors
		{"ErrRateLimit", pkgErrors.ErrRateLimit, true},
		{"ErrNetwork", pkgErrors.ErrNetwork, true},
		{"ErrTimeout", pkgErrors.ErrTimeout, true},
		{"ErrContextCancelled - NOT retryable", pkgErrors.ErrContextCancelled, false},
		{"ErrNotFound - NOT retryable", pkgErrors.ErrNotFound, false},

		// Wrapped sentinel errors (single wrap)
		{"wrapped ErrRateLimit", fmt.Errorf("API call failed: %w", pkgErrors.ErrRateLimit), true},
		{"wrapped ErrNetwork", fmt.Errorf("connection failed: %w", pkgErrors.ErrNetwork), true},
		{"wrapped ErrTimeout", fmt.Errorf("request timed out: %w", pkgErrors.ErrTimeout), true},
		{"wrapped ErrContextCancelled", fmt.Errorf("cancelled: %w", pkgErrors.ErrContextCancelled), false},
		{"wrapped ErrNotFound", fmt.Errorf("not found: %w", pkgErrors.ErrNotFound), false},

		// Double-wrapped sentinel errors
		{"double-wrapped ErrRateLimit", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", pkgErrors.ErrRateLimit)), true},
		{"double-wrapped ErrNetwork", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", pkgErrors.ErrNetwork)), true},
	}

	for _, tt := range sentinelTests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.retryable, result, "isRetryableError(%v)", tt.err)
		})
	}
}

func TestIsRetryableError_NilError(t *testing.T) {
	result := isRetryableError(nil)
	assert.False(t, result, "nil error should not be retryable")
}

func TestGetNextStep(t *testing.T) {
	tests := []struct {
		current models.TrainingStep
		next    models.TrainingStep
	}{
		{models.StepFetchSource, models.StepParseContent},
		{models.StepParseContent, models.StepChunkText},
		{models.StepChunkText, models.StepEmbedChunks},
		{models.StepEmbedChunks, models.StepStoreVectors},
	}

	for _, tt := range tests {
		result := getNextStep(tt.current)
		assert.Equal(t, tt.next, result, "getNextStep(%s)", tt.current)
	}
}
