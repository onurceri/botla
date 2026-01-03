package processing

import (
	"testing"

	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/pkg/logger"
)

func TestURLProcessor_NewURLProcessor(t *testing.T) {
	t.Parallel()

	t.Run("creates processor with injected scraper", func(t *testing.T) {
		ms := scraper.NewMockScraper()
		p := NewURLProcessor(nil, nil, nil, nil, ms)
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
		p := NewURLProcessor(nil, nil, nil, log, ms)
		if p == nil {
			t.Error("expected non-nil processor")
		}
		if p.Log != log {
			t.Error("expected logger to be set")
		}
	})
}
