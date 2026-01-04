// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresTrainingJobRepo implements TrainingJobRepository using PostgreSQL.
type PostgresTrainingJobRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresTrainingJobRepo implements TrainingJobRepository.
var _ TrainingJobRepository = (*PostgresTrainingJobRepo)(nil)

// NewPostgresTrainingJobRepo creates a new PostgresTrainingJobRepo instance.
func NewPostgresTrainingJobRepo(pool *sql.DB) *PostgresTrainingJobRepo {
	return &PostgresTrainingJobRepo{pool: pool}
}

// scanTrainingJob scans a training job from rows.
func (r *PostgresTrainingJobRepo) scanTrainingJob(rows *sql.Rows) (*models.TrainingJob, error) {
	var job models.TrainingJob
	var currentStep, errorCode, errorMessage, failedStep sql.NullString
	var startedAt, completedAt sql.NullTime

	if err := rows.Scan(
		&job.ID, &job.SourceID, &job.ChatbotID, &job.Status, &currentStep, &job.ProgressPercent,
		&errorCode, &errorMessage, &failedStep, &job.RetryCount,
		&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt,
	); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan training job")
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

// GetBySourceID retrieves the latest job for a source.
func (r *PostgresTrainingJobRepo) GetBySourceID(ctx context.Context, sourceID string) (*models.TrainingJob, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs 
		WHERE source_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`, sourceID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query job by source")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return r.scanTrainingJob(rows)
}

// GetByID retrieves a training job by its unique identifier.
func (r *PostgresTrainingJobRepo) GetByID(ctx context.Context, id string) (*models.TrainingJob, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs WHERE id = $1
	`, id)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query training job")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return r.scanTrainingJob(rows)
}

// GetByChatbotID retrieves all jobs for a chatbot.
func (r *PostgresTrainingJobRepo) GetByChatbotID(ctx context.Context, chatbotID string, limit int) ([]*models.TrainingJob, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs 
		WHERE chatbot_id = $1 
		ORDER BY created_at DESC
		LIMIT $2
	`, chatbotID, limit)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query jobs by chatbot")
	}
	defer rows.Close()

	var jobs []*models.TrainingJob
	for rows.Next() {
		job, err := r.scanTrainingJob(rows)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterate training jobs")
	}

	return jobs, nil
}

// Create creates a new training job for a data source.
func (r *PostgresTrainingJobRepo) Create(ctx context.Context, sourceID, chatbotID string) (*models.TrainingJob, error) {
	var job models.TrainingJob
	err := r.pool.QueryRowContext(ctx, `
		INSERT INTO training_jobs (source_id, chatbot_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, source_id, chatbot_id, status, progress_percent, retry_count, created_at, updated_at
	`, sourceID, chatbotID).Scan(
		&job.ID, &job.SourceID, &job.ChatbotID, &job.Status,
		&job.ProgressPercent, &job.RetryCount, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create training job")
	}
	return &job, nil
}

// UpdateJobStatus updates the status and step of a training job.
func (r *PostgresTrainingJobRepo) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, step *models.TrainingStep) error {
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

	_, err := r.pool.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = $2, current_step = $3, progress_percent = $4,
		    started_at = COALESCE($5, started_at),
		    completed_at = COALESCE($6, completed_at)
		WHERE id = $1
	`, id, status, stepStr, progress, startedAt, completedAt)
	if err != nil {
		return pkgerrors.Wrapf(err, "update job status")
	}
	return nil
}

// ResetForRetry resets a job to pending status for retry.
func (r *PostgresTrainingJobRepo) ResetForRetry(ctx context.Context, id string) error {
	_, err := r.pool.ExecContext(ctx, `
		UPDATE training_jobs
		SET status = 'pending', retry_count = retry_count + 1,
		    error_code = NULL, error_message = NULL, failed_step = NULL,
		    started_at = NULL, completed_at = NULL
		WHERE id = $1
	`, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "reset job for retry")
	}
	return nil
}

// IncrementRetryCount increments the retry count for a job and resets it to pending.
func (r *PostgresTrainingJobRepo) IncrementRetryCount(ctx context.Context, id string) (int, error) {
	var count int
	err := r.pool.QueryRowContext(ctx, `
		UPDATE training_jobs
		SET retry_count = retry_count + 1, status = 'pending',
		    error_code = NULL, error_message = NULL, failed_step = NULL,
		    started_at = NULL, completed_at = NULL
		WHERE id = $1
		RETURNING retry_count
	`, id).Scan(&count)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "increment retry count")
	}
	return count, nil
}

