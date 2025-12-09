package db

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/models"
)

// SaveMessageSources persists source usage for a message
func SaveMessageSources(ctx context.Context, pool *sql.DB, messageID string, sources []models.ChunkMetadata) error {
	if len(sources) == 0 {
		return nil
	}

	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO message_sources (message_id, source_id, chunk_index, relevance_score)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (message_id, source_id, chunk_index) DO NOTHING
    `)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for _, src := range sources {
		if src.SourceID == "" {
			continue // Skip if no source ID
		}
		_, err = stmt.ExecContext(ctx, messageID, src.SourceID, src.ChunkIndex, src.Score)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetMessageSources retrieves sources used in a specific message
func GetMessageSources(ctx context.Context, pool *sql.DB, messageID string) ([]models.MessageSource, error) {
	query := `
        SELECT id, message_id, source_id, chunk_index, relevance_score, created_at
        FROM message_sources
        WHERE message_id = $1
        ORDER BY relevance_score DESC
    `
	rows, err := pool.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var sources []models.MessageSource
	for rows.Next() {
		var s models.MessageSource
		if err := rows.Scan(&s.ID, &s.MessageID, &s.SourceID, &s.ChunkIndex, &s.RelevanceScore, &s.CreatedAt); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}
