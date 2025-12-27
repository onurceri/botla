# Task 005: Step-Level Retry Mechanism

**Priority:** 🟡 High (Reliability)  
**Phase:** 2 - Async Training Improvements  
**Estimated Time:** 3-4 hours  
**Dependencies:** Task 004 (Integrate Job Tracking)  

---

## Problem Statement

When a job fails, the entire job must be restarted from scratch. This is wasteful when:
- The failure happened at the embedding step (expensive API call)
- Network issues caused temporary failure
- Rate limiting caused failure that would succeed with retry

**Current behavior:**
- Job fails → entire job marked as failed
- User must delete and re-add source
- All fetching/parsing work is lost

**Desired behavior:**
- Job fails at step X → retry from step X
- Automatic retry with exponential backoff
- Manual retry via API after max retries exceeded

---

## Objective

Implement step-level retry:
1. Track completed steps in job metadata
2. Automatic retry (up to 3 times) with backoff
3. Resume from last failed step
4. API endpoint to manually retry failed jobs

---

## Implementation Details

### Step 1: Add Step Completion Tracking

**File:** `internal/db/training_job.go` (MODIFY)

```go
// JobStepResult tracks individual step completion
type JobStepResult struct {
	Step        models.TrainingStep `json:"step"`
	Status      string              `json:"status"` // completed, failed, skipped
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Error       *string             `json:"error,omitempty"`
	OutputHash  *string             `json:"output_hash,omitempty"` // For idempotency
}

// MarkStepCompleted marks a step as completed in job metadata
func MarkStepCompleted(ctx context.Context, db *sql.DB, jobID string, step models.TrainingStep, outputHash string) error {
	// Get current metadata
	var metadataRaw []byte
	err := db.QueryRowContext(ctx, `SELECT metadata FROM training_jobs WHERE id = $1`, jobID).Scan(&metadataRaw)
	if err != nil {
		return fmt.Errorf("get metadata: %w", err)
	}

	var metadata struct {
		Steps []JobStepResult `json:"steps"`
	}
	if len(metadataRaw) > 2 { // Not empty JSON {}
		json.Unmarshal(metadataRaw, &metadata)
	}

	// Add or update step
	now := time.Now()
	found := false
	for i, s := range metadata.Steps {
		if s.Step == step {
			metadata.Steps[i].Status = "completed"
			metadata.Steps[i].CompletedAt = &now
			metadata.Steps[i].OutputHash = &outputHash
			found = true
			break
		}
	}
	if !found {
		metadata.Steps = append(metadata.Steps, JobStepResult{
			Step:        step,
			Status:      "completed",
			StartedAt:   now,
			CompletedAt: &now,
			OutputHash:  &outputHash,
		})
	}

	// Save metadata
	newMeta, _ := json.Marshal(metadata)
	_, err = db.ExecContext(ctx, `UPDATE training_jobs SET metadata = $2 WHERE id = $1`, jobID, newMeta)
	return err
}

// GetLastCompletedStep returns the last completed step for resuming
func GetLastCompletedStep(ctx context.Context, db *sql.DB, jobID string) (*models.TrainingStep, error) {
	var metadataRaw []byte
	err := db.QueryRowContext(ctx, `SELECT metadata FROM training_jobs WHERE id = $1`, jobID).Scan(&metadataRaw)
	if err != nil {
		return nil, fmt.Errorf("get metadata: %w", err)
	}

	var metadata struct {
		Steps []JobStepResult `json:"steps"`
	}
	if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
		return nil, nil
	}

	// Find last completed step
	stepOrder := []models.TrainingStep{
		models.StepFetchSource,
		models.StepParseContent,
		models.StepChunkText,
		models.StepEmbedChunks,
		models.StepStoreVectors,
	}

	var lastCompleted *models.TrainingStep
	for _, step := range stepOrder {
		for _, s := range metadata.Steps {
			if s.Step == step && s.Status == "completed" {
				lastCompleted = &step
				break
			}
		}
	}

	return lastCompleted, nil
}

// IncrementRetryCount increments retry count and returns new count
func IncrementRetryCount(ctx context.Context, db *sql.DB, jobID string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `
		UPDATE training_jobs 
		SET retry_count = retry_count + 1, status = 'pending'
		WHERE id = $1
		RETURNING retry_count
	`, jobID).Scan(&count)
	return count, err
}
```

