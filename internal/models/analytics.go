package models

type DailyAnalytics struct {
	Date               string `json:"date"`
	TotalMessages      int    `json:"total_messages"`
	TotalConversations int    `json:"total_conversations"`
	TotalTokensUsed    int    `json:"total_tokens_used"`
	ThumbsUpCount      int    `json:"thumbs_up_count"`
	ThumbsDownCount    int    `json:"thumbs_down_count"`
	HandoffCount       int    `json:"handoff_count"`
	AvgResponseTimeMs  *int   `json:"avg_response_time_ms,omitempty"`
}

type AnalyticsOverview struct {
	TotalMessages      int     `json:"total_messages"`
	TotalConversations int     `json:"total_conversations"`
	TotalTokensUsed    int     `json:"total_tokens_used"`
	PositiveFeedback   int     `json:"positive_feedback"`
	NegativeFeedback   int     `json:"negative_feedback"`
	FeedbackRate       float64 `json:"feedback_rate"` // (positive / total) * 100
	HandoffCount       int     `json:"handoff_count"`
}

type TrendData struct {
	Daily []DailyAnalytics `json:"daily"`
}
