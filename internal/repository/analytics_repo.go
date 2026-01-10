// Package repository provides data access layer implementations for analytics.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresAnalyticsRepo implements AnalyticsRepository using PostgreSQL.
type PostgresAnalyticsRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresAnalyticsRepo implements AnalyticsRepository.
var _ AnalyticsRepository = (*PostgresAnalyticsRepo)(nil)

// NewPostgresAnalyticsRepo creates a new PostgresAnalyticsRepo instance.
func NewPostgresAnalyticsRepo(pool *sql.DB) *PostgresAnalyticsRepo {
	return &PostgresAnalyticsRepo{pool: pool}
}

// GetOverview returns aggregated stats for a chatbot for the last 30 days.
func (r *PostgresAnalyticsRepo) GetOverview(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error) {
	query, args, err := psql.
		Select(
			"COALESCE(SUM(total_messages), 0)::INT",
			"COALESCE(SUM(total_conversations), 0)::INT",
			"COALESCE(SUM(total_tokens_used), 0)::INT",
			"COALESCE(SUM(thumbs_up_count), 0)::INT",
			"COALESCE(SUM(thumbs_down_count), 0)::INT",
			"COALESCE(SUM(handoff_count), 0)::INT",
		).
		From("analytics").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		Where("analytics_date >= CURRENT_DATE - INTERVAL '29 days'").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get overview query")
	}

	var stats models.AnalyticsOverview
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
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

// GetTrends returns daily stats for a chatbot for the last N days.
func (r *PostgresAnalyticsRepo) GetTrends(ctx context.Context, chatbotID string, days int) ([]models.DailyAnalytics, error) {
	if days <= 0 {
		days = 30
	}

	// Generate date series and join with analytics
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
			a.avg_response_time_ms
		FROM dates d
		LEFT JOIN analytics a ON a.analytics_date = d.date AND a.chatbot_id = $1
		ORDER BY d.date
	`

	rows, err := r.pool.QueryContext(ctx, query, chatbotID, days-1)
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
		return nil, pkgerrors.Wrapf(err, "analytics trends rows error")
	}
	return results, nil
}

// IncrementAnalytics updates the analytics table for a given chatbot and date.
func (r *PostgresAnalyticsRepo) IncrementAnalytics(ctx context.Context, chatbotID string, isNewConversation bool, tokens int, isHandoff bool, responseTimeMs int) error {
	msgInc := 2 // User + Assistant
	convInc := 0
	if isNewConversation {
		convInc = 1
	}

	handoffInc := 0
	if isHandoff {
		handoffInc = 1
	}

	query, args, err := psql.
		Insert("analytics").
		Columns("chatbot_id", "analytics_date", "total_messages", "total_conversations", "total_tokens_used", "handoff_count").
		Values(chatbotID, sq.Expr("CURRENT_DATE"), msgInc, convInc, tokens, handoffInc).
		Suffix(`
			ON CONFLICT (chatbot_id, analytics_date) DO UPDATE SET
				total_messages = analytics.total_messages + EXCLUDED.total_messages,
				total_conversations = analytics.total_conversations + EXCLUDED.total_conversations,
				total_tokens_used = analytics.total_tokens_used + EXCLUDED.total_tokens_used,
				handoff_count = analytics.handoff_count + EXCLUDED.handoff_count
		`).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build increment analytics query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment analytics")
	}
	return nil
}

// IncrementFeedback updates the thumbs up/down count in analytics.
func (r *PostgresAnalyticsRepo) IncrementFeedback(ctx context.Context, chatbotID string, oldState *bool, newState bool) error {
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

	query, args, err := psql.
		Insert("analytics").
		Columns("chatbot_id", "analytics_date", "thumbs_up_count", "thumbs_down_count").
		Values(chatbotID, sq.Expr("CURRENT_DATE"), upInc, downInc).
		Suffix(`
			ON CONFLICT (chatbot_id, analytics_date) DO UPDATE SET
				thumbs_up_count = analytics.thumbs_up_count + EXCLUDED.thumbs_up_count,
				thumbs_down_count = analytics.thumbs_down_count + EXCLUDED.thumbs_down_count
		`).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build increment feedback query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment feedback")
	}
	return nil
}

// TrackUnansweredQuery records a query that had low confidence.
func (r *PostgresAnalyticsRepo) TrackUnansweredQuery(ctx context.Context, chatbotID, queryText string) error {
	queryText = strings.TrimSpace(queryText)
	if queryText == "" {
		return nil
	}

	query, args, err := psql.
		Insert("unanswered_queries").
		Columns("chatbot_id", "query", "occurrence_count", "last_occurred_at").
		Values(chatbotID, queryText, 1, sq.Expr("NOW()")).
		Suffix(`
			ON CONFLICT (chatbot_id, query) DO UPDATE SET
				occurrence_count = unanswered_queries.occurrence_count + 1,
				last_occurred_at = NOW()
		`).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build track unanswered query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "track unanswered query")
	}
	return nil
}

// GetUnansweredQueries returns unanswered queries for a chatbot with pagination.
func (r *PostgresAnalyticsRepo) GetUnansweredQueries(ctx context.Context, chatbotID string, limit, offset int) ([]string, error) {
	query, args, err := psql.
		Select("query").
		From("unanswered_queries").
		Where(sq.Eq{"chatbot_id": chatbotID}).
		OrderBy("occurrence_count DESC, last_occurred_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get unanswered queries query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query unanswered queries")
	}
	defer func() { _ = rows.Close() }()

	var queries []string
	for rows.Next() {
		var queryText string
		if err := rows.Scan(&queryText); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan unanswered query")
		}
		queries = append(queries, queryText)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "unanswered queries rows error")
	}
	return queries, nil
}

// GetMonthlyTokenUsage returns the total tokens used by all chatbots of a user in the current month.
func (r *PostgresAnalyticsRepo) GetMonthlyTokenUsage(ctx context.Context, userID string) (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startStr := startOfMonth.Format("2006-01-02")

	query, args, err := psql.
		Select("COALESCE(SUM(a.total_tokens_used), 0)").
		From("analytics a").
		Join("chatbots c ON a.chatbot_id = c.id").
		Where(sq.Eq{"c.user_id": userID}).
		Where(sq.Expr("a.analytics_date >= $1", startStr)).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build get monthly token usage query")
	}

	var total int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "get monthly token usage")
	}
	return total, nil
}

// GetAutoRefreshCountForMonth returns the auto_refresh_count for a user in a given month.
func (r *PostgresAnalyticsRepo) GetAutoRefreshCountForMonth(ctx context.Context, userID string, month time.Time) (int, error) {
	pm := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	query, args, err := psql.
		Select("COALESCE(auto_refresh_count, 0)").
		From("usage_ingestions").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"period_month": pm}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build get auto refresh count query")
	}

	var count int
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "get auto refresh count")
	}
	return count, nil
}

// IncrementAutoRefreshCount increments the auto_refresh_count for a user in a given month.
func (r *PostgresAnalyticsRepo) IncrementAutoRefreshCount(ctx context.Context, userID string, month time.Time, delta int) error {
	pm := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	query, args, err := psql.
		Insert("usage_ingestions").
		Columns("user_id", "period_month", "auto_refresh_count", "sources_count", "embedding_tokens", "updated_at").
		Values(userID, pm, delta, 0, 0, sq.Expr("NOW()")).
		Suffix(`
			ON CONFLICT (user_id, period_month) DO UPDATE SET
				auto_refresh_count = usage_ingestions.auto_refresh_count + EXCLUDED.auto_refresh_count,
				updated_at = NOW()
		`).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build increment auto refresh count query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "increment auto refresh count")
	}
	return nil
}

// GetGlobalAnalytics returns aggregated analytics for a scope (User, Workspace, or Org).
func (r *PostgresAnalyticsRepo) GetGlobalAnalytics(ctx context.Context, userID string, orgID, wsID *string) ([]AnalyticsPoint, error) {
	var whereClause string
	var args []interface{}

	switch {
	case wsID != nil:
		whereClause = "c.workspace_id = $1"
		args = append(args, *wsID)
	case orgID != nil:
		whereClause = "(c.organization_id = $1 OR c.workspace_id IN (SELECT id FROM workspaces WHERE organization_id = $1))"
		args = append(args, *orgID)
	default:
		whereClause = "c.user_id = $1 AND c.workspace_id IS NULL AND c.organization_id IS NULL"
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
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
			WHERE (%s) AND a.analytics_date >= CURRENT_DATE - INTERVAL '29 days'
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
	`, whereClause)

	rows, err := r.pool.QueryContext(ctx, query, args...)
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
		return nil, pkgerrors.Wrapf(err, "global analytics rows error")
	}
	return data, nil
}

