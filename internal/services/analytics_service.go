package services

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/logger"
)

type AnalyticsService struct {
	DB  *sql.DB
	Log *logger.Logger
}

func NewAnalyticsService(db *sql.DB, log *logger.Logger) *AnalyticsService {
	return &AnalyticsService{
		DB:  db,
		Log: log,
	}
}

// GetChatbotOverview returns aggregated analytics for a specific chatbot
func (s *AnalyticsService) GetChatbotOverview(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error) {
	return db.GetAnalyticsOverview(ctx, s.DB, chatbotID)
}

// GetChatbotTrends returns daily trends for a chatbot
func (s *AnalyticsService) GetChatbotTrends(ctx context.Context, chatbotID string, days int) (*models.TrendData, error) {
	daily, err := db.GetAnalyticsTrends(ctx, s.DB, chatbotID, days)
	if err != nil {
		return nil, err
	}
	return &models.TrendData{Daily: daily}, nil
}

// TrackUnansweredQuery records a low confidence query
func (s *AnalyticsService) TrackUnansweredQuery(ctx context.Context, chatbotID, query string) error {
	return db.TrackUnansweredQuery(ctx, s.DB, chatbotID, query)
}
