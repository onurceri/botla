# Task 004: Integrate Job Tracking into Source Queue

**Priority:** 🔴 Critical (Core Functionality)  
**Phase:** 2 - Async Training Improvements  
**Estimated Time:** 3-4 hours  
**Dependencies:** Task 002 (Job State Table), Task 003 (Job Progress API)  

---

## Problem Statement

Currently, the `SourceQueue` processes sources without creating job records. The new job tracking system from Task 002 needs to be integrated into the processing pipeline.

**Current flow:**
1. Source created with `pending` status
2. `Enqueue(sourceID)` called
3. Worker processes source
4. Source status updated to `completed` or `failed`

**Desired flow:**
1. Source created with `pending` status
2. **TrainingJob created**
3. `Enqueue(jobID)` called
4. Worker processes job, **updating step status along the way**
5. Job and source status updated to `completed` or `failed`

---

## Objective

Modify `SourceQueue` to:
1. Create a `TrainingJob` when enqueueing a source
2. Update job step/progress as processing progresses
3. Track which step failed if processing fails
4. Support future step-level retry

---

## Implementation Details

### Step 1: Modify Enqueue to Create Job

**File:** `internal/processing/sources_queue.go` (MODIFY)

```go
// EnqueueSource creates a job and enqueues it for processing
// Returns the job ID for tracking
func (q *SourceQueue) EnqueueSource(ctx context.Context, sourceID, chatbotID string) (string, error) {
	if q == nil || q.ch == nil {
		return "", fmt.Errorf("queue not initialized")
	}

	// Create training job record
	job, err := db.CreateTrainingJob(ctx, q.db, sourceID, chatbotID)
	if err != nil {
		return "", fmt.Errorf("create training job: %w", err)
	}

	// Enqueue the job ID (not source ID)
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
		_ = db.FailJob(ctx, q.db, job.ID, models.StepFetchSource, "QUEUE_FULL", "Processing queue is full")
		return "", fmt.Errorf("queue full")
	}
}

// Enqueue is deprecated, use EnqueueSource instead
// Kept for backward compatibility during migration
func (q *SourceQueue) Enqueue(id string) {
	if q == nil || q.ch == nil {
		return
	}
	select {
	case q.ch <- id:
	default:
		if q.log != nil {
			q.log.Warn("source_queue_full", map[string]any{"dropped_id": id})
		}
	}
}
```

### Step 2: Modify Worker to Process Jobs

**File:** `internal/processing/sources_queue.go` (MODIFY)

```go
// worker processes jobs from the queue
func (q *SourceQueue) worker() {
	if q.ch == nil {
		return
	}
	for {
		select {
		case <-q.stopCh:
			if q.log != nil {
				q.log.Info("source_queue_shutdown", nil)
			}
			return
		case jobID := <-q.ch:
			if os.Getenv("GO_ENV") != "test" {
				time.Sleep(500 * time.Millisecond)
			}
			q.processJob(jobID)
		}
	}
}

// processJob handles processing of a single job
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

	// Update job to running
	step := models.StepFetchSource
	if err := db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusRunning, &step); err != nil {
		q.log.Error("update_job_failed", map[string]any{"job_id": jobID, "error": err.Error()})
	}

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
		})
	}

	// Process with step tracking
	var result ProcessResult
	switch source.SourceType {
	case "url":
		result = q.urlProcessor.ProcessWithSteps(ctx, source, bot, langCode, plan, func(step models.TrainingStep) {
			_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusRunning, &step)
		})
	case "pdf":
		result = q.pdfProcessor.ProcessWithSteps(ctx, source, bot, langCode, plan, func(step models.TrainingStep) {
			_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusRunning, &step)
		})
	case "text":
		result = q.textProcessor.ProcessWithSteps(ctx, source, bot, langCode, plan, func(step models.TrainingStep) {
			_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusRunning, &step)
		})
	default:
		_ = db.FailJob(ctx, q.db, jobID, models.StepFetchSource, "UNKNOWN_TYPE", "Unknown source type: "+source.SourceType)
		q.fail(job.SourceID, "unknown_source_type")
		return
	}

	if result.Error != nil {
		failedStep := result.FailedStep
		if failedStep == "" {
			failedStep = models.StepFetchSource
		}
		_ = db.FailJob(ctx, q.db, jobID, failedStep, "PROCESSING_ERROR", result.Error.Error())
		q.fail(job.SourceID, result.Error.Error())
		return
	}

	// Mark job as completed
	completedStep := models.StepStoreVectors
	_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusCompleted, &completedStep)
	q.complete(job.SourceID, result.ChunkCount)

	// Enqueue discovered sources
	for _, newID := range result.NewSourceIDs {
		q.Enqueue(newID) // Legacy method for discovered sources
	}
}
```

