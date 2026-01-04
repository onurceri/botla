package processing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/scraper"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/stretchr/testify/mock"
)

// TestEnqueueSource_CreatesJob tests that EnqueueSource creates a job record
func TestEnqueueSource_CreatesJob(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	ctx := context.Background()

	log := logger.New("DEBUG")
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{Content: `{"capability_summary":"test"}`}, nil).Maybe()
	mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
	mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockVC.On("DeleteBySourceID", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil).Maybe()

	sourceRepo := repository.NewMockSourceRepo()
	trainingJobRepo := &MockTrainingJobRepo{}
	chatbotRepo := repository.NewMockChatbotRepo()
	planRepo := repository.NewMockPlanRepo()
	usageRepo := &MockUsageRepo{}

	processor := NewJobProcessor(JobProcessorConfig{
		TrainingJobRepo: trainingJobRepo,
		SourceRepo:      sourceRepo,
		ChatbotRepo:     chatbotRepo,
		PlanRepo:        planRepo,
		UsageRepo:       usageRepo,
		OpenAIClient:    mockOAI,
		VectorClient:    mockVC,
		Log:             log,
		Processors: map[string]SourceProcessor{
			"url": NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, log, &testMockScraper{}, nil),
		},
	})

	queueManager := NewQueueManager(1, log, processor)

	queue := &SourceQueue{
		queue:           queueManager,
		processor:       processor,
		trainingJobRepo: trainingJobRepo,
		log:             log,
	}

	// Create test chatbot first (FK dependency)
	chatbotID := "test-chatbot-id"
	sourceID := "test-source-id"

	// Enqueue source
	jobID, err := queue.EnqueueSource(ctx, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("EnqueueSource failed: %v", err)
	}

	if jobID == "" {
		t.Error("expected non-empty job ID")
	}
}

// TestQueueManager_EnqueueAndWorker tests queue manager enqueue and worker processing
func TestQueueManager_EnqueueAndWorker(t *testing.T) {
	log := logger.New("DEBUG")
	processed := make(chan string, 10)

	handler := &mockJobHandler{processed: processed}
	qm := NewQueueManager(1, log, handler)
	qm.Start()
	defer qm.Stop()

	// Enqueue a job
	if !qm.Enqueue("test-job-1") {
		t.Error("expected enqueue to succeed")
	}

	// Wait for processing
	select {
	case jobID := <-processed:
		if jobID != "test-job-1" {
			t.Errorf("expected test-job-1, got %s", jobID)
		}
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for job processing")
	}
}

// TestQueueManager_QueueFull tests behavior when queue is full
func TestQueueManager_QueueFull(t *testing.T) {
	log := logger.New("DEBUG")
	handler := &mockJobHandler{processed: make(chan string, 100)}

	// Create queue manager but don't start workers
	qm := &QueueManager{
		ch:      make(chan string, 2), // Small buffer
		stopCh:  make(chan struct{}),
		workers: 0,
		log:     log,
		handler: handler,
	}

	// Fill the queue
	qm.Enqueue("job-1")
	qm.Enqueue("job-2")

	// This should fail (queue full)
	if qm.Enqueue("job-3") {
		t.Error("expected enqueue to fail when queue is full")
	}
}

// mockJobHandler is a test implementation of JobHandler
type mockJobHandler struct {
	processed chan string
}

func (h *mockJobHandler) HandleJob(jobID string) {
	h.processed <- jobID
}

// testMockScraper implements scraper.Scraper for testing
type testMockScraper struct {
	content    string
	html       string
	shouldFail bool
}

func (m *testMockScraper) FetchRawHTML(url string) (string, error) {
	if m.shouldFail {
		return "", &testScraperError{"fetch failed"}
	}
	return m.html, nil
}

func (m *testMockScraper) ScrapeURLWithFallback(task scraper.ScrapingTask, dynamicEnabled bool, scrapeConfig *scraper.ScrapeConfig) (string, error) {
	if m.shouldFail {
		return "", &testScraperError{"scrape failed"}
	}
	return m.content, nil
}

func (m *testMockScraper) ExtractLinks(htmlContent, baseURL string, filter *scraper.PathFilter) ([]string, error) {
	return nil, nil
}

type testScraperError struct {
	msg string
}

func (e *testScraperError) Error() string {
	return e.msg
}

// =============================================================================
// Additional Edge Case Tests
// =============================================================================