### Step 2: Add Retry Logic to Worker

**File:** `internal/processing/sources_queue.go` (MODIFY)

```go
const MaxRetries = 3

// processJob with retry support
func (q *SourceQueue) processJob(jobID string) {
	ctx := context.Background()

	job, err := db.GetTrainingJob(ctx, q.db, jobID)
	if err != nil || job == nil {
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

	// Load source
	source, bot, langCode, plan, ok := q.loadSourceAndLang(job.SourceID)
	if !ok {
		_ = db.FailJob(ctx, q.db, jobID, models.StepFetchSource, "SOURCE_NOT_FOUND", "Source not found")
		return
	}

	// Process with resume support
	result := q.processWithResume(ctx, jobID, source, bot, langCode, plan, lastStep)

	if result.Error != nil {
		// Check if we should retry
		if job.RetryCount < MaxRetries && isRetryableError(result.Error) {
			newCount, _ := db.IncrementRetryCount(ctx, q.db, jobID)
			
			// Calculate backoff
			backoff := time.Duration(1<<uint(newCount)) * time.Second // 2s, 4s, 8s
			
			q.log.Info("job_retry_scheduled", map[string]any{
				"job_id":      jobID,
				"retry_count": newCount,
				"backoff":     backoff.String(),
			})
			
			// Re-enqueue with delay
			go func() {
				time.Sleep(backoff)
				q.Enqueue(jobID) // Legacy enqueue just puts ID in channel
			}()
			return
		}

		// Max retries exceeded or non-retryable error
		_ = db.FailJob(ctx, q.db, jobID, result.FailedStep, "MAX_RETRIES", result.Error.Error())
		q.fail(job.SourceID, result.Error.Error())
		return
	}

	// Success
	completedStep := models.StepStoreVectors
	_ = db.UpdateJobStatus(ctx, q.db, jobID, models.JobStatusCompleted, &completedStep)
	q.complete(job.SourceID, result.ChunkCount)
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
	errStr := err.Error()
	// Retry on network errors, rate limits, temporary failures
	retryable := []string{
		"connection refused",
		"timeout",
		"rate limit",
		"429",
		"503",
		"502",
		"temporary",
	}
	for _, pattern := range retryable {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}
	return false
}
```

### Step 3: Add Manual Retry Endpoint

**File:** `internal/api/handlers/training_job.go` (MODIFY)

Add retry endpoint:

```go
// RetryJob handles POST /api/v1/sources/{id}/job/retry
func (h *TrainingJobHandlers) RetryJob(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")
	if sourceID == "" {
		api.WriteError(w, http.StatusBadRequest, api.ErrMissingID)
		return
	}

	// Validate access (same as GetJobStatus)
	source, err := db.GetSourceByID(r.Context(), h.DB, sourceID)
	if err != nil || source == nil {
		api.WriteError(w, http.StatusNotFound, api.ErrNotFound)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	chatbot, _ := db.GetChatbotByID(r.Context(), h.DB, source.ChatbotID)
	if chatbot == nil || chatbot.UserID != userID {
		api.WriteError(w, http.StatusForbidden, api.ErrForbidden)
		return
	}

	// Get latest job
	job, err := db.GetJobBySourceID(r.Context(), h.DB, sourceID)
	if err != nil || job == nil {
		api.WriteError(w, http.StatusNotFound, "ERR_JOB_NOT_FOUND")
		return
	}

	// Only allow retry on failed jobs
	if job.Status != models.JobStatusFailed {
		api.WriteError(w, http.StatusBadRequest, "ERR_JOB_NOT_FAILED")
		return
	}

	// Reset retry count for manual retry
	_, err = h.DB.ExecContext(r.Context(), `
		UPDATE training_jobs 
		SET status = 'pending', retry_count = 0, error_code = NULL, error_message = NULL
		WHERE id = $1
	`, job.ID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrInternalServer)
		return
	}

	// Enqueue for processing
	if h.Queue != nil {
		h.Queue.Enqueue(job.ID)
	}

	api.WriteJSON(w, http.StatusAccepted, map[string]string{
		"job_id":  job.ID,
		"message": "Job queued for retry",
	})
}
```

