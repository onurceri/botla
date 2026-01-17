package processing

import (
	"testing"

	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/stretchr/testify/mock"
)

// Reusing MockVectorClient from sources_queue_error_test.go if available.
// If not, we might need to define it here or in a common test file.
// Since Go tests in the same package share Scope, it should be available.

func TestWorkerPool_MultipleWorkers(t *testing.T) {
	mockVC := &MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)

	// Start with 5 workers
	q, err := StartSourceQueue(nil, nil, nil, nil, nil, nil, storage.NewMemoryStorage(), nil, mockVC, nil, nil, 5)
	if err != nil {
		t.Fatalf("StartSourceQueue failed: %v", err)
	}
	defer q.Stop()

	if q.WorkerCount() != 5 {
		t.Errorf("expected worker count 5, got %d", q.WorkerCount())
	}
}

func TestWorkerPool_MaxWorkerLimit(t *testing.T) {
	mockVC := &MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)

	// Request 20 workers, should be capped at 16
	q, err := StartSourceQueue(nil, nil, nil, nil, nil, nil, storage.NewMemoryStorage(), nil, mockVC, nil, nil, 20)
	if err != nil {
		t.Fatalf("StartSourceQueue failed: %v", err)
	}
	defer q.Stop()

	if q.WorkerCount() != 16 {
		t.Errorf("expected worker count capped at 16, got %d", q.WorkerCount())
	}
}
