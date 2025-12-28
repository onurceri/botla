package processing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

// JobProcessor handles the actual processing of training jobs.
// It orchestrates URL, PDF, and Text processors with retry and resume support.
type JobProcessor struct {
	db            *sql.DB
	storage       storage.StorageService
	openaiClient  rag.LLMClient
	vectorClient  rag.VectorClient
	log           *logger.Logger
	urlProcessor  *URLProcessor
	pdfProcessor  *PDFProcessor
	textProcessor *TextProcessor

	// Callback for re-enqueueing jobs with delay (for retries)
	enqueueWithDelay func(jobID string, delay time.Duration)
}

// JobProcessorConfig contains the configuration for creating a JobProcessor.
type JobProcessorConfig struct {
	DB               *sql.DB
	Storage          storage.StorageService
	OpenAIClient     rag.LLMClient
	VectorClient     rag.VectorClient
	Log              *logger.Logger
	EnqueueWithDelay func(jobID string, delay time.Duration)
}

// NewJobProcessor creates a new JobProcessor with the given configuration.
func NewJobProcessor(cfg JobProcessorConfig) *JobProcessor {
	return &JobProcessor{
		db:               cfg.DB,
		storage:          cfg.Storage,
		openaiClient:     cfg.OpenAIClient,
		vectorClient:     cfg.VectorClient,
		log:              cfg.Log,
		urlProcessor:     NewURLProcessor(cfg.DB, cfg.OpenAIClient, cfg.VectorClient, cfg.Log, nil),
		pdfProcessor:     NewPDFProcessor(cfg.DB, cfg.Storage, cfg.OpenAIClient, cfg.VectorClient, cfg.Log),
		textProcessor:    NewTextProcessor(cfg.DB, cfg.Storage, cfg.OpenAIClient, cfg.VectorClient, cfg.Log),
		enqueueWithDelay: cfg.EnqueueWithDelay,
	}
}

// HandleJob implements JobHandler interface and processes a single job by ID.
func (p *JobProcessor) HandleJob(jobID string) {
	p.processJob(jobID)
}

// processJob handles the complete job processing lifecycle.
func (p *JobProcessor) processJob(jobID string) {
	ctx := context.Background()

	// Load job
	job, err := db.GetTrainingJob(ctx, p.db, jobID)
	if err != nil || job == nil {
		if p.log != nil {
			p.log.Error("job_not_found", map[string]any{"job_id": jobID, "error": err})
		}
		return
	}

	// Check if this is a retry
	lastStep, _ := db.GetLastCompletedStep(ctx, p.db, jobID)

	// Update status to running
	var startStep models.TrainingStep
	if lastStep != nil {
		startStep = getNextStep(*lastStep)
	} else {
		startStep = models.StepFetchSource
	}

	_ = db.UpdateJobStatus(ctx, p.db, jobID, models.JobStatusRunning, &startStep)

	// Load source and dependencies
	source, bot, langCode, plan, ok := p.loadSourceAndLang(job.SourceID)
	if !ok {
		_ = db.FailJob(ctx, p.db, jobID, models.StepFetchSource, "SOURCE_NOT_FOUND", "Source not found")
		return
	}

	// Mark source as processing
	p.markProcessing(job.SourceID)

	if p.log != nil {
		p.log.Info("job_processing_start", map[string]any{
			"job_id":      jobID,
			"source_id":   job.SourceID,
			"source_type": source.SourceType,
			"chatbot_id":  job.ChatbotID,
			"retry_count": job.RetryCount,
			"last_step":   lastStep,
		})
	}

	// Process with resume support
	result := p.processWithResume(ctx, jobID, source, bot, langCode, plan, lastStep)

	if result.Error != nil {
		// Check if we should retry
		if job.RetryCount < MaxRetries && isRetryableError(result.Error) {
			newCount, _ := db.IncrementRetryCount(ctx, p.db, jobID)

			backoff := p.calculateBackoff(newCount)

			if p.log != nil {
				p.log.Info("job_retry_scheduled", map[string]any{
					"job_id":      jobID,
					"retry_count": newCount,
					"backoff":     backoff.String(),
					"error":       result.Error.Error(),
				})
			}

			// Re-enqueue with delay
			if p.enqueueWithDelay != nil {
				p.enqueueWithDelay(jobID, backoff)
			}
			return
		}

		// Max retries exceeded or non-retryable error
		failedStep := result.FailedStep
		if failedStep == "" {
			failedStep = models.StepFetchSource
		}
		_ = db.FailJob(ctx, p.db, jobID, failedStep, "MAX_RETRIES", result.Error.Error())
		p.fail(job.SourceID, result.Error.Error())
		return
	}

	if result.Skipped {
		completedStep := models.StepStoreVectors
		_ = db.UpdateJobStatus(ctx, p.db, jobID, models.JobStatusCompleted, &completedStep)
		p.complete(job.SourceID, result.ChunkCount)
		return
	}

	// Success
	completedStep := models.StepStoreVectors
	_ = db.UpdateJobStatus(ctx, p.db, jobID, models.JobStatusCompleted, &completedStep)
	p.complete(job.SourceID, result.ChunkCount)

	// Return discovered sources (handled by orchestrator)
	// Note: NewSourceIDs are stored in result and will be handled externally
}