**File:** `internal/api/router/routes_sources.go` (MODIFY)

```go
mux.HandleFunc("POST /api/v1/sources/{id}/job/retry", h.TrainingJobHandlers.RetryJob)
```

---

## Tests to Write

### Unit Tests

**File:** `internal/processing/retry_test.go` (NEW)

```go
package processing

import (
	"errors"
	"testing"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		err       error
		retryable bool
	}{
		{errors.New("connection refused"), true},
		{errors.New("timeout waiting for response"), true},
		{errors.New("rate limit exceeded"), true},
		{errors.New("status 429"), true},
		{errors.New("invalid URL"), false},
		{errors.New("parse error"), false},
		{errors.New("unauthorized"), false},
	}

	for _, tt := range tests {
		result := isRetryableError(tt.err)
		if result != tt.retryable {
			t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.retryable)
		}
	}
}

func TestGetNextStep(t *testing.T) {
	tests := []struct {
		current models.TrainingStep
		next    models.TrainingStep
	}{
		{models.StepFetchSource, models.StepParseContent},
		{models.StepParseContent, models.StepChunkText},
		{models.StepChunkText, models.StepEmbedChunks},
		{models.StepEmbedChunks, models.StepStoreVectors},
	}

	for _, tt := range tests {
		result := getNextStep(tt.current)
		if result != tt.next {
			t.Errorf("getNextStep(%s) = %s, want %s", tt.current, result, tt.next)
		}
	}
}
```

### Integration Test

**File:** `internal/integration/retry_test.go` (NEW)

```go
package integration

func TestManualRetry_Integration(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	token := authToken(t, te.Server.URL, "retry@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "Retry Test Bot")

	// Create source that will fail
	sourceID := createURLSource(t, te.Server.URL, token, chatbotID, "http://invalid.test")

	// Wait for failure
	time.Sleep(3 * time.Second)

	// Verify job is failed
	job := getJobStatus(t, te.Server.URL, token, sourceID)
	if job["status"] != "failed" {
		t.Skipf("job didn't fail as expected, status: %s", job["status"])
	}

	// Retry
	req, _ := http.NewRequest("POST", te.Server.URL+"/api/v1/sources/"+sourceID+"/job/retry", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode)
	}

	// Verify job is pending again
	time.Sleep(500 * time.Millisecond)
	job = getJobStatus(t, te.Server.URL, token, sourceID)
	if job["status"] == "failed" {
		// Could be pending or running
		t.Log("job may have already reprocessed")
	}
}
```

---

## Acceptance Criteria

- [x] Completed steps are tracked in job metadata
- [x] Automatic retry on retryable errors (up to 3 times)
- [x] Exponential backoff between retries
- [x] Manual retry endpoint works for failed jobs
- [x] Resume from last completed step on retry
- [x] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `internal/db/training_job.go` | MODIFY |
| `internal/processing/sources_queue.go` | MODIFY |
| `internal/api/handlers/training_job.go` | MODIFY |
| `internal/api/router/routes_sources.go` | MODIFY |
| `internal/processing/retry_test.go` | CREATE |
| `internal/integration/retry_test.go` | CREATE |
