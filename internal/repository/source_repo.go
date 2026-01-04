// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresSourceRepo implements SourceRepository using PostgreSQL.
// SQL queries are built using Squirrel for type safety and maintainability.
type PostgresSourceRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresSourceRepo implements SourceRepository.
var _ SourceRepository = (*PostgresSourceRepo)(nil)

// NewPostgresSourceRepo creates a new PostgresSourceRepo instance.
func NewPostgresSourceRepo(pool *sql.DB) *PostgresSourceRepo {
	return &PostgresSourceRepo{pool: pool}
}

// Pool returns the underlying database pool for use in tests.
func (r *PostgresSourceRepo) Pool() *sql.DB {
	return r.pool
}

// scanSource scans a single data source row from the result set.
func (r *PostgresSourceRepo) scanSource(rows *sql.Rows) (*models.DataSource, error) {
	var s models.DataSource
	var isDiscovered sql.NullBool
	if err := rows.Scan(
		&s.ID, &s.ChatbotID, &s.SourceType, &s.SourceURL, &s.FilePath, &s.OriginalFilename,
		&s.Status, &s.ErrorMessage, &s.ChunkCount, &s.ProcessedAt, &s.CreatedAt, &s.Hash,
		&s.DeletedAt, &s.SizeBytes, &s.LastRefreshedAt, &isDiscovered, &s.CapabilitySummary,
	); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan source")
	}
	s.IsDiscovered = isDiscovered.Bool
	return &s, nil
}

// scanSources scans multiple data source rows from the result set.
func (r *PostgresSourceRepo) scanSources(rows *sql.Rows) ([]models.DataSource, error) {
	var out []models.DataSource
	for rows.Next() {
		s, err := r.scanSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan sources rows")
	}
	return out, nil
}

// scanSourceMinimal scans a source with minimal columns (for GetURLSources).
func (r *PostgresSourceRepo) scanSourceMinimal(rows *sql.Rows) (*models.DataSource, error) {
	var s models.DataSource
	if err := rows.Scan(
		&s.ID, &s.ChatbotID, &s.SourceType, &s.SourceURL, &s.Status, &s.ErrorMessage,
		&s.ChunkCount, &s.CreatedAt, &s.Hash, &s.DeletedAt, &s.LastRefreshedAt,
	); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan source minimal")
	}
	return &s, nil
}

// scanSourcesMinimal scans multiple sources with minimal columns.
func (r *PostgresSourceRepo) scanSourcesMinimal(rows *sql.Rows) ([]models.DataSource, error) {
	var out []models.DataSource
	for rows.Next() {
		s, err := r.scanSourceMinimal(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan sources minimal rows")
	}
	return out, nil
}

// scanSourceByHash scans a source retrieved by hash (limited columns).
func (r *PostgresSourceRepo) scanSourceByHash(rows *sql.Rows) (*models.DataSource, error) {
	var s models.DataSource
	if err := rows.Scan(
		&s.ID, &s.ChatbotID, &s.SourceType, &s.Status, &s.CreatedAt,
	); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan source by hash")
	}
	return &s, nil
}

// GetByID retrieves a data source by its unique identifier.
// Returns nil, nil if the source is not found.
func (r *PostgresSourceRepo) GetByID(ctx context.Context, id string) (*models.DataSource, error) {
	query, args, err := psql.
		Select(
			"id", "chatbot_id", "source_type", "source_url", "file_path", "original_filename",
			"status", "error_message", "chunk_count", "processed_at", "created_at", "hash",
			"deleted_at", "size_bytes", "last_refreshed_at", "COALESCE(is_discovered, false)",
			"capability_summary",
		).
		From("data_sources").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by id query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query source")
	}
	defer rows.Close()

	sources, err := r.scanSources(rows)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, nil
	}
	return &sources[0], nil
}

