package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

func CreateSuggestionJob(ctx context.Context, db *sql.DB, chatbotID string) (*models.SuggestionJob, error) {
	var job models.SuggestionJob
	var status sql.NullString
	err := db.QueryRowContext(ctx, `
		INSERT INTO suggestion_jobs (chatbot_id, status)
		VALUES ($1, 'pending')
		RETURNING id, chatbot_id, status, created_at, updated_at
	`, chatbotID).Scan(
		&job.ID, &job.ChatbotID, &status, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create suggestion job")
	}
	if status.Valid {
		job.Status = models.SuggestionJobStatus(status.String)
	}
	return &job, nil
}

func GetSuggestionJob(ctx context.Context, db *sql.DB, id string) (*models.SuggestionJob, error) {
	var job models.SuggestionJob
	var status sql.NullString
	var errorMessage sql.NullString
	var suggestions pq.StringArray
	var startedAt, completedAt sql.NullTime

	err := db.QueryRowContext(ctx, `
		SELECT id, chatbot_id, status, error_message, suggested_questions,
		       created_at, started_at, completed_at, updated_at
		FROM suggestion_jobs WHERE id = $1
	`, id).Scan(
		&job.ID, &job.ChatbotID, &status, &errorMessage, &suggestions,
		&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get suggestion job")
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

func UpdateSuggestionJobStatus(ctx context.Context, db *sql.DB, id string, status models.SuggestionJobStatus) error {
	var startedAt, completedAt interface{}
	if status == models.SuggestionJobStatusRunning {
		startedAt = time.Now()
	}
	if status == models.SuggestionJobStatusCompleted || status == models.SuggestionJobStatusFailed {
		completedAt = time.Now()
	}

	_, err := db.ExecContext(ctx, `
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

func CompleteSuggestionJob(ctx context.Context, db *sql.DB, id string, suggestions []string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE suggestion_jobs
		SET status = 'completed', suggested_questions = $2, completed_at = NOW()
		WHERE id = $1
	`, id, pq.Array(suggestions))
	if err != nil {
		return pkgerrors.Wrapf(err, "complete suggestion job")
	}
	return nil
}

func FailSuggestionJob(ctx context.Context, db *sql.DB, id string, errMsg string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE suggestion_jobs
		SET status = 'failed', error_message = $2, completed_at = NOW()
		WHERE id = $1
	`, id, errMsg)
	if err != nil {
		return pkgerrors.Wrapf(err, "fail suggestion job")
	}
	return nil
}

func GetLatestSuggestionJobForChatbot(ctx context.Context, db *sql.DB, chatbotID string) (*models.SuggestionJob, error) {
	var job models.SuggestionJob
	var status sql.NullString
	var errorMessage sql.NullString
	var suggestions pq.StringArray
	var startedAt, completedAt sql.NullTime

	err := db.QueryRowContext(ctx, `
		SELECT id, chatbot_id, status, error_message, suggested_questions,
		       created_at, started_at, completed_at, updated_at
		FROM suggestion_jobs
		WHERE chatbot_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, chatbotID).Scan(
		&job.ID, &job.ChatbotID, &status, &errorMessage, &suggestions,
		&job.CreatedAt, &startedAt, &completedAt, &job.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get latest suggestion job for chatbot")
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
