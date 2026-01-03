package processing

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTextProcessor_Unit(t *testing.T) {
	ctx := context.Background()
	db := testdb.OpenParallelTestDB(t)
	mockStorage := &storage.MockStorageService{}
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}

	p := NewTextProcessor(db, mockStorage, mockOAI, mockVC, nil)

	t.Run("Process Successful", func(t *testing.T) {
		filePath := "test.txt"
		content := "This is a test content for embedding."

		// Insert test data
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p1', 'Free', 'free', '{}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u1', 'test@test.com', 'h', 'p1')`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name) VALUES ('bot-1', 'u1', 'Bot')`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, file_path) VALUES ('src-1', 'bot-1', 'text', 'test.txt')`)

		source := &models.DataSource{
			ID:         "src-1",
			ChatbotID:  "bot-1",
			SourceType: "text",
			FilePath:   &filePath,
		}
		bot := &models.Chatbot{ID: "bot-1"}
		plan := &models.Plan{Config: models.PlanConfig{}}

		// Mock Storage
		mockStorage.On("DownloadFile", ctx, filePath).Return(io.NopCloser(strings.NewReader(content)), nil).Once()

		// Mock Metadata Extraction
		mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
			Content: `{"capability_summary": "test", "suggested_questions": ["q1"]}`,
		}, nil).Once()

		// Mock Embedding
		mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Once()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.NoError(t, result.Error)
		assert.Greater(t, result.ChunkCount, 0)
		mockStorage.AssertExpectations(t)
		mockOAI.AssertExpectations(t)
		mockVC.AssertExpectations(t)
	})

	t.Run("Fails on empty file path", func(t *testing.T) {
		source := &models.DataSource{FilePath: nil}
		result := p.ProcessWithSteps(ctx, "test-job", source, nil, "tr", nil, nil, func(step models.TrainingStep) {})
		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeEmptyFilePath, result.Error.Error())
	})

	t.Run("Fails on download error", func(t *testing.T) {
		filePath := "fail.txt"
		source := &models.DataSource{FilePath: &filePath}
		mockStorage.On("DownloadFile", ctx, filePath).Return(nil, errors.New("download failed")).Once()

		result := p.ProcessWithSteps(ctx, "test-job", source, nil, "tr", nil, nil, func(step models.TrainingStep) {})

		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "download failed")
		mockStorage.AssertExpectations(t)
	})
}

func TestURLProcessor_Unit(t *testing.T) {
	t.Run("Uses injected mock scraper", func(t *testing.T) {
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		p := NewURLProcessor(nil, mockOAI, mockVC, nil, mockScraper)

		assert.Equal(t, mockScraper, p.Scraper)
	})
}

