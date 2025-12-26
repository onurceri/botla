package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// monthStart normalizes a time to the first day of the month (UTC date component)
func monthStart(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

// IncrementSuccessfulIngestion increments sources_count for the current month
func IncrementSuccessfulIngestion(ctx context.Context, pool *sql.DB, userID string, at time.Time, delta int) error {
	pm := monthStart(at)
	_, err := pool.ExecContext(ctx, `
        INSERT INTO usage_ingestions(user_id, period_month, sources_count, embedding_tokens, updated_at)
        VALUES ($1, $2, $3, 0, NOW())
        ON CONFLICT (user_id, period_month)
        DO UPDATE SET sources_count = usage_ingestions.sources_count + EXCLUDED.sources_count,
                      updated_at = NOW()
    `, userID, pm, delta)
	if err != nil {
		return fmt.Errorf("increment successful ingestion: %w", err)
	}
	return nil
}

// AddEmbeddingTokens adds to embedding_tokens counter for the current month
func AddEmbeddingTokens(ctx context.Context, pool *sql.DB, userID string, at time.Time, tokens int) error {
	pm := monthStart(at)
	_, err := pool.ExecContext(ctx, `
        INSERT INTO usage_ingestions(user_id, period_month, sources_count, embedding_tokens, updated_at)
        VALUES ($1, $2, 0, $3, NOW())
        ON CONFLICT (user_id, period_month)
        DO UPDATE SET embedding_tokens = usage_ingestions.embedding_tokens + EXCLUDED.embedding_tokens,
                      updated_at = NOW()
    `, userID, pm, tokens)
	if err != nil {
		return fmt.Errorf("add embedding tokens: %w", err)
	}
	return nil
}

// GetMonthlyIngestionUsage returns sources_count and embedding_tokens for current month
func GetMonthlyIngestionUsage(ctx context.Context, pool *sql.DB, userID string, at time.Time) (int, int, error) {
	pm := monthStart(at)
	var sources, tokens int
	err := pool.QueryRowContext(ctx, `
        SELECT COALESCE(sources_count, 0), COALESCE(embedding_tokens, 0)
        FROM usage_ingestions WHERE user_id=$1 AND period_month=$2
    `, userID, pm).Scan(&sources, &tokens)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	if err != nil {
		return sources, tokens, fmt.Errorf("get monthly ingestion usage: %w", err)
	}
	return sources, tokens, nil
}

// GetAutoRefreshCountForMonth returns the auto_refresh_count for a user in a given month
func GetAutoRefreshCountForMonth(ctx context.Context, pool *sql.DB, userID string, month time.Time) (int, error) {
	pm := monthStart(month)
	var count int
	err := pool.QueryRowContext(ctx, `
        SELECT COALESCE(auto_refresh_count, 0) 
        FROM usage_ingestions 
        WHERE user_id = $1 AND period_month = $2`,
		userID, pm).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get auto refresh count: %w", err)
	}
	return count, nil
}

// IncrementAutoRefreshCount increments the auto_refresh_count for a user in a given month
func IncrementAutoRefreshCount(ctx context.Context, pool *sql.DB, userID string, month time.Time, delta int) error {
	pm := monthStart(month)
	_, err := pool.ExecContext(ctx, `
        INSERT INTO usage_ingestions (user_id, period_month, auto_refresh_count, sources_count, embedding_tokens, updated_at)
        VALUES ($1, $2, $3, 0, 0, NOW())
        ON CONFLICT(user_id, period_month) 
        DO UPDATE SET auto_refresh_count = usage_ingestions.auto_refresh_count + $3,
                      updated_at = NOW()`,
		userID, pm, delta)
	if err != nil {
		return fmt.Errorf("increment auto refresh count: %w", err)
	}
	return nil
}