### Step 3: Add Step Callback to Processors

**File:** `internal/processing/url_processor.go` (MODIFY)

```go
// ProcessResult now includes FailedStep
type ProcessResult struct {
	ChunkCount   int
	NewSourceIDs []string
	Error        error
	Skipped      bool
	FailedStep   models.TrainingStep // NEW: Which step failed
}

// StepCallback is called when a step is started
type StepCallback func(step models.TrainingStep)

// ProcessWithSteps processes a URL source with step callbacks
func (p *URLProcessor) ProcessWithSteps(
	ctx context.Context,
	source *models.DataSource,
	bot *models.Chatbot,
	langCode string,
	plan *models.Plan,
	onStep StepCallback,
) ProcessResult {
	// Step 1: Fetch
	onStep(models.StepFetchSource)
	content, err := p.fetchURL(ctx, *source.SourceURL)
	if err != nil {
		return ProcessResult{Error: err, FailedStep: models.StepFetchSource}
	}

	// Step 2: Parse
	onStep(models.StepParseContent)
	text, err := p.parseContent(content)
	if err != nil {
		return ProcessResult{Error: err, FailedStep: models.StepParseContent}
	}

	// Check if content changed (hash comparison)
	newHash := computeHash([]byte(text))
	if source.Hash != nil && *source.Hash == newHash {
		return ProcessResult{Skipped: true}
	}

	// Step 3: Chunk
	onStep(models.StepChunkText)
	chunks, err := p.chunkText(text, langCode)
	if err != nil {
		return ProcessResult{Error: err, FailedStep: models.StepChunkText}
	}

	// Step 4: Embed
	onStep(models.StepEmbedChunks)
	embeddings, err := p.embedChunks(ctx, chunks)
	if err != nil {
		return ProcessResult{Error: err, FailedStep: models.StepEmbedChunks}
	}

	// Step 5: Store
	onStep(models.StepStoreVectors)
	if err := p.storeVectors(ctx, source.ID, embeddings); err != nil {
		return ProcessResult{Error: err, FailedStep: models.StepStoreVectors}
	}

	// Update hash
	_ = db.UpdateSourceHash(ctx, p.db, source.ID, newHash)

	return ProcessResult{ChunkCount: len(chunks)}
}

// Process maintains backward compatibility
func (p *URLProcessor) Process(ctx context.Context, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan) ProcessResult {
	return p.ProcessWithSteps(ctx, source, bot, langCode, plan, func(step models.TrainingStep) {
		// No-op callback for backward compatibility
	})
}
```

### Step 4: Update Handler to Use New Method

**File:** `internal/api/handlers/source_create.go` (MODIFY)

Update `persistAndEnqueueInternal` to use `EnqueueSource`:

```go
func (h *SourcesHandlers) persistAndEnqueueInternal(r *http.Request, ds *models.DataSource) (string, error) {
	ctx := r.Context()

	// Insert source
	newID, err := db.CreateDataSource(ctx, h.DB, ds)
	if err != nil {
		return "", err
	}

	// Create job and enqueue
	if h.Queue != nil && h.Queue.SourceQueue != nil {
		jobID, err := h.Queue.SourceQueue.EnqueueSource(ctx, newID, ds.ChatbotID)
		if err != nil {
			h.logError("enqueue_source_failed", map[string]any{
				"source_id": newID,
				"error":     err.Error(),
			})
			// Source is created, job failed to enqueue
			// Mark source as failed
			_ = db.UpdateSourceProcessing(ctx, h.DB, newID, "failed", ptrStr("queue_failed"), 0, nil)
			return newID, nil // Return ID but log error
		}
		h.logInfo("source_enqueued", map[string]any{
			"source_id": newID,
			"job_id":    jobID,
		})
	}

	return newID, nil
}

func ptrStr(s string) *string {
	return &s
}
```

---

## Tests to Write

### Unit Tests for Step Tracking

