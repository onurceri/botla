package scraper

import "errors"

// MockCall records a call to a mock method for verification.
type MockCall struct {
	Method string
	Args   []interface{}
}

// MockScraper is a configurable mock implementation of the Scraper interface for testing.
type MockScraper struct {
	// Function overrides for custom behavior per test
	ScrapeURLWithFallbackFunc func(task ScrapingTask, allowDynamic bool, scrapeConfig *ScrapeConfig) (string, error)
	FetchRawHTMLFunc          func(url string) (string, error)
	ExtractLinksFunc          func(htmlContent, baseURL string, filter *PathFilter) ([]string, error)

	// Responses maps URLs to predefined content responses
	Responses map[string]string

	// HTMLResponses maps URLs to predefined raw HTML responses
	HTMLResponses map[string]string

	// LinkResponses maps URLs to predefined link lists
	LinkResponses map[string][]string

	// Errors maps URLs to predefined errors
	Errors map[string]error

	// Calls tracks all method invocations for verification
	Calls []MockCall
}

// NewMockScraper creates a new MockScraper with initialized maps.
func NewMockScraper() *MockScraper {
	return &MockScraper{
		Responses:     make(map[string]string),
		HTMLResponses: make(map[string]string),
		LinkResponses: make(map[string][]string),
		Errors:        make(map[string]error),
		Calls:         make([]MockCall, 0),
	}
}

// ScrapeURLWithFallback implements Scraper.ScrapeURLWithFallback.
// It first checks for a custom function, then for configured error, then for configured response.
// If none are found, it returns default mock content.
func (m *MockScraper) ScrapeURLWithFallback(task ScrapingTask, allowDynamic bool, scrapeConfig *ScrapeConfig) (string, error) {
	m.Calls = append(m.Calls, MockCall{
		Method: "ScrapeURLWithFallback",
		Args:   []interface{}{task, allowDynamic, scrapeConfig},
	})

	if m.ScrapeURLWithFallbackFunc != nil {
		return m.ScrapeURLWithFallbackFunc(task, allowDynamic, scrapeConfig)
	}

	if err, ok := m.Errors[task.URL]; ok {
		return "", err
	}

	if resp, ok := m.Responses[task.URL]; ok {
		return resp, nil
	}

	return "Mock scraped content for " + task.URL, nil
}

// FetchRawHTML implements Scraper.FetchRawHTML.
func (m *MockScraper) FetchRawHTML(url string) (string, error) {
	m.Calls = append(m.Calls, MockCall{
		Method: "FetchRawHTML",
		Args:   []interface{}{url},
	})

	if m.FetchRawHTMLFunc != nil {
		return m.FetchRawHTMLFunc(url)
	}

	if err, ok := m.Errors[url]; ok {
		return "", err
	}

	if html, ok := m.HTMLResponses[url]; ok {
		return html, nil
	}

	return "<html><body><h1>Mock HTML</h1></body></html>", nil
}

// ExtractLinks implements Scraper.ExtractLinks.
func (m *MockScraper) ExtractLinks(htmlContent, baseURL string, filter *PathFilter) ([]string, error) {
	m.Calls = append(m.Calls, MockCall{
		Method: "ExtractLinks",
		Args:   []interface{}{htmlContent, baseURL, filter},
	})

	if m.ExtractLinksFunc != nil {
		return m.ExtractLinksFunc(htmlContent, baseURL, filter)
	}

	if err, ok := m.Errors[baseURL]; ok {
		return nil, err
	}

	if links, ok := m.LinkResponses[baseURL]; ok {
		return links, nil
	}

	return []string{}, nil
}

// SetResponse configures a content response for a specific URL.
func (m *MockScraper) SetResponse(url, content string) {
	m.Responses[url] = content
}

// SetHTMLResponse configures a raw HTML response for a specific URL.
func (m *MockScraper) SetHTMLResponse(url, html string) {
	m.HTMLResponses[url] = html
}

// SetLinks configures link extraction response for a specific base URL.
func (m *MockScraper) SetLinks(baseURL string, links []string) {
	m.LinkResponses[baseURL] = links
}

// SetError configures an error response for a specific URL.
func (m *MockScraper) SetError(url string, err error) {
	m.Errors[url] = err
}

// Reset clears all configured responses, errors, and call tracking.
func (m *MockScraper) Reset() {
	m.Responses = make(map[string]string)
	m.HTMLResponses = make(map[string]string)
	m.LinkResponses = make(map[string][]string)
	m.Errors = make(map[string]error)
	m.Calls = make([]MockCall, 0)
}

// AssertCalled verifies that a method was called with any arguments.
func (m *MockScraper) AssertCalled(method string) bool {
	for _, call := range m.Calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// CallCount returns the number of times a method was called.
func (m *MockScraper) CallCount(method string) int {
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// ErrMockScraper is a sentinel error for mock failures.
var ErrMockScraper = errors.New("mock scraper error")

// Compile-time check: MockScraper must implement Scraper interface
var _ Scraper = (*MockScraper)(nil)
