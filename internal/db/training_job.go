package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

// CreateTrainingJob creates a new training job for a data source
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

// CompleteJob marks a job as completed
func CompleteJob(ctx context.Context, db *sql.DB, id string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = 'completed', progress_percent = 100, completed_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("complete job: %w", err)
	}
	return nil
}

// CancelJob marks a job as cancelled
func CancelJob(ctx context.Context, db *sql.DB, id string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = 'cancelled', completed_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("cancel job: %w", err)
	}
	return nil
}

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
		if uerr := json.Unmarshal(metadataRaw, &metadata); uerr != nil {
			// If invalid JSON, start fresh
			metadata.Steps = []JobStepResult{}
		}
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
	if err != nil {
		return fmt.Errorf("update metadata: %w", err)
	}
	return nil
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
	if len(metadataRaw) <= 2 {
		return nil, nil
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
				copyStep := step
				lastCompleted = &copyStep
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
		SET retry_count = retry_count + 1, status = 'pending',
		    error_code = NULL, error_message = NULL, failed_step = NULL,
		    started_at = NULL, completed_at = NULL
		WHERE id = $1
		RETURNING retry_count
	`, jobID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("increment retry count: %w", err)
	}
	return count, nil
}

// GetJobBySourceID retrieves the latest job for a source
func GetJobBySourceID(ctx context.Context, db *sql.DB, sourceID string) (*models.TrainingJob, error) {
	var job models.TrainingJob
	var currentStep, errorCode, errorMessage, failedStep sql.NullString
	var startedAt, completedAt sql.NullTime

	err := db.QueryRowContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs 
		WHERE source_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`, sourceID).Scan(
		&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &currentStep, &job.ProgressPercent,
		&errorCode, &errorMessage, &failedStep, &job.RetryCount,
		&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job by source: %w", err)
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

// GetJobsByChatbotID retrieves all jobs for a chatbot ordered by creation date (most recent first)
func GetJobsByChatbotID(ctx context.Context, db *sql.DB, chatbotID string, limit int) ([]*models.TrainingJob, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs 
		WHERE chatbot_id = $1 
		ORDER BY created_at DESC
		LIMIT $2
	`, chatbotID, limit)
	if err != nil {
		return nil, fmt.Errorf("get jobs by chatbot: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []*models.TrainingJob
	for rows.Next() {
		var job models.TrainingJob
		var currentStep, errorCode, errorMessage, failedStep sql.NullString
		var startedAt, completedAt sql.NullTime

		if err := rows.Scan(
			&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &currentStep, &job.ProgressPercent,
			&errorCode, &errorMessage, &failedStep, &job.RetryCount,
			&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt,
		); err != nil {
			continue
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

		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error for jobs by chatbot: %w", err)
	}
	return jobs, nil
}

// GetPendingJobs retrieves jobs in pending status for recovery
func GetPendingJobs(ctx context.Context, db *sql.DB, limit int) ([]*models.TrainingJob, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, created_at, updated_at
		FROM training_jobs 
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []*models.TrainingJob
	for rows.Next() {
		var job models.TrainingJob
		if err := rows.Scan(&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &job.CreatedAt, &job.UpdatedAt); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error for pending jobs: %w", err)
	}
	return jobs, nil
}

// GetRetryableJobs retrieves failed jobs that can be retried (retry_count < maxRetries)
func GetRetryableJobs(ctx context.Context, db *sql.DB, maxRetries, limit int) ([]*models.TrainingJob, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, failed_step, retry_count, created_at, updated_at
		FROM training_jobs 
		WHERE status = 'failed' AND retry_count < $1
		ORDER BY created_at ASC
		LIMIT $2
	`, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("get retryable jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []*models.TrainingJob
	for rows.Next() {
		var job models.TrainingJob
		var failedStep sql.NullString
		if err := rows.Scan(&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &failedStep, &job.RetryCount, &job.CreatedAt, &job.UpdatedAt); err != nil {
			continue
		}
		if failedStep.Valid {
			step := models.TrainingStep(failedStep.String)
			job.FailedStep = &step
		}
		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error for retryable jobs: %w", err)
	}
	return jobs, nil
}

// GetRunningJobs retrieves jobs currently running (for stale job detection)
func GetRunningJobs(ctx context.Context, db *sql.DB, limit int) ([]*models.TrainingJob, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, started_at, created_at, updated_at
		FROM training_jobs 
		WHERE status = 'running'
		ORDER BY started_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("get running jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []*models.TrainingJob
	for rows.Next() {
		var job models.TrainingJob
		var currentStep sql.NullString
		var startedAt sql.NullTime
		if err := rows.Scan(&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &currentStep, &startedAt, &job.CreatedAt, &job.UpdatedAt); err != nil {
			continue
		}
		if currentStep.Valid {
			step := models.TrainingStep(currentStep.String)
			job.CurrentStep = &step
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error for running jobs: %w", err)
	}
	return jobs, nil
}

// UpdateJobMetadata updates the metadata field for a training job
func UpdateJobMetadata(ctx context.Context, db *sql.DB, id string, metadata []byte) error {
	_, err := db.ExecContext(ctx, `
		UPDATE training_jobs SET metadata = $2 WHERE id = $1
	`, id, metadata)
	if err != nil {
		return fmt.Errorf("update job metadata: %w", err)
	}
	return nil
}
