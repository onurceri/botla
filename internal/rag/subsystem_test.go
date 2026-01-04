package rag

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/stretchr/testify/mock"
)

func TestRAGSubsystemInterface(t *testing.T) {
	var _ RAGSubsystem = (*ragSubsystem)(nil)
}

func TestNewRAGSubsystem(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	vector := &MockVectorClient{}
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)
	if subsystem == nil {
		t.Fatal("expected non-nil subsystem")
	}
}

func TestRAGSubsystem_Embed(t *testing.T) {
	embedder := new(MockEmbeddingClient)
	embedder.On("CreateEmbedding", mock.Anything, "test text").Return([]float32{0.9, 0.8, 0.7}, nil)
	vector := &MockVectorClient{}
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	result, err := subsystem.Embed(context.Background(), "test text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 dimensions, got %d", len(result))
	}
	if result[0] != 0.9 {
		t.Errorf("expected first dimension to be 0.9, got %f", result[0])
	}
	embedder.AssertExpectations(t)
}

func TestRAGSubsystem_EmbedBatch(t *testing.T) {
	texts := []string{"text1", "text2"}
	embedder := new(MockEmbeddingClient)
	embedder.On("CreateEmbeddingsBatch", mock.Anything, texts).Return([][]float32{
		{0.1, 0.2},
		{0.3, 0.4},
	}, nil)
	vector := &MockVectorClient{}
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	result, err := subsystem.EmbedBatch(context.Background(), texts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(result))
	}
	embedder.AssertExpectations(t)
}

func TestRAGSubsystem_Store(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	vector := new(MockVectorClient)
	vector.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	chunk := ChunkWithEmbedding{
		Chunk:     "test chunk",
		Embedding: []float32{0.1, 0.2, 0.3},
	}
	err := subsystem.Store(context.Background(), chunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vector.AssertExpectations(t)
}

func TestRAGSubsystem_Search(t *testing.T) {
	embedder := new(MockEmbeddingClient)
	embedder.On("CreateEmbedding", mock.Anything, "test query").Return([]float32{0.1, 0.2, 0.3}, nil)
	vector := new(MockVectorClient)
	expectedResults := []SearchResult{
		{Score: 0.95, Payload: EmbeddingPayload{OriginalText: "found content"}},
	}
	vector.On("SearchSimilar", mock.Anything, []float32{0.1, 0.2, 0.3}, "bot-123", 5).Return(expectedResults, nil)
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	req := SearchRequest{Query: "test query", ChatbotID: "bot-123", Limit: 5}
	results, err := subsystem.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score != 0.95 {
		t.Errorf("expected score 0.95, got %f", results[0].Score)
	}
	embedder.AssertExpectations(t)
	vector.AssertExpectations(t)
}

func TestRAGSubsystem_DeleteBySource(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	vector := new(MockVectorClient)
	vector.On("DeleteBySourceID", mock.Anything, "source-123").Return(nil)
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	err := subsystem.DeleteBySource(context.Background(), "source-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vector.AssertExpectations(t)
}

func TestRAGSubsystem_Complete(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	vector := &MockVectorClient{}
	llm := new(MockLLMClient)
	expectedResult := &models.CompletionResult{Content: "AI response", UsageTokens: 50}
	llm.On("CreateCompletion", mock.Anything, mock.Anything).Return(expectedResult, nil)

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	params := models.CompletionParams{
		SystemPrompt: "You are a helpful assistant.",
		UserMessage:  "Hello!",
		Model:        "gpt-4o-mini",
	}
	result, err := subsystem.Complete(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "AI response" {
		t.Errorf("expected 'AI response', got '%s'", result.Content)
	}
	llm.AssertExpectations(t)
}

func TestRAGSubsystem_Ready(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	vector := &MockVectorClient{}
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	if !subsystem.Ready() {
		t.Error("expected Ready() to return true")
	}
}

func TestRAGSubsystem_NilEmbedder(t *testing.T) {
	vector := &MockVectorClient{}
	llm := &MockLLMClient{}

	subsystem := ragSubsystem{
		embedder: nil,
		vector:   vector,
		llm:      llm,
	}

	_, err := subsystem.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected error when embedder is nil")
	}
}

func TestRAGSubsystem_NilVector(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	llm := &MockLLMClient{}

	subsystem := ragSubsystem{
		embedder: embedder,
		vector:   nil,
		llm:      llm,
	}

	_, err := subsystem.Search(context.Background(), SearchRequest{})
	if err == nil {
		t.Error("expected error when vector is nil")
	}
}

func TestRAGSubsystem_NilLLM(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	vector := &MockVectorClient{}

	subsystem := ragSubsystem{
		embedder: embedder,
		vector:   vector,
		llm:      nil,
	}

	_, err := subsystem.Complete(context.Background(), models.CompletionParams{})
	if err == nil {
		t.Error("expected error when llm is nil")
	}
}

func TestRAGSubsystem_EmbedError(t *testing.T) {
	embedder := new(MockEmbeddingClient)
	embedder.On("CreateEmbedding", mock.Anything, "failing text").Return(nil, context.DeadlineExceeded)
	vector := &MockVectorClient{}
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	_, err := subsystem.Embed(context.Background(), "failing text")
	if err == nil {
		t.Error("expected error from embedder")
	}
	embedder.AssertExpectations(t)
}

func TestRAGSubsystem_SearchError(t *testing.T) {
	embedder := new(MockEmbeddingClient)
	embedder.On("CreateEmbedding", mock.Anything, "test query").Return([]float32{0.1, 0.2, 0.3}, nil)
	vector := new(MockVectorClient)
	vector.On("SearchSimilar", mock.Anything, []float32{0.1, 0.2, 0.3}, "bot-123", 5).Return(nil, context.DeadlineExceeded)
	llm := &MockLLMClient{}

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	req := SearchRequest{Query: "test query", ChatbotID: "bot-123", Limit: 5}
	_, err := subsystem.Search(context.Background(), req)
	if err == nil {
		t.Error("expected error from vector search")
	}
	embedder.AssertExpectations(t)
	vector.AssertExpectations(t)
}

func TestRAGSubsystem_CompleteError(t *testing.T) {
	embedder := &MockEmbeddingClient{}
	vector := &MockVectorClient{}
	llm := new(MockLLMClient)
	llm.On("CreateCompletion", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded)

	subsystem := NewRAGSubsystem(embedder, vector, llm)

	_, err := subsystem.Complete(context.Background(), models.CompletionParams{})
	if err == nil {
		t.Error("expected error from LLM")
	}
	llm.AssertExpectations(t)
}
