package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SourceUsageStats represents usage statistics for a data source
type SourceUsageStats struct {
	SourceID         string    `json:"source_id"`
	SourceName       string    `json:"source_name"`
	SourceType       string    `json:"source_type"`
	TimesUsed        int       `json:"times_used"`
	AvgRelevance     float64   `json:"avg_relevance"`
	PositiveFeedback int       `json:"positive_feedback"`
	NegativeFeedback int       `json:"negative_feedback"`
	LastUsed         time.Time `json:"last_used"`
}

// GetSourceUsageStats returns usage statistics for sources of a chatbot
func GetSourceUsageStats(ctx context.Context, pool *sql.DB, chatbotID string, days int) ([]SourceUsageStats, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	query := `
        SELECT 
            ds.id as source_id,
            COALESCE(ds.original_filename, ds.source_url, 'Unknown Source') as source_name,
            ds.source_type,
            COUNT(DISTINCT ms.message_id) as times_used,
            AVG(ms.relevance_score) as avg_relevance,
            COUNT(CASE WHEN m.thumbs_up = true THEN 1 END) as positive_feedback,
            COUNT(CASE WHEN m.thumbs_up = false THEN 1 END) as negative_feedback,
            MAX(ms.created_at) as last_used
        FROM data_sources ds
        INNER JOIN message_sources ms ON ds.id = ms.source_id
        INNER JOIN messages m ON ms.message_id = m.id
        INNER JOIN conversations c ON m.conversation_id = c.id
        WHERE c.chatbot_id = $1
          AND ms.created_at >= $2
        GROUP BY ds.id, ds.original_filename, ds.source_url, ds.source_type
        ORDER BY times_used DESC
    `

	rows, err := pool.QueryContext(ctx, query, chatbotID, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query source usage stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []SourceUsageStats
	for rows.Next() {
		var s SourceUsageStats
		if err := rows.Scan(&s.SourceID, &s.SourceName, &s.SourceType, &s.TimesUsed,
			&s.AvgRelevance, &s.PositiveFeedback, &s.NegativeFeedback, &s.LastUsed); err != nil {
			return nil, fmt.Errorf("scan source usage stat: %w", err)
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("source usage stats rows err: %w", err)
	}
	return stats, nil
}
