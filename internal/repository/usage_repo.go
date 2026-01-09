// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresUsageRepo implements UsageRepository using PostgreSQL.
type PostgresUsageRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresUsageRepo implements UsageRepository.
var _ UsageRepository = (*PostgresUsageRepo)(nil)

// NewPostgresUsageRepo creates a new PostgresUsageRepo instance.
func NewPostgresUsageRepo(pool *sql.DB) *PostgresUsageRepo {
	return &PostgresUsageRepo{pool: pool}
}

// monthStart normalizes a time to the first day of the month (UTC date component).
func monthStart(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

// CountChatbotsByUserID returns the number of chatbots owned by a user.
func (r *PostgresUsageRepo) CountChatbotsByUserID(ctx context.Context, userID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("chatbots").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count query")
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return count, pkgerrors.Wrapf(err, "count chatbots")
	}
	return count, nil
}

// CountChatbotsByWorkspace returns the number of chatbots in a workspace.
func (r *PostgresUsageRepo) CountChatbotsByWorkspace(ctx context.Context, workspaceID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("chatbots").
		Where(sq.Eq{"workspace_id": workspaceID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count query")
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return count, pkgerrors.Wrapf(err, "count chatbots by workspace")
	}
	return count, nil
}

// GetFileCountByUserID returns the total number of file sources (pdf, text) for a user's chatbots.
func (r *PostgresUsageRepo) GetFileCountByUserID(ctx context.Context, userID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("data_sources ds").
		Join("chatbots c ON ds.chatbot_id = c.id").
		Where(sq.Eq{"c.user_id": userID}).
		Where(sq.Eq{"ds.source_type": []string{"pdf", "text"}}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build query")
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return count, pkgerrors.Wrapf(err, "get file count")
	}
	return count, nil
}

// GetURLCountByUserID returns the total number of URL sources for a user's chatbots.
func (r *PostgresUsageRepo) GetURLCountByUserID(ctx context.Context, userID string) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("data_sources ds").
		Join("chatbots c ON ds.chatbot_id = c.id").
		Where(sq.Eq{"c.user_id": userID}).
		Where(sq.Eq{"ds.source_type": "url"}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build query")
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return count, pkgerrors.Wrapf(err, "get url count")
	}
	return count, nil
}

// GetStorageUsedMBByUserID returns the total storage used by a user's sources in MB.
func (r *PostgresUsageRepo) GetStorageUsedMBByUserID(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(ds.size_bytes), 0)
		FROM data_sources ds
		JOIN chatbots c ON ds.chatbot_id = c.id
		WHERE c.user_id = $1 AND ds.deleted_at IS NULL
	`
	var totalBytes int64
	if err := r.pool.QueryRowContext(ctx, query, userID).Scan(&totalBytes); err != nil {
		return 0, pkgerrors.Wrapf(err, "get storage used")
	}
	return int(totalBytes / (1024 * 1024)), nil
}

// GetMaxFileCountInAnyBot returns the maximum number of file sources in any single chatbot.
func (r *PostgresUsageRepo) GetMaxFileCountInAnyBot(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COALESCE(MAX(cnt), 0) FROM (
			SELECT COUNT(*) as cnt
			FROM data_sources ds
			JOIN chatbots c ON ds.chatbot_id = c.id
			WHERE c.user_id = $1 AND ds.source_type IN ('pdf', 'text')
			GROUP BY ds.chatbot_id
		) as counts
	`
	var count int
	if err := r.pool.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return count, pkgerrors.Wrapf(err, "get max file count")
	}
	return count, nil
}

