package scraper

// DefaultScraper implements the Scraper interface by delegating to existing
// package-level functions. This is the production implementation.
type DefaultScraper struct {
	cfg      CollectorConfig
	bScraper *BrowserScraper
}

// NewDefaultScraper creates a new DefaultScraper instance.
func NewDefaultScraper(cfg CollectorConfig, bScraper *BrowserScraper) *DefaultScraper {
	return &DefaultScraper{cfg: cfg, bScraper: bScraper}
}

// ScrapeURLWithFallback implements Scraper.ScrapeURLWithFallback by delegating
// to the package-level ScrapeURLWithFallback function.
func (s *DefaultScraper) ScrapeURLWithFallback(task ScrapingTask, allowDynamic bool, scrapeConfig *ScrapeConfig) (string, error) {
	return ScrapeURLWithFallback(task, s.cfg, s.bScraper, allowDynamic, scrapeConfig)
}

// FetchRawHTML implements Scraper.FetchRawHTML by delegating to the
// package-level FetchRawHTML function.
func (s *DefaultScraper) FetchRawHTML(url string) (string, error) {
	return FetchRawHTML(url, s.cfg)
}

// ExtractLinks implements Scraper.ExtractLinks by delegating to the
// package-level ExtractLinks function.
func (s *DefaultScraper) ExtractLinks(htmlContent, baseURL string, filter *PathFilter) ([]string, error) {
	return ExtractLinks(htmlContent, baseURL, filter)
}

// Compile-time check: DefaultScraper must implement Scraper interface
var _ Scraper = (*DefaultScraper)(nil)
