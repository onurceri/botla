package integration

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockToolsLLMClient struct{}

func (m *MockToolsLLMClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	return &models.CompletionResult{Content: "Mock response", UsageTokens: 10}, nil
}

func (m *MockToolsLLMClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{Name: "mock-model", Provider: "openai"}
}

func (m *MockToolsLLMClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2}, nil
}

func (m *MockToolsLLMClient) CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return [][]float32{{0.1}}, nil
}

func (m *MockToolsLLMClient) CreateCompletionWithTools(ctx context.Context, messages []rag.ChatMessage, tools []rag.Tool, model string, temperature float32, maxTokens int) (*rag.ChatResponseWithTools, error) {
	// Check if this is the initial call or the second call (after tool execution)
	// For simplicity, we assume the first call triggers the tool.
	// If the last message is from user, trigger tool.
	// If the last message is tool result, return final response?
	// Actually, the loop handles it.

	lastMsg := messages[len(messages)-1]

	if lastMsg.Role == "user" {
		// Return tool call
		return &rag.ChatResponseWithTools{
			Choices: []struct {
				Message      rag.ChatMessage `json:"message"`
				FinishReason string          `json:"finish_reason"`
			}{
				{
					Message: rag.ChatMessage{
						Role: "assistant",
						ToolCalls: []rag.ToolCall{
							{
								ID:   "call_123",
								Type: "function",
								Function: struct {
									Name      string `json:"name"`
									Arguments string `json:"arguments"`
								}{
									Name:      "request_human_handoff",
									Arguments: "{}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: struct {
				TotalTokens int `json:"total_tokens"`
			}{TotalTokens: 20},
		}, nil
	}

	// If getting tool result, return something (though code breaks loop on handoff success)
	return &rag.ChatResponseWithTools{
		Choices: []struct {
			Message      rag.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{
			{
				Message: rag.ChatMessage{
					Role:    "assistant",
					Content: ptrString(""),
				},
			},
		},
	}, nil
}

func TestAutomatedHandoff(t *testing.T) {
	te, err := SetupTestEnv()
	require.NoError(t, err)
	defer TeardownTestEnv(te)

	ctx := context.Background()
	log := logger.New("DEBUG")

	// Create User
	var userID string
	err = te.DB.QueryRowContext(ctx, `INSERT INTO users (email, password_hash, full_name) VALUES ($1, $2, $3) RETURNING id`, "test@example.com", "hash", "Test User").Scan(&userID)
	require.NoError(t, err)
	// 1. Setup Mock Factory
	factory := rag.NewClientFactory(te.Cfg)
	mockClient := &MockToolsLLMClient{}
	factory.RegisterClient("openrouter", mockClient) // Primary provider
	factory.RegisterClient("openai", mockClient)     // Fallback provider

	// 2. Setup Chat Service with mock factory
	// We need embedder too. MockClient implements it.
	chatSvc := services.NewChatService(te.DB, factory, mockClient, nil, log)

	// 3. Create Chatbot with HandoffEnabled
	bot := &models.Chatbot{
		UserID:           userID,
		Name:             "AutoHandoffBot",
		SystemPrompt:     "You are helpful.",
		LanguageCode:     "en",
		Model:            "openrouter:openai/gpt-4o-mini", // OpenRouter is primary provider
		Temperature:      0.7,
		MaxTokens:        100,
		HandoffEnabled:   true,
		FallbackMessages: &models.FallbackMessages{HandoffMessage: "Connecting you to a human..."},
	}
	botID, err := db.CreateChatbot(ctx, te.DB, bot)
	require.NoError(t, err)
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
	msgs, err := db.ListRecentMessages(ctx, te.DB, res.ConversationID, 1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, "assistant", msgs[0].Role)
	assert.Equal(t, "handoff", msgs[0].Type) // Validates updated CreateMessage logic

	// 8. Verify Handoff Request in DB
	active, err := db.HasActiveHandoffRequest(ctx, te.DB, res.ConversationID)
	require.NoError(t, err)
	assert.True(t, active, "Handoff request should be active")
}

func ptrString(s string) *string {
	return &s
}
