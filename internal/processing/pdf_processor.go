package processing

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/pdf"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/text"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/onurceri/botla-app/pkg/tokenizer"
)

// PDFProcessor handles PDF source processing
type PDFProcessor struct {
	sourceRepo       repository.SourceRepository
	usageRepo        repository.UsageRepository
	Storage          storage.StorageService
	OpenAIClient     rag.LLMClient
	VectorClient     rag.VectorClient
	Log              *logger.Logger
	EmbeddingService *rag.EmbeddingService
	Loader           *tokenizer.Loader
}

// NewPDFProcessor creates a new PDFProcessor
func NewPDFProcessor(sourceRepo repository.SourceRepository, usageRepo repository.UsageRepository, st storage.StorageService, oai rag.LLMClient, vc rag.VectorClient, log *logger.Logger, loader *tokenizer.Loader) *PDFProcessor {
	// Create EmbeddingService if we have an EmbeddingClient
	var embSvc *rag.EmbeddingService
	if emb, ok := oai.(rag.EmbeddingClient); ok {
		embSvc = rag.NewEmbeddingService(emb, vc, log)
	}
	return &PDFProcessor{
		sourceRepo:       sourceRepo,
		usageRepo:        usageRepo,
		Storage:          st,
		OpenAIClient:     oai,
		VectorClient:     vc,
		Log:              log,
		EmbeddingService: embSvc,
		Loader:           loader,
	}
}

// ProcessWithSteps processes a PDF source with step callbacks
func (p *PDFProcessor) ProcessWithSteps(ctx context.Context, jobID string, s *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
	if s.FilePath == nil || *s.FilePath == "" {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeEmptyFilePath}, FailedStep: models.StepFetchSource}
	}

	// Step 1: Fetch
	if lastStep == nil || (models.IsStepAtOrAfter(models.StepFetchSource, *lastStep) && models.StepFetchSource != *lastStep) {
		onStep(models.StepFetchSource)
	}

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

	// Step 2: Parse
	if lastStep == nil || (models.IsStepAtOrAfter(models.StepParseContent, *lastStep) && models.StepParseContent != *lastStep) {
		onStep(models.StepParseContent)
	}

	// Extract text from PDF
	content, err := pdf.ExtractPDFText(localPath, langCode, false)
	if err != nil {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodePDFParseFailed}, FailedStep: models.StepParseContent}
	}

	if content == "" {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeEmptyContent}, FailedStep: models.StepParseContent}
	}

	// Step 3: Chunk
	if lastStep == nil || (models.IsStepAtOrAfter(models.StepChunkText, *lastStep) && models.StepChunkText != *lastStep) {
		onStep(models.StepChunkText)
	}

	content = text.NormalizeTR(content)
	rc, rerr := rag.ChunkText(p.Loader, content, 512, langCode)
	if rerr != nil {
		p.logWarn("pdf_chunking_failed", map[string]any{"error": rerr.Error()})
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeChunkingFailed}, FailedStep: models.StepChunkText}
	}

	// Step 4: Embed
	if lastStep == nil || (models.IsStepAtOrAfter(models.StepEmbedChunks, *lastStep) && models.StepEmbedChunks != *lastStep) {
		onStep(models.StepEmbedChunks)
	}

	if p.EmbeddingService == nil {
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeLLMNotSupported}, FailedStep: models.StepEmbedChunks}
	}

	if embedErr := p.EmbeddingService.GenerateForSource(ctx, rc, s.ChatbotID, s.ID, s.SourceType); embedErr != nil {
		p.logWarn("pdf_embedding_failed", map[string]any{"error": embedErr.Error()})
		return ProcessResult{Error: &ProcessingError{Msg: ErrCodeEmbeddingFailed}, FailedStep: models.StepEmbedChunks}
	}

	// Step 5: Store
	if lastStep == nil || (models.IsStepAtOrAfter(models.StepStoreVectors, *lastStep) && models.StepStoreVectors != *lastStep) {
		onStep(models.StepStoreVectors)
	}

	// Calculate token usage
	var tokens int
	for _, ch := range rc {
		tokens += ch.TokenCount
	}

	now := time.Now()
	_ = p.sourceRepo.UpdateSourceProcessing(ctx, s.ID, "completed", nil, len(rc), &now)

	// Update usage statistics
	if err := p.usageRepo.IncrementSuccessfulIngestion(ctx, bot.UserID, now, 1); err != nil {
		p.logWarn("pdf_stats_increment_failed", map[string]any{"error": err.Error()})
	}
	if err := p.usageRepo.AddEmbeddingTokens(ctx, bot.UserID, now, tokens); err != nil {
		p.logWarn("pdf_stats_tokens_failed", map[string]any{"error": err.Error()})
	}

	return ProcessResult{ChunkCount: len(rc)}
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
