package processing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

const MaxRetries = 3

// SourceQueue manages background processing of data sources
type SourceQueue struct {
	ch           chan string
	stopCh       chan struct{}
	wg           sync.WaitGroup
	db           *sql.DB
	storage      storage.StorageService
	openaiClient rag.LLMClient
	vectorClient rag.VectorClient
	log          *logger.Logger

	// Processors
	urlProcessor  *URLProcessor
	pdfProcessor  *PDFProcessor
	textProcessor *TextProcessor

	workerCount int
}

// StartSourceQueue creates and starts a new source processing queue
func StartSourceQueue(dbpool *sql.DB, st storage.StorageService, oai rag.LLMClient, vc rag.VectorClient, workerCount int) (*SourceQueue, error) {
	c := make(chan string, 64)
	stop := make(chan struct{})

	log := logger.New("INFO")

	if workerCount > 16 {
		log.Warn("worker_count_capped", map[string]any{"requested": workerCount, "capped": 16})
		workerCount = 16
	}

	q := &SourceQueue{
		ch:           c,
		stopCh:       stop,
		db:           dbpool,
		storage:      st,
		openaiClient: oai,
		vectorClient: vc,
		log:          log,

		// Initialize processors
		urlProcessor:  NewURLProcessor(dbpool, oai, vc, log, nil),
		pdfProcessor:  NewPDFProcessor(dbpool, st, oai, vc, log),
		textProcessor: NewTextProcessor(dbpool, st, oai, vc, log),
		workerCount:   workerCount,
	}

	for i := 0; i < workerCount; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	// Ensure collection exists at startup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := vc.EnsureEmbeddingsCollection(ctx); err != nil {
		return nil, fmt.Errorf("ensure embeddings collection: %w", err)
	}

	// Recover pending jobs at startup
	go q.recoverPendingJobs()

	return q, nil
}

// EnqueueSource creates a training job and enqueues it for processing
func (q *SourceQueue) EnqueueSource(ctx context.Context, sourceID, chatbotID string) (string, error) {
	if q == nil || q.ch == nil {
		return "", fmt.Errorf("queue not initialized")
	}

	// Create training job record
	job, err := db.CreateTrainingJob(ctx, q.db, sourceID, chatbotID)
	if err != nil {
		return "", fmt.Errorf("create training job: %w", err)
	}

	// Enqueue the job ID
	select {
	case q.ch <- job.ID:
		if q.log != nil {
			q.log.Info("job_enqueued", map[string]any{
				"job_id":    job.ID,
				"source_id": sourceID,
			})
		}
		return job.ID, nil
	default:
		// Queue full, mark job as failed
		failedStep := models.StepFetchSource
		_ = db.FailJob(ctx, q.db, job.ID, failedStep, "QUEUE_FULL", "Processing queue is full")
		return "", fmt.Errorf("queue full")
	}
}

// Enqueue puts a job ID into the processing queue without creating a new job record
func (q *SourceQueue) Enqueue(jobID string) {
	if q == nil || q.ch == nil {
		return
	}
	select {
	case q.ch <- jobID:
		if q.log != nil {
			q.log.Info("job_enqueued_manually", map[string]any{
				"job_id": jobID,
			})
		}
	default:
		if q.log != nil {
			q.log.Warn("job_enqueue_failed_queue_full", map[string]any{
				"job_id": jobID,
			})
		}
	}
}

// Stop gracefully shuts down the queue worker
func (q *SourceQueue) Stop() {
	if q == nil || q.stopCh == nil {
		return
	}
	close(q.stopCh)
	q.wg.Wait()
}

