package processing

import (
	"os"
	"sync"
	"time"

	"github.com/onurceri/botla-app/pkg/logger"
)

// QueueManager handles the job queue lifecycle and worker management.
// It is responsible for enqueueing jobs, managing workers, and graceful shutdown.
type QueueManager struct {
	ch      chan string
	stopCh  chan struct{}
	wg      sync.WaitGroup
	workers int
	log     *logger.Logger
	handler JobHandler
}

// JobHandler is the interface for processing jobs from the queue.
type JobHandler interface {
	HandleJob(jobID string)
}

// NewQueueManager creates a new QueueManager with the specified worker count.
func NewQueueManager(workerCount int, log *logger.Logger, handler JobHandler) *QueueManager {
	if workerCount > 16 {
		if log != nil {
			log.Warn("worker_count_capped", map[string]any{"requested": workerCount, "capped": 16})
		}
		workerCount = 16
	}

	return &QueueManager{
		ch:      make(chan string, 64),
		stopCh:  make(chan struct{}),
		workers: workerCount,
		log:     log,
		handler: handler,
	}
}

// Start begins processing jobs with the configured number of workers.
func (q *QueueManager) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// Enqueue adds a job ID to the processing queue.
// Returns true if the job was enqueued, false if the queue is full.
func (q *QueueManager) Enqueue(jobID string) bool {
	if q == nil || q.ch == nil {
		return false
	}

	select {
	case q.ch <- jobID:
		if q.log != nil {
			q.log.Info("job_enqueued", map[string]any{"job_id": jobID})
		}
		return true
	default:
		if q.log != nil {
			q.log.Warn("job_enqueue_failed_queue_full", map[string]any{"job_id": jobID})
		}
		return false
	}
}

// EnqueueWithDelay adds a job ID to the queue after a delay.
// This is used for retry backoff.
func (q *QueueManager) EnqueueWithDelay(jobID string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		q.Enqueue(jobID)
	}()
}

// Stop gracefully shuts down the queue and waits for all workers to finish.
func (q *QueueManager) Stop() {
	if q == nil || q.stopCh == nil {
		return
	}
	close(q.stopCh)
	q.wg.Wait()
}

// WorkerCount returns the number of active workers.
func (q *QueueManager) WorkerCount() int {
	return q.workers
}

// QueueLength returns the current number of jobs waiting in the queue.
func (q *QueueManager) QueueLength() int {
	if q == nil || q.ch == nil {
		return 0
	}
	return len(q.ch)
}

// worker is the main loop for processing jobs.
func (q *QueueManager) worker(workerID int) {
	defer q.wg.Done()

	if q.ch == nil {
		return
	}

	for {
		select {
		case <-q.stopCh:
			if q.log != nil {
				q.log.Info("source_queue_shutdown", map[string]any{"worker_id": workerID})
			}
			return
		case jobID := <-q.ch:
			// Add a small delay to prevent rapid-fire API calls (rate limiting)
			// Skip in tests to avoid timeouts
			if os.Getenv("GO_ENV") != "test" {
				time.Sleep(500 * time.Millisecond)
			}
			if q.handler != nil {
				q.handler.HandleJob(jobID)
			}
		}
	}
}