// GetNewSourceIDs processes a job and returns any newly discovered source IDs.
// This is used by the orchestrator to enqueue discovered sources.
func (p *JobProcessor) ProcessAndGetDiscoveredSources(jobID string) []string {
	ctx := context.Background()

	job, err := db.GetTrainingJob(ctx, p.db, jobID)
	if err != nil || job == nil {
		return nil
	}

	lastStep, _ := db.GetLastCompletedStep(ctx, p.db, jobID)
	source, bot, langCode, plan, ok := p.loadSourceAndLang(job.SourceID)
	if !ok {
		return nil
	}

	result := p.processWithResume(ctx, jobID, source, bot, langCode, plan, lastStep)
	return result.NewSourceIDs
}

// calculateBackoff returns the backoff duration for a given retry count.
func (p *JobProcessor) calculateBackoff(retryCount int) time.Duration {
	backoff := 2 * time.Second
	for i := 1; i < retryCount && i < 10; i++ {
		backoff *= 2
	}
	return backoff
}

// processWithResume handles processing with step resume support.
func (p *JobProcessor) processWithResume(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep) ProcessResult {
	onStep := func(step models.TrainingStep) {
		_ = db.UpdateJobStatus(ctx, p.db, jobID, models.JobStatusRunning, &step)
	}

	switch source.SourceType {
	case "url":
		return p.urlProcessor.ProcessWithSteps(ctx, jobID, source, bot, langCode, plan, lastStep, func(s models.TrainingStep) {
			onStep(s)
		})
	case "pdf":
		return p.pdfProcessor.ProcessWithSteps(ctx, jobID, source, bot, langCode, plan, lastStep, onStep)
	case "text":
		return p.textProcessor.ProcessWithSteps(ctx, jobID, source, bot, langCode, plan, lastStep, onStep)
	default:
		return ProcessResult{
			Error:      fmt.Errorf("unknown source type: %s", source.SourceType),
			FailedStep: models.StepFetchSource,
		}
	}
}

// loadSourceAndLang loads source, chatbot, and plan data.
func (p *JobProcessor) loadSourceAndLang(sourceID string) (*models.DataSource, *models.Chatbot, string, *models.Plan, bool) {
	ctx := context.Background()

	s, err := db.GetSourceByID(ctx, p.db, sourceID)
	if err != nil || s == nil {
		p.fail(sourceID, "source_not_found")
		return nil, nil, "", nil, false
	}

	bot, err := db.GetChatbotByID(ctx, p.db, s.ChatbotID)
	if err != nil || bot == nil {
		p.fail(sourceID, "chatbot_not_found")
		return nil, nil, "", nil, false
	}

	plan, err := db.GetPlanByUserID(ctx, p.db, bot.UserID)
	if err != nil {
		p.fail(sourceID, "plan_error")
		return nil, nil, "", nil, false
	}

	// Fallback to empty plan if nil
	if plan == nil {
		plan = &models.Plan{}
	}

	return s, bot, defaultLang(bot.LanguageCode), plan, true
}

// markProcessing marks a source as processing.
func (p *JobProcessor) markProcessing(id string) {
	if p.log != nil {
		p.log.Info("source_processing_start", map[string]any{"source_id": id})
	}
	ctx := context.Background()
	chunkCount := 0
	if err := p.db.QueryRowContext(ctx, `SELECT chunk_count FROM data_sources WHERE id=$1`, id).Scan(&chunkCount); err != nil {
		chunkCount = 0
	}
	_ = db.UpdateSourceProcessing(ctx, p.db, id, "processing", nil, chunkCount, nil)
}

// fail marks a source as failed.
func (p *JobProcessor) fail(id string, msg string) {
	if p.log != nil {
		p.log.Warn("source_processing_fail", map[string]any{"source_id": id, "reason": msg})
	}
	_ = db.UpdateSourceProcessing(context.Background(), p.db, id, "failed", &msg, 0, nil)
}

// complete marks a source as completed.
func (p *JobProcessor) complete(id string, chunks int) {
	if p.log != nil {
		p.log.Info("source_processing_complete", map[string]any{"source_id": id, "chunks": chunks})
	}
	now := time.Now()
	_ = db.UpdateSourceProcessing(context.Background(), p.db, id, "completed", nil, chunks, &now)
}

// isRetryableError determines if an error should trigger a retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// Retry on network errors, rate limits, temporary failures
	retryable := []string{
		"connection refused",
		"timeout",
		"rate limit",
		"429",
		"503",
		"502",
		"temporary",
		"context deadline exceeded",
	}
	for _, pattern := range retryable {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// getNextStep returns the next step after the given step.
func getNextStep(current models.TrainingStep) models.TrainingStep {
	switch current {
	case models.StepFetchSource:
		return models.StepParseContent
	case models.StepParseContent:
		return models.StepChunkText
	case models.StepChunkText:
		return models.StepEmbedChunks
	case models.StepEmbedChunks:
		return models.StepStoreVectors
	default:
		return models.StepFetchSource
	}
}

// defaultLang extracts base language code.
func defaultLang(code string) string {
	s := strings.TrimSpace(code)
	if s == "" {
		return "tr"
	}
	if i := strings.Index(s, "-"); i > 0 {
		s = s[:i]
	}
	return s
}
