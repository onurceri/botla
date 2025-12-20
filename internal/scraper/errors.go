package scraper

import "fmt"

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

func (e *ScrapeError) Unwrap() error {
	return e.Err
}
