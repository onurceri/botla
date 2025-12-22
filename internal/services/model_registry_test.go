package services

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
)

func TestModelService_GetAvailableModels(t *testing.T) {
	db := testdb.OpenTestDB(t)

	// Ensure seed data is present (testdb should have migrations applied).
	// After migration 000040, models use bare names (e.g., "gpt-4o-mini").

	svc := NewModelService(db)

	t.Run("Returns all models when list is empty", func(t *testing.T) {
		models, err := svc.GetAvailableModels(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, models)
		assert.True(t, len(models) >= 3)
	})

	t.Run("Returns only allowed models", func(t *testing.T) {
		allowed := []string{"gpt-4o-mini"}
		models, err := svc.GetAvailableModels(context.Background(), allowed)
		assert.NoError(t, err)
		assert.Len(t, models, 1)
		assert.Equal(t, "gpt-4o-mini", models[0].ID)
	})

	t.Run("Returns multiple allowed models", func(t *testing.T) {
		allowed := []string{"gpt-4o-mini", "gpt-4o"}
		models, err := svc.GetAvailableModels(context.Background(), allowed)
		assert.NoError(t, err)
		assert.Len(t, models, 2)
	})

	t.Run("Ignores unknown models", func(t *testing.T) {
		allowed := []string{"gpt-4o-mini", "unknown-model"}
		models, err := svc.GetAvailableModels(context.Background(), allowed)
		assert.NoError(t, err)
		assert.Len(t, models, 1)
		assert.Equal(t, "gpt-4o-mini", models[0].ID)
	})
}
