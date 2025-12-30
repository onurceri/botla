package fixtures_test

import (
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelpers_CreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(env)

	email := "helper_test@example.com"
	user, err := env.CreateUser(email)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, email, user.Email)
	assert.NotNil(t, user.PlanID)

	var planCode string
	err = env.DB.QueryRow("SELECT code FROM plans WHERE id = $1", user.PlanID).Scan(&planCode)
	require.NoError(t, err)
	assert.Equal(t, policy.PlanFree.String(), planCode)
}

func TestHelpers_CreateChatbot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(env)

	user, err := env.CreateUser("bot_owner@example.com")
	require.NoError(t, err)

	botName := "MyTestBot"
	bot, err := env.CreateChatbot(user, botName)
	require.NoError(t, err)

	assert.NotEmpty(t, bot.ID)
	assert.Equal(t, botName, bot.Name)
	assert.Equal(t, user.ID, bot.UserID)
	assert.Equal(t, "manual", bot.RefreshPolicy)
}

func TestHelpers_CreateSource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(env)

	user, err := env.CreateUser("source_owner@example.com")
	require.NoError(t, err)

	bot, err := env.CreateChatbot(user, "SourceBot")
	require.NoError(t, err)

	sourceURL := "http://example.com"
	source, err := env.CreateSource(bot, sourceURL)
	require.NoError(t, err)

	assert.NotEmpty(t, source.ID)
	assert.Equal(t, bot.ID, source.ChatbotID)
	assert.Equal(t, "website", source.SourceType)
	assert.NotNil(t, source.SourceURL)
	assert.Equal(t, sourceURL, *source.SourceURL)
	assert.Equal(t, "completed", source.Status)
}

func TestHelpers_AuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(env)

	email := "auth_test@example.com"
	token, err := env.AuthToken(email)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestHelpers_UpdateUserPlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(env)

	user, err := env.CreateUser("plan_test@example.com")
	require.NoError(t, err)

	err = env.UpdateUserPlan(user.Email, "pro")
	require.NoError(t, err)

	var planCode string
	err = env.DB.QueryRow("SELECT code FROM plans WHERE id = (SELECT plan_id FROM users WHERE email = $1)", user.Email).Scan(&planCode)
	require.NoError(t, err)
	assert.Equal(t, "pro", planCode)
}

func TestHelpers_CreateChatbotWithConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env, err := fixtures.SetupTestEnv()
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(env)

	user, err := env.CreateUser("config_test@example.com")
	require.NoError(t, err)

	opts := map[string]any{
		"discovery_mode":  "pending",
		"handoff_enabled": true,
	}
	bot, err := env.CreateChatbotWithConfig(user, "Config Bot", opts)
	require.NoError(t, err)

	assert.NotEmpty(t, bot.ID)
	assert.Equal(t, "Config Bot", bot.Name)
	assert.Equal(t, "pending", bot.DiscoveryMode)
	assert.True(t, bot.HandoffEnabled)
}