**File:** `internal/processing/sources_queue_step_test.go` (NEW)

```go
package processing

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
)

func TestProcessJob_UpdatesSteps(t *testing.T) {
	testDB := setupTestDB(t)
	queue := setupTestQueue(t, testDB)

	source := createTestSource(t, testDB)
	chatbot := createTestChatbot(t, testDB)

	// Enqueue and get job ID
	jobID, err := queue.EnqueueSource(context.Background(), source.ID, chatbot.ID)
	if err != nil {
		t.Fatalf("EnqueueSource failed: %v", err)
	}

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Check job was updated through steps
	job, err := db.GetTrainingJob(context.Background(), testDB.db, jobID)
	if err != nil {
		t.Fatalf("GetTrainingJob failed: %v", err)
	}

	if job.Status != models.JobStatusCompleted && job.Status != models.JobStatusFailed {
		t.Errorf("expected terminal status, got %s", job.Status)
	}

	if job.StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}

func TestProcessJob_FailedStepTracked(t *testing.T) {
	testDB := setupTestDB(t)
	queue := setupTestQueue(t, testDB)

	// Create source with invalid URL that will fail
	source := &models.DataSource{
		ID:         "test-source",
		ChatbotID:  "test-bot",
		SourceType: "url",
		SourceURL:  ptr("http://invalid-url-that-will-fail.test"),
	}
	createTestSourceWithData(t, testDB, source)

	jobID, _ := queue.EnqueueSource(context.Background(), source.ID, source.ChatbotID)

	// Wait for processing
	time.Sleep(2 * time.Second)

	job, _ := db.GetTrainingJob(context.Background(), testDB.db, jobID)

	if job.Status != models.JobStatusFailed {
		t.Errorf("expected failed, got %s", job.Status)
	}
	if job.FailedStep == nil {
		t.Error("expected failed_step to be set")
	}
	if job.ErrorMessage == nil {
		t.Error("expected error_message to be set")
	}
}
```

### Integration Test

**File:** `internal/integration/job_tracking_test.go` (NEW)

```go
package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestJobTracking_FullFlow(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	token := authToken(t, te.Server.URL, "jobtrack@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "Job Track Bot")

	// Create source
	sourceID := createURLSource(t, te.Server.URL, token, chatbotID, "https://example.com")

	// Poll for job status
	var lastStatus string
	var attempts int
	for attempts < 10 {
		time.Sleep(500 * time.Millisecond)
		
		req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/sources/"+sourceID+"/job", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		resp, _ := http.DefaultClient.Do(req)
		var job map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&job)
		resp.Body.Close()
		
		lastStatus = job["status"].(string)
		if lastStatus == "completed" || lastStatus == "failed" {
			break
		}
		attempts++
	}

	if lastStatus != "completed" && lastStatus != "failed" {
		t.Errorf("job did not complete in time, last status: %s", lastStatus)
	}
}
```

---

## Verification Steps

1. **Run unit tests:**
   ```bash
   go test ./internal/processing/... -v -run TestProcessJob
   ```

2. **Run integration tests:**
   ```bash
   go test ./internal/integration/... -v -run TestJobTracking
   ```

3. **Manual verification:**
   ```bash
   # Create a source
   SOURCE_ID=$(curl -s -X POST http://localhost:8080/api/v1/chatbots/{id}/sources \
     -H "Authorization: Bearer $TOKEN" \
     -F "source_type=url" \
     -F "source_url=https://example.com" | jq -r '.id')
   
   # Poll status (should see progress change)
   watch -n 1 'curl -s -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/sources/'$SOURCE_ID'/job | jq'
   ```

---

## Acceptance Criteria

- [x] `EnqueueSource` creates job record
- [x] Worker updates job status as steps progress
- [x] Failed step is tracked in job record
- [x] Job has `started_at` and `completed_at` timestamps
- [x] Progress percentage updates with each step
- [x] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `internal/processing/sources_queue.go` | MODIFY |
| `internal/processing/url_processor.go` | MODIFY |
| `internal/processing/pdf_processor.go` | MODIFY |
| `internal/processing/text_processor.go` | MODIFY |
| `internal/api/handlers/source_create.go` | MODIFY |
| `internal/processing/sources_queue_step_test.go` | CREATE |
| `internal/integration/job_tracking_test.go` | CREATE |
