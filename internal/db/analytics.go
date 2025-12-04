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

	// We don't have a total_tokens column in the schema provided in docs/chatbot_saas_implementation_guide.md
	// The schema has: total_conversations, total_messages, unanswered_messages, thumbs_up_count, thumbs_down_count, average_tokens_per_message
	// To keep it simple and consistent with the schema, we will just update message and conversation counts for now.
	// If we want to track tokens, we should probably add a total_tokens_used column.
	// For now, let's assume we want to track what we can.

	query := `
		INSERT INTO analytics (chatbot_id, analytics_date, total_messages, total_conversations)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chatbot_id, analytics_date)
		DO UPDATE SET
			total_messages = analytics.total_messages + EXCLUDED.total_messages,
			total_conversations = analytics.total_conversations + EXCLUDED.total_conversations
	`

	_, err := pool.ExecContext(ctx, query, chatbotID, dateStr, msgInc, convInc)
	return err
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
