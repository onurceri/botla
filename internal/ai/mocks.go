package ai

import "context"

// MockVectorStore is a mock implementation of VectorStore for testing
type MockVectorStore struct {
	EnsureCollectionFunc func(ctx context.Context) error
	UpsertFunc           func(ctx context.Context, id interface{}, vector []float32, payload VectorPayload) error
	SearchFunc           func(ctx context.Context, vector []float32, filter SearchFilter, limit int) ([]SearchResult, error)
	DeleteFunc           func(ctx context.Context, filter DeleteFilter) error
	ScrollFunc           func(ctx context.Context, filter SearchFilter, limit int, offset interface{}) ([]SearchResult, interface{}, error)
}

// Verify interface compliance
var _ VectorStore = (*MockVectorStore)(nil)

func (m *MockVectorStore) EnsureCollection(ctx context.Context) error {
	if m.EnsureCollectionFunc != nil {
		return m.EnsureCollectionFunc(ctx)
	}
	return nil
}

func (m *MockVectorStore) Upsert(ctx context.Context, id interface{}, vector []float32, payload VectorPayload) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, id, vector, payload)
	}
	return nil
}

func (m *MockVectorStore) Search(ctx context.Context, vector []float32, filter SearchFilter, limit int) ([]SearchResult, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, vector, filter, limit)
	}
	return []SearchResult{}, nil
}

func (m *MockVectorStore) Delete(ctx context.Context, filter DeleteFilter) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, filter)
	}
	return nil
}

func (m *MockVectorStore) Scroll(ctx context.Context, filter SearchFilter, limit int, offset interface{}) ([]SearchResult, interface{}, error) {
	if m.ScrollFunc != nil {
		return m.ScrollFunc(ctx, filter, limit, offset)
	}
	return []SearchResult{}, nil, nil
}

// MockEmbedder is a mock implementation of Embedder for testing
type MockEmbedder struct {
	EmbedFunc      func(ctx context.Context, text string) ([]float32, error)
	EmbedBatchFunc func(ctx context.Context, texts []string) ([][]float32, error)
	DimensionFunc  func() int
}

// Verify interface compliance
var _ Embedder = (*MockEmbedder)(nil)

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.EmbedFunc != nil {
		return m.EmbedFunc(ctx, text)
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if m.EmbedBatchFunc != nil {
		return m.EmbedBatchFunc(ctx, texts)
	}
	results := make([][]float32, len(texts))
	for i := range results {
		results[i] = []float32{0.1, 0.2, 0.3}
	}
	return results, nil
}

func (m *MockEmbedder) Dimension() int {
	if m.DimensionFunc != nil {
		return m.DimensionFunc()
	}
	return 1536
}
