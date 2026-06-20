package processing

import (
	"context"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/onurceri/botla-app/pkg/tokenizer"
)

func TestSourceProcessorInterface_Implementation(t *testing.T) {
	t.Parallel()

	// Create mock repositories for interface testing
	sourceRepo := repository.NewMockSourceRepo()
	usageRepo := &MockUsageRepo{}
	pendingURLRepo := &MockPendingURLRepo{}
	planRepo := repository.NewMockPlanRepo()
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockStorage := &storage.MockStorageService{}
	loader := &tokenizer.Loader{}

	t.Run("URLProcessor implements SourceProcessor", func(t *testing.T) {
		var _ SourceProcessor = NewURLProcessor(sourceRepo, usageRepo, pendingURLRepo, planRepo, mockOAI, mockVC, nil, nil, loader)
	})

	t.Run("PDFProcessor implements SourceProcessor", func(t *testing.T) {
		var _ SourceProcessor = NewPDFProcessor(sourceRepo, usageRepo, mockStorage, mockOAI, mockVC, nil, loader)
	})

	t.Run("TextProcessor implements SourceProcessor", func(t *testing.T) {
		var _ SourceProcessor = NewTextProcessor(sourceRepo, usageRepo, mockStorage, mockOAI, mockVC, nil, loader)
	})
}

func TestProcessorMap_Registry(t *testing.T) {
	t.Parallel()

	// Create mock repositories for interface testing
	sourceRepo := repository.NewMockSourceRepo()
	pendingURLRepo := &MockPendingURLRepo{}
	planRepo := repository.NewMockPlanRepo()
	usageRepo := &MockUsageRepo{}
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockStorage := &storage.MockStorageService{}
	loader := &tokenizer.Loader{}

	urlProc := NewURLProcessor(sourceRepo, usageRepo, pendingURLRepo, planRepo, mockOAI, mockVC, nil, nil, loader)
	pdfProc := NewPDFProcessor(sourceRepo, usageRepo, mockStorage, mockOAI, mockVC, nil, loader)
	textProc := NewTextProcessor(sourceRepo, usageRepo, mockStorage, mockOAI, mockVC, nil, loader)

	processors := map[string]SourceProcessor{
		"url":  urlProc,
		"pdf":  pdfProc,
		"text": textProc,
	}

	t.Run("URL processor is registered", func(t *testing.T) {
		proc, ok := processors["url"]
		if !ok {
			t.Error("url processor not found in map")
		}
		if proc != urlProc {
			t.Error("url processor is not the expected instance")
		}
	})

	t.Run("PDF processor is registered", func(t *testing.T) {
		proc, ok := processors["pdf"]
		if !ok {
			t.Error("pdf processor not found in map")
		}
		if proc != pdfProc {
			t.Error("pdf processor is not the expected instance")
		}
	})

	t.Run("Text processor is registered", func(t *testing.T) {
		proc, ok := processors["text"]
		if !ok {
			t.Error("text processor not found in map")
		}
		if proc != textProc {
			t.Error("text processor is not the expected instance")
		}
	})
}

func TestProcessorMap_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("Unknown source type returns false from lookup", func(t *testing.T) {
		processors := map[string]SourceProcessor{
			"url":  nil,
			"pdf":  nil,
			"text": nil,
		}

		_, ok := processors["unknown_type"]
		if ok {
			t.Error("unknown source type should not be found in map")
		}
	})

	t.Run("Nil processor in map is retrievable", func(t *testing.T) {
		processors := map[string]SourceProcessor{
			"url": nil,
		}

		proc, ok := processors["url"]
		if !ok {
			t.Error("url key not found in map")
		}
		if proc != nil {
			t.Error("expected nil processor")
		}
	})

	t.Run("Empty map returns false for any lookup", func(t *testing.T) {
		processors := make(map[string]SourceProcessor)

		_, ok := processors["url"]
		if ok {
			t.Error("empty map should not return true for any key")
		}
	})
}

// MockPendingURLRepo is a minimal mock for PendingURLRepository
type MockPendingURLRepo struct{}

func (m *MockPendingURLRepo) InsertPendingURL(ctx context.Context, chatbotID string, sourceID *string, url string) error {
	return nil
}

func (m *MockPendingURLRepo) ListPendingURLs(ctx context.Context, chatbotID string, limit, offset int) ([]models.PendingURL, error) {
	return nil, nil
}

func (m *MockPendingURLRepo) CountPendingURLs(ctx context.Context, chatbotID string) (int, error) {
	return 0, nil
}

func (m *MockPendingURLRepo) UpdatePendingURLStatus(ctx context.Context, chatbotID string, urlIDs []string, status string) (int, error) {
	return 0, nil
}

func (m *MockPendingURLRepo) GetPendingURLsByIDs(ctx context.Context, chatbotID string, urlIDs []string) ([]models.PendingURL, error) {
	return nil, nil
}

func (m *MockPendingURLRepo) DeletePendingURLsByChatbot(ctx context.Context, chatbotID string) (int, error) {
	return 0, nil
}

// MockUsageRepo is a minimal mock for UsageRepository
type MockUsageRepo struct{}

func (m *MockUsageRepo) CountChatbotsByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) CountChatbotsByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) GetFileCountByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) GetURLCountByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) GetStorageUsedMBByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) GetMaxFileCountInAnyBot(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) GetMaxURLCountInAnyBot(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) GetMonthlyTokenUsage(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) GetMonthlyIngestionUsage(ctx context.Context, userID string, month time.Time) (int, int, error) {
	return 0, 0, nil
}

func (m *MockUsageRepo) GetMonthlyRefreshCount(ctx context.Context, userID string, month time.Time) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) IncrementRefreshCount(ctx context.Context, userID string, month time.Time) error {
	return nil
}

func (m *MockUsageRepo) IncrementSuccessfulIngestion(ctx context.Context, userID string, at time.Time, delta int) error {
	return nil
}

func (m *MockUsageRepo) AddEmbeddingTokens(ctx context.Context, userID string, at time.Time, tokens int) error {
	return nil
}

func (m *MockUsageRepo) GetAutoRefreshCountForMonth(ctx context.Context, userID string, month time.Time) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) IncrementAutoRefreshCount(ctx context.Context, userID string, month time.Time, delta int) error {
	return nil
}

func (m *MockUsageRepo) ReserveChatTokens(ctx context.Context, userID string, estimatedTokens int, maxMonthlyTokens int) error {
	return nil
}

func (m *MockUsageRepo) AdjustChatTokens(ctx context.Context, userID string, deltaTokens int) error {
	return nil
}

func (m *MockUsageRepo) GetMonthlyChatTokens(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockUsageRepo) IncrementChatTokens(ctx context.Context, userID string, tokens int) error {
	return nil
}