// GetByChatbot retrieves all non-deleted data sources for a chatbot.
func (r *PostgresSourceRepo) GetByChatbot(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
	query, args, err := psql.
		Select(
			"id", "chatbot_id", "source_type", "source_url", "file_path", "original_filename",
			"status", "error_message", "chunk_count", "processed_at", "created_at", "hash",
			"deleted_at", "size_bytes", "last_refreshed_at", "COALESCE(is_discovered, false)",
			"capability_summary",
		).
		From("data_sources").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by chatbot query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query sources")
	}
	defer rows.Close()
	return r.scanSources(rows)
}

// GetURLSources retrieves all URL-type sources for a chatbot.
func (r *PostgresSourceRepo) GetURLSources(ctx context.Context, chatbotID string) ([]models.DataSource, error) {
	query, args, err := psql.
		Select(
			"id", "chatbot_id", "source_type", "source_url", "status", "error_message",
			"chunk_count", "created_at", "hash", "deleted_at", "last_refreshed_at",
		).
		From("data_sources").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"source_type": "url"}).
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get url sources query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query url sources")
	}
	defer rows.Close()
	return r.scanSourcesMinimal(rows)
}

// Create persists a new data source and returns its generated ID.
func (r *PostgresSourceRepo) Create(ctx context.Context, source *models.DataSource) (string, error) {
	query, args, err := psql.
		Insert("data_sources").
		Columns(
			"chatbot_id", "source_type", "source_url", "file_path", "original_filename",
			"status", "error_message", "chunk_count", "processed_at", "hash",
			"deleted_at", "size_bytes", "is_discovered", "capability_summary",
		).
		Values(
			source.ChatbotID, source.SourceType, source.SourceURL, source.FilePath, source.OriginalFilename,
			source.Status, source.ErrorMessage, source.ChunkCount, source.ProcessedAt, source.Hash,
			source.DeletedAt, source.SizeBytes, source.IsDiscovered, source.CapabilitySummary,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", pkgerrors.Wrapf(err, "build create query")
	}

	var id string
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		// Fallback for legacy schemas without new columns
		return r.createLegacy(ctx, source)
	}
	return id, nil
}

// createLegacy handles insertion for schemas without the newer columns.
func (r *PostgresSourceRepo) createLegacy(ctx context.Context, source *models.DataSource) (string, error) {
	query, args, err := psql.
		Insert("data_sources").
		Columns(
			"chatbot_id", "source_type", "source_url", "file_path", "original_filename",
			"status", "error_message", "chunk_count",
		).
		Values(
			source.ChatbotID, source.SourceType, source.SourceURL, source.FilePath, source.OriginalFilename,
			source.Status, source.ErrorMessage, source.ChunkCount,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", pkgerrors.Wrapf(err, "build legacy create query")
	}

	var id string
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "create source legacy")
	}
	return id, nil
}

// SoftDelete marks a source as deleted by setting deleted_at timestamp.
func (r *PostgresSourceRepo) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	query, args, err := psql.
		Update("data_sources").
		Set("deleted_at", now).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build soft delete query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "soft delete source")
	}
	return nil
}

// Delete permanently removes a source by its ID.
func (r *PostgresSourceRepo) Delete(ctx context.Context, id string) error {
	query, args, err := psql.
		Delete("data_sources").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build delete query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "delete source")
	}
	return nil
}

// Exists checks if a source with the given URL already exists for a chatbot.
func (r *PostgresSourceRepo) Exists(ctx context.Context, chatbotID, url string) (bool, error) {
	query, args, err := psql.
		Select("1").
		From("data_sources").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"source_url": url}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, pkgerrors.Wrapf(err, "build exists query")
	}

	var exists int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, pkgerrors.Wrapf(err, "check source exists")
	}
	return exists == 1, nil
}

// ExistsByHash checks if a source with the same content hash exists for a chatbot.
func (r *PostgresSourceRepo) ExistsByHash(ctx context.Context, chatbotID, hash string) (bool, error) {
	query, args, err := psql.
		Select("1").
		From("data_sources").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"hash": hash}).
		Where(sq.Eq{"deleted_at": nil}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, pkgerrors.Wrapf(err, "build exists by hash query")
	}

	var exists int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, pkgerrors.Wrapf(err, "check source exists by hash")
	}
	return exists == 1, nil
}

