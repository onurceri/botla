package processing

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/stretchr/testify/mock"
)

// TestEnqueueSource_CreatesJob tests that EnqueueSource creates a job record
func TestEnqueueSource_CreatesJob(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	log := logger.New("DEBUG")
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{Content: `{"capability_summary":"test"}`}, nil).Maybe()
	mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
	mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockVC.On("DeleteBySourceID", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil).Maybe()

	// Create processor
	processor := NewJobProcessor(JobProcessorConfig{
		DB:           tdb,
		OpenAIClient: mockOAI,
		VectorClient: mockVC,
		Log:          log,
		Processors: map[string]SourceProcessor{
			"url": &URLProcessor{
				DB:           tdb,
				Log:          log,
				Scraper:      &testMockScraper{},
				OpenAIClient: mockOAI,
				VectorClient: mockVC,
			},
		},
	})

	// Create queue manager
	queueManager := NewQueueManager(1, log, processor)

	queue := &SourceQueue{
		queue:     queueManager,
		processor: processor,
		db:        tdb,
		log:       log,
	}

	// Create test chatbot first (FK dependency)
	chatbotID := createTestChatbot(t, tdb)
	sourceID := createTestSource(t, tdb, "url", chatbotID)

	// Enqueue source
	jobID, err := queue.EnqueueSource(ctx, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("EnqueueSource failed: %v", err)
	}

	if jobID == "" {
		t.Error("expected non-empty job ID")
	}

	// Verify job was created
	job, err := db.GetTrainingJob(ctx, tdb, jobID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}

	if job == nil {
		t.Fatal("expected job to be created")
	}

	if job.SourceID != sourceID {
		t.Errorf("expected source ID %s, got %s", sourceID, job.SourceID)
	}

	if job.ChatbotID != chatbotID {
		t.Errorf("expected chatbot ID %s, got %s", chatbotID, job.ChatbotID)
	}

	if job.Status != models.JobStatusPending {
		t.Errorf("expected status pending, got %s", job.Status)
	}
}

// TestProcessJob_UpdatesSteps tests that job steps are tracked during processing
func TestProcessJob_UpdatesSteps(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	log := logger.New("DEBUG")
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{Content: `{"capability_summary":"test"}`}, nil).Maybe()
	mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
	mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockVC.On("DeleteBySourceID", mock.Anything, mock.Anything).Return(nil).Maybe()

	sc := &testMockScraper{}

	// Create processor with mock implementations
	processor := NewJobProcessor(JobProcessorConfig{
		DB:           tdb,
		OpenAIClient: mockOAI,
		VectorClient: mockVC,
		Log:          log,
		Processors: map[string]SourceProcessor{
			"url": &URLProcessor{
				DB:           tdb,
				Log:          log,
				Scraper:      sc,
				OpenAIClient: mockOAI,
				VectorClient: mockVC,
			},
			"pdf": &PDFProcessor{
				DB:           tdb,
				Log:          log,
				OpenAIClient: mockOAI,
				VectorClient: mockVC,
			},
			"text": &TextProcessor{
				DB:           tdb,
				Log:          log,
				OpenAIClient: mockOAI,
				VectorClient: mockVC,
			},
		},
	})

	// Create queue manager
	queueManager := NewQueueManager(1, log, processor)

	queue := &SourceQueue{
		queue:     queueManager,
		processor: processor,
		db:        tdb,
		log:       log,
	}

	// Create test chatbot first (FK dependency)
	chatbotID := createTestChatbot(t, tdb)
	sourceID := createTestSourceWithURL(t, tdb, "https://example.com", chatbotID)

	// Set mock to return valid content
	sc.content = "This is test content for processing"
	sc.html = "<html><body>Test</body></html>"

	// Enqueue source
	jobID, err := queue.EnqueueSource(ctx, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("EnqueueSource failed: %v", err)
	}

	// Process the job directly via processor
	processor.HandleJob(jobID)

	// Verify job status
	job, err := db.GetTrainingJob(ctx, tdb, jobID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}

	if job == nil {
		t.Fatal("job should exist")
	}

	if job.Status != models.JobStatusCompleted && job.Status != models.JobStatusFailed {
		t.Errorf("expected terminal status, got %s", job.Status)
	}

	if job.StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}

// TestProcessJob_FailedStepTracked tests that failed step is tracked
func TestProcessJob_FailedStepTracked(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	log := logger.New("DEBUG")
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{Content: `{"capability_summary":"test"}`}, nil).Maybe()
	mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
	mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	// Create processor with failing mock scraper
	processor := NewJobProcessor(JobProcessorConfig{
		DB:           tdb,
		OpenAIClient: mockOAI,
		VectorClient: mockVC,
		Log:          log,
		Processors: map[string]SourceProcessor{
			"url": &URLProcessor{
				DB:           tdb,
				Log:          log,
				Scraper:      &testMockScraper{shouldFail: true},
				OpenAIClient: mockOAI,
				VectorClient: mockVC,
			},
		},
	})

	// Create queue manager
	queueManager := NewQueueManager(1, log, processor)

	queue := &SourceQueue{
		queue:     queueManager,
		processor: processor,
		db:        tdb,
		log:       log,
	}

	// Create test chatbot first (FK dependency)
	chatbotID := createTestChatbot(t, tdb)
	sourceID := createTestSourceWithURL(t, tdb, "http://invalid-url-that-will-fail.test", chatbotID)

	// Enqueue source
	jobID, err := queue.EnqueueSource(ctx, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("EnqueueSource failed: %v", err)
	}

	// Process the job directly via processor
	processor.HandleJob(jobID)

	// Verify job failed
	job, err := db.GetTrainingJob(ctx, tdb, jobID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}

	if job.Status != models.JobStatusFailed {
		t.Errorf("expected failed, got %s", job.Status)
	}

	if job.FailedStep == nil {
		t.Error("expected failed_step to be set")
	}
}

