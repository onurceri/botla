package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

func CreateDataSource(ctx context.Context, pool *sql.DB, s *models.DataSource) (string, error) {
	var id string
	err := pool.QueryRowContext(
		ctx,
		`INSERT INTO data_sources (
            chatbot_id, source_type, source_url, file_path, original_filename,
            status, error_message, chunk_count
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		s.ChatbotID, s.SourceType, s.SourceURL, s.FilePath, s.OriginalFilename,
		s.Status, s.ErrorMessage, s.ChunkCount,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ListSourcesByChatbotID(ctx context.Context, pool *sql.DB, chatbotID string) ([]models.DataSource, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT id, chatbot_id, source_type, source_url, file_path, original_filename,
               status, error_message, chunk_count, processed_at, created_at
        FROM data_sources
        WHERE chatbot_id=$1
        ORDER BY created_at DESC`, chatbotID)
	if err != nil {
		return nil, err
	}
    defer func() { _ = rows.Close() }()
	out := []models.DataSource{}
	for rows.Next() {
		var d models.DataSource
		if err := rows.Scan(
			&d.ID, &d.ChatbotID, &d.SourceType, &d.SourceURL, &d.FilePath, &d.OriginalFilename,
			&d.Status, &d.ErrorMessage, &d.ChunkCount, &d.ProcessedAt, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func GetSourceByID(ctx context.Context, pool *sql.DB, id string) (*models.DataSource, error) {
	var d models.DataSource
	err := pool.QueryRowContext(ctx, `
        SELECT id, chatbot_id, source_type, source_url, file_path, original_filename,
               status, error_message, chunk_count, processed_at, created_at
        FROM data_sources WHERE id=$1`, id).Scan(
		&d.ID, &d.ChatbotID, &d.SourceType, &d.SourceURL, &d.FilePath, &d.OriginalFilename,
		&d.Status, &d.ErrorMessage, &d.ChunkCount, &d.ProcessedAt, &d.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func UpdateSourceProcessing(ctx context.Context, pool *sql.DB, id string, status string, errorMessage *string, chunkCount int, processedAt *time.Time) error {
	_, err := pool.ExecContext(ctx, `
        UPDATE data_sources SET
            status=$1,
            error_message=$2,
            chunk_count=$3,
            processed_at=$4
        WHERE id=$5`,
		status, errorMessage, chunkCount, processedAt, id,
	)
	return err
}

func DeleteSource(ctx context.Context, pool *sql.DB, id string) error {
	_, err := pool.ExecContext(ctx, `DELETE FROM data_sources WHERE id=$1`, id)
	return err
}

func UpdateSourceCapability(ctx context.Context, pool *sql.DB, id string, summary string) error {
	_, err := pool.ExecContext(ctx, `UPDATE data_sources SET capability_summary=$1 WHERE id=$2`, summary, id)
	return err
}

func UpdateSourceSuggestions(ctx context.Context, pool *sql.DB, id string, suggestions []string) error {
	var js any
	if suggestions == nil {
		js = nil
	} else {
		js = suggestions
	}
	_, err := pool.ExecContext(ctx, `UPDATE data_sources SET suggested_questions=$1 WHERE id=$2`, js, id)
	return err
}