// GetByHash retrieves a source by its content hash within a chatbot.
// Returns nil, nil if no matching source is found.
func (r *PostgresSourceRepo) GetByHash(ctx context.Context, chatbotID, hash string) (*models.DataSource, error) {
	query, args, err := psql.
		Select("id", "chatbot_id", "source_type", "status", "created_at").
		From("data_sources").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"hash": hash}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by hash query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query source by hash")
	}
	defer rows.Close()

	for rows.Next() {
		return r.scanSourceByHash(rows)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan source by hash")
	}
	return nil, nil
}

// CountByType counts non-deleted, non-failed sources of a specific type.
func (r *PostgresSourceRepo) CountByType(ctx context.Context, chatbotID, sourceType string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("data_sources").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"source_type": sourceType}).
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.NotEq{"status": "failed"}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count by type query")
	}

	var count int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "count sources by type")
	}
	return count, nil
}

// UpdateForRefresh sets status to pending and clears error_message for a source refresh.
func (r *PostgresSourceRepo) UpdateForRefresh(ctx context.Context, id string) error {
	query, args, err := psql.
		Update("data_sources").
		Set("status", "pending").
		Set("error_message", nil).
		Set("last_refreshed_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update for refresh query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source for refresh")
	}
	return nil
}

// UpdateSourceHash updates the content hash for a source.
func (r *PostgresSourceRepo) UpdateSourceHash(ctx context.Context, id string, hash string) error {
	query, args, err := psql.
		Update("data_sources").
		Set("hash", hash).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update hash query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source hash")
	}
	return nil
}

// UpdateSourceProcessing updates processing status, error, chunk count, and processed_at.
func (r *PostgresSourceRepo) UpdateSourceProcessing(ctx context.Context, id string, status string, errorMessage *string, chunkCount int, processedAt *time.Time) error {
	query, args, err := psql.
		Update("data_sources").
		Set("status", status).
		Set("error_message", errorMessage).
		Set("chunk_count", chunkCount).
		Set("processed_at", processedAt).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update processing query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source processing")
	}
	return nil
}

// UpdateSourceCapability updates the capability summary for a source.
func (r *PostgresSourceRepo) UpdateSourceCapability(ctx context.Context, id string, summary string) error {
	query, args, err := psql.
		Update("data_sources").
		Set("capability_summary", summary).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update capability query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source capability")
	}
	return nil
}

// UpdateSourceSuggestions updates the suggested questions for a source.
func (r *PostgresSourceRepo) UpdateSourceSuggestions(ctx context.Context, id string, suggestions []string) error {
	var js []byte
	var err error
	if suggestions != nil {
		js, err = json.Marshal(suggestions)
		if err != nil {
			return pkgerrors.Wrapf(err, "marshal suggestions")
		}
	}

	query, args, err := psql.
		Update("data_sources").
		Set("suggested_questions", js).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update suggestions query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update source suggestions")
	}
	return nil
}

// GetLastDeletedAtForURL returns the most recent deleted_at timestamp for a given URL in a chatbot.
// Returns zero time, false, nil if no deleted source is found.
func (r *PostgresSourceRepo) GetLastDeletedAtForURL(ctx context.Context, chatbotID, url string) (time.Time, bool, error) {
	query, args, err := psql.
		Select("deleted_at").
		From("data_sources").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"source_url": url}).
		Where(sq.NotEq{"deleted_at": nil}).
		OrderBy("deleted_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return time.Time{}, false, pkgerrors.Wrapf(err, "build get last deleted at query")
	}

	var deletedAt sql.NullTime
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&deletedAt)
	if err == sql.ErrNoRows {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, pkgerrors.Wrapf(err, "get last deleted at for url")
	}
	return deletedAt.Time, deletedAt.Valid, nil
}
