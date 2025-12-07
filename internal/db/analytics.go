package db

import (
	"context"
	"database/sql"
	"time"
)

// IncrementAnalytics updates the analytics table for a given chatbot and date.
// It uses an UPSERT (INSERT ... ON CONFLICT DO UPDATE) to ensure the row exists.
func IncrementAnalytics(ctx context.Context, pool *sql.DB, chatbotID string, date time.Time, isNewConversation bool, tokens int) error {
	// Format date as YYYY-MM-DD
	dateStr := date.Format("2006-01-02")

	// Calculate increments
	msgInc := 2 // User + Assistant
	convInc := 0
	if isNewConversation {
		convInc = 1
	}

	query := `
		INSERT INTO analytics (chatbot_id, analytics_date, total_messages, total_conversations, total_tokens_used)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (chatbot_id, analytics_date)
		DO UPDATE SET
			total_messages = analytics.total_messages + EXCLUDED.total_messages,
			total_conversations = analytics.total_conversations + EXCLUDED.total_conversations,
			total_tokens_used = analytics.total_tokens_used + EXCLUDED.total_tokens_used
	`

	_, err := pool.ExecContext(ctx, query, chatbotID, dateStr, msgInc, convInc, tokens)
	return err
}

// GetMonthlyTokenUsage returns the total tokens used by all chatbots of a user in the current month.
func GetMonthlyTokenUsage(ctx context.Context, pool *sql.DB, userID string) (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startStr := startOfMonth.Format("2006-01-02")

	var total int
	query := `
		SELECT COALESCE(SUM(a.total_tokens_used), 0)
		FROM analytics a
		JOIN chatbots c ON a.chatbot_id = c.id
		WHERE c.user_id = $1 AND a.analytics_date >= $2
	`
	err := pool.QueryRowContext(ctx, query, userID, startStr).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// IncrementFeedback updates the thumbs up/down count in analytics
func IncrementFeedback(ctx context.Context, pool *sql.DB, chatbotID string, date time.Time, isThumbsUp bool) error {
	dateStr := date.Format("2006-01-02")

	upInc := 0
	downInc := 0
	if isThumbsUp {
		upInc = 1
	} else {
		downInc = 1
	}

	query := `
		INSERT INTO analytics (chatbot_id, analytics_date, thumbs_up_count, thumbs_down_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chatbot_id, analytics_date)
		DO UPDATE SET
			thumbs_up_count = analytics.thumbs_up_count + EXCLUDED.thumbs_up_count,
			thumbs_down_count = analytics.thumbs_down_count + EXCLUDED.thumbs_down_count
	`

	_, err := pool.ExecContext(ctx, query, chatbotID, dateStr, upInc, downInc)
	return err
}
