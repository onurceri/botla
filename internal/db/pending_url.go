package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

// InsertPendingURL adds a URL to the pending list for approval
func InsertPendingURL(ctx context.Context, pool *sql.DB, chatbotID string, sourceID *string, url string) error {
	_, err := pool.ExecContext(ctx, `
		INSERT INTO pending_discovered_urls (chatbot_id, source_id, url)
		VALUES ($1, $2, $3)
		ON CONFLICT (chatbot_id, url) DO NOTHING`,
		chatbotID, sourceID, url)
	if err != nil {
		return pkgerrors.Wrapf(err, "insert pending url")
	}
	return nil
}

// ListPendingURLs returns pending URLs for a chatbot with pagination
func ListPendingURLs(ctx context.Context, pool *sql.DB, chatbotID string, limit, offset int) ([]models.PendingURL, error) {
	rows, err := pool.QueryContext(ctx, `
		SELECT id, chatbot_id, source_id, url, discovered_at, status
		FROM pending_discovered_urls
		WHERE chatbot_id = $1 AND status = 'pending'
		ORDER BY discovered_at DESC
		LIMIT $2 OFFSET $3`,
		chatbotID, limit, offset)
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

// CountPendingURLs returns the total count of pending URLs for a chatbot
func CountPendingURLs(ctx context.Context, pool *sql.DB, chatbotID string) (int, error) {
	var count int
	err := pool.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pending_discovered_urls
		WHERE chatbot_id = $1 AND status = 'pending'`,
		chatbotID).Scan(&count)
	if err != nil {
		return count, pkgerrors.Wrapf(err, "count pending urls")
	}
	return count, nil
}

// UpdatePendingURLStatus updates the status of multiple pending URLs
func UpdatePendingURLStatus(ctx context.Context, pool *sql.DB, chatbotID string, urlIDs []string, status string) (int, error) {
	if len(urlIDs) == 0 {
		return 0, nil
	}

	result, err := pool.ExecContext(ctx, `
		UPDATE pending_discovered_urls
		SET status = $2
		WHERE chatbot_id = $1 AND id = ANY($3::uuid[])`,
		chatbotID, status, urlIDs)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "update pending url status")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "rows affected")
	}
	return int(affected), nil
}

// GetPendingURLsByIDs returns pending URLs by their IDs
func GetPendingURLsByIDs(ctx context.Context, pool *sql.DB, chatbotID string, urlIDs []string) ([]models.PendingURL, error) {
	if len(urlIDs) == 0 {
		return nil, nil
	}

	rows, err := pool.QueryContext(ctx, `
		SELECT id, chatbot_id, source_id, url, discovered_at, status
		FROM pending_discovered_urls
		WHERE chatbot_id = $1 AND id = ANY($2::uuid[]) AND status = 'pending'`,
		chatbotID, urlIDs)
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

// DeletePendingURLsByChatbot clears all pending URLs for a chatbot
func DeletePendingURLsByChatbot(ctx context.Context, pool *sql.DB, chatbotID string) (int, error) {
	result, err := pool.ExecContext(ctx, `
		DELETE FROM pending_discovered_urls
		WHERE chatbot_id = $1`,
		chatbotID)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "delete pending urls")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "rows affected")
	}
	return int(affected), nil
}
