package db

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

// IncrementAnalytics updates the analytics table for a given chatbot and date.
// It uses an UPSERT (INSERT ... ON CONFLICT DO UPDATE) to ensure the row exists.
func IncrementAnalytics(ctx context.Context, pool *sql.DB, chatbotID string, isNewConversation bool, tokens int, isHandoff bool, responseTimeMs int) error {
	// Calculate increments
	msgInc := 2 // User + Assistant
	convInc := 0
	if isNewConversation {
		convInc = 1
	}

	handoffInc := 0
	if isHandoff {
		handoffInc = 1
	}

	query := `
		INSERT INTO analytics (chatbot_id, analytics_date, total_messages, total_conversations, total_tokens_used, handoff_count)
		VALUES ($1, CURRENT_DATE, $2, $3, $4, $5)
		ON CONFLICT (chatbot_id, analytics_date)
		DO UPDATE SET
			total_messages = analytics.total_messages + EXCLUDED.total_messages,
			total_conversations = analytics.total_conversations + EXCLUDED.total_conversations,
			total_tokens_used = analytics.total_tokens_used + EXCLUDED.total_tokens_used,
			handoff_count = analytics.handoff_count + EXCLUDED.handoff_count
	`

	_, err := pool.ExecContext(ctx, query, chatbotID, msgInc, convInc, tokens, handoffInc)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment analytics")
	}
	return nil
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
		return 0, pkgerrors.Wrapf(err, "get monthly token usage")
	}
	return total, nil
}

// IncrementFeedback updates the thumbs up/down count in analytics
func IncrementFeedback(ctx context.Context, pool *sql.DB, chatbotID string, oldState *bool, newState bool) error {
	upInc := 0
	downInc := 0

	if oldState == nil {
		// New feedback
		if newState {
			upInc = 1
		} else {
			downInc = 1
		}
	} else {
		if *oldState == newState {
			// No change
			return nil
		}
		if *oldState && !newState {
			// Was up, now down
			upInc = -1
			downInc = 1
		} else if !*oldState && newState {
			// Was down, now up
			downInc = -1
			upInc = 1
		}
	}

	if upInc == 0 && downInc == 0 {
		return nil
	}

	query := `
		INSERT INTO analytics (chatbot_id, analytics_date, thumbs_up_count, thumbs_down_count)
		VALUES ($1, CURRENT_DATE, $2, $3)
		ON CONFLICT (chatbot_id, analytics_date)
		DO UPDATE SET
			thumbs_up_count = analytics.thumbs_up_count + EXCLUDED.thumbs_up_count,
			thumbs_down_count = analytics.thumbs_down_count + EXCLUDED.thumbs_down_count
	`

	_, err := pool.ExecContext(ctx, query, chatbotID, upInc, downInc)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment feedback")
	}
	return nil
}

// GetAnalyticsOverview returns aggregated stats for a chatbot for the last 30 days
func GetAnalyticsOverview(ctx context.Context, pool *sql.DB, chatbotID string) (*models.AnalyticsOverview, error) {
	query := `
		SELECT 
			COALESCE(SUM(total_messages), 0)::INT,
			COALESCE(SUM(total_conversations), 0)::INT,
			COALESCE(SUM(total_tokens_used), 0)::INT,
			COALESCE(SUM(thumbs_up_count), 0)::INT,
			COALESCE(SUM(thumbs_down_count), 0)::INT,
			COALESCE(SUM(handoff_count), 0)::INT
		FROM analytics
		WHERE chatbot_id = $1 AND analytics_date >= CURRENT_DATE - INTERVAL '29 days'
	`
	var stats models.AnalyticsOverview
	row := pool.QueryRowContext(ctx, query, chatbotID)
	err := row.Scan(
		&stats.TotalMessages,
		&stats.TotalConversations,
		&stats.TotalTokensUsed,
		&stats.PositiveFeedback,
		&stats.NegativeFeedback,
		&stats.HandoffCount,
	)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get analytics overview")
	}

	totalFeedback := stats.PositiveFeedback + stats.NegativeFeedback
	if totalFeedback > 0 {
		stats.FeedbackRate = (float64(stats.PositiveFeedback) / float64(totalFeedback)) * 100
	}

	return &stats, nil
}

