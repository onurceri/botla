package processing

import (
	"context"
	"fmt"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/scraper"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/onurceri/botla-app/pkg/tokenizer"
)

// MaxRetries is the maximum number of retry attempts for a failed job.
const MaxRetries = 3

// SourceQueue orchestrates background processing of data sources.
// It combines QueueManager for worker lifecycle and JobProcessor for business logic.
type SourceQueue struct {
	queue           *QueueManager
	processor       *JobProcessor
	trainingJobRepo repository.TrainingJobRepository
	log             *logger.Logger
	loader          *tokenizer.Loader
}

// StartSourceQueue creates and starts a new source processing queue.
func StartSourceQueue(trainingJobRepo repository.TrainingJobRepository, st storage.StorageService, oai rag.LLMClient, vc rag.VectorClient, sc scraper.Scraper, loader *tokenizer.Loader, workerCount int) (*SourceQueue, error) {
	log := logger.New("INFO")

	// Create the orchestrator first (needed for circular dependency with processor)
	sq := &SourceQueue{
		trainingJobRepo: trainingJobRepo,
		log:             log,
	}

	// Create processor with enqueue callback
	processor := NewJobProcessor(JobProcessorConfig{
		TrainingJobRepo: trainingJobRepo,
		Storage:         st,
		OpenAIClient:    oai,
		VectorClient:    vc,
		Log:             log,
		Scraper:         sc,
		Loader:          loader,
		EnqueueWithDelay: func(jobID string, delay time.Duration) {
			sq.queue.EnqueueWithDelay(jobID, delay)
		},
	})

	sq.processor = processor

	// Create queue manager with processor as handler
	queueManager := NewQueueManager(workerCount, log, processor)
	sq.queue = queueManager

	// Ensure collection exists at startup, but don't block startup if it fails
	// The worker will retry operations if needed
	if vc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := vc.EnsureEmbeddingsCollection(ctx); err != nil {
			log.Warn("ensure_embeddings_collection_failed", map[string]any{"error": err.Error()})
		}
	}

	// Start workers
	queueManager.Start()

	// Recover pending jobs at startup
	go sq.recoverPendingJobs()

	return sq, nil
}

// EnqueueSource creates a training job and enqueues it for processing.
func (sq *SourceQueue) EnqueueSource(ctx context.Context, sourceID, chatbotID string) (string, error) {
	if sq == nil || sq.queue == nil {
		return "", fmt.Errorf("queue not initialized")
	}

	// Create training job record
	job, err := sq.trainingJobRepo.Create(ctx, sourceID, chatbotID)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "create training job")
	}

	// Enqueue the job ID
	if sq.queue.Enqueue(job.ID) {
		if sq.log != nil {
			sq.log.Info("job_enqueued", map[string]any{
				"job_id":    job.ID,
				"source_id": sourceID,
			})
		}
		return job.ID, nil
	}

	// Queue full, mark job as failed
	failedStep := models.StepFetchSource
	_ = sq.trainingJobRepo.Fail(ctx, job.ID, failedStep, "QUEUE_FULL", "Processing queue is full")
	return "", fmt.Errorf("queue full")
}

// Enqueue puts a job ID into the processing queue without creating a new job record.
func (sq *SourceQueue) Enqueue(jobID string) {
	if sq == nil || sq.queue == nil {
		return
	}
	sq.queue.Enqueue(jobID)
}

// Stop gracefully shuts down the queue worker.
func (sq *SourceQueue) Stop() {
	if sq == nil || sq.queue == nil {
		return
	}
	sq.queue.Stop()
}

// WorkerCount returns the number of active workers.
func (sq *SourceQueue) WorkerCount() int {
	if sq == nil || sq.queue == nil {
		return 0
	}
	return sq.queue.WorkerCount()
}

// QueueLength returns the current number of jobs in the queue.
func (sq *SourceQueue) QueueLength() int {
	if sq == nil || sq.queue == nil {
		return 0
	}
	return sq.queue.QueueLength()
}

// recoverPendingJobs finds and enqueues jobs stuck in 'pending' status at startup.
func (sq *SourceQueue) recoverPendingJobs() {
	defer func() {
		if r := recover(); r != nil {
			if sq.log != nil {
				sq.log.Error("recover_pending_jobs_panic", map[string]any{"panic": r})
			}
		}
	}()

	if sq == nil || sq.trainingJobRepo == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find jobs with pending status
	jobs, err := sq.trainingJobRepo.GetPendingJobs(ctx, 100)
	if err != nil {
		if sq.log != nil {
			sq.log.Warn("recover_pending_jobs_query_failed", map[string]any{"error": err.Error()})
		}
		return
	}

	var recovered int
	for _, job := range jobs {
		if sq.queue.Enqueue(job.ID) {
			recovered++
		} else {
			if sq.log != nil {
				sq.log.Warn("recover_pending_jobs_queue_full", map[string]any{"job_id": job.ID})
			}
			break
		}
	}

	if recovered > 0 && sq.log != nil {
		sq.log.Info("recover_pending_jobs_completed", map[string]any{
			"recovered_count": recovered,
		})
	}
}
