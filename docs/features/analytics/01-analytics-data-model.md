> Extracted from: analytics.md — Section: 1.1 Analytics Data Model

## 1.1 Analytics Data Model

- [ ] `DailyAnalytics` fields: `Date`, `TotalMessages`, `TotalConversations`, `TotalTokensUsed`, `ThumbsUpCount`, `ThumbsDownCount`, `HandoffCount`, `AvgResponseTimeMs`
- [ ] `AnalyticsOverview` fields: `TotalMessages`, `TotalConversations`, `TotalTokensUsed`, `PositiveFeedback`, `NegativeFeedback`, `FeedbackRate`, `HandoffCount`
- [ ] `TrendData` wraps `[]DailyAnalytics` in `Daily` field
- [ ] Data aggregated per chatbot per day in `analytics` table
- [ ] Overview computed over last 30 days (`CURRENT_DATE - INTERVAL '29 days'`)
- [ ] Trends support configurable lookback window (1–365 days)

