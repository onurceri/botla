package integration

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/policy"
	"github.com/stretchr/testify/assert"
)

func TestPlanValidationStartupIntegration(t *testing.T) {
t.Parallel()
	db := testdb.OpenParallelTestDB(t)
	defer db.Close()

	// Ensure we start with valid plans
	fixtures.RestorePlans(db)

	// Update the 'free' plan to be invalid
	// We use a negative value for max_chatbots which is invalid according to PlanConfig.Validate()
	_, err := db.Exec(`UPDATE plans SET config = jsonb_set(config, '{max_chatbots}', '-1'::jsonb) WHERE code = $1`, policy.PlanFree.String())
	assert.NoError(t, err)

	// This simulates the validation call in cmd/server/main.go:newApplication()
	// In the real app, this is called right after DB initialization.
	planSvc := services.NewPlanService(db, nil)
	err = planSvc.ValidateAllPlans(context.Background())

	assert.Error(t, err, "Should fail when plans are invalid")
	assert.Contains(t, err.Error(), "plan \"free\": config.max_chatbots must be >= 1")
}

func TestPlanValidationStartupSuccess(t *testing.T) {
t.Parallel()
	db := testdb.OpenParallelTestDB(t)
	defer db.Close()

	// Ensure we start with valid plans
	fixtures.RestorePlans(db)

	// The default seeded plans should be valid
	// In the real app, this ensures the server starts correctly with standard plans.
	planSvc := services.NewPlanService(db, nil)
	err := planSvc.ValidateAllPlans(context.Background())
	assert.NoError(t, err, "Seeded plans should be valid")
}
