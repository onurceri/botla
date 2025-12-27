package processing

import (
	"errors"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		err       error
		retryable bool
	}{
		{errors.New("connection refused"), true},
		{errors.New("timeout waiting for response"), true},
		{errors.New("rate limit exceeded"), true},
		{errors.New("status 429"), true},
		{errors.New("invalid URL"), false},
		{errors.New("parse error"), false},
		{errors.New("unauthorized"), false},
		{errors.New("context deadline exceeded"), true},
	}

	for _, tt := range tests {
		result := isRetryableError(tt.err)
		assert.Equal(t, tt.retryable, result, "isRetryableError(%v)", tt.err)
	}
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
