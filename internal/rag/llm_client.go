package rag

import (
	"context"

	"github.com/onurceri/botla-co/internal/models"
)

// LLMClient defines the interface for interacting with LLM providers for chat completions
type LLMClient interface {
	CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error)
	GetModelInfo() models.ModelInfo
}

// EmbeddingClient defines the interface for creating embeddings
type EmbeddingClient interface {
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)
	CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// FullLLMClient is a helper interface that combines both LLM and Embedding capabilities
// Most providers (like OpenAI) will implement both
type FullLLMClient interface {
	LLMClient
	EmbeddingClient
}

// ToolsLLMClient defines the interface for LLMs that support tool calling
type ToolsLLMClient interface {
	LLMClient
	CreateCompletionWithTools(ctx context.Context, messages []ChatMessage, tools []Tool, model string, temperature float32, maxTokens int) (*ChatResponseWithTools, error)
}
