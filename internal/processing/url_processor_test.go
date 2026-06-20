package processing

import (
	"testing"

	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/scraper"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/tokenizer"
)

func TestURLProcessor_NewURLProcessor(t *testing.T) {
	t.Parallel()

	t.Run("creates processor with injected scraper", func(t *testing.T) {
		ms := scraper.NewMockScraper()
		loader := &tokenizer.Loader{}
		p := NewURLProcessor(nil, nil, nil, nil, nil, nil, nil, ms, loader)
		if p == nil {
			t.Error("expected non-nil processor")
		}
		if p.Scraper != ms {
			t.Error("expected injected scraper to be set")
		}
	})

	t.Run("creates processor with custom logger", func(t *testing.T) {
		log := logger.New("test")
		ms := scraper.NewMockScraper()
		sourceRepo := repository.NewMockSourceRepo()
		usageRepo := &MockUsageRepo{}
		pendingURLRepo := &MockPendingURLRepo{}
		planRepo := repository.NewMockPlanRepo()
		loader := &tokenizer.Loader{}
		p := NewURLProcessor(sourceRepo, usageRepo, pendingURLRepo, planRepo, nil, nil, log, ms, loader)
		if p == nil {
			t.Error("expected non-nil processor")
		}
		if p.Log != log {
			t.Error("expected logger to be set")
		}
	})
}
