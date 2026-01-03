package scraper

// Scraper defines the interface for web scraping operations.
// This interface allows for dependency injection and mock-based testing.
type Scraper interface {
	// ScrapeURLWithFallback tries static scraping first, then falls back to dynamic if enabled.
	// If scrapeConfig is provided and contains Selectors, only content from matching elements is extracted.
	ScrapeURLWithFallback(task ScrapingTask, allowDynamic bool, scrapeConfig *ScrapeConfig) (string, error)

	// FetchRawHTML fetches raw HTML content from a URL for link discovery purposes.
	// This is separate from ScrapeURLWithFallback which extracts visible text.
	FetchRawHTML(url string) (string, error)

	// ExtractLinks finds all links in the HTML content that belong to the same domain as baseURL.
	// It returns a list of absolute URLs, optionally filtered by the provided PathFilter.
	ExtractLinks(htmlContent, baseURL string, filter *PathFilter) ([]string, error)
}
