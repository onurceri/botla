package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

func CreateDataSource(ctx context.Context, pool *sql.DB, s *models.DataSource) (string, error) {
	var id string
	err := pool.QueryRowContext(
		ctx,
		`INSERT INTO data_sources (
            chatbot_id, source_type, source_url, file_path, original_filename,
            status, error_message, chunk_count, processed_at, hash, deleted_at, size_bytes
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		s.ChatbotID, s.SourceType, s.SourceURL, s.FilePath, s.OriginalFilename,
		s.Status, s.ErrorMessage, s.ChunkCount, s.ProcessedAt, s.Hash, s.DeletedAt, s.SizeBytes,
	).Scan(&id)
	if err != nil {
		// Fallback for legacy schemas without new columns
		// SQLSTATE 42703: undefined_column
		var legacyID string
		if e := pool.QueryRowContext(
			ctx,
			`INSERT INTO data_sources (
                chatbot_id, source_type, source_url, file_path, original_filename,
                status, error_message, chunk_count
            ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
			s.ChatbotID, s.SourceType, s.SourceURL, s.FilePath, s.OriginalFilename,
			s.Status, s.ErrorMessage, s.ChunkCount,
		).Scan(&legacyID); e == nil {
			return legacyID, nil
		}
		return "", pkgerrors.Wrapf(err, "scan data source id")
	}
	return id, nil
}

func CreateSource(ctx context.Context, pool *sql.DB, chatbotID, sourceType string, sourceURL, filePath, originalFilename *string) (string, error) {
	ds := models.DataSource{
		ChatbotID:        chatbotID,
		SourceType:       sourceType,
		SourceURL:        sourceURL,
		FilePath:         filePath,
		OriginalFilename: originalFilename,
		Status:           "pending",
		IsDiscovered:     false,
	}
	return CreateDataSource(ctx, pool, &ds)
}

// CreateDiscoveredSource creates a source that was discovered via URL crawling
// These sources will not perform further sub-page discovery
func CreateDiscoveredSource(ctx context.Context, pool *sql.DB, chatbotID string, sourceURL string) (string, error) {
	var id string
	err := pool.QueryRowContext(
		ctx,
		`INSERT INTO data_sources (chatbot_id, source_type, source_url, status, is_discovered)
		 VALUES ($1, 'url', $2, 'pending', true) RETURNING id`,
		chatbotID, sourceURL,
	).Scan(&id)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "create discovered source")
	}
	return id, nil
}