// GetPendingJobs retrieves jobs in pending status for recovery.
func (r *PostgresTrainingJobRepo) GetPendingJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs 
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get pending jobs")
	}
	defer rows.Close()

	var jobs []*models.TrainingJob
	for rows.Next() {
		job, err := r.scanTrainingJob(rows)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterate pending jobs")
	}
	return jobs, nil
}

// JobStepResult tracks individual step completion.
type JobStepResult struct {
	Step        models.TrainingStep `json:"step"`
	Status      string              `json:"status"` // completed, failed, skipped
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Error       *string             `json:"error,omitempty"`
	OutputHash  *string             `json:"output_hash,omitempty"` // For idempotency
}

// MarkStepCompleted marks a step as completed in job metadata.
func (r *PostgresTrainingJobRepo) MarkStepCompleted(ctx context.Context, jobID string, step models.TrainingStep, outputHash string) error {
	// Get current metadata
	var metadataRaw []byte
	err := r.pool.QueryRowContext(ctx, `SELECT metadata FROM training_jobs WHERE id = $1`, jobID).Scan(&metadataRaw)
	if err != nil {
		return pkgerrors.Wrapf(err, "get metadata")
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
	_, err = r.pool.ExecContext(ctx, `UPDATE training_jobs SET metadata = $2 WHERE id = $1`, jobID, newMeta)
	if err != nil {
		return pkgerrors.Wrapf(err, "update metadata")
	}
	return nil
}

// GetLastCompletedStep returns the last completed step for resuming.
func (r *PostgresTrainingJobRepo) GetLastCompletedStep(ctx context.Context, jobID string) (*models.TrainingStep, error) {
	var metadataRaw []byte
	err := r.pool.QueryRowContext(ctx, `SELECT metadata FROM training_jobs WHERE id = $1`, jobID).Scan(&metadataRaw)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get metadata")
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

// Fail marks a job as failed with error details.
func (r *PostgresTrainingJobRepo) Fail(ctx context.Context, id string, step models.TrainingStep, errCode, errMsg string) error {
	_, err := r.pool.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = 'failed', failed_step = $2, error_code = $3, error_message = $4,
		    completed_at = NOW()
		WHERE id = $1
	`, id, step, errCode, errMsg)
	if err != nil {
		return pkgerrors.Wrapf(err, "fail job")
	}
	return nil
}

// Complete marks a job as completed.
func (r *PostgresTrainingJobRepo) Complete(ctx context.Context, id string) error {
	_, err := r.pool.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = 'completed', progress_percent = 100, completed_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "complete job")
	}
	return nil
}

// Cancel marks a job as cancelled.
func (r *PostgresTrainingJobRepo) Cancel(ctx context.Context, id string) error {
	_, err := r.pool.ExecContext(ctx, `
		UPDATE training_jobs 
		SET status = 'cancelled', completed_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "cancel job")
	}
	return nil
}

// GetRetryableJobs retrieves failed jobs that can be retried.
func (r *PostgresTrainingJobRepo) GetRetryableJobs(ctx context.Context, maxRetries, limit int) ([]*models.TrainingJob, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs 
		WHERE status = 'failed' AND retry_count < $1
		ORDER BY created_at ASC
		LIMIT $2
	`, maxRetries, limit)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get retryable jobs")
	}
	defer rows.Close()

	var jobs []*models.TrainingJob
	for rows.Next() {
		job, err := r.scanTrainingJob(rows)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterate retryable jobs")
	}
	return jobs, nil
}

// GetRunningJobs retrieves jobs currently running.
func (r *PostgresTrainingJobRepo) GetRunningJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, source_id, chatbot_id, status, current_step, progress_percent,
		       error_code, error_message, failed_step, retry_count,
		       created_at, started_at, completed_at, updated_at
		FROM training_jobs 
		WHERE status = 'running'
		ORDER BY started_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get running jobs")
	}
	defer rows.Close()

	var jobs []*models.TrainingJob
	for rows.Next() {
		job, err := r.scanTrainingJob(rows)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterate running jobs")
	}
	return jobs, nil
}
