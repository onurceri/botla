package processing

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/text"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

// TextProcessor handles text source processing
type TextProcessor struct {
	DB           *sql.DB
	Storage      storage.StorageService
	OpenAIClient rag.LLMClient
	VectorClient rag.VectorClient
	Log          *logger.Logger
}

// NewTextProcessor creates a new TextProcessor
func NewTextProcessor(db *sql.DB, st storage.StorageService, oai rag.LLMClient, vc rag.VectorClient, log *logger.Logger) *TextProcessor {
	return &TextProcessor{
		DB:           db,
		Storage:      st,
		OpenAIClient: oai,
		VectorClient: vc,
		Log:          log,
	}
}

// Process processes a text source
func (p *TextProcessor) Process(ctx context.Context, s *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan) ProcessResult {
	if s.FilePath == nil || *s.FilePath == "" {
		return ProcessResult{Error: &ProcessingError{Msg: "empty_file_path"}}
	}

	var content string
	if p.Storage != nil {
		rc, err := p.Storage.DownloadFile(ctx, *s.FilePath)
		if err != nil {
			return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
		}
		content = string(b)
	} else {
		return ProcessResult{Error: &ProcessingError{Msg: "storage_required"}}
	}

	if content == "" {
		return ProcessResult{ChunkCount: 0}
	}

	content = text.NormalizeTR(content)

	// Extract and persist metadata
	maxQuestions := 0
	if plan != nil && plan.Config.Chat.MaxSuggestedQuestions > 0 {
		maxQuestions = plan.Config.Chat.MaxSuggestedQuestions
	}
	p.persistIngestionMetadata(ctx, content, langCode, s, maxQuestions)

	// Chunk and embed
	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		return ProcessResult{Error: &ProcessingError{Msg: rerr.Error()}}
	}

	emb, ok := p.OpenAIClient.(rag.EmbeddingClient)
	if !ok {
		return ProcessResult{Error: &ProcessingError{Msg: "llm_client_does_not_support_embeddings"}}
	}

	if err := rag.GenerateEmbeddingsForSource(ctx, emb, p.VectorClient, rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
		return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
	}

	// Calculate token usage
	var tokens int
	for _, ch := range rc {
		tokens += ch.TokenCount
	}
	_ = db.IncrementSuccessfulIngestion(ctx, p.DB, bot.UserID, time.Now(), 1)
	_ = db.AddEmbeddingTokens(ctx, p.DB, bot.UserID, time.Now(), tokens)

	return ProcessResult{ChunkCount: len(rc)}
}

// persistIngestionMetadata extracts and saves metadata for the source
func (p *TextProcessor) persistIngestionMetadata(ctx context.Context, content, langCode string, s *models.DataSource, maxQuestions int) {
	meta, err := rag.ExtractIngestionMetadata(ctx, p.OpenAIClient, content, langCode, maxQuestions)
	if err != nil {
		p.logWarn("extract_metadata_failed", map[string]any{"source_id": s.ID, "error": err.Error()})
		return
	}

	if len(meta.SuggestedQuestions) == 0 {
		p.logWarn("extract_metadata_empty_questions", map[string]any{"source_id": s.ID})
	} else {
		p.logInfo("extract_metadata_success", map[string]any{
			"source_id":       s.ID,
			"questions_count": len(meta.SuggestedQuestions),
			"questions":       meta.SuggestedQuestions,
		})
	}

	if err := db.UpdateSourceCapability(ctx, p.DB, s.ID, meta.CapabilitySummary); err != nil {
		p.logWarn("update_source_capability_failed", map[string]any{"source_id": s.ID, "error": err.Error()})
	}
	if err := db.UpdateSourceSuggestions(ctx, p.DB, s.ID, meta.SuggestedQuestions); err != nil {
		p.logWarn("update_source_suggestions_failed", map[string]any{"source_id": s.ID, "error": err.Error()})
	}

	go AggregateAndPersistChatbotSuggestions(ctx, p.DB, s.ChatbotID, p.Log)
}

func (p *TextProcessor) logInfo(event string, data map[string]any) {
	if p.Log != nil {
		p.Log.Info(event, data)
	}
}

func (p *TextProcessor) logWarn(event string, data map[string]any) {
	if p.Log != nil {
		p.Log.Warn(event, data)
	}
}
