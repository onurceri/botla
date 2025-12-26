package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrTokenQuotaExceeded is returned when a user has exceeded their monthly token quota.
var ErrTokenQuotaExceeded = errors.New("monthly token quota exceeded")

// ReserveChatTokens atomically reserves tokens for a chat request.
// It checks if the user has enough quota and increments the token count in a single
// atomic operation. This prevents the TOCTOU race condition where concurrent requests
// could bypass the quota check.
//
// Parameters:
//   - userID: The user to reserve tokens for
//   - estimatedTokens: Estimated tokens this request will use (e.g., chatbot's max_tokens)
//   - maxMonthlyTokens: The user's monthly token limit from their plan
//
// Returns:
//   - nil if tokens were successfully reserved
//   - ErrTokenQuotaExceeded if the reservation would exceed the limit
//   - Other errors for database failures
//
// After processing, call AdjustChatTokens to correct the reservation to actual usage.
func ReserveChatTokens(ctx context.Context, pool *sql.DB, userID string, estimatedTokens int, maxMonthlyTokens int) error {
	pm := monthStart(time.Now())

	// Atomic check-and-increment using a conditional UPSERT.
	// First, we try to INSERT for new users (who have no row yet)
	// If conflict, we UPDATE but only if the limit wouldn't be exceeded.
	//
	// We use a CTE to define parameters with explicit types to avoid ambiguity.
	// usage_ingestions.chat_tokens defaults to 0 on insert
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
	err := pool.QueryRowContext(ctx, query, userID, pm, estimatedTokens, maxMonthlyTokens).Scan(&newTokens)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// The WHERE clause rejected the insert or update - limit exceeded
			return ErrTokenQuotaExceeded
		}
		return fmt.Errorf("reserve chat tokens: %w", err)
	}

	return nil
}

// AdjustChatTokens adjusts the token count after a chat request completes.
// Call this with the difference between actual tokens used and estimated tokens.
//
// Parameters:
//   - deltaTokens: The adjustment amount (positive to add more, negative to refund)
//
// This allows for the pattern:
//  1. ReserveChatTokens(estimated) - blocks if would exceed
//  2. Process chat, get actualTokens
//  3. AdjustChatTokens(actualTokens - estimated) - correct the count
func AdjustChatTokens(ctx context.Context, pool *sql.DB, userID string, deltaTokens int) error {
	if deltaTokens == 0 {
		return nil
	}

	pm := monthStart(time.Now())
	_, err := pool.ExecContext(ctx, `
		UPDATE usage_ingestions 
		SET chat_tokens = chat_tokens + $3, updated_at = NOW()
		WHERE user_id = $1 AND period_month = $2::date
	`, userID, pm, deltaTokens)
	if err != nil {
		return fmt.Errorf("adjust chat tokens: %w", err)
	}
	return nil
}

// GetMonthlyChatTokens returns the current monthly chat token usage for a user
// from the usage_ingestions table.
func GetMonthlyChatTokens(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	pm := monthStart(time.Now())
	var tokens int
	err := pool.QueryRowContext(ctx, `
		SELECT COALESCE(chat_tokens, 0)
		FROM usage_ingestions
		WHERE user_id = $1 AND period_month = $2::date
	`, userID, pm).Scan(&tokens)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get monthly chat tokens: %w", err)
	}
	return tokens, nil
}

// IncrementChatTokens adds to the chat_tokens counter for the current month.
// This is a simpler version that doesn't check limits - useful for recording
// usage when limits don't apply or have already been checked.
func IncrementChatTokens(ctx context.Context, pool *sql.DB, userID string, tokens int) error {
	pm := monthStart(time.Now())
	_, err := pool.ExecContext(ctx, `
		INSERT INTO usage_ingestions (user_id, period_month, chat_tokens, sources_count, embedding_tokens, updated_at)
		VALUES ($1, $2, $3, 0, 0, NOW())
		ON CONFLICT (user_id, period_month)
		DO UPDATE SET 
			chat_tokens = usage_ingestions.chat_tokens + EXCLUDED.chat_tokens,
			updated_at = NOW()
	`, userID, pm, tokens)
	if err != nil {
		return fmt.Errorf("increment chat tokens: %w", err)
	}
	return nil
}
