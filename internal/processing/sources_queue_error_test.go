package processing

import (
	"context"
	"errors"
	"testing"

	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/storage"
	"github.com/stretchr/testify/mock"
)

type MockVectorClient struct {
	mock.Mock
}

func (m *MockVectorClient) EnsureEmbeddingsCollection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVectorClient) SearchSimilar(ctx context.Context, vector []float32, chatbotID string, limit int) ([]rag.SearchResult, error) {
	args := m.Called(ctx, vector, chatbotID, limit)
	return args.Get(0).([]rag.SearchResult), args.Error(1)
}

func (m *MockVectorClient) UpsertEmbedding(ctx context.Context, id interface{}, vector []float32, payload rag.EmbeddingPayload) error {
	args := m.Called(ctx, id, vector, payload)
	return args.Error(0)
}

func (m *MockVectorClient) DeleteSourceEmbeddings(ctx context.Context, sourceID string) error {
	args := m.Called(ctx, sourceID)
	return args.Error(0)
}

func (m *MockVectorClient) DeleteBySourceID(ctx context.Context, sourceID string) error {
	args := m.Called(ctx, sourceID)
	return args.Error(0)
}

func (m *MockVectorClient) ScrollChunks(ctx context.Context, sourceID string, limit int, offset interface{}) ([]rag.SearchResult, *string, error) {
	args := m.Called(ctx, sourceID, limit, offset)
	return args.Get(0).([]rag.SearchResult), args.Get(1).(*string), args.Error(2)
}

func TestStartSourceQueue_NoFailOnVectorStoreError(t *testing.T) {
	// Create mock that fails on EnsureEmbeddingsCollection
	mockVC := &MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(errors.New("connection failed"))

	// We don't need real DB/ Storage/LLM for this test as it fails before using them
	q, err := StartSourceQueue(nil, storage.NewMemoryStorage(), nil, mockVC, nil, nil, 1)

	if err != nil {
		t.Fatalf("expected no error (non-blocking), got %v", err)
	}
	if q == nil {
		t.Fatal("expected non-nil queue")
	}
	defer q.Stop()
}
