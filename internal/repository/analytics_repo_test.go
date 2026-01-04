package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
)

// TestMockAnalyticsRepo_InterfaceCompliance ensures MockAnalyticsRepo implements AnalyticsRepository
func TestMockAnalyticsRepo_InterfaceCompliance(t *testing.T) {
	var _ AnalyticsRepository = (*MockAnalyticsRepo)(nil)
}

// TestNewMockAnalyticsRepo verifies that NewMockAnalyticsRepo creates a valid mock
func TestNewMockAnalyticsRepo(t *testing.T) {
	mock := NewMockAnalyticsRepo()
	if mock == nil {
		t.Fatal("NewMockAnalyticsRepo returned nil")
	}
}

// TestNewPostgresAnalyticsRepo verifies that NewPostgresAnalyticsRepo creates a valid repo
func TestNewPostgresAnalyticsRepo(t *testing.T) {
	repo := NewPostgresAnalyticsRepo(nil)
	if repo == nil {
		t.Fatal("NewPostgresAnalyticsRepo returned nil")
	}
}

// TestMockAnalyticsRepo_GetOverview tests the GetOverview mock functionality
func TestMockAnalyticsRepo_GetOverview(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		result, err := mock.GetOverview(context.Background(), "chatbot-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})

	t.Run("custom function returns overview", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		expectedOverview := &models.AnalyticsOverview{
			TotalMessages:      100,
			TotalConversations: 50,
			TotalTokensUsed:    50000,
			PositiveFeedback:   20,
			NegativeFeedback:   5,
			HandoffCount:       3,
			FeedbackRate:       80.0,
		}
		mock.GetOverviewFunc = func(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error) {
			return expectedOverview, nil
		}

		result, err := mock.GetOverview(context.Background(), "chatbot-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != expectedOverview {
			t.Errorf("expected overview, got: %v", result)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		expectedErr := errors.New("database connection failed")
		mock.GetOverviewFunc = func(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error) {
			return nil, expectedErr
		}

		_, err := mock.GetOverview(context.Background(), "any-id")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockAnalyticsRepo_GetTrends tests the GetTrends mock functionality
func TestMockAnalyticsRepo_GetTrends(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		result, err := mock.GetTrends(context.Background(), "chatbot-123", 30)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})

	t.Run("custom function returns trends", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		expectedTrends := []models.DailyAnalytics{
			{Date: "2024-01-01", TotalMessages: 10},
			{Date: "2024-01-02", TotalMessages: 15},
		}
		mock.GetTrendsFunc = func(ctx context.Context, chatbotID string, days int) ([]models.DailyAnalytics, error) {
			return expectedTrends, nil
		}

		result, err := mock.GetTrends(context.Background(), "chatbot-123", 2)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 trends, got: %d", len(result))
		}
	})
}

// TestMockAnalyticsRepo_IncrementAnalytics tests the IncrementAnalytics mock functionality
func TestMockAnalyticsRepo_IncrementAnalytics(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		err := mock.IncrementAnalytics(context.Background(), "chatbot-123", true, 100, false, 50)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		expectedErr := errors.New("increment failed")
		mock.IncrementAnalyticsFunc = func(ctx context.Context, chatbotID string, isNewConversation bool, tokens int, isHandoff bool, responseTimeMs int) error {
			return expectedErr
		}

		err := mock.IncrementAnalytics(context.Background(), "chatbot-123", true, 100, false, 50)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockAnalyticsRepo_TrackUnansweredQuery tests the TrackUnansweredQuery mock functionality
func TestMockAnalyticsRepo_TrackUnansweredQuery(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		err := mock.TrackUnansweredQuery(context.Background(), "chatbot-123", "test query")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		expectedErr := errors.New("track failed")
		mock.TrackUnansweredQueryFunc = func(ctx context.Context, chatbotID, queryText string) error {
			return expectedErr
		}

		err := mock.TrackUnansweredQuery(context.Background(), "chatbot-123", "test query")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})

	t.Run("empty query returns nil without calling func", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		called := false
		mock.TrackUnansweredQueryFunc = func(ctx context.Context, chatbotID, queryText string) error {
			called = true
			return nil
		}

		err := mock.TrackUnansweredQuery(context.Background(), "chatbot-123", "")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if called {
			t.Error("expected function not to be called for empty query")
		}
	})
}

// TestMockAnalyticsRepo_GetMonthlyTokenUsage tests the GetMonthlyTokenUsage mock functionality
func TestMockAnalyticsRepo_GetMonthlyTokenUsage(t *testing.T) {
	t.Run("default returns 0", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		result, err := mock.GetMonthlyTokenUsage(context.Background(), "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != 0 {
			t.Errorf("expected 0, got: %d", result)
		}
	})

	t.Run("custom function returns value", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		mock.GetMonthlyTokenUsageFunc = func(ctx context.Context, userID string) (int, error) {
			return 100000, nil
		}

		result, err := mock.GetMonthlyTokenUsage(context.Background(), "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != 100000 {
			t.Errorf("expected 100000, got: %d", result)
		}
	})
}

// TestMockAnalyticsRepo_GetUnansweredQueries tests the GetUnansweredQueries mock functionality
func TestMockAnalyticsRepo_GetUnansweredQueries(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		result, err := mock.GetUnansweredQueries(context.Background(), "chatbot-123", 10, 0)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})

	t.Run("custom function returns queries", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		expectedQueries := []string{"query1", "query2"}
		mock.GetUnansweredQueriesFunc = func(ctx context.Context, chatbotID string, limit, offset int) ([]string, error) {
			return expectedQueries, nil
		}

		result, err := mock.GetUnansweredQueries(context.Background(), "chatbot-123", 10, 0)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 queries, got: %d", len(result))
		}
	})
}

// TestMockAnalyticsRepo_IncrementFeedback tests the IncrementFeedback mock functionality
func TestMockAnalyticsRepo_IncrementFeedback(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		err := mock.IncrementFeedback(context.Background(), "chatbot-123", nil, true)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		expectedErr := errors.New("increment feedback failed")
		mock.IncrementFeedbackFunc = func(ctx context.Context, chatbotID string, oldState *bool, newState bool) error {
			return expectedErr
		}

		err := mock.IncrementFeedback(context.Background(), "chatbot-123", nil, true)
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})

	t.Run("with oldState nil", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		var calledWithOldState *bool
		mock.IncrementFeedbackFunc = func(ctx context.Context, chatbotID string, oldState *bool, newState bool) error {
			calledWithOldState = oldState
			return nil
		}

		mock.IncrementFeedback(context.Background(), "chatbot-123", nil, true)
		if calledWithOldState != nil {
			t.Error("expected oldState to be nil")
		}
	})

	t.Run("with oldState set", func(t *testing.T) {
		mock := NewMockAnalyticsRepo()
		var capturedOldState *bool
		oldVal := true
		mock.IncrementFeedbackFunc = func(ctx context.Context, chatbotID string, oldState *bool, newState bool) error {
			capturedOldState = oldState
			return nil
		}

		mock.IncrementFeedback(context.Background(), "chatbot-123", &oldVal, false)
		if capturedOldState == nil {
			t.Error("expected oldState to be passed")
		}
		if capturedOldState == nil || *capturedOldState != true {
			t.Error("expected oldState to be true")
		}
	})
}
