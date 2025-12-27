# Task 002: Job State Table Migration

**Priority:** 🔴 Critical (Foundation)  
**Phase:** 1 - Observability Foundation  
**Estimated Time:** 2-3 hours  
**Dependencies:** None  

---

## Problem Statement

Currently, the training job state is not persisted. If the server restarts during job processing:
- In-flight jobs are lost
- No way to track job history
- No visibility into job progress from frontend

**Evidence from codebase:**
- `sources_queue.go` uses in-memory channels
- `recoverPendingSources()` only recovers `pending` status from `data_sources` table
- No dedicated job tracking table exists

---

## Objective

Create a `training_jobs` table to:
1. Persist all job states (PENDING, RUNNING, COMPLETED, FAILED)
2. Track current step and progress
3. Enable job history queries
4. Support future step-level retry

---

## Implementation Details

### Step 1: Create Migration File

**File:** `db/migrations/YYYYMMDDHHMMSS_create_training_jobs.up.sql` (NEW)

```sql
-- Training jobs table for async job tracking
CREATE TABLE training_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    
    -- Job status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    current_step VARCHAR(50),
    progress_percent INTEGER DEFAULT 0,
    
    -- Error tracking
    error_code VARCHAR(100),
    error_message TEXT,
    failed_step VARCHAR(50),
    retry_count INTEGER DEFAULT 0,
    
    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    CONSTRAINT valid_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'))
);

-- Indexes for common queries
CREATE INDEX idx_training_jobs_source_id ON training_jobs(source_id);
CREATE INDEX idx_training_jobs_chatbot_id ON training_jobs(chatbot_id);
CREATE INDEX idx_training_jobs_status ON training_jobs(status);
CREATE INDEX idx_training_jobs_created_at ON training_jobs(created_at DESC);

-- Index for finding jobs to retry
CREATE INDEX idx_training_jobs_retry ON training_jobs(status, retry_count) 
    WHERE status = 'failed' AND retry_count < 3;

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_training_jobs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER training_jobs_updated_at
    BEFORE UPDATE ON training_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_training_jobs_updated_at();

-- Add comment
COMMENT ON TABLE training_jobs IS 'Tracks async training job execution for data sources';
```

**File:** `db/migrations/YYYYMMDDHHMMSS_create_training_jobs.down.sql` (NEW)

```sql
DROP TRIGGER IF EXISTS training_jobs_updated_at ON training_jobs;
DROP FUNCTION IF EXISTS update_training_jobs_updated_at();
DROP TABLE IF EXISTS training_jobs;
```

### Step 2: Create Job Model

**File:** `internal/models/training_job.go` (NEW)

```go
package models

import (
	"encoding/json"
	"time"
)

// JobStatus represents the status of a training job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// TrainingStep represents a step in the training pipeline
type TrainingStep string

const (
	StepFetchSource   TrainingStep = "fetch_source"
	StepParseContent  TrainingStep = "parse_content"
	StepChunkText     TrainingStep = "chunk_text"
	StepEmbedChunks   TrainingStep = "embed_chunks"
	StepStoreVectors  TrainingStep = "store_vectors"
)

// TrainingJob represents a training job
type TrainingJob struct {
	ID              string           `json:"id"`
	SourceID        string           `json:"source_id"`
	ChatbotID       string           `json:"chatbot_id"`
	Status          JobStatus        `json:"status"`
	CurrentStep     *TrainingStep    `json:"current_step,omitempty"`
	ProgressPercent int              `json:"progress_percent"`
	ErrorCode       *string          `json:"error_code,omitempty"`
	ErrorMessage    *string          `json:"error_message,omitempty"`
	FailedStep      *TrainingStep    `json:"failed_step,omitempty"`
	RetryCount      int              `json:"retry_count"`
	CreatedAt       time.Time        `json:"created_at"`
	StartedAt       *time.Time       `json:"started_at,omitempty"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Metadata        json.RawMessage  `json:"metadata,omitempty"`
}

