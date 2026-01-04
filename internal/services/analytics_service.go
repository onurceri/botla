package services

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/logger"
)

type AnalyticsService struct {
	AnalyticsRepo repository.AnalyticsRepository
	Log           *logger.Logger
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository, log *logger.Logger) *AnalyticsService {
	return &AnalyticsService{
		AnalyticsRepo: analyticsRepo,
		Log:           log,
	}
}

// GetChatbotOverview returns aggregated analytics for a specific chatbot
func (s *AnalyticsService) GetChatbotOverview(ctx context.Context, chatbotID string) (*models.AnalyticsOverview, error) {
	res, err := s.AnalyticsRepo.GetOverview(ctx, chatbotID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get analytics overview")
	}
	return res, nil
}

// GetChatbotTrends returns daily trends for a chatbot
func (s *AnalyticsService) GetChatbotTrends(ctx context.Context, chatbotID string, days int) (*models.TrendData, error) {
	daily, err := s.AnalyticsRepo.GetTrends(ctx, chatbotID, days)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get analytics trends")
	}
	return &models.TrendData{Daily: daily}, nil
}

// TrackUnansweredQuery records a low confidence query
func (s *AnalyticsService) TrackUnansweredQuery(ctx context.Context, chatbotID, query string) error {
	if err := s.AnalyticsRepo.TrackUnansweredQuery(ctx, chatbotID, query); err != nil {
		return pkgerrors.Wrapf(err, "track unanswered query")
	}
	return nil
}