// recoverPendingJobs finds and enqueues jobs stuck in 'pending' status at startup
func (q *SourceQueue) recoverPendingJobs() {
	defer func() {
		if r := recover(); r != nil {
			if q.log != nil {
				q.log.Error("recover_pending_jobs_panic", map[string]any{"panic": r})
			}
		}
	}()

	if q == nil || q.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if DB is reachable
	if err := q.db.PingContext(ctx); err != nil {
		if q.log != nil {
			q.log.Warn("recover_pending_jobs_db_unreachable", map[string]any{"error": err.Error()})
		}
		return
	}

	// Find jobs with pending status
	jobs, err := db.GetPendingJobs(ctx, q.db, 100)
	if err != nil {
		if q.log != nil {
			q.log.Warn("recover_pending_jobs_query_failed", map[string]any{"error": err.Error()})
		}
		return
	}

	var recovered int
RecoverLoop:
	for _, job := range jobs {
		select {
		case q.ch <- job.ID:
			recovered++
		default:
			if q.log != nil {
				q.log.Warn("recover_pending_jobs_queue_full", map[string]any{"job_id": job.ID})
			}
			break RecoverLoop
		}
	}

	if recovered > 0 && q.log != nil {
		q.log.Info("recover_pending_jobs_completed", map[string]any{
			"recovered_count": recovered,
		})
	}
}

// worker processes jobs from the queue
func (q *SourceQueue) worker(workerID int) {
	if q.ch == nil {
		return
	}
	for {
		select {
		case <-q.stopCh:
			if q.log != nil {
				q.log.Info("source_queue_shutdown", map[string]any{"worker_id": workerID})
			}
			q.wg.Done()
			return
		case jobID := <-q.ch:
			// Add a small delay to prevent rapid-fire API calls (rate limiting)
			// Skip in tests to avoid timeouts
			if os.Getenv("GO_ENV") != "test" {
				time.Sleep(500 * time.Millisecond)
			}
			q.processJob(jobID)
		}
	}
}

// processJob with retry support
func (q *SourceQueue) processJob(jobID string) {
	ctx := context.Background()

	// Load job
	job, err := db.GetTrainingJob(ctx, q.db, jobID)
	if err != nil || job == nil {
		if q.log != nil {
			q.log.Error("job_not_found", map[string]any{"job_id": jobID, "error": err})
		}
		return
	}

	// Check if this is a retry
	lastStep, _ := db.GetLastCompletedStep(ctx, q.db, jobID)

	// Update status to running
	var startStep models.TrainingStep
	if lastStep != nil {
		startStep = getNextStep(*lastStep)
	} else {
		startStep = models.StepFetchSource
	}

	_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusRunning, &startStep)

	// Load source and dependencies
	source, bot, langCode, plan, ok := q.loadSourceAndLang(job.SourceID)
	if !ok {
		_ = db.FailJob(ctx, q.db, jobID, models.StepFetchSource, "SOURCE_NOT_FOUND", "Source not found")
		return
	}

	// Mark source as processing
	q.markProcessing(job.SourceID)

	if q.log != nil {
		q.log.Info("job_processing_start", map[string]any{
			"job_id":      jobID,
			"source_id":   job.SourceID,
			"source_type": source.SourceType,
			"chatbot_id":  job.ChatbotID,
			"retry_count": job.RetryCount,
			"last_step":   lastStep,
		})
	}

	// Process with resume support
	result := q.processWithResume(ctx, jobID, source, bot, langCode, plan, lastStep)

	if result.Error != nil {
		// Check if we should retry
		if job.RetryCount < MaxRetries && isRetryableError(result.Error) {
			newCount, _ := db.IncrementRetryCount(ctx, q.db, jobID)

			backoff := 2 * time.Second
			for i := 1; i < newCount && i < 10; i++ {
				backoff *= 2
			}

			if q.log != nil {
				q.log.Info("job_retry_scheduled", map[string]any{
					"job_id":      jobID,
					"retry_count": newCount,
					"backoff":     backoff.String(),
					"error":       result.Error.Error(),
				})
			}

			// Re-enqueue with delay
			go func() {
				time.Sleep(backoff)
				q.ch <- jobID
			}()
			return
		}

		// Max retries exceeded or non-retryable error
		failedStep := result.FailedStep
		if failedStep == "" {
			failedStep = models.StepFetchSource
		}
		_ = db.FailJob(ctx, q.db, jobID, failedStep, "MAX_RETRIES", result.Error.Error())
		q.fail(job.SourceID, result.Error.Error())
		return
	}

	if result.Skipped {
		completedStep := models.StepStoreVectors
		_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusCompleted, &completedStep)
		q.complete(job.SourceID, result.ChunkCount)
		return
	}

	// Success
	completedStep := models.StepStoreVectors
	_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusCompleted, &completedStep)
	q.complete(job.SourceID, result.ChunkCount)

	// Enqueue any newly discovered sources
	for _, newSourceID := range result.NewSourceIDs {
		if _, err := q.EnqueueSource(ctx, newSourceID, source.ChatbotID); err != nil {
			if q.log != nil {
				q.log.Warn("enqueue_discovered_source_failed", map[string]any{
					"source_id": newSourceID,
					"error":     err.Error(),
				})
			}
		}
	}
}

