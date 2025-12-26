package scraper

// DefaultScraper implements the Scraper interface by delegating to existing
// package-level functions. This is the production implementation.
type DefaultScraper struct{}

// NewDefaultScraper creates a new DefaultScraper instance.
func NewDefaultScraper() *DefaultScraper {
	return &DefaultScraper{}
}

// ScrapeURLWithFallback implements Scraper.ScrapeURLWithFallback by delegating
// to the package-level ScrapeURLWithFallback function.
func (s *DefaultScraper) ScrapeURLWithFallback(task ScrapingTask, cfg CollectorConfig, allowDynamic bool, scrapeConfig *ScrapeConfig) (string, error) {
	return ScrapeURLWithFallback(task, cfg, allowDynamic, scrapeConfig)
}

// FetchRawHTML implements Scraper.FetchRawHTML by delegating to the
// package-level FetchRawHTML function.
func (s *DefaultScraper) FetchRawHTML(url string, cfg CollectorConfig) (string, error) {
	return FetchRawHTML(url, cfg)
}

// ExtractLinks implements Scraper.ExtractLinks by delegating to the
// package-level ExtractLinks function.
func (s *DefaultScraper) ExtractLinks(htmlContent, baseURL string, filter *PathFilter) ([]string, error) {
	return ExtractLinks(htmlContent, baseURL, filter)
}

// Compile-time check: DefaultScraper must implement Scraper interface
var _ Scraper = (*DefaultScraper)(nil)