// GetMaxURLCountInAnyBot returns the maximum number of URL sources in any single chatbot.
func (r *PostgresUsageRepo) GetMaxURLCountInAnyBot(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COALESCE(MAX(cnt), 0) FROM (
			SELECT COUNT(*) as cnt
			FROM data_sources ds
			JOIN chatbots c ON ds.chatbot_id = c.id
			WHERE c.user_id = $1 AND ds.source_type = 'url'
			GROUP BY ds.chatbot_id
		) as counts
	`
	var count int
	if err := r.pool.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return count, pkgerrors.Wrapf(err, "get max url count")
	}
	return count, nil
}

// GetMonthlyTokenUsage returns the total tokens used by all chatbots of a user in the current month.
func (r *PostgresUsageRepo) GetMonthlyTokenUsage(ctx context.Context, userID string) (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startStr := startOfMonth.Format("2006-01-02")

	query := `
		SELECT COALESCE(SUM(a.total_tokens_used), 0)
		FROM analytics a
		JOIN chatbots c ON a.chatbot_id = c.id
		WHERE c.user_id = $1 AND a.analytics_date >= $2
	`
	var total int
	if err := r.pool.QueryRowContext(ctx, query, userID, startStr).Scan(&total); err != nil {
		return 0, pkgerrors.Wrapf(err, "get monthly token usage")
	}
	return total, nil
}

// GetMonthlyIngestionUsage returns sources_count and embedding_tokens for current month.
func (r *PostgresUsageRepo) GetMonthlyIngestionUsage(ctx context.Context, userID string, at time.Time) (int, int, error) {
	pm := monthStart(at)
	query := `
		SELECT COALESCE(sources_count, 0), COALESCE(embedding_tokens, 0)
		FROM usage_ingestions WHERE user_id=$1 AND period_month=$2
	`
	var sources, tokens int
	err := r.pool.QueryRowContext(ctx, query, userID, pm).Scan(&sources, &tokens)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	if err != nil {
		return sources, tokens, pkgerrors.Wrapf(err, "get monthly ingestion usage")
	}
	return sources, tokens, nil
}

// GetMonthlyRefreshCount returns the number of refreshes for a user in a given month.
func (r *PostgresUsageRepo) GetMonthlyRefreshCount(ctx context.Context, userID string, month time.Time) (int, error) {
	periodMonth := monthStart(month)
	query := `
		SELECT COALESCE(refresh_count, 0) FROM usage_ingestions
		WHERE user_id=$1 AND period_month=$2
	`
	var count int
	err := r.pool.QueryRowContext(ctx, query, userID, periodMonth).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "get monthly refresh count")
	}
	return count, nil
}

// IncrementRefreshCount increments the refresh_count for a user in a given month.
func (r *PostgresUsageRepo) IncrementRefreshCount(ctx context.Context, userID string, month time.Time) error {
	periodMonth := monthStart(month)
	query := `
		INSERT INTO usage_ingestions (user_id, period_month, refresh_count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, period_month)
		DO UPDATE SET refresh_count = usage_ingestions.refresh_count + 1, updated_at = NOW()
	`
	if _, err := r.pool.ExecContext(ctx, query, userID, periodMonth); err != nil {
		return pkgerrors.Wrapf(err, "increment refresh count")
	}
	return nil
}

// IncrementSuccessfulIngestion increments sources_count for the current month.
func (r *PostgresUsageRepo) IncrementSuccessfulIngestion(ctx context.Context, userID string, at time.Time, delta int) error {
	pm := monthStart(at)
	query := `
		INSERT INTO usage_ingestions(user_id, period_month, sources_count, embedding_tokens, updated_at)
		VALUES ($1, $2, $3, 0, NOW())
		ON CONFLICT (user_id, period_month)
		DO UPDATE SET sources_count = usage_ingestions.sources_count + EXCLUDED.sources_count,
					  updated_at = NOW()
	`
	if _, err := r.pool.ExecContext(ctx, query, userID, pm, delta); err != nil {
		return pkgerrors.Wrapf(err, "increment successful ingestion")
	}
	return nil
}

// AddEmbeddingTokens adds to embedding_tokens counter for the current month.
func (r *PostgresUsageRepo) AddEmbeddingTokens(ctx context.Context, userID string, at time.Time, tokens int) error {
	pm := monthStart(at)
	query := `
		INSERT INTO usage_ingestions(user_id, period_month, sources_count, embedding_tokens, updated_at)
		VALUES ($1, $2, 0, $3, NOW())
		ON CONFLICT (user_id, period_month)
		DO UPDATE SET embedding_tokens = usage_ingestions.embedding_tokens + EXCLUDED.embedding_tokens,
					  updated_at = NOW()
	`
	if _, err := r.pool.ExecContext(ctx, query, userID, pm, tokens); err != nil {
		return pkgerrors.Wrapf(err, "add embedding tokens")
	}
	return nil
}

// GetAutoRefreshCountForMonth returns the auto_refresh_count for a user in a given month.
func (r *PostgresUsageRepo) GetAutoRefreshCountForMonth(ctx context.Context, userID string, month time.Time) (int, error) {
	pm := monthStart(month)
	query := `
		SELECT COALESCE(auto_refresh_count, 0) 
		FROM usage_ingestions 
		WHERE user_id = $1 AND period_month = $2
	`
	var count int
	err := r.pool.QueryRowContext(ctx, query, userID, pm).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "get auto refresh count")
	}
	return count, nil
}

// IncrementAutoRefreshCount increments the auto_refresh_count for a user in a given month.
func (r *PostgresUsageRepo) IncrementAutoRefreshCount(ctx context.Context, userID string, month time.Time, delta int) error {
	pm := monthStart(month)
	query := `
		INSERT INTO usage_ingestions (user_id, period_month, auto_refresh_count, sources_count, embedding_tokens, updated_at)
		VALUES ($1, $2, $3, 0, 0, NOW())
		ON CONFLICT(user_id, period_month) 
		DO UPDATE SET auto_refresh_count = usage_ingestions.auto_refresh_count + $3,
					  updated_at = NOW()
	`
	if _, err := r.pool.ExecContext(ctx, query, userID, pm, delta); err != nil {
		return pkgerrors.Wrapf(err, "increment auto refresh count")
	}
	return nil
}

// ErrTokenQuotaExceeded is returned when a user has exceeded their monthly token quota.
var ErrTokenQuotaExceeded = errors.New("monthly token quota exceeded")

// ReserveChatTokens atomically reserves tokens for a chat request.
// Returns ErrTokenQuotaExceeded if the reservation would exceed the limit.
func (r *PostgresUsageRepo) ReserveChatTokens(ctx context.Context, userID string, estimatedTokens int, maxMonthlyTokens int) error {
	pm := monthStart(time.Now())

	query := `
		WITH params AS (
			SELECT $1::text AS user_id, $2::date AS period_month, $3::int AS est_tokens, $4::int AS max_tokens
		)
		INSERT INTO usage_ingestions (user_id, period_month, chat_tokens, sources_count, embedding_tokens, updated_at)
		SELECT user_id, period_month, est_tokens, 0, 0, NOW()
		FROM params
		WHERE est_tokens <= max_tokens
		ON CONFLICT (user_id, period_month)
		DO UPDATE SET 
			chat_tokens = usage_ingestions.chat_tokens + EXCLUDED.chat_tokens,
			updated_at = NOW()
		WHERE usage_ingestions.chat_tokens + EXCLUDED.chat_tokens <= (SELECT max_tokens FROM params)
		RETURNING chat_tokens
	`

	var newTokens int
	err := r.pool.QueryRowContext(ctx, query, userID, pm, estimatedTokens, maxMonthlyTokens).Scan(&newTokens)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTokenQuotaExceeded
		}
		return pkgerrors.Wrapf(err, "reserve chat tokens")
	}

	return nil
}

// AdjustChatTokens adjusts the token count after a chat request completes.
func (r *PostgresUsageRepo) AdjustChatTokens(ctx context.Context, userID string, deltaTokens int) error {
	if deltaTokens == 0 {
		return nil
	}

	pm := monthStart(time.Now())
	_, err := r.pool.ExecContext(ctx, `
		UPDATE usage_ingestions 
		SET chat_tokens = chat_tokens + $3, updated_at = NOW()
		WHERE user_id = $1 AND period_month = $2::date
	`, userID, pm, deltaTokens)
	if err != nil {
		return pkgerrors.Wrapf(err, "adjust chat tokens")
	}
	return nil
}

// GetMonthlyChatTokens returns the current monthly chat token usage for a user.
func (r *PostgresUsageRepo) GetMonthlyChatTokens(ctx context.Context, userID string) (int, error) {
	pm := monthStart(time.Now())
	var tokens int
	err := r.pool.QueryRowContext(ctx, `
		SELECT COALESCE(chat_tokens, 0)
		FROM usage_ingestions
		WHERE user_id = $1 AND period_month = $2::date
	`, userID, pm).Scan(&tokens)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "get monthly chat tokens")
	}
	return tokens, nil
}

// IncrementChatTokens adds to the chat_tokens counter for the current month.
func (r *PostgresUsageRepo) IncrementChatTokens(ctx context.Context, userID string, tokens int) error {
	pm := monthStart(time.Now())
	_, err := r.pool.ExecContext(ctx, `
		INSERT INTO usage_ingestions (user_id, period_month, chat_tokens, sources_count, embedding_tokens, updated_at)
		VALUES ($1, $2, $3, 0, 0, NOW())
		ON CONFLICT (user_id, period_month)
		DO UPDATE SET 
			chat_tokens = usage_ingestions.chat_tokens + EXCLUDED.chat_tokens,
			updated_at = NOW()
	`, userID, pm, tokens)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment chat tokens")
	}
	return nil
}

