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
	VectorClient rag.VectorClient
	Log          *logger.Logger
}

// NewPDFProcessor creates a new PDFProcessor
func NewPDFProcessor(db *sql.DB, st storage.StorageService, oai rag.LLMClient, vc rag.VectorClient, log *logger.Logger) *PDFProcessor {
	return &PDFProcessor{
		DB:           db,
		Storage:      st,
		OpenAIClient: oai,
		VectorClient: vc,
		Log:          log,
	}
}

// ProcessWithSteps processes a PDF source with step callbacks
func (p *PDFProcessor) ProcessWithSteps(ctx context.Context, jobID string, s *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
	if s.FilePath == nil || *s.FilePath == "" {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeEmptyFilePath}, FailedStep: models.StepFetchSource}
	}

	// Step 1: Fetch
	onStep(models.StepFetchSource)

	localPath := *s.FilePath

	// Download file from storage if available
	if p.Storage != nil {
		rc, err := p.Storage.DownloadFile(ctx, *s.FilePath)
		if err != nil {
			return ProcessResult{Error: &ProcessingError{Msg: ErrCodePDFDownloadFailed}, FailedStep: models.StepFetchSource}
		}

		tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
		if err != nil {
			_ = rc.Close()
			return ProcessResult{Error: &ProcessingError{Msg: ErrCodePDFDownloadFailed}, FailedStep: models.StepFetchSource}
		}

		_, err = io.Copy(tmpFile, rc)
		_ = rc.Close()
		_ = tmpFile.Close()
		if err != nil {
			return ProcessResult{Error: &ProcessingError{Msg: ErrCodePDFDownloadFailed}, FailedStep: models.StepFetchSource}
		}

		localPath = tmpFile.Name()
		defer func() { _ = os.Remove(localPath) }()
	}

	_ = db.MarkStepCompleted(ctx, p.DB, jobID, models.StepFetchSource, "")

	// Step 2: Parse
	onStep(models.StepParseContent)

	// Extract text from PDF
	content, err := pdf.ExtractPDFText(localPath, langCode, plan.Config.Files.OCREnabled)
	if err != nil {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodePDFParseFailed}, FailedStep: models.StepParseContent}
	}

	_ = db.MarkStepCompleted(ctx, p.DB, jobID, models.StepParseContent, "")
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

	// Step 3: Chunk
	onStep(models.StepChunkText)

	// Chunk and embed
	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeChunkingFailed}, FailedStep: models.StepChunkText}
	}

	_ = db.MarkStepCompleted(ctx, p.DB, jobID, models.StepChunkText, "")

	// Step 4: Embed
	onStep(models.StepEmbedChunks)

	emb, ok := p.OpenAIClient.(rag.EmbeddingClient)
	if !ok {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeLLMNotSupported}, FailedStep: models.StepEmbedChunks}
	}

	if err := rag.GenerateEmbeddingsForSource(ctx, emb, p.VectorClient, rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeEmbeddingFailed}, FailedStep: models.StepEmbedChunks}
	}

	_ = db.MarkStepCompleted(ctx, p.DB, jobID, models.StepEmbedChunks, "")

	// Step 5: Store
	onStep(models.StepStoreVectors)

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
func (p *PDFProcessor) persistIngestionMetadata(ctx context.Context, content, langCode string, s *models.DataSource, maxQuestions int) {
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