// StepProgress maps steps to progress percentages
var StepProgress = map[TrainingStep]int{
	StepFetchSource:  10,
	StepParseContent: 30,
	StepChunkText:    50,
	StepEmbedChunks:  80,
	StepStoreVectors: 100,
}
```

### Step 3: Create Job Database Operations

**File:** `internal/db/training_job.go` (NEW)

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

// CreateTrainingJob creates a new training job
func CreateTrainingJob(ctx context.Context, db *sql.DB, sourceID, chatbotID string) (*models.TrainingJob, error) {
	var job models.TrainingJob
	err := db.QueryRowContext(ctx, `
		INSERT INTO training_jobs (source_id, chatbot_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, source_id, chatbot_id, status, progress_percent, retry_count, created_at, updated_at
	`, sourceID, chatbotID).Scan(
		&job.ID, &job.SourceID, &job.ChatbotID, &job.Status,
		&job.ProgressPercent, &job.RetryCount, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create training job: %w", err)
	}
	return &job, nil
}

// GetTrainingJob retrieves a training job by ID
func GetTrainingJob(ctx context.Context, db *sql.DB, id string) (*models.TrainingJob, error) {
	var job models.TrainingJob
	var currentStep, errorCode, errorMessage, failedStep sql.NullString
	var startedAt, completedAt sql.NullTime

	err := db.QueryRowContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at, metadata
		FROM training_jobs WHERE id = $1
	`, id).Scan(
		&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &currentStep, &job.ProgressPercent,
		&errorCode, &errorMessage, &failedStep, &job.RetryCount,
		&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt, &job.Metadata,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get training job: %w", err)
	}

	if currentStep.Valid {
		step := models.TrainingStep(currentStep.String)
		job.CurrentStep = &step
	}
	if errorCode.Valid {
		job.ErrorCode = &errorCode.String
	}
	if errorMessage.Valid {
		job.ErrorMessage = &errorMessage.String
	}
	if failedStep.Valid {
		step := models.TrainingStep(failedStep.String)
		job.FailedStep = &step
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// UpdateJobStatus updates the status and step of a training job
func UpdateJobStatus(ctx context.Context, db *sql.DB, id string, status models.JobStatus, step *models.TrainingStep) error {
	var stepStr *string
	var progress int
	if step != nil {
		s := string(*step)
		stepStr = &s
		progress = models.StepProgress[*step]
	}

	var startedAt interface{}
	var completedAt interface{}
	if status == models.JobStatusRunning {
		startedAt = time.Now()
	}
	if status == models.JobStatusCompleted || status == models.JobStatusFailed {
		completedAt = time.Now()
	}

	_, err := db.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = $2, current_step = $3, progress_percent = $4,
		    started_at = COALESCE($5, started_at),
		    completed_at = COALESCE($6, completed_at)
		WHERE id = $1
	`, id, status, stepStr, progress, startedAt, completedAt)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}
	return nil
}

// FailJob marks a job as failed with error details
func FailJob(ctx context.Context, db *sql.DB, id string, step models.TrainingStep, errCode, errMsg string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = 'failed', failed_step = $2, error_code = $3, error_message = $4,
		    completed_at = NOW()
		WHERE id = $1
	`, id, step, errCode, errMsg)
	if err != nil {
		return fmt.Errorf("fail job: %w", err)
	}
	return nil
}

// GetJobBySourceID retrieves the latest job for a source
func GetJobBySourceID(ctx context.Context, db *sql.DB, sourceID string) (*models.TrainingJob, error) {
	var job models.TrainingJob
	err := db.QueryRowContext(ctx, `
		SELECT id, source_id, chatbot_id, status, progress_percent, created_at, updated_at
		FROM training_jobs 
		WHERE source_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`, sourceID).Scan(
		&job.ID, &job.SourceID, &job.ChatbotID, &job.Status,
		&job.ProgressPercent, &job.CreatedAt, &job.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job by source: %w", err)
	}
	return &job, nil
}

