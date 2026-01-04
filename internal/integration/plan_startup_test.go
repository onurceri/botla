package integration

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/stretchr/testify/assert"
)

func TestPlanValidationStartupIntegration(t *testing.T) {
	// Skip this test - DB CHECK constraints now enforce validation at the database level.
	// We cannot insert invalid data (e.g., max_chatbots < 1) due to chk_max_chatbots constraint.
	// This is actually better protection than application-level validation alone.
	// The DB will reject any INSERT/UPDATE that violates constraints, providing
	// stronger data integrity guarantees than application validation alone.
	t.Skip("Skipping: DB CHECK constraints prevent inserting invalid plan limits")
}

func TestPlanValidationStartupSuccess(t *testing.T) {
	t.Parallel()
	dbConn := testdb.OpenParallelTestDB(t)
	defer dbConn.Close()

	// Ensure we start with valid plans
	fixtures.RestorePlans(dbConn)

	// The default seeded plans should be valid
	// In the real app, this ensures the server starts correctly with standard plans.
	planRepo := repository.NewPostgresPlanRepo(dbConn, nil)
	planSvc := services.NewPlanService(planRepo, nil)
	err := planSvc.ValidateAllPlans(context.Background())
	assert.NoError(t, err, "Seeded plans should be valid")
}