func (q *SourceQueue) processWithResume(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep) ProcessResult {
	onStep := func(step models.TrainingStep) {
		_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusRunning, &step)
		// Mark the PREVIOUS step as completed when we move to the next one
		// This is a bit tricky because MarkStepCompleted wants an outputHash
		// For now, the processors will handle calling MarkStepCompleted internally 
		// if we update them, or we can do it here if we refactor.
	}

	switch source.SourceType {
	case "url":
		return q.urlProcessor.ProcessWithSteps(ctx, jobID, source, bot, langCode, plan, lastStep, func(s models.TrainingStep) {
			onStep(s)
			// We can't easily get the hash here without changing ProcessWithSteps return/callbacks
		})
	case "pdf":
		return q.pdfProcessor.ProcessWithSteps(ctx, jobID, source, bot, langCode, plan, lastStep, onStep)
	case "text":
		return q.textProcessor.ProcessWithSteps(ctx, jobID, source, bot, langCode, plan, lastStep, onStep)
	default:
		return ProcessResult{
			Error:      fmt.Errorf("unknown source type: %s", source.SourceType),
			FailedStep: models.StepFetchSource,
		}
	}
}

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

// loadSourceAndLang loads source, chatbot, and plan data
func (q *SourceQueue) loadSourceAndLang(sourceID string) (*models.DataSource, *models.Chatbot, string, *models.Plan, bool) {
	ctx := context.Background()

	s, err := db.GetSourceByID(ctx, q.db, sourceID)
	if err != nil || s == nil {
		q.fail(sourceID, "source_not_found")
		return nil, nil, "", nil, false
	}

	bot, err := db.GetChatbotByID(ctx, q.db, s.ChatbotID)
	if err != nil || bot == nil {
		q.fail(sourceID, "chatbot_not_found")
		return nil, nil, "", nil, false
	}

	plan, err := db.GetPlanByUserID(ctx, q.db, bot.UserID)
	if err != nil {
		q.fail(sourceID, "plan_error")
		return nil, nil, "", nil, false
	}

	// Fallback to empty plan if nil
	if plan == nil {
		plan = &models.Plan{}
	}

	return s, bot, defaultLang(bot.LanguageCode), plan, true
}

// markProcessing marks a source as processing
func (q *SourceQueue) markProcessing(id string) {
	if q.log != nil {
		q.log.Info("source_processing_start", map[string]any{"source_id": id})
	}
	ctx := context.Background()
	chunkCount := 0
	if err := q.db.QueryRowContext(ctx, `SELECT chunk_count FROM data_sources WHERE id=$1`, id).Scan(&chunkCount); err != nil {
		chunkCount = 0
	}
	_ = db.UpdateSourceProcessing(ctx, q.db, id, "processing", nil, chunkCount, nil)
}

// fail marks a source as failed
func (q *SourceQueue) fail(id string, msg string) {
	if q.log != nil {
		q.log.Warn("source_processing_fail", map[string]any{"source_id": id, "reason": msg})
	}
	_ = db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
}

// complete marks a source as completed
func (q *SourceQueue) complete(id string, chunks int) {
	if q.log != nil {
		q.log.Info("source_processing_complete", map[string]any{"source_id": id, "chunks": chunks})
	}
	now := time.Now()
	_ = db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, chunks, &now)
}

// defaultLang extracts base language code
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

func (q *SourceQueue) WorkerCount() int {
	return q.workerCount
}

func (q *SourceQueue) QueueLength() int {
	return len(q.ch)
}