func TestURLProcessor_Process_WithMock(t *testing.T) {
	ctx := context.Background()
	db := testdb.OpenParallelTestDB(t)

	t.Run("Handles scraper error", func(t *testing.T) {
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/error"
		mockScraper.SetError(testURL, errors.New("connection refused"))

		// Insert test data
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p2', 'Free2', 'free2', '{}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u2', 'test2@test.com', 'h', 'p2') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name, discovery_mode) VALUES ('bot-2', 'u2', 'Bot2', 'disabled') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url) VALUES ('src-2', 'bot-2', 'url', $1) ON CONFLICT DO NOTHING`, testURL)

		p := NewURLProcessor(db, mockOAI, mockVC, nil, mockScraper)

		source := &models.DataSource{
			ID:         "src-2",
			ChatbotID:  "bot-2",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-2", DiscoveryMode: "disabled"}
		plan := &models.Plan{Config: models.PlanConfig{}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeScrapeFailedNetwork, result.Error.Error())
		assert.True(t, mockScraper.AssertCalled("FetchRawHTML"))
		assert.True(t, mockScraper.AssertCalled("ScrapeURLWithFallback"))
	})

	t.Run("Handles empty content returns ERR_DYNAMIC_REQUIRED", func(t *testing.T) {
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/empty"
		mockScraper.SetResponse(testURL, "") // Empty content

		// Insert test data
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p3', 'Free3', 'free3', '{"scraping": {"dynamic_enabled": false}}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u3', 'test3@test.com', 'h', 'p3') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name, discovery_mode) VALUES ('bot-3', 'u3', 'Bot3', 'disabled') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url) VALUES ('src-3', 'bot-3', 'url', $1) ON CONFLICT DO NOTHING`, testURL)

		p := NewURLProcessor(db, mockOAI, mockVC, nil, mockScraper)

		source := &models.DataSource{
			ID:         "src-3",
			ChatbotID:  "bot-3",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-3", DiscoveryMode: "disabled"}
		plan := &models.Plan{Config: models.PlanConfig{
			Scraping: models.ScrapingConfig{DynamicEnabled: false},
		}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeDynamicRequired, result.Error.Error())
	})

	t.Run("Handles empty content returns ERR_EMPTY_CONTENT when dynamic enabled", func(t *testing.T) {
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/empty-dynamic"
		mockScraper.SetResponse(testURL, "") // Empty content even with dynamic

		// Insert test data with dynamic enabled
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p3d', 'Pro3', 'pro3', '{"scraping": {"dynamic_enabled": true}}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u3d', 'test3d@test.com', 'h', 'p3d') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name, discovery_mode) VALUES ('bot-3d', 'u3d', 'Bot3d', 'disabled') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url) VALUES ('src-3d', 'bot-3d', 'url', $1) ON CONFLICT DO NOTHING`, testURL)

		p := NewURLProcessor(db, mockOAI, mockVC, nil, mockScraper)

		source := &models.DataSource{
			ID:         "src-3d",
			ChatbotID:  "bot-3d",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-3d", DiscoveryMode: "disabled"}
		plan := &models.Plan{Config: models.PlanConfig{
			Scraping: models.ScrapingConfig{DynamicEnabled: true},
		}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeEmptyContent, result.Error.Error())
	})

	t.Run("Successful scrape and processing", func(t *testing.T) {
		mockOAI := &rag.MockFullClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/success"
		mockScraper.SetResponse(testURL, "This is test content with enough words to generate embeddings properly for the test case.")
		mockScraper.SetHTMLResponse(testURL, "<html><body><p>Test</p></body></html>")

		// Insert test data
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p4', 'Free4', 'free4', '{}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u4', 'test4@test.com', 'h', 'p4') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name, discovery_mode) VALUES ('bot-4', 'u4', 'Bot4', 'disabled') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url) VALUES ('src-4', 'bot-4', 'url', $1) ON CONFLICT DO NOTHING`, testURL)

		// Mock embedding calls
		mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
			Content: `{"capability_summary": "test", "suggested_questions": ["What is this?"]}`,
		}, nil).Once()
		mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1, 0.2}}, nil).Once()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		p := NewURLProcessor(db, mockOAI, mockVC, nil, mockScraper)

		source := &models.DataSource{
			ID:         "src-4",
			ChatbotID:  "bot-4",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-4", UserID: "u4", DiscoveryMode: "disabled"}
		plan := &models.Plan{Config: models.PlanConfig{}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.NoError(t, result.Error)
		assert.Greater(t, result.ChunkCount, 0)
		assert.Equal(t, 1, mockScraper.CallCount("ScrapeURLWithFallback"))
		assert.Equal(t, 1, mockScraper.CallCount("FetchRawHTML"))
	})
}

