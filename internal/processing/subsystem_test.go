package processing

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
)

func TestProcessingSubsystemInterface(t *testing.T) {
	var _ ProcessingSubsystem = (*SourceQueue)(nil)
}

func TestSourceQueue_Status(t *testing.T) {
	log := logger.New("INFO")
	queue := NewQueueManager(2, log, nil)
	queue.Start()
	defer queue.Stop()

	sq := &SourceQueue{
		queue: queue,
		log:   log,
	}

	status := sq.Status()
	if status.WorkerCount != 2 {
		t.Errorf("expected 2 workers, got %d", status.WorkerCount)
	}
	if !status.IsRunning {
		t.Error("expected IsRunning to be true")
	}
	if status.QueueLength != 0 {
		t.Errorf("expected empty queue, got %d", status.QueueLength)
	}
}

func TestSourceQueue_StatusNilQueue(t *testing.T) {
	sq := &SourceQueue{
		queue: nil,
		log:   nil,
	}

	status := sq.Status()
	if status.WorkerCount != 0 {
		t.Errorf("expected 0 workers, got %d", status.WorkerCount)
	}
	if status.IsRunning {
		t.Error("expected IsRunning to be false")
	}
	if status.QueueLength != 0 {
		t.Errorf("expected 0 queue length, got %d", status.QueueLength)
	}
}

func TestSourceQueue_StatusNilLog(t *testing.T) {
	queue := NewQueueManager(2, nil, nil)
	queue.Start()
	defer queue.Stop()

	sq := &SourceQueue{
		queue: queue,
		log:   nil,
	}

	status := sq.Status()
	if status.WorkerCount != 2 {
		t.Errorf("expected 2 workers, got %d", status.WorkerCount)
	}
}

func TestSourceQueue_EnqueueJob(t *testing.T) {
	log := logger.New("INFO")
	queue := NewQueueManager(1, log, nil)
	queue.Start()
	defer queue.Stop()

	sq := &SourceQueue{
		queue: queue,
		log:   log,
	}

	sq.EnqueueJob("job-123")

	if queue.QueueLength() != 1 {
		t.Errorf("expected 1 job in queue, got %d", queue.QueueLength())
	}
}

func TestSourceQueue_EnqueueJobNilQueue(t *testing.T) {
	sq := &SourceQueue{
		queue: nil,
		log:   nil,
	}

	sq.EnqueueJob("job-123")
}

func TestSourceQueue_StatusWithJobs(t *testing.T) {
	log := logger.New("INFO")
	queue := NewQueueManager(2, log, nil)
	queue.Start()
	defer queue.Stop()

	queue.Enqueue("job-1")
	queue.Enqueue("job-2")

	if queue.QueueLength() != 2 {
		t.Errorf("expected 2 jobs in queue immediately after enqueue, got %d", queue.QueueLength())
	}
}

func TestQueueStatusStruct(t *testing.T) {
	status := QueueStatus{
		WorkerCount: 4,
		QueueLength: 10,
		IsRunning:   true,
	}

	if status.WorkerCount != 4 {
		t.Errorf("expected WorkerCount 4, got %d", status.WorkerCount)
	}
	if status.QueueLength != 10 {
		t.Errorf("expected QueueLength 10, got %d", status.QueueLength)
	}
	if !status.IsRunning {
		t.Error("expected IsRunning to be true")
	}
}

func TestSourceQueue_StopNilQueue(t *testing.T) {
	sq := &SourceQueue{
		queue: nil,
		log:   nil,
	}

	sq.Stop()
}

func TestSourceQueue_EnqueueSourceNilQueue(t *testing.T) {
	sq := &SourceQueue{
		queue: nil,
		log:   nil,
	}

	_, err := sq.EnqueueSource(context.Background(), "source-1", "chatbot-1")
	if err == nil {
		t.Error("expected error when queue is nil")
	}
}

func TestSourceQueue_WorkerCountNilQueue(t *testing.T) {
	sq := &SourceQueue{
		queue: nil,
		log:   nil,
	}

	if sq.WorkerCount() != 0 {
		t.Errorf("expected 0, got %d", sq.WorkerCount())
	}
}

func TestSourceQueue_QueueLengthNilQueue(t *testing.T) {
	sq := &SourceQueue{
		queue: nil,
		log:   nil,
	}

	if sq.QueueLength() != 0 {
		t.Errorf("expected 0, got %d", sq.QueueLength())
	}
}

func TestNewRAGSubsystem(t *testing.T) {
	var _ rag.EmbeddingClient = &rag.MockEmbeddingClient{}
	var _ rag.VectorClient = &rag.MockVectorClient{}
	var _ rag.LLMClient = &rag.MockLLMClient{}

	embedder := &rag.MockEmbeddingClient{}
	vector := &rag.MockVectorClient{}
	llm := &rag.MockLLMClient{}

	subsystem := rag.NewRAGSubsystem(embedder, vector, llm)
	if subsystem == nil {
		t.Fatal("expected non-nil subsystem")
	}

	if !subsystem.Ready() {
		t.Error("expected Ready() to return true")
	}
}
