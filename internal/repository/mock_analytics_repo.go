package repository

import (
	"context"
	"strings"
	"time"

	"github.com/onurceri/botla-app/internal/models"
)

// MockAnalyticsRepo is a mock implementation of AnalyticsRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockAnalyticsRepo struct {
	// GetOverviewFunc is called when GetOverview is invoked.
	GetOverviewFunc func(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error)

	// GetTrendsFunc is called when GetTrends is invoked.
	GetTrendsFunc func(ctx context.Context, chatbotID string, days int) ([]models.DailyAnalytics, error)

	// IncrementAnalyticsFunc is called when IncrementAnalytics is invoked.
	IncrementAnalyticsFunc func(ctx context.Context, chatbotID string, isNewConversation bool, tokens int, isHandoff bool, responseTimeMs int) error

	// IncrementFeedbackFunc is called when IncrementFeedback is invoked.
	IncrementFeedbackFunc func(ctx context.Context, chatbotID string, oldState *bool, newState bool) error

	// TrackUnansweredQueryFunc is called when TrackUnansweredQuery is invoked.
	TrackUnansweredQueryFunc func(ctx context.Context, chatbotID, queryText string) error

	// GetUnansweredQueriesFunc is called when GetUnansweredQueries is invoked.
	GetUnansweredQueriesFunc func(ctx context.Context, chatbotID string, limit, offset int) ([]string, error)

	// GetMonthlyTokenUsageFunc is called when GetMonthlyTokenUsage is invoked.
	GetMonthlyTokenUsageFunc func(ctx context.Context, userID string) (int, error)

	// GetAutoRefreshCountForMonthFunc is called when GetAutoRefreshCountForMonth is invoked.
	GetAutoRefreshCountForMonthFunc func(ctx context.Context, userID string, month time.Time) (int, error)

	// IncrementAutoRefreshCountFunc is called when IncrementAutoRefreshCount is invoked.
	IncrementAutoRefreshCountFunc func(ctx context.Context, userID string, month time.Time, delta int) error

	// GetGlobalAnalyticsFunc is called when GetGlobalAnalytics is invoked.
	GetGlobalAnalyticsFunc func(ctx context.Context, userID string, orgID, wsID *string) ([]AnalyticsPoint, error)

	// GetSourceUsageStatsFunc is called when GetSourceUsageStats is invoked.
	GetSourceUsageStatsFunc func(ctx context.Context, chatbotID string, days int) ([]SourceUsageStat, error)

	// UpdateMessageFeedbackFunc is called when UpdateMessageFeedback is invoked.
	UpdateMessageFeedbackFunc func(ctx context.Context, messageID string, thumbsUp bool) (string, bool, error)

	// Invocation tracking for test assertions
	Calls struct {
		GetOverview                 []GetOverviewCall
		GetTrends                   []GetTrendsCall
		IncrementAnalytics          []IncrementAnalyticsCall
		IncrementFeedback           []IncrementFeedbackCall
		TrackUnansweredQuery        []TrackUnansweredQueryCall
		GetUnansweredQueries        []GetUnansweredQueriesCall
		GetMonthlyTokenUsage        []GetMonthlyTokenUsageCall
		GetAutoRefreshCountForMonth []GetAutoRefreshCountForMonthCall
		IncrementAutoRefreshCount   []IncrementAutoRefreshCountCall
		GetGlobalAnalytics          []GetGlobalAnalyticsCall
		GetSourceUsageStats         []GetSourceUsageStatsCall
		UpdateMessageFeedback       []UpdateMessageFeedbackCall
	}
}

// Call recording types for test verification
type GetOverviewCall struct {
	ChatbotID string
}

type GetTrendsCall struct {
	ChatbotID string
	Days      int
}

type IncrementAnalyticsCall struct {
	ChatbotID         string
	IsNewConversation bool
	Tokens            int
	IsHandoff         bool
	ResponseTimeMs    int
}

type IncrementFeedbackCall struct {
	ChatbotID string
	OldState  *bool
	NewState  bool
}

type TrackUnansweredQueryCall struct {
	ChatbotID string
	QueryText string
}

type GetUnansweredQueriesCall struct {
	ChatbotID string
	Limit     int
	Offset    int
}

type GetMonthlyTokenUsageCall struct {
	UserID string
}