// TestQueueManager_Stop tests graceful shutdown
func TestQueueManager_Stop(t *testing.T) {
	log := logger.New("DEBUG")
	processed := make(chan string, 10)

	handler := &mockJobHandler{processed: processed}
	qm := NewQueueManager(2, log, handler)
	qm.Start()

	// Enqueue a job
	qm.Enqueue("test-job-1")

	// Wait for processing
	select {
	case <-processed:
		// Job processed
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for job")
	}

	// Stop should complete without hanging
	done := make(chan struct{})
	go func() {
		qm.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Stop completed
	case <-time.After(5 * time.Second):
		t.Error("Stop() timed out - possible deadlock")
	}
}

// TestQueueManager_MultipleWorkers tests concurrent job processing
func TestQueueManager_MultipleWorkers(t *testing.T) {
	log := logger.New("DEBUG")
	processed := make(chan string, 100)

	handler := &mockJobHandler{processed: processed}
	qm := NewQueueManager(4, log, handler)
	qm.Start()
	defer qm.Stop()

	// Enqueue multiple jobs
	for i := 0; i < 10; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		if !qm.Enqueue(jobID) {
			t.Errorf("failed to enqueue %s", jobID)
		}
	}

	// Wait for all jobs to be processed
	processedJobs := make(map[string]bool)
	timeout := time.After(10 * time.Second)

	for len(processedJobs) < 10 {
		select {
		case jobID := <-processed:
			processedJobs[jobID] = true
		case <-timeout:
			t.Errorf("timeout: only processed %d/10 jobs", len(processedJobs))
			return
		}
	}

	if len(processedJobs) != 10 {
		t.Errorf("expected 10 processed jobs, got %d", len(processedJobs))
	}
}

// TestQueueManager_WorkerCount tests worker count getter
func TestQueueManager_WorkerCount(t *testing.T) {
	log := logger.New("DEBUG")
	handler := &mockJobHandler{processed: make(chan string, 10)}

	qm := NewQueueManager(5, log, handler)
	if qm.WorkerCount() != 5 {
		t.Errorf("expected 5 workers, got %d", qm.WorkerCount())
	}
}

// TestQueueManager_QueueLength tests queue length getter
func TestQueueManager_QueueLength(t *testing.T) {
	log := logger.New("DEBUG")
	handler := &mockJobHandler{processed: make(chan string, 100)}

	// Create but don't start (so jobs stay in queue)
	qm := &QueueManager{
		ch:      make(chan string, 10),
		stopCh:  make(chan struct{}),
		workers: 0,
		log:     log,
		handler: handler,
	}

	qm.Enqueue("job-1")
	qm.Enqueue("job-2")
	qm.Enqueue("job-3")

	if qm.QueueLength() != 3 {
		t.Errorf("expected queue length 3, got %d", qm.QueueLength())
	}
}

// TestQueueManager_EnqueueWithDelay tests delayed enqueue
func TestQueueManager_EnqueueWithDelay(t *testing.T) {
	log := logger.New("DEBUG")
	processed := make(chan string, 10)

	handler := &mockJobHandler{processed: processed}
	qm := NewQueueManager(1, log, handler)
	qm.Start()
	defer qm.Stop()

	start := time.Now()
	qm.EnqueueWithDelay("delayed-job", 100*time.Millisecond)

	// Wait for processing
	select {
	case jobID := <-processed:
		elapsed := time.Since(start)
		if elapsed < 100*time.Millisecond {
			t.Errorf("job processed too early: %v", elapsed)
		}
		if jobID != "delayed-job" {
			t.Errorf("expected delayed-job, got %s", jobID)
		}
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for delayed job")
	}
}

// TestJobProcessor_CalculateBackoff tests exponential backoff calculation
func TestJobProcessor_CalculateBackoff(t *testing.T) {
	jp := &JobProcessor{}

	tests := []struct {
		retryCount int
		expected   time.Duration
	}{
		{0, 2 * time.Second},                 // base = 2s, no doublings
		{1, 2 * time.Second},                 // loop doesn't run for i=0
		{2, 4 * time.Second},                 // 2s * 2^1 = 4s
		{3, 8 * time.Second},                 // 2s * 2^2 = 8s
		{4, 16 * time.Second},                // 2s * 2^3 = 16s
		{5, 32 * time.Second},                // 2s * 2^4 = 32s
		{10, 17*time.Minute + 4*time.Second}, // 2s * 2^9 = 1024s = 17m4s
	}

	for _, tt := range tests {
		result := jp.calculateBackoff(tt.retryCount)
		if result != tt.expected {
			t.Errorf("backoff(%d) = %v, expected %v", tt.retryCount, result, tt.expected)
		}
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" + hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" + hex.EncodeToString(b[10:])
}
