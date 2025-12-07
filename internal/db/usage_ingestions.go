package db

import (
    "context"
    "database/sql"
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
    return err
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
    return err
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
    return sources, tokens, err
}

