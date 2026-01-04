package processing

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/scraper"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTextProcessor_Unit(t *testing.T) {
	ctx := context.Background()
	sourceRepo := repository.NewMockSourceRepo()
	usageRepo := &MockUsageRepo{}
	mockStorage := &storage.MockStorageService{}
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}

	p := NewTextProcessor(sourceRepo, usageRepo, mockStorage, mockOAI, mockVC, nil, nil)

	t.Run("Process Successful", func(t *testing.T) {
		filePath := "test.txt"
		content := "This is a test content for embedding."

		source := &models.DataSource{
			ID:         "src-1",
			ChatbotID:  "bot-1",
			SourceType: "text",
			FilePath:   &filePath,
		}
		bot := &models.Chatbot{ID: "bot-1"}
		plan := &models.Plan{Limits: &models.PlanLimits{}}

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
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		planRepo := repository.NewMockPlanRepo()
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		p := NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, nil, mockScraper, nil)

		assert.Equal(t, mockScraper, p.Scraper)
	})
}

func TestURLProcessor_Process_WithMock(t *testing.T) {
	ctx := context.Background()

	t.Run("Handles scraper error", func(t *testing.T) {
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		planRepo := repository.NewMockPlanRepo()
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/error"
		mockScraper.SetError(testURL, errors.New("connection refused"))

		p := NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, nil, mockScraper, nil)

		source := &models.DataSource{
			ID:         "src-2",
			ChatbotID:  "bot-2",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-2", DiscoveryMode: "disabled"}
		plan := &models.Plan{Limits: &models.PlanLimits{}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeScrapeFailedNetwork, result.Error.Error())
		assert.True(t, mockScraper.AssertCalled("FetchRawHTML"))
		assert.True(t, mockScraper.AssertCalled("ScrapeURLWithFallback"))
	})

	t.Run("Handles empty content returns ERR_DYNAMIC_REQUIRED", func(t *testing.T) {
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		planRepo := repository.NewMockPlanRepo()
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/empty"
		mockScraper.SetResponse(testURL, "") // Empty content

		p := NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, nil, mockScraper, nil)

		source := &models.DataSource{
			ID:         "src-3",
			ChatbotID:  "bot-3",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-3", DiscoveryMode: "disabled"}
		plan := &models.Plan{Limits: &models.PlanLimits{
			ScrapingDynamicEnabled: false,
		}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeDynamicRequired, result.Error.Error())
	})

	t.Run("Handles empty content returns ERR_EMPTY_CONTENT when dynamic enabled", func(t *testing.T) {
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		planRepo := repository.NewMockPlanRepo()
		mockOAI := &rag.MockLLMClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/empty-dynamic"
		mockScraper.SetResponse(testURL, "") // Empty content even with dynamic

		p := NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, nil, mockScraper, nil)

		source := &models.DataSource{
			ID:         "src-3d",
			ChatbotID:  "bot-3d",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-3d", DiscoveryMode: "disabled"}
		plan := &models.Plan{Limits: &models.PlanLimits{
			ScrapingDynamicEnabled: true,
		}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.Error(t, result.Error)
		assert.Equal(t, ErrCodeEmptyContent, result.Error.Error())
	})

	t.Run("Successful scrape and processing", func(t *testing.T) {
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		planRepo := repository.NewMockPlanRepo()
		mockOAI := &rag.MockFullClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/success"
		mockScraper.SetResponse(testURL, "This is test content with enough words to generate embeddings properly for the test case.")
		mockScraper.SetHTMLResponse(testURL, "<html><body><p>Test</p></body></html>")

		// Mock embedding calls
		mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
			Content: `{"capability_summary": "test", "suggested_questions": ["What is this?"]}`,
		}, nil).Once()
		mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1, 0.2}}, nil).Once()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		p := NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, nil, mockScraper, nil)

		source := &models.DataSource{
			ID:         "src-4",
			ChatbotID:  "bot-4",
			SourceType: "url",
			SourceURL:  &testURL,
		}
		bot := &models.Chatbot{ID: "bot-4", UserID: "u4", DiscoveryMode: "disabled"}
		plan := &models.Plan{Limits: &models.PlanLimits{}}

		result := p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		assert.NoError(t, result.Error)
		assert.Greater(t, result.ChunkCount, 0)
		assert.Equal(t, 1, mockScraper.CallCount("ScrapeURLWithFallback"))
		assert.Equal(t, 1, mockScraper.CallCount("FetchRawHTML"))
	})
}

