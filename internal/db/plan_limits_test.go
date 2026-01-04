package db_test

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPlanLimitsByPlanID(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	// Get the free plan ID
	var planID string
	err := dbConn.QueryRow(`SELECT id FROM plans WHERE code = 'free'`).Scan(&planID)
	require.NoError(t, err)

	// Fetch limits
	limits, err := db.GetPlanLimitsByPlanID(ctx, dbConn, planID)
	require.NoError(t, err)
	require.NotNil(t, limits)

	// Verify values match Free plan defaults
	assert.Equal(t, planID, limits.PlanID)
	assert.Equal(t, 1, limits.MaxChatbots)
	assert.Equal(t, 50, limits.MaxMonthlyIngestions)
	assert.False(t, limits.ScrapingDynamicEnabled)
	assert.NotEmpty(t, limits.ChatDefaultModel) // May vary based on migrated data
	assert.False(t, limits.SecuritySecureEmbedEnabled)

}

func TestGetPlanLimitsByPlanID_NotFound(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	limits, err := db.GetPlanLimitsByPlanID(ctx, dbConn, "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Nil(t, limits)
}

func TestGetPlanLimitsByCode(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	t.Run("free_plan", func(t *testing.T) {
		limits, err := db.GetPlanLimitsByCode(ctx, dbConn, "free")
		require.NoError(t, err)
		require.NotNil(t, limits)
		assert.Equal(t, 1, limits.MaxChatbots)
		assert.False(t, limits.SecuritySecureEmbedEnabled)
		assert.False(t, limits.ScrapingDynamicEnabled)
		assert.Equal(t, 100, limits.RateLimitsRequestsPerMinute)
	})

	t.Run("pro_plan", func(t *testing.T) {
		limits, err := db.GetPlanLimitsByCode(ctx, dbConn, "pro")
		require.NoError(t, err)
		require.NotNil(t, limits)
		assert.Equal(t, 10, limits.MaxChatbots)
		assert.True(t, limits.SecuritySecureEmbedEnabled)
		assert.True(t, limits.ScrapingDynamicEnabled)
		assert.Equal(t, 500, limits.RateLimitsRequestsPerMinute)
	})
}

func TestGetPlanLimitsByCode_NotFound(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	limits, err := db.GetPlanLimitsByCode(ctx, dbConn, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, limits)
}

func TestUpdatePlanLimitField(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	t.Run("update_int_field", func(t *testing.T) {
		// Get original value
		originalLimits, err := db.GetPlanLimitsByCode(ctx, dbConn, "free")
		require.NoError(t, err)
		originalValue := originalLimits.MaxChatbots

		// Update max_chatbots
		err = db.UpdatePlanLimitField(ctx, dbConn, "free", "max_chatbots", 5)
		require.NoError(t, err)

		// Verify
		limits, err := db.GetPlanLimitsByCode(ctx, dbConn, "free")
		require.NoError(t, err)
		assert.Equal(t, 5, limits.MaxChatbots)

		// Reset
		_ = db.UpdatePlanLimitField(ctx, dbConn, "free", "max_chatbots", originalValue)
	})

	t.Run("update_bool_field", func(t *testing.T) {
		// Get original value
		originalLimits, err := db.GetPlanLimitsByCode(ctx, dbConn, "free")
		require.NoError(t, err)
		originalValue := originalLimits.SecuritySecureEmbedEnabled

		// Update
		err = db.UpdatePlanLimitField(ctx, dbConn, "free", "security_secure_embed_enabled", true)
		require.NoError(t, err)

		// Verify
		limits, err := db.GetPlanLimitsByCode(ctx, dbConn, "free")
		require.NoError(t, err)
		assert.True(t, limits.SecuritySecureEmbedEnabled)

		// Reset
		_ = db.UpdatePlanLimitField(ctx, dbConn, "free", "security_secure_embed_enabled", originalValue)
	})

	t.Run("update_string_field", func(t *testing.T) {
		// Get original value
		originalLimits, err := db.GetPlanLimitsByCode(ctx, dbConn, "free")
		require.NoError(t, err)
		originalValue := originalLimits.ChatDefaultModel

		// Update
		err = db.UpdatePlanLimitField(ctx, dbConn, "free", "chat_default_model", "openai/gpt-4o")
		require.NoError(t, err)

		// Verify
		limits, err := db.GetPlanLimitsByCode(ctx, dbConn, "free")
		require.NoError(t, err)
		assert.Equal(t, "openai/gpt-4o", limits.ChatDefaultModel)

		// Reset
		_ = db.UpdatePlanLimitField(ctx, dbConn, "free", "chat_default_model", originalValue)
	})
}

func TestUpdatePlanLimitField_InvalidField(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	err := db.UpdatePlanLimitField(ctx, dbConn, "free", "invalid_field", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field name")
}

func TestUpdatePlanLimitField_NonexistentPlan(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	err := db.UpdatePlanLimitField(ctx, dbConn, "nonexistent", "max_chatbots", 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no plan found")
}

func TestUpdatePlanLimitField_PreventsSQLInjection(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	// Attempt SQL injection via field name
	err := db.UpdatePlanLimitField(ctx, dbConn, "free", "max_chatbots; DROP TABLE plans;--", 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field name")

	// Verify table still exists
	var count int
	err = dbConn.QueryRow(`SELECT COUNT(*) FROM plans`).Scan(&count)
	assert.NoError(t, err)
	assert.Greater(t, count, 0)
}

func TestGetPlanWithLimits(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	t.Run("free_plan", func(t *testing.T) {
		plan, err := db.GetPlanWithLimits(ctx, dbConn, "free")
		require.NoError(t, err)
		require.NotNil(t, plan)

		// Verify plan fields
		assert.Equal(t, "free", plan.Code)
		assert.Equal(t, "active", plan.Status)

		// Verify limits attached
		require.NotNil(t, plan.Limits)
		assert.Equal(t, 1, plan.Limits.MaxChatbots)
		assert.Equal(t, plan.ID, plan.Limits.PlanID)
	})

	t.Run("pro_plan", func(t *testing.T) {
		plan, err := db.GetPlanWithLimits(ctx, dbConn, "pro")
		require.NoError(t, err)
		require.NotNil(t, plan)

		// Verify plan fields
		assert.Equal(t, "pro", plan.Code)

		// Verify limits attached
		require.NotNil(t, plan.Limits)
		assert.Equal(t, 10, plan.Limits.MaxChatbots)
		assert.True(t, plan.Limits.SecuritySecureEmbedEnabled)
	})

	t.Run("not_found", func(t *testing.T) {
		plan, err := db.GetPlanWithLimits(ctx, dbConn, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, plan)
	})
}

func TestGetAllPlansWithLimits(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	ctx := context.Background()

	plans, err := db.GetAllPlansWithLimits(ctx, dbConn)
	require.NoError(t, err)
	require.NotEmpty(t, plans)

	// Should have at least free and pro plans
	assert.GreaterOrEqual(t, len(plans), 2)

	// Plans should be sorted by price (free first)
	assert.Equal(t, "free", plans[0].Code)

	// Each plan should have limits attached
	for _, plan := range plans {
		require.NotNil(t, plan.Limits, "plan %s should have limits", plan.Code)
		assert.Equal(t, plan.ID, plan.Limits.PlanID)
		assert.Greater(t, plan.Limits.MaxChatbots, 0)
	}
}
