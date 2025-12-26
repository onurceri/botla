package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockVectorStore_InterfaceCompliance(t *testing.T) {
	var _ VectorStore = (*MockVectorStore)(nil)
}

func TestMockVectorStore_DefaultBehavior(t *testing.T) {
	mock := &MockVectorStore{}
	ctx := context.Background()

	// Test EnsureCollection
	err := mock.EnsureCollection(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Upsert
	payload := VectorPayload{
		ChatbotID:    "test-bot",
		SourceID:     "test-source",
		ChunkIndex:   0,
		OriginalText: "test",
		SourceType:   "text",
		CreatedAt:    time.Now(),
	}
	err = mock.Upsert(ctx, "id-1", []float32{0.1}, payload)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Search
	results, err := mock.Search(ctx, []float32{0.1}, SearchFilter{ChatbotID: "test"}, 10)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if results == nil {
		t.Error("expected results to be non-nil")
	}

	// Test Delete
	err = mock.Delete(ctx, DeleteFilter{SourceID: "test"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Scroll
	scrollResults, offset, err := mock.Scroll(ctx, SearchFilter{ChatbotID: "test"}, 10, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if scrollResults == nil {
		t.Error("expected results to be non-nil")
	}
	if offset != nil {
		t.Errorf("expected offset to be nil, got %v", offset)
	}
}

func TestMockVectorStore_CustomBehavior(t *testing.T) {
	expectedErr := errors.New("custom error")
	mock := &MockVectorStore{
		SearchFunc: func(ctx context.Context, vector []float32, filter SearchFilter, limit int) ([]SearchResult, error) {
			return []SearchResult{
				{
					ID:    "result-1",
					Score: 0.95,
					Payload: VectorPayload{
						ChatbotID:    filter.ChatbotID,
						SourceID:     "src-1",
						ChunkIndex:   0,
						OriginalText: "custom result",
						SourceType:   "text",
						CreatedAt:    time.Now(),
					},
				},
			}, nil
		},
		DeleteFunc: func(ctx context.Context, filter DeleteFilter) error {
			return expectedErr
		},
	}

	ctx := context.Background()

	// Test custom Search
	results, err := mock.Search(ctx, []float32{0.1}, SearchFilter{ChatbotID: "bot-123"}, 10)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Payload.ChatbotID != "bot-123" {
		t.Errorf("expected chatbotID 'bot-123', got %s", results[0].Payload.ChatbotID)
	}

	// Test custom Delete
	err = mock.Delete(ctx, DeleteFilter{SourceID: "test"})
	if err != expectedErr {
		t.Errorf("expected custom error, got %v", err)
	}
}

func TestMockEmbedder_InterfaceCompliance(t *testing.T) {
	var _ Embedder = (*MockEmbedder)(nil)
}

func TestMockEmbedder_DefaultBehavior(t *testing.T) {
	mock := &MockEmbedder{}
	ctx := context.Background()

	// Test Embed
	embedding, err := mock.Embed(ctx, "test text")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(embedding) != 3 {
		t.Errorf("expected 3 dimensions, got %d", len(embedding))
	}

	// Test EmbedBatch
	embeddings, err := mock.EmbedBatch(ctx, []string{"text1", "text2"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(embeddings) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(embeddings))
	}

	// Test Dimension
	dim := mock.Dimension()
	if dim != 1536 {
		t.Errorf("expected dimension 1536, got %d", dim)
	}
}

func TestMockEmbedder_CustomBehavior(t *testing.T) {
	expectedErr := errors.New("custom error")
	mock := &MockEmbedder{
		EmbedFunc: func(ctx context.Context, text string) ([]float32, error) {
			if text == "error" {
				return nil, expectedErr
			}
			return []float32{0.5, 0.6, 0.7}, nil
		},
		DimensionFunc: func() int {
			return 3072
		},
	}

	ctx := context.Background()

	// Test custom Embed
	embedding, err := mock.Embed(ctx, "test")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(embedding) != 3 || embedding[0] != 0.5 {
		t.Errorf("unexpected embedding: %v", embedding)
	}

	// Test custom error
	_, err = mock.Embed(ctx, "error")
	if err != expectedErr {
		t.Errorf("expected custom error, got %v", err)
	}

	// Test custom Dimension
	dim := mock.Dimension()
	if dim != 3072 {
		t.Errorf("expected dimension 3072, got %d", dim)
	}
}
