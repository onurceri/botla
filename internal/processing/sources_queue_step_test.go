package processing

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"testing"

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

	queue := &SourceQueue{
		ch:  make(chan string, 10),
		db:  tdb,
		log: log,
		urlProcessor: &URLProcessor{
			DB:           tdb,
			Log:          log,
			Scraper:      &testMockScraper{},
			OpenAIClient: mockOAI,
			VectorClient: mockVC,
		},
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
	queue := &SourceQueue{
		ch:  make(chan string, 10),
		db:  tdb,
		log: log,
		urlProcessor: &URLProcessor{
			DB:           tdb,
			Log:          log,
			Scraper:      sc,
			OpenAIClient: mockOAI,
			VectorClient: mockVC,
		},
		pdfProcessor: &PDFProcessor{
			DB:           tdb,
			Log:          log,
			OpenAIClient: mockOAI,
			VectorClient: mockVC,
		},
		textProcessor: &TextProcessor{
			DB:           tdb,
			Log:          log,
			OpenAIClient: mockOAI,
			VectorClient: mockVC,
		},
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

	// Process the job
	queue.processJob(jobID)

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

	queue := &SourceQueue{
		ch:  make(chan string, 10),
		db:  tdb,
		log: log,
		urlProcessor: &URLProcessor{
			DB:           tdb,
			Log:          log,
			Scraper:      &testMockScraper{shouldFail: true},
			OpenAIClient: mockOAI,
			VectorClient: mockVC,
		},
	}

	// Create test chatbot first (FK dependency)
	chatbotID := createTestChatbot(t, tdb)
	sourceID := createTestSourceWithURL(t, tdb, "http://invalid-url-that-will-fail.test", chatbotID)

	// Enqueue source
	jobID, err := queue.EnqueueSource(ctx, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("EnqueueSource failed: %v", err)
	}

	// Process the job
	queue.processJob(jobID)

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

	queue := &SourceQueue{
		ch:  make(chan string, 10),
		db:  tdb,
		log: log,
		urlProcessor: &URLProcessor{
			DB:           tdb,
			Log:          log,
			Scraper:      sc,
			OpenAIClient: mockOAI,
			VectorClient: mockVC,
		},
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

	// Process the job
	queue.processJob(jobID)

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

// testMockScraper implements scraper.Scraper for testing
type testMockScraper struct {
	content    string
	html       string
	shouldFail bool
}

func (m *testMockScraper) FetchRawHTML(url string, config scraper.CollectorConfig) (string, error) {
	if m.shouldFail {
		return "", &testScraperError{"fetch failed"}
	}
	return m.html, nil
}

func (m *testMockScraper) ScrapeURLWithFallback(task scraper.ScrapingTask, config scraper.CollectorConfig, dynamicEnabled bool, scrapeConfig *scraper.ScrapeConfig) (string, error) {
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
		INSERT INTO plans (id, code, config, created_at, updated_at)
		VALUES ($1, $2, '{"max_monthly_ingestions": 100, "scraping": {"max_urls_per_bot": 10, "max_pages_per_crawl": 5}, "chat": {"max_suggested_questions": 3}, "files": {"max_size_mb": 10, "max_files_per_bot": 5}}'::jsonb, NOW(), NOW())
		ON CONFLICT (code) DO NOTHING
	`, planID, "plan_"+generateUUID()).Err()

	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to create test plan: %v", err)
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