// GetAnalyticsTrends returns daily stats for a chatbot for the last N days
func GetAnalyticsTrends(ctx context.Context, pool *sql.DB, chatbotID string, days int) ([]models.DailyAnalytics, error) {
	if days <= 0 {
		days = 30
	}
	query := `
		WITH dates AS (
			SELECT generate_series(
				CURRENT_DATE - make_interval(days => $2),
				CURRENT_DATE,
				'1 day'::interval
			)::date AS date
		)
		SELECT 
			to_char(d.date, 'YYYY-MM-DD') as date,
			COALESCE(a.total_messages, 0),
			COALESCE(a.total_conversations, 0),
			COALESCE(a.total_tokens_used, 0),
			COALESCE(a.thumbs_up_count, 0),
			COALESCE(a.thumbs_down_count, 0),
			COALESCE(a.handoff_count, 0),
			a.avg_response_time_ms -- can be null
		FROM dates d
		LEFT JOIN analytics a ON a.analytics_date = d.date AND a.chatbot_id = $1
		ORDER BY d.date
	`
	rows, err := pool.QueryContext(ctx, query, chatbotID, days-1) // days-1 because inclusive range
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query analytics trends")
	}
	defer func() { _ = rows.Close() }()

	var results []models.DailyAnalytics
	for rows.Next() {
		var da models.DailyAnalytics
		if err := rows.Scan(
			&da.Date,
			&da.TotalMessages,
			&da.TotalConversations,
			&da.TotalTokensUsed,
			&da.ThumbsUpCount,
			&da.ThumbsDownCount,
			&da.HandoffCount,
			&da.AvgResponseTimeMs,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan daily analytics")
		}
		results = append(results, da)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "analytics trends rows err")
	}
	return results, nil
}

// TrackUnansweredQuery records a query that had low confidence
func TrackUnansweredQuery(ctx context.Context, pool *sql.DB, chatbotID, queryText string) error {
	queryText = strings.TrimSpace(queryText)
	if queryText == "" {
		return nil
	}

	query := `
		INSERT INTO unanswered_queries (chatbot_id, query, occurrence_count, last_occurred_at)
		VALUES ($1, $2, 1, NOW())
		ON CONFLICT (chatbot_id, query)
		DO UPDATE SET
			occurrence_count = unanswered_queries.occurrence_count + 1,
			last_occurred_at = NOW()
	`
	_, err := pool.ExecContext(ctx, query, chatbotID, queryText)
	if err != nil {
		return pkgerrors.Wrapf(err, "track unanswered query")
	}
	return nil
}

// GetGlobalAnalytics returns aggregated analytics for a scope (User, Workspace, or Org)
func GetGlobalAnalytics(ctx context.Context, pool *sql.DB, userID string, orgID, wsID *string) ([]AnalyticsPoint, error) {
	// Construct WHERE clause based on scope
	var whereClause string
	var args []interface{}

	switch {
	case wsID != nil:
		whereClause = "c.workspace_id = $1"
		args = append(args, *wsID)
	case orgID != nil:
		// Chatbots can be linked to an organization directly OR via a workspace
		whereClause = "(c.organization_id = $1 OR c.workspace_id IN (SELECT id FROM workspaces WHERE organization_id = $1))"
		args = append(args, *orgID)
	default:
		// Personal scope: UserID matches AND not in any workspace/org
		whereClause = "c.user_id = $1 AND c.workspace_id IS NULL AND c.organization_id IS NULL"
		args = append(args, userID)
	}

	//nolint:gosec // whereClause is constructed from constant strings
	query := `
		WITH dates AS (
			SELECT generate_series(
				CURRENT_DATE - INTERVAL '29 days',
				CURRENT_DATE,
				'1 day'::interval
			)::date AS date
		),
		user_analytics AS (
			SELECT a.analytics_date, a.total_messages, a.total_conversations, a.total_tokens_used, a.thumbs_up_count, a.thumbs_down_count, a.handoff_count
			FROM analytics a
			JOIN chatbots c ON a.chatbot_id = c.id
			WHERE (` + whereClause + `) AND a.analytics_date >= CURRENT_DATE - INTERVAL '29 days'
		)
		SELECT 
			to_char(d.date, 'YYYY-MM-DD') as date,
			COALESCE(SUM(ua.total_messages), 0)::INTEGER as messages,
			COALESCE(SUM(ua.total_conversations), 0)::INTEGER as conversations,
			COALESCE(SUM(ua.total_tokens_used), 0)::INTEGER as tokens,
			COALESCE(SUM(ua.thumbs_up_count), 0)::INTEGER as thumbs_up,
			COALESCE(SUM(ua.thumbs_down_count), 0)::INTEGER as thumbs_down,
			COALESCE(SUM(ua.handoff_count), 0)::INTEGER as handoffs
		FROM dates d
		LEFT JOIN user_analytics ua ON ua.analytics_date = d.date
		GROUP BY d.date
		ORDER BY d.date
	`

	rows, err := pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query global analytics")
	}
	defer func() { _ = rows.Close() }()

	var data []AnalyticsPoint
	for rows.Next() {
		var p AnalyticsPoint
		if err := rows.Scan(&p.Date, &p.Messages, &p.Conversations, &p.Tokens, &p.ThumbsUp, &p.ThumbsDown, &p.Handoffs); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan analytics point")
		}
		data = append(data, p)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "global analytics rows err")
	}
	return data, nil
}

type AnalyticsPoint struct {
	Date          string `json:"date"`
	Messages      int    `json:"messages"`
	Conversations int    `json:"conversations"`
	Tokens        int    `json:"tokens"`
	ThumbsUp      int    `json:"thumbs_up"`
	ThumbsDown    int    `json:"thumbs_down"`
	Handoffs      int    `json:"handoffs"`
}