type GetAutoRefreshCountForMonthCall struct {
	UserID string
	Month  time.Time
}

type IncrementAutoRefreshCountCall struct {
	UserID string
	Month  time.Time
	Delta  int
}

type GetGlobalAnalyticsCall struct {
	UserID string
	OrgID  *string
	WsID   *string
}

type GetSourceUsageStatsCall struct {
	ChatbotID string
	Days      int
}

type UpdateMessageFeedbackCall struct {
	MessageID string
	ThumbsUp  bool
}

// Compile-time check that MockAnalyticsRepo implements AnalyticsRepository.
var _ AnalyticsRepository = (*MockAnalyticsRepo)(nil)

// NewMockAnalyticsRepo creates a new MockAnalyticsRepo with default no-op behavior.
func NewMockAnalyticsRepo() *MockAnalyticsRepo {
	return &MockAnalyticsRepo{}
}

// GetOverview returns aggregated stats for a chatbot.
func (m *MockAnalyticsRepo) GetOverview(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error) {
	m.Calls.GetOverview = append(m.Calls.GetOverview, GetOverviewCall{ChatbotID: chatbotID})
	if m.GetOverviewFunc != nil {
		return m.GetOverviewFunc(ctx, chatbotID)
	}
	return nil, nil
}

// GetTrends returns daily stats for a chatbot.
func (m *MockAnalyticsRepo) GetTrends(ctx context.Context, chatbotID string, days int) ([]models.DailyAnalytics, error) {
	m.Calls.GetTrends = append(m.Calls.GetTrends, GetTrendsCall{ChatbotID: chatbotID, Days: days})
	if m.GetTrendsFunc != nil {
		return m.GetTrendsFunc(ctx, chatbotID, days)
	}
	return nil, nil
}

// IncrementAnalytics updates analytics counters.
func (m *MockAnalyticsRepo) IncrementAnalytics(ctx context.Context, chatbotID string, isNewConversation bool, tokens int, isHandoff bool, responseTimeMs int) error {
	m.Calls.IncrementAnalytics = append(m.Calls.IncrementAnalytics, IncrementAnalyticsCall{
		ChatbotID:         chatbotID,
		IsNewConversation: isNewConversation,
		Tokens:            tokens,
		IsHandoff:         isHandoff,
		ResponseTimeMs:    responseTimeMs,
	})
	if m.IncrementAnalyticsFunc != nil {
		return m.IncrementAnalyticsFunc(ctx, chatbotID, isNewConversation, tokens, isHandoff, responseTimeMs)
	}
	return nil
}

// IncrementFeedback updates thumbs up/down counts.
func (m *MockAnalyticsRepo) IncrementFeedback(ctx context.Context, chatbotID string, oldState *bool, newState bool) error {
	m.Calls.IncrementFeedback = append(m.Calls.IncrementFeedback, IncrementFeedbackCall{
		ChatbotID: chatbotID,
		OldState:  oldState,
		NewState:  newState,
	})
	if m.IncrementFeedbackFunc != nil {
		return m.IncrementFeedbackFunc(ctx, chatbotID, oldState, newState)
	}
	return nil
}

// TrackUnansweredQuery records a low confidence query.
func (m *MockAnalyticsRepo) TrackUnansweredQuery(ctx context.Context, chatbotID, queryText string) error {
	// Mirror real implementation behavior: skip empty queries
	if strings.TrimSpace(queryText) == "" {
		return nil
	}
	m.Calls.TrackUnansweredQuery = append(m.Calls.TrackUnansweredQuery, TrackUnansweredQueryCall{
		ChatbotID: chatbotID,
		QueryText: queryText,
	})
	if m.TrackUnansweredQueryFunc != nil {
		return m.TrackUnansweredQueryFunc(ctx, chatbotID, queryText)
	}
	return nil
}

// GetUnansweredQueries returns unanswered queries for a chatbot.
func (m *MockAnalyticsRepo) GetUnansweredQueries(ctx context.Context, chatbotID string, limit, offset int) ([]string, error) {
	m.Calls.GetUnansweredQueries = append(m.Calls.GetUnansweredQueries, GetUnansweredQueriesCall{
		ChatbotID: chatbotID,
		Limit:     limit,
		Offset:    offset,
	})
	if m.GetUnansweredQueriesFunc != nil {
		return m.GetUnansweredQueriesFunc(ctx, chatbotID, limit, offset)
	}
	return nil, nil
}