// TestProcessJob_SkipsUnchangedContent tests that unchanged content is skipped
func TestProcessJob_SkipsUnchangedContent(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	log := logger.New("DEBUG")
	sc := &testMockScraper{}
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{Content: `{"capability_summary":"test"}`}, nil).Maybe()
	mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
	mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockVC.On("DeleteBySourceID", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Create processor with mock scraper
	processor := NewJobProcessor(JobProcessorConfig{
		DB:           tdb,
		OpenAIClient: mockOAI,
		VectorClient: mockVC,
		Log:          log,
		Processors: map[string]SourceProcessor{
			"url": &URLProcessor{
				DB:           tdb,
				Log:          log,
				Scraper:      sc,
				OpenAIClient: mockOAI,
				VectorClient: mockVC,
			},
		},
	})

	// Create queue manager
	queueManager := NewQueueManager(1, log, processor)

	queue := &SourceQueue{
		queue:     queueManager,
		processor: processor,
		db:        tdb,
		log:       log,
	}

	// Create test chatbot first (FK dependency)
	chatbotID := createTestChatbot(t, tdb)
	sourceID := createTestSourceWithHash(t, tdb, "https://example.com", "726df8bcc21cb319dde031e10a3ab40ee5ce4979cef01451a9be341fec8e8153", chatbotID)

	// Set mock to return same content (hash will match)
	sc.content = "This is test content"
	sc.html = "<html><body>Test</body></html>"

	// Enqueue source
	jobID, err := queue.EnqueueSource(ctx, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("EnqueueSource failed: %v", err)
	}

	// Process the job directly via processor
	processor.HandleJob(jobID)

	// Verify job completed with skipped status
	job, err := db.GetTrainingJob(ctx, tdb, jobID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}

	// Job should be completed (skipped is a type of completion)
	if job.Status != models.JobStatusCompleted {
		t.Errorf("expected completed, got %s", job.Status)
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

// Helper functions

func createTestSource(t *testing.T, db *sql.DB, sourceType, chatbotID string) string {
	t.Helper()

	ctx := context.Background()
	sourceID := generateUUID()

	err := db.QueryRowContext(ctx, `
		INSERT INTO data_sources (id, chatbot_id, source_type, status, created_at)
		VALUES ($1, $2, $3, 'pending', NOW())
		ON CONFLICT (id) DO NOTHING
	`, sourceID, chatbotID, sourceType).Err()

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to create test source: %v", err)
	}

	return sourceID
}

func createTestSourceWithURL(t *testing.T, db *sql.DB, url, chatbotID string) string {
	t.Helper()

	ctx := context.Background()
	sourceID := generateUUID()

	err := db.QueryRowContext(ctx, `
		INSERT INTO data_sources (id, chatbot_id, source_type, status, source_url, created_at)
		VALUES ($1, $2, 'url', 'pending', $3, NOW())
		ON CONFLICT (id) DO NOTHING
	`, sourceID, chatbotID, url).Err()

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to create test source: %v", err)
	}

	return sourceID
}

func createTestSourceWithHash(t *testing.T, db *sql.DB, url, hash, chatbotID string) string {
	t.Helper()

	ctx := context.Background()
	sourceID := generateUUID()

	err := db.QueryRowContext(ctx, `
		INSERT INTO data_sources (id, chatbot_id, source_type, status, source_url, hash, chunk_count, created_at)
		VALUES ($1, $2, 'url', 'pending', $3, $4, 5, NOW())
		ON CONFLICT (id) DO NOTHING
	`, sourceID, chatbotID, url, hash).Err()

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to create test source: %v", err)
	}

	return sourceID
}

func createTestChatbot(t *testing.T, db *sql.DB) string {
	t.Helper()

	ctx := context.Background()
	chatbotID := generateUUID()
	userID := generateUUID()
	planID := generateUUID()

	// Create plan first
	err := db.QueryRowContext(ctx, `
		INSERT INTO plans (id, code, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (code) DO NOTHING
	`, planID, "plan_"+generateUUID()).Err()

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to create test plan: %v", err)
	}

	// Create plan limits
	_, err = db.ExecContext(ctx, `
		INSERT INTO plan_limits (plan_id, max_monthly_ingestions, scraping_max_urls_per_bot, scraping_max_pages_per_crawl,
			chat_max_suggested_questions, files_max_size_mb, files_max_files_per_bot, created_at)
		VALUES ($1, 100, 10, 5, 3, 10, 5, NOW())
		ON CONFLICT (plan_id) DO NOTHING
	`, planID)

	if err != nil {
		t.Fatalf("failed to create test plan limits: %v", err)
	}

	// Create user
	err = db.QueryRowContext(ctx, `
		INSERT INTO users (id, email, password_hash, plan_id)
		VALUES ($1, $2, 'hash', $3)
		ON CONFLICT (id) DO NOTHING
	`, userID, "user_"+generateUUID()+"@test.com", planID).Err()

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to create test user: %v", err)
	}

	err = db.QueryRowContext(ctx, `
		INSERT INTO chatbots (id, user_id, name, created_at, updated_at)
		VALUES ($1, $2, 'test-bot', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, chatbotID, userID).Err()

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to create test chatbot: %v", err)
	}

	return chatbotID
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" + hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" + hex.EncodeToString(b[10:])
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