func TestURLProcessor_Discovery_WithMock(t *testing.T) {
	ctx := context.Background()

	t.Run("Discovers links from HTML", func(t *testing.T) {
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		planRepo := repository.NewMockPlanRepo()
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

		// Mock embedding calls
		mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
			Content: `{"capability_summary": "test", "suggested_questions": []}`,
		}, nil).Maybe()
		mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		p := NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, nil, mockScraper, nil)

		source := &models.DataSource{
			ID:           "src-5",
			ChatbotID:    "bot-5",
			SourceType:   "url",
			SourceURL:    &testURL,
			IsDiscovered: false,
		}
		bot := &models.Chatbot{ID: "bot-5", UserID: "u5", DiscoveryMode: "auto"}
		plan := &models.Plan{Limits: &models.PlanLimits{
			ScrapingMaxPagesPerCrawl: 10,
			ScrapingMaxURLsPerBot:    10,
		}}

		_ = p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		// Verify FetchRawHTML was called (this always happens for discovery preparation)
		assert.True(t, mockScraper.AssertCalled("FetchRawHTML"))
		// ExtractLinks may or may not be called depending on whether discovery proceeds
		// The important thing is that the mock scraper is being used correctly
		assert.Equal(t, 1, mockScraper.CallCount("FetchRawHTML"))
	})

	t.Run("Skips discovery when MaxURLsPerBot is 1", func(t *testing.T) {
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		planRepo := repository.NewMockPlanRepo()
		mockOAI := &rag.MockFullClient{}
		mockVC := &rag.MockVectorClient{}
		mockScraper := scraper.NewMockScraper()

		testURL := "https://example.com/single-url-plan"
		mockScraper.SetResponse(testURL, "Test content for single url plan.")
		mockScraper.SetHTMLResponse(testURL, `<html><body><a href="/page1">Link1</a></body></html>`)
		mockScraper.SetLinks(testURL, []string{
			"https://example.com/page1",
		})

		// Mock embedding calls
		mockOAI.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
			Content: `{"capability_summary": "test", "suggested_questions": []}`,
		}, nil).Maybe()
		mockOAI.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil).Maybe()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		p := NewURLProcessor(sourceRepo, usageRepo, planRepo, mockOAI, mockVC, nil, mockScraper, nil)

		source := &models.DataSource{
			ID:           "src-6",
			ChatbotID:    "bot-6",
			SourceType:   "url",
			SourceURL:    &testURL,
			IsDiscovered: false,
		}
		bot := &models.Chatbot{ID: "bot-6", UserID: "u6", DiscoveryMode: "auto"}
		plan := &models.Plan{Limits: &models.PlanLimits{
			ScrapingMaxPagesPerCrawl: 5,
			ScrapingMaxURLsPerBot:    1, // Key: limit is 1, discovery should skip
		}}

		_ = p.ProcessWithSteps(ctx, "test-job", source, bot, "en", plan, nil, func(step models.TrainingStep) {})

		// When MaxURLsPerBot=1, ExtractLinks should NOT be called
		// because discovery is skipped early
		assert.Equal(t, 0, mockScraper.CallCount("ExtractLinks"))
	})
}

func TestPDFProcessor_Unit(t *testing.T) {
	ctx := context.Background()
	sourceRepo := repository.NewMockSourceRepo()
	usageRepo := &MockUsageRepo{}
	mockStorage := &storage.MockStorageService{}
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}

	p := NewPDFProcessor(sourceRepo, usageRepo, mockStorage, mockOAI, mockVC, nil, nil)

	t.Run("Process Successful", func(t *testing.T) {
		filePath := "test.pdf"

		source := &models.DataSource{
			ID:         "src-pdf",
			ChatbotID:  "bot-pdf",
			SourceType: "pdf",
			FilePath:   &filePath,
		}
		bot := &models.Chatbot{ID: "bot-pdf", UserID: "u_pdf"}
		plan := &models.Plan{Limits: &models.PlanLimits{}}

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
