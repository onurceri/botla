package scraper

import (
	"fmt"

	pkgErrors "github.com/onurceri/botla-app/pkg/errors"
)

// ScrapeError represents an error during scraping
type ScrapeError struct {
	StatusCode int
	URL        string
	Err        error
}

func (e *ScrapeError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("HTTP %d: %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("Scraping failed: %v", e.Err)
}

// Unwrap returns the underlying error, wrapped with a sentinel error
// based on the HTTP status code for type-safe error checking.
func (e *ScrapeError) Unwrap() error {
	if e.Err == nil {
		return nil
	}
	// Wrap with sentinel error based on status code
	switch e.StatusCode {
	case 429: // Too Many Requests
		return fmt.Errorf("%w: %w", pkgErrors.ErrRateLimit, e.Err)
	case 404:
		return fmt.Errorf("%w: %w", pkgErrors.ErrNotFound, e.Err)
	case 500, 502, 503, 504:
		return fmt.Errorf("%w: %w", pkgErrors.ErrNetwork, e.Err)
	default:
		return e.Err
	}
}