// GetPendingJobs retrieves jobs in pending status for recovery
func GetPendingJobs(ctx context.Context, db *sql.DB, limit int) ([]*models.TrainingJob, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, created_at
		FROM training_jobs 
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.TrainingJob
	for rows.Next() {
		var job models.TrainingJob
		if err := rows.Scan(&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &job.CreatedAt); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}
	return jobs, rows.Err()
}
```

---

## Tests to Write

### Unit Tests

**File:** `internal/db/training_job_test.go` (NEW)

```go
package db

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
)

func TestCreateTrainingJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test source and chatbot first
	sourceID := createTestSource(t, db)
	chatbotID := createTestChatbot(t, db)

	job, err := CreateTrainingJob(context.Background(), db, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("CreateTrainingJob failed: %v", err)
	}

	if job.ID == "" {
		t.Error("expected job ID, got empty")
	}
	if job.Status != models.JobStatusPending {
		t.Errorf("expected pending status, got %s", job.Status)
	}
	if job.ProgressPercent != 0 {
		t.Errorf("expected 0 progress, got %d", job.ProgressPercent)
	}
}

func TestUpdateJobStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	sourceID := createTestSource(t, db)
	chatbotID := createTestChatbot(t, db)
	job, _ := CreateTrainingJob(context.Background(), db, sourceID, chatbotID)

	// Update to running
	step := models.StepFetchSource
	err := UpdateJobStatus(context.Background(), db, job.ID, models.JobStatusRunning, &step)
	if err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}

	// Verify
	updated, _ := GetTrainingJob(context.Background(), db, job.ID)
	if updated.Status != models.JobStatusRunning {
		t.Errorf("expected running, got %s", updated.Status)
	}
	if updated.StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}

func TestFailJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	sourceID := createTestSource(t, db)
	chatbotID := createTestChatbot(t, db)
	job, _ := CreateTrainingJob(context.Background(), db, sourceID, chatbotID)

	err := FailJob(context.Background(), db, job.ID, models.StepEmbedChunks, "EMBED_ERROR", "OpenAI API rate limit")
	if err != nil {
		t.Fatalf("FailJob failed: %v", err)
	}

	updated, _ := GetTrainingJob(context.Background(), db, job.ID)
	if updated.Status != models.JobStatusFailed {
		t.Errorf("expected failed, got %s", updated.Status)
	}
	if *updated.FailedStep != models.StepEmbedChunks {
		t.Errorf("expected embed_chunks, got %s", *updated.FailedStep)
	}
	if *updated.ErrorCode != "EMBED_ERROR" {
		t.Errorf("expected EMBED_ERROR, got %s", *updated.ErrorCode)
	}
}
```

---

## Verification Steps

1. **Run migrations:**
   ```bash
   make migrate-up
   ```

2. **Verify table created:**
   ```bash
   make psql
   \d training_jobs
   ```

3. **Run unit tests:**
   ```bash
   go test ./internal/db/... -v -run TestTrainingJob
   ```

4. **Verify rollback works:**
   ```bash
   make migrate-down
   make migrate-up
   ```

---

## Acceptance Criteria

- [ ] Migration creates `training_jobs` table successfully
- [ ] Rollback migration drops table cleanly
- [ ] `TrainingJob` model with all fields
- [ ] CRUD operations work correctly
- [ ] All unit tests pass
- [ ] Indexes verified with `\d training_jobs`

---

## Files Changed

| File | Action |
|------|--------|
| `db/migrations/YYYYMMDDHHMMSS_create_training_jobs.up.sql` | CREATE |
| `db/migrations/YYYYMMDDHHMMSS_create_training_jobs.down.sql` | CREATE |
| `internal/models/training_job.go` | CREATE |
| `internal/db/training_job.go` | CREATE |
| `internal/db/training_job_test.go` | CREATE |
