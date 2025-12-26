package rag //nolint:wrapcheck

import (
	"context"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockLLMClient is a mock implementation of LLMClient
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CompletionResult), args.Error(1)
}

func (m *MockLLMClient) GetModelInfo() models.ModelInfo {
	args := m.Called()
	return args.Get(0).(models.ModelInfo)
}

// MockEmbeddingClient is a mock implementation of EmbeddingClient
type MockEmbeddingClient struct {
	mock.Mock
}

func (m *MockEmbeddingClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
}

func (m *MockEmbeddingClient) CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error) {
	args := m.Called(ctx, texts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]float32), args.Error(1)
}

// MockVectorClient is a mock implementation of VectorClient
type MockVectorClient struct {
	mock.Mock
}

func (m *MockVectorClient) EnsureEmbeddingsCollection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVectorClient) UpsertEmbedding(ctx context.Context, id interface{}, vector []float32, payload EmbeddingPayload) error {
	args := m.Called(ctx, id, vector, payload)
	return args.Error(0)
}

func (m *MockVectorClient) SearchSimilar(ctx context.Context, embedding []float32, chatbotID string, topK int) ([]SearchResult, error) {
	args := m.Called(ctx, embedding, chatbotID, topK)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]SearchResult), args.Error(1)
}

func (m *MockVectorClient) DeleteBySourceID(ctx context.Context, sourceID string) error {
	args := m.Called(ctx, sourceID)
	return args.Error(0)
}

func (m *MockVectorClient) ScrollChunks(ctx context.Context, sourceID string, limit int, offset interface{}) ([]SearchResult, *string, error) {
	args := m.Called(ctx, sourceID, limit, offset)
	var nextOffset *string
	if args.Get(1) != nil {
		nextOffset = args.Get(1).(*string)
	}
	if args.Get(0) == nil {
		return nil, nextOffset, args.Error(2)
	}
	return args.Get(0).([]SearchResult), nextOffset, args.Error(2)
}

// MockFullClient implements both LLMClient and EmbeddingClient
type MockFullClient struct {
	MockLLMClient
}

func (m *MockFullClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
}

func (m *MockFullClient) CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error) {
	args := m.Called(ctx, texts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]float32), args.Error(1)
}

func (m *MockFullClient) CreateCompletionWithTools(ctx context.Context, messages []ChatMessage, tools []Tool, model string, temperature float32, maxTokens int) (*ChatResponseWithTools, error) {
	args := m.Called(ctx, messages, tools, model, temperature, maxTokens)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ChatResponseWithTools), args.Error(1)
}

// MockToolsLLMClient combines MockLLMClient with tool calling
type MockToolsLLMClient struct {
	MockLLMClient
}

func (m *MockToolsLLMClient) CreateCompletionWithTools(ctx context.Context, messages []ChatMessage, tools []Tool, model string, temperature float32, maxTokens int) (*ChatResponseWithTools, error) {
	args := m.Called(ctx, messages, tools, model, temperature, maxTokens)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ChatResponseWithTools), args.Error(1)
}
