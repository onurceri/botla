package processing

import (
	"testing"

	"github.com/onurceri/botla-co/pkg/logger"
)

func TestURLProcessor_NewURLProcessor(t *testing.T) {
	t.Parallel()

	t.Run("creates processor with default scraper", func(t *testing.T) {
		p := NewURLProcessor(nil, nil, nil, nil, nil)
		if p == nil {
			t.Error("expected non-nil processor")
		}
		if p.Scraper == nil {
			t.Error("expected non-nil default scraper")
		}
	})

	t.Run("creates processor with custom logger", func(t *testing.T) {
		log := logger.New("test")
		p := NewURLProcessor(nil, nil, nil, log, nil)
		if p == nil {
			t.Error("expected non-nil processor")
		}
		if p.Log != log {
			t.Error("expected logger to be set")
		}
	})
}
