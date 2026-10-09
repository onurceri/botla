package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestPostgresPlanRepo_UpdatePlanLimits_RejectsInvalidFields verifies that the
// repository refuses to build SQL for keys outside the plan_limits allowlist.
// A nil pool is used: any attempt to reach the database would panic.
func TestPostgresPlanRepo_UpdatePlanLimits_RejectsInvalidFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		updates map[string]interface{}
	}{
		{name: "empty", updates: map[string]interface{}{}},
		{name: "unknown column", updates: map[string]interface{}{"unknown": 1}},
		{name: "injection in key", updates: map[string]interface{}{"max_chatbots = 1; DROP TABLE plans; --": 1}},
		{name: "non-updatable column", updates: map[string]interface{}{"plan_id": "other"}},
		{name: "wrong value type", updates: map[string]interface{}{"refresh_enabled": "yes"}},
	}

	repo := NewPostgresPlanRepo(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := repo.UpdatePlanLimits(context.Background(), "plan-1", tt.updates)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, models.ErrInvalidPlanLimitUpdate))
		})
	}
}