func ListSourcesByChatbotID(ctx context.Context, pool *sql.DB, chatbotID string) ([]models.DataSource, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT id, chatbot_id, source_type, source_url, file_path, original_filename,
               status, error_message, chunk_count, processed_at, created_at, hash, deleted_at, 
               size_bytes, last_refreshed_at, COALESCE(is_discovered, false), capability_summary
        FROM data_sources
        WHERE chatbot_id=$1 AND deleted_at IS NULL
        ORDER BY created_at DESC`, chatbotID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query sources by chatbot id")
	}
	defer func() { _ = rows.Close() }()
	out := []models.DataSource{}
	for rows.Next() {
		var d models.DataSource
		if err := rows.Scan(
			&d.ID, &d.ChatbotID, &d.SourceType, &d.SourceURL, &d.FilePath, &d.OriginalFilename,
			&d.Status, &d.ErrorMessage, &d.ChunkCount, &d.ProcessedAt, &d.CreatedAt, &d.Hash, &d.DeletedAt,
			&d.SizeBytes, &d.LastRefreshedAt, &d.IsDiscovered, &d.CapabilitySummary,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan source")
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "rows err")
	}
	return out, nil
}

func GetSourceByID(ctx context.Context, pool *sql.DB, id string) (*models.DataSource, error) {
	var d models.DataSource
	err := pool.QueryRowContext(ctx, `
        SELECT id, chatbot_id, source_type, source_url, file_path, original_filename,
               status, error_message, chunk_count, processed_at, created_at, hash, deleted_at, size_bytes, last_refreshed_at, COALESCE(is_discovered, false)
        FROM data_sources WHERE id=$1`, id).Scan(
		&d.ID, &d.ChatbotID, &d.SourceType, &d.SourceURL, &d.FilePath, &d.OriginalFilename,
		&d.Status, &d.ErrorMessage, &d.ChunkCount, &d.ProcessedAt, &d.CreatedAt, &d.Hash, &d.DeletedAt, &d.SizeBytes, &d.LastRefreshedAt, &d.IsDiscovered,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get source by id")
	}
	return &d, nil
}

// GetURLSourcesForChatbot returns all URL sources for a chatbot
func GetURLSourcesForChatbot(ctx context.Context, pool *sql.DB, chatbotID string) ([]models.DataSource, error) {
	rows, err := pool.QueryContext(ctx, `
        SELECT id, chatbot_id, source_type, source_url, status, error_message, 
               chunk_count, created_at, hash, deleted_at, last_refreshed_at
        FROM data_sources
        WHERE chatbot_id = $1 AND source_type = 'url' AND deleted_at IS NULL
        ORDER BY created_at DESC`, chatbotID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query url sources")
	}
	defer func() { _ = rows.Close() }()
	var out []models.DataSource
	for rows.Next() {
		var s models.DataSource
		if err := rows.Scan(&s.ID, &s.ChatbotID, &s.SourceType, &s.SourceURL, &s.Status, &s.ErrorMessage,
			&s.ChunkCount, &s.CreatedAt, &s.Hash, &s.DeletedAt, &s.LastRefreshedAt); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan url source")
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "rows err")
	}
	return out, nil
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
	if err != nil {
		return pkgerrors.Wrapf(err, "update source processing")
	}
	return nil
}

func DeleteSource(ctx context.Context, pool *sql.DB, id string) error {
	_, err := pool.ExecContext(ctx, `DELETE FROM data_sources WHERE id=$1`, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete source")
	}
	return nil
}

// SoftDeleteSource sets deleted_at timestamp without removing the row
func SoftDeleteSource(ctx context.Context, pool *sql.DB, id string) error {
	now := time.Now()
	_, err := pool.ExecContext(ctx, `UPDATE data_sources SET deleted_at=$2 WHERE id=$1`, id, now)
	if err != nil {
		return pkgerrors.Wrapf(err, "soft delete source")
	}
	return nil
}

func UpdateSourceCapability(ctx context.Context, pool *sql.DB, id string, summary string) error {
	_, err := pool.ExecContext(ctx, `UPDATE data_sources SET capability_summary=$1 WHERE id=$2`, summary, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source capability")
	}
	return nil
}

func UpdateSourceSuggestions(ctx context.Context, pool *sql.DB, id string, suggestions []string) error {
	var js []byte
	var err error
	if suggestions != nil {
		js, err = json.Marshal(suggestions)
		if err != nil {
			return pkgerrors.Wrapf(err, "marshal suggestions")
		}
	}
	_, err = pool.ExecContext(ctx, `UPDATE data_sources SET suggested_questions=$1 WHERE id=$2`, js, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source suggestions")
	}
	return nil
}

// CountSourcesByType counts non-deleted, non-failed sources of a specific type for a chatbot.
// Failed sources do not count toward plan limits so users aren't penalized for errors.
func CountSourcesByType(ctx context.Context, pool *sql.DB, chatbotID string, sourceType string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM data_sources 
		WHERE chatbot_id=$1 AND source_type=$2 AND deleted_at IS NULL AND status != 'failed'
	`, chatbotID, sourceType).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "count sources by type")
	}
	return count, nil
}

func GetFileCountByUserID(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM data_sources ds
		JOIN chatbots c ON ds.chatbot_id = c.id
		WHERE c.user_id = $1 AND ds.source_type IN ('pdf', 'text')
	`, userID).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "get file count by user id")
	}
	return count, nil
}

func GetURLCountByUserID(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM data_sources ds
		JOIN chatbots c ON ds.chatbot_id = c.id
		WHERE c.user_id = $1 AND ds.source_type = 'url'
	`, userID).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "get url count by user id")
	}
	return count, nil
}

func GetMaxFileCountInAnyBot(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(cnt), 0) FROM (
			SELECT COUNT(*) as cnt
			FROM data_sources ds
			JOIN chatbots c ON ds.chatbot_id = c.id
			WHERE c.user_id = $1 AND ds.source_type IN ('pdf', 'text')
			GROUP BY ds.chatbot_id
		) as counts
	`, userID).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "get max file count in any bot")
	}
	return count, nil
}

func GetMaxURLCountInAnyBot(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(cnt), 0) FROM (
			SELECT COUNT(*) as cnt
			FROM data_sources ds
			JOIN chatbots c ON ds.chatbot_id = c.id
			WHERE c.user_id = $1 AND ds.source_type = 'url'
			GROUP BY ds.chatbot_id
		) as counts
	`, userID).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "get max url count in any bot")
	}
	return count, nil
}

