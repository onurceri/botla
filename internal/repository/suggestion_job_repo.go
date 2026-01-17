// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresSuggestionJobRepo implements SuggestionJobRepository using PostgreSQL.
type PostgresSuggestionJobRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresSuggestionJobRepo implements SuggestionJobRepository.
var _ SuggestionJobRepository = (*PostgresSuggestionJobRepo)(nil)

// NewPostgresSuggestionJobRepo creates a new PostgresSuggestionJobRepo instance.
func NewPostgresSuggestionJobRepo(pool *sql.DB) *PostgresSuggestionJobRepo {
	return &PostgresSuggestionJobRepo{pool: pool}
}

// scanSuggestionJob scans a suggestion job from rows.
func (r *PostgresSuggestionJobRepo) scanSuggestionJob(rows *sql.Row) (*models.SuggestionJob, error) {
	var job models.SuggestionJob
	var status sql.NullString
	var errorMessage sql.NullString
	var suggestions pq.StringArray
	var startedAt, completedAt sql.NullTime

	err := rows.Scan(
		&job.ID, &job.ChatbotID, &status, &errorMessage, &suggestions,
		&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "scan suggestion job")
	}

	if status.Valid {
		job.Status = models.SuggestionJobStatus(status.String)
	}
	if errorMessage.Valid {
		job.ErrorMessage = &errorMessage.String
	}
	if len(suggestions) > 0 {
		job.SuggestedQuestions = suggestions
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// scanSuggestionJobFromRows scans a suggestion job from rows (for queries with multiple results).
func (r *PostgresSuggestionJobRepo) scanSuggestionJobFromRows(rows *sql.Rows) (*models.SuggestionJob, error) {
	var job models.SuggestionJob
	var status sql.NullString
	var errorMessage sql.NullString
	var suggestions pq.StringArray
	var startedAt, completedAt sql.NullTime

	err := rows.Scan(
		&job.ID, &job.ChatbotID, &status, &errorMessage, &suggestions,
		&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "scan suggestion job")
	}

	if status.Valid {
		job.Status = models.SuggestionJobStatus(status.String)
	}
	if errorMessage.Valid {
		job.ErrorMessage = &errorMessage.String
	}
	if len(suggestions) > 0 {
		job.SuggestedQuestions = suggestions
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// Create creates a new suggestion job for a chatbot.
func (r *PostgresSuggestionJobRepo) Create(ctx context.Context, chatbotID string) (*models.SuggestionJob, error) {
	row := r.pool.QueryRowContext(ctx, `
		INSERT INTO suggestion_jobs (chatbot_id, status)
		VALUES ($1, 'pending')
		RETURNING id, chatbot_id, status, error_message, suggested_questions,
		          created_at, started_at, completed_at, updated_at
	`, chatbotID)
	return r.scanSuggestionJob(row)
}

// GetByID retrieves a suggestion job by its unique identifier.
func (r *PostgresSuggestionJobRepo) GetByID(ctx context.Context, id string) (*models.SuggestionJob, error) {
	row := r.pool.QueryRowContext(ctx, `
		SELECT id, chatbot_id, status, error_message, suggested_questions,
		       created_at, started_at, completed_at, updated_at
		FROM suggestion_jobs WHERE id = $1
	`, id)
	return r.scanSuggestionJob(row)
}

// GetLatestForChatbot retrieves the most recent suggestion job for a chatbot.
func (r *PostgresSuggestionJobRepo) GetLatestForChatbot(ctx context.Context, chatbotID string) (*models.SuggestionJob, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, chatbot_id, status, error_message, suggested_questions,
		       created_at, started_at, completed_at, updated_at
		FROM suggestion_jobs
		WHERE chatbot_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, chatbotID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query latest suggestion job")
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return nil, nil
	}

	return r.scanSuggestionJobFromRows(rows)
}

// UpdateStatus updates the status of a suggestion job.
func (r *PostgresSuggestionJobRepo) UpdateStatus(ctx context.Context, id string, status models.SuggestionJobStatus) error {
	var startedAt, completedAt interface{}
	if status == models.SuggestionJobStatusRunning {
		startedAt = time.Now()
	}
	if status == models.SuggestionJobStatusCompleted || status == models.SuggestionJobStatusFailed {
		completedAt = time.Now()
	}

	_, err := r.pool.ExecContext(ctx, `
		UPDATE suggestion_jobs
		SET status = $2,
		    started_at = COALESCE($3, started_at),
		    completed_at = COALESCE($4, completed_at)
		WHERE id = $1
	`, id, status, startedAt, completedAt)
	if err != nil {
		return pkgerrors.Wrapf(err, "update suggestion job status")
	}
	return nil
}

// Complete marks a suggestion job as completed with suggestions.
func (r *PostgresSuggestionJobRepo) Complete(ctx context.Context, id string, suggestions []string) error {
	_, err := r.pool.ExecContext(ctx, `
		UPDATE suggestion_jobs
		SET status = 'completed', suggested_questions = $2, completed_at = NOW()
		WHERE id = $1
	`, id, pq.Array(suggestions))
	if err != nil {
		return pkgerrors.Wrapf(err, "complete suggestion job")
	}
	return nil
}

// Fail marks a suggestion job as failed with an error message.
func (r *PostgresSuggestionJobRepo) Fail(ctx context.Context, id string, errMsg string) error {
	_, err := r.pool.ExecContext(ctx, `
		UPDATE suggestion_jobs
		SET status = 'failed', error_message = $2, completed_at = NOW()
		WHERE id = $1
	`, id, errMsg)
	if err != nil {
		return pkgerrors.Wrapf(err, "fail suggestion job")
	}
	return nil
}