// GetSourceUsageStats returns source usage analytics for a chatbot.
func (r *PostgresAnalyticsRepo) GetSourceUsageStats(ctx context.Context, chatbotID string, days int) ([]SourceUsageStat, error) {
	query := `
		SELECT 
			ds.id,
			ds.source_type,
			ds.source_url,
			ds.original_filename,
			COALESCE(SUM(a.total_messages), 0)::INTEGER as message_count
		FROM data_sources ds
		LEFT JOIN messages m ON m.source_id = ds.id
		LEFT JOIN analytics a ON a.chatbot_id = ds.chatbot_id AND a.analytics_date >= CURRENT_DATE - make_interval(days => $2)
		WHERE ds.chatbot_id = $1 AND ds.deleted_at IS NULL
		GROUP BY ds.id, ds.source_type, ds.source_url, ds.original_filename
		ORDER BY message_count DESC
	`

	rows, err := r.pool.QueryContext(ctx, query, chatbotID, days)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query source usage stats")
	}
	defer func() { _ = rows.Close() }()

	var stats []SourceUsageStat
	for rows.Next() {
		var stat SourceUsageStat
		if err := rows.Scan(&stat.SourceID, &stat.SourceType, &stat.SourceURL, &stat.OriginalFilename, &stat.MessageCount); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan source usage stat")
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "source usage stats rows error")
	}
	return stats, nil
}

func (r *PostgresAnalyticsRepo) UpdateMessageFeedback(ctx context.Context, messageID string, thumbsUp bool) (string, *bool, error) {
	var chatbotIDVar string
	var oldThumbsUpVar sql.NullBool

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return "", nil, pkgerrors.Wrapf(err, "begin tx")
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, `
		SELECT m.thumbs_up, c.chatbot_id
		FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE m.id = $1 FOR UPDATE
	`, messageID).Scan(&oldThumbsUpVar, &chatbotIDVar)
	if err != nil {
		return "", nil, pkgerrors.Wrapf(err, "query current feedback state")
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE messages SET thumbs_up = $2
		WHERE id = $1
	`, messageID, thumbsUp)
	if err != nil {
		return "", nil, pkgerrors.Wrapf(err, "update message feedback")
	}

	if err := tx.Commit(); err != nil {
		return "", nil, pkgerrors.Wrapf(err, "commit tx")
	}

	var oldVal *bool
	if oldThumbsUpVar.Valid {
		v := oldThumbsUpVar.Bool
		oldVal = &v
	}
	return chatbotIDVar, oldVal, nil
}