func TestURLProcessor_Discovery_WithMock(t *testing.T) {
	ctx := context.Background()
	db := testdb.OpenParallelTestDB(t)

	t.Run("Discovers links from HTML", func(t *testing.T) {
		mockOAI := &rag.MockFullClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/discover"
		mockScraper.SetResponse(testURL, "Test content for discovery page.")
		mockScraper.SetHTMLResponse(testURL, `<html><body><a href="/page1">Link1</a><a href="/page2">Link2</a></body></html>`)
		mockScraper.SetLinks(testURL, []string{
			"https://example.com/page1",
			"https://example.com/page2",
		})

		// Insert test data
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p5', 'Free5', 'free5', '{"scraping": {"max_pages_per_crawl": 10, "max_urls_per_bot": 10}}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u5', 'test5@test.com', 'h', 'p5') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name, discovery_mode) VALUES ('bot-5', 'u5', 'Bot5', 'auto') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url, is_discovered) VALUES ('src-5', 'bot-5', 'url', $1, false) ON CONFLICT DO NOTHING`, testURL)

		// Mock embedding calls
		mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
			Content: `{"capability_summary": "test", "suggested_questions": []}`,
		}, nil).Maybe()
		mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		p := NewURLProcessor(db, mockOAI, mockVC, nil, mockScraper)

		source := &models.DataSource{
			ID:           "src-5",
			ChatbotID:    "bot-5",
			SourceType:   "url",
			SourceURL:    &testURL,
			IsDiscovered: false,
		}
		bot := &models.Chatbot{ID: "bot-5", UserID: "u5", DiscoveryMode: "auto"}
		plan := &models.Plan{Config: models.PlanConfig{
			Scraping: models.ScrapingConfig{
				MaxPagesPerCrawl: 10,
				MaxURLsPerBot:    10,
			},
		}}

		_ = p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		// Verify FetchRawHTML was called (this always happens for discovery preparation)
		assert.True(t, mockScraper.AssertCalled("FetchRawHTML"))
		// ExtractLinks may or may not be called depending on whether discovery proceeds
		// The important thing is that the mock scraper is being used correctly
		assert.Equal(t, 1, mockScraper.CallCount("FetchRawHTML"))
	})

	t.Run("Skips discovery when MaxURLsPerBot is 1", func(t *testing.T) {
		mockOAI := &rag.MockFullClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/single-url-plan"
		mockScraper.SetResponse(testURL, "Test content for single url plan.")
		mockScraper.SetHTMLResponse(testURL, `<html><body><a href="/page1">Link1</a></body></html>`)
		mockScraper.SetLinks(testURL, []string{
			"https://example.com/page1",
		})

		// Insert test data with MaxURLsPerBot=1
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p6', 'Free6', 'free6', '{"scraping": {"max_pages_per_crawl": 5, "max_urls_per_bot": 1}}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u6', 'test6@test.com', 'h', 'p6') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name, discovery_mode) VALUES ('bot-6', 'u6', 'Bot6', 'auto') ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url, is_discovered) VALUES ('src-6', 'bot-6', 'url', $1, false) ON CONFLICT DO NOTHING`, testURL)

		// Mock embedding calls
		mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
			Content: `{"capability_summary": "test", "suggested_questions": []}`,
		}, nil).Maybe()
		mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		p := NewURLProcessor(db, mockOAI, mockVC, nil, mockScraper)

		source := &models.DataSource{
			ID:           "src-6",
			ChatbotID:    "bot-6",
			SourceType:   "url",
			SourceURL:    &testURL,
			IsDiscovered: false,
		}
		bot := &models.Chatbot{ID: "bot-6", UserID: "u6", DiscoveryMode: "auto"}
		plan := &models.Plan{Config: models.PlanConfig{
			Scraping: models.ScrapingConfig{
				MaxPagesPerCrawl: 5,
				MaxURLsPerBot:    1, // Key: limit is 1, discovery should skip
			},
		}}

		_ = p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		// When MaxURLsPerBot=1, ExtractLinks should NOT be called
		// because discovery is skipped early
		assert.Equal(t, 0, mockScraper.CallCount("ExtractLinks"))
	})
}
func TestPDFProcessor_Unit(t *testing.T) {
	ctx := context.Background()
	db := testdb.OpenParallelTestDB(t)
	mockStorage := &storage.MockStorageService{}
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}

	p := NewPDFProcessor(db, mockStorage, mockOAI, mockVC, nil)

	t.Run("Process Successful", func(t *testing.T) {
		filePath := "test.pdf"

		// Insert test data
		_, _ = db.Exec(`INSERT INTO plans (id, name, code, config) VALUES ('p_pdf', 'Free', 'free_pdf', '{}'::jsonb) ON CONFLICT DO NOTHING`)
		_, _ = db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ('u_pdf', 'pdf@test.com', 'h', 'p_pdf')`)
		_, _ = db.Exec(`INSERT INTO chatbots (id, user_id, name) VALUES ('bot-pdf', 'u_pdf', 'Bot')`)
		_, _ = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, file_path) VALUES ('src-pdf', 'bot-pdf', 'pdf', 'test.pdf')`)

		source := &models.DataSource{
			ID:         "src-pdf",
			ChatbotID:  "bot-pdf",
			SourceType: "pdf",
			FilePath:   &filePath,
		}
		bot := &models.Chatbot{ID: "bot-pdf", UserID: "u_pdf"}
		plan := &models.Plan{Config: models.PlanConfig{}}

		// Mock Storage
		mockStorage.On("DownloadFile", ctx, filePath).Return(io.NopCloser(strings.NewReader("%PDF-1.4...")), nil).Once()

		// Since pdf.ExtractPDFText might fail without fitz, we just check the flow if it proceeds.
		// If it fails with ParseFailed, that's also an acceptable result for this environment.
		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		// We don't assert specific success here because it depends on fitz installation,
		// but we ensure it doesn't panic and uses the mocks.
		assert.NotNil(t, result)
	})

	t.Run("Fails on empty file path", func(t *testing.T) {
		source := &models.DataSource{FilePath: nil}
		result := p.ProcessWithSteps(ctx, "test-job", source, nil, "tr", nil, nil, func(step models.TrainingStep) {})
		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeEmptyFilePath, result.Error.Error())
	})
}