func SourceExists(ctx context.Context, pool *sql.DB, chatbotID, url string) (bool, error) {
	var exists bool
	err := pool.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM data_sources WHERE chatbot_id=$1 AND source_url=$2)`, chatbotID, url).Scan(&exists)
	if err != nil {
		return exists, pkgerrors.Wrapf(err, "check source existence")
	}
	return exists, nil
}

// GetLastDeletedAtForURL returns most recent deleted_at for a given URL in a chatbot
func GetLastDeletedAtForURL(ctx context.Context, pool *sql.DB, chatbotID, url string) (sql.NullTime, error) {
	var t sql.NullTime
	err := pool.QueryRowContext(ctx, `
        SELECT deleted_at FROM data_sources
        WHERE chatbot_id=$1 AND source_url=$2 AND deleted_at IS NOT NULL
        ORDER BY deleted_at DESC LIMIT 1
    `, chatbotID, url).Scan(&t)
	if err == sql.ErrNoRows {
		return sql.NullTime{}, nil
	}
	if err != nil {
		return t, pkgerrors.Wrapf(err, "get last deleted at for url")
	}
	return t, nil
}

// GetStorageUsedMBByUserID sums size_bytes for user's sources
func GetStorageUsedMBByUserID(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	var totalBytes int64
	err := pool.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(size_bytes),0)
        FROM data_sources ds
        JOIN chatbots c ON ds.chatbot_id = c.id
        WHERE c.user_id = $1 AND ds.deleted_at IS NULL
    `, userID).Scan(&totalBytes)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "get storage used")
	}
	// convert to MB
	return int(totalBytes / (1024 * 1024)), nil
}

// UpdateSourceHash updates the content hash for a source
func UpdateSourceHash(ctx context.Context, pool *sql.DB, id string, hash string) error {
	_, err := pool.ExecContext(ctx, `UPDATE data_sources SET hash=$1 WHERE id=$2`, hash, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source hash")
	}
	return nil
}

// UpdateSourceForRefresh sets status to pending and updates last_refreshed_at
func UpdateSourceForRefresh(ctx context.Context, pool *sql.DB, id string) error {
	now := time.Now()
	_, err := pool.ExecContext(ctx, `
		UPDATE data_sources 
		SET status='pending', last_refreshed_at=$1, error_message=NULL
		WHERE id=$2`, now, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source for refresh")
	}
	return nil
}

// GetMonthlyRefreshCount returns the number of refreshes for a user in a given month
func GetMonthlyRefreshCount(ctx context.Context, pool *sql.DB, userID string, month time.Time) (int, error) {
	periodMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COALESCE(refresh_count, 0) FROM usage_ingestions
		WHERE user_id=$1 AND period_month=$2
	`, userID, periodMonth).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "get monthly refresh count")
	}
	return count, nil
}

func IncrementRefreshCount(ctx context.Context, pool *sql.DB, userID string, month time.Time) error {
	periodMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	_, err := pool.ExecContext(ctx, `
		INSERT INTO usage_ingestions (user_id, period_month, refresh_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, period_month)
		DO UPDATE SET refresh_count = usage_ingestions.refresh_count + 1, updated_at = NOW()
	`, userID, periodMonth)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment refresh count")
	}
	return nil
}

// SourceExistsByHash checks if a source with the same hash exists for a chatbot
func SourceExistsByHash(ctx context.Context, db *sql.DB, chatbotID, hash string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM data_sources 
			WHERE chatbot_id = $1 
			  AND hash = $2 
			  AND deleted_at IS NULL
		)
	`, chatbotID, hash).Scan(&exists)
	if err != nil {
		return false, pkgerrors.Wrapf(err, "check source hash")
	}
	return exists, nil
}

// GetSourceByHash returns the existing source with the same hash
func GetSourceByHash(ctx context.Context, db *sql.DB, chatbotID, hash string) (*models.DataSource, error) {
	var s models.DataSource
	err := db.QueryRowContext(ctx, `
		SELECT id, chatbot_id, source_type, status, created_at
		FROM data_sources 
		WHERE chatbot_id = $1 AND hash = $2 AND deleted_at IS NULL
		LIMIT 1
	`, chatbotID, hash).Scan(&s.ID, &s.ChatbotID, &s.SourceType, &s.Status, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get source by hash")
	}
	return &s, nil
}
