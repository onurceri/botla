package processing

import (
	"context"
	"database/sql"
	"io"
	"os"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/pdf"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/text"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

// PDFProcessor handles PDF source processing
type PDFProcessor struct {
	DB           *sql.DB
	Storage      storage.StorageService
	OpenAIClient rag.LLMClient
	Log          *logger.Logger
}

// NewPDFProcessor creates a new PDFProcessor
func NewPDFProcessor(db *sql.DB, st storage.StorageService, oai rag.LLMClient, log *logger.Logger) *PDFProcessor {
	return &PDFProcessor{
		DB:           db,
		Storage:      st,
		OpenAIClient: oai,
		Log:          log,
	}
}

// Process processes a PDF source
func (p *PDFProcessor) Process(ctx context.Context, s *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan) ProcessResult {
	if s.FilePath == nil || *s.FilePath == "" {
		return ProcessResult{Error: &ProcessingError{Msg: "empty_file_path"}}
	}

	localPath := *s.FilePath

	// Download file from storage if available
	if p.Storage != nil {
		rc, err := p.Storage.DownloadFile(ctx, *s.FilePath)
		if err != nil {
			return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
		}

		tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
		if err != nil {
			_ = rc.Close()
			return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
		}

		_, err = io.Copy(tmpFile, rc)
		_ = rc.Close()
		_ = tmpFile.Close()
		if err != nil {
			return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
		}

		localPath = tmpFile.Name()
		defer func() { _ = os.Remove(localPath) }()
	}

	// Extract text from PDF
	content, err := pdf.ExtractPDFText(localPath, langCode, plan.Config.Files.OCREnabled)
	if err != nil {
		return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
	}
	if content == "" {
		return ProcessResult{ChunkCount: 0}
	}

	content = text.NormalizeTR(content)

	// Extract and persist metadata
	p.persistIngestionMetadata(ctx, content, langCode, s)

	// Chunk and embed
	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		return ProcessResult{Error: &ProcessingError{Msg: rerr.Error()}}
	}
	if err := rag.GenerateEmbeddingsForSource(rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
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
func (p *PDFProcessor) persistIngestionMetadata(ctx context.Context, content, langCode string, s *models.DataSource) {
	meta, err := rag.ExtractIngestionMetadata(ctx, p.OpenAIClient, content, langCode)
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

func (p *PDFProcessor) logInfo(event string, data map[string]any) {
	if p.Log != nil {
		p.Log.Info(event, data)
	}
}

func (p *PDFProcessor) logWarn(event string, data map[string]any) {
	if p.Log != nil {
		p.Log.Warn(event, data)
	}
}
