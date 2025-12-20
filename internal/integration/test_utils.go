package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
)

type mockToolsClient struct{}

func (m *mockToolsClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	return &models.CompletionResult{Content: "mock response"}, nil
}

func (m *mockToolsClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{Name: "mock-model"}
}

func (m *mockToolsClient) CreateCompletionWithTools(ctx context.Context, messages []rag.ChatMessage, tools []rag.Tool, model string, temperature float32, maxTokens int) (*rag.ChatResponseWithTools, error) {
	// Generate unique name to avoid DB unique constraint violations
	content := fmt.Sprintf("mock_tool_%d", time.Now().UnixNano())
	return &rag.ChatResponseWithTools{
		Choices: []struct {
			Message      rag.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{{
			Message: rag.ChatMessage{Content: &content},
		}},
	}, nil
}
