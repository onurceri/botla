package processing

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
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
		
		result := p.Process(ctx, source, bot, "en", plan)
		
		assert.NoError(t, result.Error)
		assert.Greater(t, result.ChunkCount, 0)
		mockStorage.AssertExpectations(t)
		mockOAI.AssertExpectations(t)
		mockVC.AssertExpectations(t)
	})
}

func TestURLProcessor_Unit(t *testing.T) {
	// Note: URLProcessor uses scraper.ScrapeURLWithFallback which is package-level and hard to mock without more refactoring.
	// For this track, we'll focus on the interface injection part.
	mockOAI := &rag.MockLLMClient{}
	mockVC := &rag.MockVectorClient{}
	
	p := NewURLProcessor(nil, mockOAI, mockVC, nil)
	assert.NotNil(t, p.OpenAIClient)
	assert.NotNil(t, p.VectorClient)
}
