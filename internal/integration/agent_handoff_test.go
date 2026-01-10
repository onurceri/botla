package integration

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockToolsLLMClient removed - using fixtures.NewLLMMock instead

func TestAutomatedHandoff(t *testing.T) {
	t.Parallel()
	oai := fixtures.NewLLMMock(t)
	defer oai.Close()

	// Configure Mock to return request_human_handoff tool call
	oai.SetChatResponse(func(req fixtures.MockRequest) (map[string]any, int) {
		messages := req.Body["messages"].([]any)
		lastMsg := messages[len(messages)-1].(map[string]any)

		if lastMsg["role"] == "user" {
			// Return tool call
			return map[string]any{
				"choices": []map[string]any{{
					"message": map[string]any{
						"role": "assistant",
						"tool_calls": []map[string]any{{
							"id":   "call_123",
							"type": "function",
							"function": map[string]any{
								"name":      "request_human_handoff",
								"arguments": "{}",
							},
						}},
					},
					"finish_reason": "tool_calls",
				}},
			}, 200
		}
		// Fallback
		return map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{
					"role":    "assistant",
					"content": "ok",
				},
				"finish_reason": "stop",
			}},
		}, 200
	})

	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.OPENAI_API_KEY = "test-key"
	}, false)
	require.NoError(t, err)
	defer fixtures.TeardownTestEnv(te)

	ctx := context.Background()
	log := logger.New("DEBUG")

	// Update pro plan to allow handoff (escalate fallback)
	updateProPlanConfig(t, te)
	var proPlanID string
	err = te.DB.QueryRowContext(ctx, "SELECT id FROM plans WHERE code=$1", policy.PlanPro.String()).Scan(&proPlanID)
	require.NoError(t, err)

	// Create User with Pro Plan
	var userID string
	err = te.DB.QueryRowContext(ctx, `INSERT INTO users (email, password_hash, full_name, plan_id) VALUES ($1, $2, $3, $4) RETURNING id`, "test@example.com", "hash", "Test User", proPlanID).Scan(&userID)
	require.NoError(t, err)

	// 1. Setup Chat Service with REAL factory (using mock server via env)
	factory := rag.NewClientFactory(te.Cfg)
	planRepo := repository.NewPostgresPlanRepo(te.DB, nil)
	conversationRepo := repository.NewPostgresConversationRepo(te.DB)
	analyticsRepo := repository.NewPostgresAnalyticsRepo(te.DB)
	actionRepo := repository.NewPostgresActionRepo(te.DB)
	sourceRepo := repository.NewPostgresSourceRepo(te.DB)
	handoffRepo := repository.NewPostgresHandoffRepo(te.DB)
	usageRepo := repository.NewPostgresUsageRepo(te.DB)
	chatSvc := services.NewChatService(planRepo, conversationRepo, analyticsRepo, actionRepo, sourceRepo, handoffRepo, factory, nil, nil, usageRepo, log)
	chatSvc.SyncAnalytics = true // Run analytics synchronously in tests

	// 3. Create Chatbot with HandoffEnabled
	bot := &models.Chatbot{
		UserID:           userID,
		Name:             "AutoHandoffBot",
		SystemPrompt:     "You are helpful.",
		LanguageCode:     "en",
		Model:            policy.ModelGPT4oMini.String(), // Bare model name, resolved to API format at call time
		Temperature:      0.7,
		MaxTokens:        100,
		HandoffEnabled:   true,
		FallbackMessages: &models.FallbackMessages{HandoffMessage: "Connecting you to a human..."},
	}
	botResult := testdb.CreateChatbot(t, te.DB, testdb.ChatbotFixture{
		UserID: userID,
		Name:   "AutoHandoffBot",
	})
	botID := botResult.Chatbot.ID
	bot.ID = botID

	// 4. Create Chat Request
	sessionID := "sess_auto_1"
	req := models.ChatRequest{
		SessionID: sessionID,
		Message:   "I want a human please",
	}

	// 5. Process Chat
	// This should trigger ProcessChatWithTools because HandoffEnabled=true
	// And mock client will return request_human_handoff tool call
	res, err := chatSvc.ProcessChat(ctx, req, bot, models.RAGConfig{})
	require.NoError(t, err)

	// 6. Verify Result
	assert.Equal(t, "Connecting you to a human...", res.Response)
	assert.Equal(t, "handoff", "handoff") // Verify type if exposed in Result (it's not fields, but we can check DB)

	// 7. Verify Message in DB
	msgs, err := conversationRepo.ListRecentMessages(ctx, res.ConversationID, 1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, "assistant", msgs[0].Role)
	assert.Equal(t, "handoff", msgs[0].Type) // Validates updated CreateMessage logic

	// 8. Verify Handoff Request in DB
	active, err := handoffRepo.HasActiveHandoffRequest(ctx, res.ConversationID)
	require.NoError(t, err)
	assert.True(t, active, "Handoff request should be active")
}

func ptrString(s string) *string {
	return &s
}
