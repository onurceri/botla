// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresPendingURLRepo implements PendingURLRepository using PostgreSQL.
type PostgresPendingURLRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresPendingURLRepo implements PendingURLRepository.
var _ PendingURLRepository = (*PostgresPendingURLRepo)(nil)

// NewPostgresPendingURLRepo creates a new PostgresPendingURLRepo instance.
func NewPostgresPendingURLRepo(pool *sql.DB) *PostgresPendingURLRepo {
	return &PostgresPendingURLRepo{pool: pool}
}

// InsertPendingURL adds a URL to the pending list for approval.
func (r *PostgresPendingURLRepo) InsertPendingURL(ctx context.Context, chatbotID string, sourceID *string, url string) error {
	query, args, err := psql.
		Insert("pending_discovered_urls").
		Columns("chatbot_id", "source_id", "url").
		Values(chatbotID, sourceID, url).
		Suffix("ON CONFLICT (chatbot_id, url) DO NOTHING").
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build insert pending url query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "insert pending url")
	}
	return nil
}

// ListPendingURLs returns pending URLs for a chatbot with pagination.
func (r *PostgresPendingURLRepo) ListPendingURLs(ctx context.Context, chatbotID string, limit, offset int) ([]models.PendingURL, error) {
	limit64, offset64, err := ValidatePagination(limit, offset)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "validate pagination")
	}

	query, args, err := psql.
		Select("id", "chatbot_id", "source_id", "url", "discovered_at", "status").
		From("pending_discovered_urls").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"status": "pending"}).
		OrderBy("discovered_at DESC").
		Limit(limit64).
		Offset(offset64).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build list pending urls query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query pending urls")
	}
	defer func() { _ = rows.Close() }()

	var urls []models.PendingURL
	for rows.Next() {
		var u models.PendingURL
		var discoveredAt time.Time
		if err := rows.Scan(&u.ID, &u.ChatbotID, &u.SourceID, &u.URL, &discoveredAt, &u.Status); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan pending url")
		}
		u.DiscoveredAt = discoveredAt
		urls = append(urls, u)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "pending urls rows err")
	}
	return urls, nil
}

// CountPendingURLs returns the total count of pending URLs for a chatbot.
func (r *PostgresPendingURLRepo) CountPendingURLs(ctx context.Context, chatbotID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("pending_discovered_urls").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where(sq.Eq{"status": "pending"}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count pending urls query")
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, pkgerrors.Wrapf(err, "count pending urls")
	}
	return count, nil
}

// UpdatePendingURLStatus updates the status of multiple pending URLs.
func (r *PostgresPendingURLRepo) UpdatePendingURLStatus(ctx context.Context, chatbotID string, urlIDs []string, status string) (int, error) {
	if len(urlIDs) == 0 {
		return 0, nil
	}

	// Use raw SQL for the ANY array operator which Squirrel doesn't handle well
	result, err := r.pool.ExecContext(ctx, `
		UPDATE pending_discovered_urls
		SET status = $2
		WHERE chatbot_id = $1 AND id = ANY($3::uuid[])
	`, chatbotID, status, pq.Array(urlIDs))
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "update pending url status")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "rows affected")
	}
	return int(affected), nil
}

// GetPendingURLsByIDs returns pending URLs by their IDs.
func (r *PostgresPendingURLRepo) GetPendingURLsByIDs(ctx context.Context, chatbotID string, urlIDs []string) ([]models.PendingURL, error) {
	if len(urlIDs) == 0 {
		return nil, nil
	}

	// Use raw SQL for the ANY array operator
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, chatbot_id, source_id, url, discovered_at, status
		FROM pending_discovered_urls
		WHERE chatbot_id = $1 AND id = ANY($2::uuid[]) AND status = 'pending'
	`, chatbotID, pq.Array(urlIDs))
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query pending urls by ids")
	}
	defer func() { _ = rows.Close() }()

	var urls []models.PendingURL
	for rows.Next() {
		var u models.PendingURL
		var discoveredAt time.Time
		if err := rows.Scan(&u.ID, &u.ChatbotID, &u.SourceID, &u.URL, &discoveredAt, &u.Status); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan pending url by id")
		}
		u.DiscoveredAt = discoveredAt
		urls = append(urls, u)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "pending urls by ids rows err")
	}
	return urls, nil
}

// DeletePendingURLsByChatbot clears all pending URLs for a chatbot.
func (r *PostgresPendingURLRepo) DeletePendingURLsByChatbot(ctx context.Context, chatbotID string) (int, error) {
	query, args, err := psql.
		Delete("pending_discovered_urls").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build delete pending urls query")
	}

	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "delete pending urls")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "rows affected")
	}
	return int(affected), nil
}