// GetMonthlyTokenUsage returns monthly token usage for a user.
func (m *MockAnalyticsRepo) GetMonthlyTokenUsage(ctx context.Context, userID string) (int, error) {
	m.Calls.GetMonthlyTokenUsage = append(m.Calls.GetMonthlyTokenUsage, GetMonthlyTokenUsageCall{UserID: userID})
	if m.GetMonthlyTokenUsageFunc != nil {
		return m.GetMonthlyTokenUsageFunc(ctx, userID)
	}
	return 0, nil
}

// GetAutoRefreshCountForMonth returns the auto_refresh_count for a user in a given month.
func (m *MockAnalyticsRepo) GetAutoRefreshCountForMonth(ctx context.Context, userID string, month time.Time) (int, error) {
	m.Calls.GetAutoRefreshCountForMonth = append(m.Calls.GetAutoRefreshCountForMonth, GetAutoRefreshCountForMonthCall{UserID: userID, Month: month})
	if m.GetAutoRefreshCountForMonthFunc != nil {
		return m.GetAutoRefreshCountForMonthFunc(ctx, userID, month)
	}
	return 0, nil
}

// IncrementAutoRefreshCount increments the auto_refresh_count for a user in a given month.
func (m *MockAnalyticsRepo) IncrementAutoRefreshCount(ctx context.Context, userID string, month time.Time, delta int) error {
	m.Calls.IncrementAutoRefreshCount = append(m.Calls.IncrementAutoRefreshCount, IncrementAutoRefreshCountCall{UserID: userID, Month: month, Delta: delta})
	if m.IncrementAutoRefreshCountFunc != nil {
		return m.IncrementAutoRefreshCountFunc(ctx, userID, month, delta)
	}
	return nil
}

// GetGlobalAnalytics returns aggregated analytics for a scope.
func (m *MockAnalyticsRepo) GetGlobalAnalytics(ctx context.Context, userID string, orgID, wsID *string) ([]AnalyticsPoint, error) {
	m.Calls.GetGlobalAnalytics = append(m.Calls.GetGlobalAnalytics, GetGlobalAnalyticsCall{UserID: userID, OrgID: orgID, WsID: wsID})
	if m.GetGlobalAnalyticsFunc != nil {
		return m.GetGlobalAnalyticsFunc(ctx, userID, orgID, wsID)
	}
	return nil, nil
}

// GetSourceUsageStats returns source usage statistics for a chatbot.
func (m *MockAnalyticsRepo) GetSourceUsageStats(ctx context.Context, chatbotID string, days int) ([]SourceUsageStat, error) {
	m.Calls.GetSourceUsageStats = append(m.Calls.GetSourceUsageStats, GetSourceUsageStatsCall{ChatbotID: chatbotID, Days: days})
	if m.GetSourceUsageStatsFunc != nil {
		return m.GetSourceUsageStatsFunc(ctx, chatbotID, days)
	}
	return nil, nil
}

// UpdateMessageFeedback updates feedback for a message and returns affected chatbot ID.
func (m *MockAnalyticsRepo) UpdateMessageFeedback(ctx context.Context, messageID string, thumbsUp bool) (string, bool, error) {
	m.Calls.UpdateMessageFeedback = append(m.Calls.UpdateMessageFeedback, UpdateMessageFeedbackCall{MessageID: messageID, ThumbsUp: thumbsUp})
	if m.UpdateMessageFeedbackFunc != nil {
		return m.UpdateMessageFeedbackFunc(ctx, messageID, thumbsUp)
	}
	return "", false, nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockAnalyticsRepo) Reset() {
	m.Calls.GetOverview = nil
	m.Calls.GetTrends = nil
	m.Calls.IncrementAnalytics = nil
	m.Calls.IncrementFeedback = nil
	m.Calls.TrackUnansweredQuery = nil
	m.Calls.GetUnansweredQueries = nil
	m.Calls.GetMonthlyTokenUsage = nil
	m.Calls.GetAutoRefreshCountForMonth = nil
	m.Calls.IncrementAutoRefreshCount = nil
}
