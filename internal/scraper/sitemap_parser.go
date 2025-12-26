package scraper

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SitemapURL represents a single URL entry in a sitemap
type SitemapURL struct {
	Loc        string  `xml:"loc" json:"loc"`
	LastMod    string  `xml:"lastmod" json:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq" json:"changefreq,omitempty"`
	Priority   float64 `xml:"priority" json:"priority,omitempty"`
}

// SitemapIndex represents a sitemap index file containing references to other sitemaps
type SitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	} `xml:"sitemap"`
}

// URLSet represents a standard sitemap containing URL entries
type URLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapParseResult contains the result of parsing a sitemap
type SitemapParseResult struct {
	URLs           []SitemapURL `json:"urls"`
	TotalCount     int          `json:"total_count"`
	IsSitemapIndex bool         `json:"is_sitemap_index"`
	SubSitemaps    int          `json:"sub_sitemaps,omitempty"`
}

// HTTPClient interface for testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SitemapParser handles parsing of sitemap XML files
type SitemapParser struct {
	client   HTTPClient
	maxDepth int
	timeout  time.Duration
	maxURLs  int
}

// DefaultSitemapParser creates a parser with default settings
func DefaultSitemapParser() *SitemapParser {
	return &SitemapParser{
		client:   &http.Client{Timeout: 30 * time.Second},
		maxDepth: 3,
		timeout:  30 * time.Second,
		maxURLs:  10000,
	}
}

// NewSitemapParser creates a parser with custom settings
func NewSitemapParser(client HTTPClient, maxDepth int, timeout time.Duration, maxURLs int) *SitemapParser {
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	if maxDepth <= 0 {
		maxDepth = 3
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if maxURLs <= 0 {
		maxURLs = 10000
	}
	return &SitemapParser{
		client:   client,
		maxDepth: maxDepth,
		timeout:  timeout,
		maxURLs:  maxURLs,
	}
}

// ParseSitemap fetches and parses a sitemap from the given URL.
// If the sitemap is an index, it recursively fetches all referenced sitemaps.
func (p *SitemapParser) ParseSitemap(ctx context.Context, sitemapURL string) (*SitemapParseResult, error) {
	// Validate URL
	parsedURL, err := url.Parse(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("invalid sitemap URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: must be http or https")
	}

	return p.parseSitemapRecursive(ctx, sitemapURL, 0)
}

func (p *SitemapParser) parseSitemapRecursive(ctx context.Context, sitemapURL string, depth int) (*SitemapParseResult, error) {
	if depth > p.maxDepth {
		return nil, fmt.Errorf("maximum sitemap depth exceeded (%d)", p.maxDepth)
	}

	// Fetch the sitemap
	content, err := p.fetchURL(ctx, sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}

	// Try to parse as sitemap index first
	if isSitemapIndex(content) {
		return p.parseSitemapIndex(ctx, content, depth)
	}

	// Parse as regular sitemap
	return p.parseURLSet(content)
}

func (p *SitemapParser) fetchURL(ctx context.Context, targetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request for %s: %w", targetURL, err)
	}
	req.Header.Set("User-Agent", "Botla-Sitemap-Parser/1.0")
	req.Header.Set("Accept", "application/xml, text/xml, */*")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching sitemap from %s: %w", targetURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Limit read to 10MB
	limitedReader := io.LimitReader(resp.Body, 10*1024*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("reading sitemap content from %s: %w", targetURL, err)
	}
	return body, nil
}

// isSitemapIndex checks if the XML content is a sitemap index
func isSitemapIndex(content []byte) bool {
	return strings.Contains(string(content), "<sitemapindex") ||
		strings.Contains(string(content), ":sitemapindex")
}

func (p *SitemapParser) parseSitemapIndex(ctx context.Context, content []byte, depth int) (*SitemapParseResult, error) {
	var index SitemapIndex
	if err := xml.Unmarshal(content, &index); err != nil {
		return nil, fmt.Errorf("failed to parse sitemap index: %w", err)
	}

	result := &SitemapParseResult{
		URLs:           make([]SitemapURL, 0),
		IsSitemapIndex: true,
		SubSitemaps:    len(index.Sitemaps),
	}

	// Fetch each sub-sitemap
	for _, sitemap := range index.Sitemaps {
		if len(result.URLs) >= p.maxURLs {
			break
		}

		subResult, err := p.parseSitemapRecursive(ctx, sitemap.Loc, depth+1)
		if err != nil {
			// Log error but continue with other sitemaps
			continue
		}

		// Append URLs up to maxURLs
		remaining := p.maxURLs - len(result.URLs)
		if remaining > 0 {
			if len(subResult.URLs) <= remaining {
				result.URLs = append(result.URLs, subResult.URLs...)
			} else {
				result.URLs = append(result.URLs, subResult.URLs[:remaining]...)
			}
		}
	}

	result.TotalCount = len(result.URLs)
	return result, nil
}

func (p *SitemapParser) parseURLSet(content []byte) (*SitemapParseResult, error) {
	var urlset URLSet
	if err := xml.Unmarshal(content, &urlset); err != nil {
		return nil, fmt.Errorf("failed to parse sitemap: %w", err)
	}

	urls := urlset.URLs
	if len(urls) > p.maxURLs {
		urls = urls[:p.maxURLs]
	}

	return &SitemapParseResult{
		URLs:       urls,
		TotalCount: len(urls),
	}, nil
}

// FilterURLsByPath filters sitemap URLs using a PathFilter
func FilterURLsByPath(urls []SitemapURL, filter *PathFilter) []SitemapURL {
	if filter == nil {
		return urls
	}

	filtered := make([]SitemapURL, 0, len(urls))
	for _, surl := range urls {
		parsedURL, err := url.Parse(surl.Loc)
		if err != nil {
			continue
		}
		if filter.Match(parsedURL.Path) {
			filtered = append(filtered, surl)
		}
	}
	return filtered
}

// DeduplicateURLs removes duplicate URLs from the list
func DeduplicateURLs(urls []SitemapURL) []SitemapURL {
	seen := make(map[string]bool)
	result := make([]SitemapURL, 0, len(urls))

	for _, u := range urls {
		normalized := strings.TrimSuffix(u.Loc, "/")
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, u)
		}
	}
	return result
}

// ValidateSitemapURL checks if a URL looks like a valid sitemap URL
func ValidateSitemapURL(sitemapURL string) error {
	parsed, err := url.Parse(sitemapURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

// DiscoverSitemapURL attempts to find sitemap URL from common locations
func DiscoverSitemapURL(ctx context.Context, baseURL string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Common sitemap locations to try
	commonPaths := []string{
		"/sitemap.xml",
		"/sitemap_index.xml",
		"/sitemap-index.xml",
		"/sitemaps/sitemap.xml",
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		// Prevent redirect following for HEAD requests
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	for _, path := range commonPaths {
		// Check context cancellation between iterations
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("context error: %w", err)
		}

		// Per-request timeout
		reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		
		testURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, path)
		found, err := probeSitemap(reqCtx, client, testURL)
		cancel() // Always cancel to free resources
		
		if err != nil {
			continue
		}
		if found {
			return testURL, nil
		}
	}

	return "", fmt.Errorf("no sitemap found at common locations")
}

// probeSitemap checks if a sitemap exists at the given URL using HEAD then GET
func probeSitemap(ctx context.Context, client *http.Client, targetURL string) (bool, error) {
	// Try HEAD first
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, targetURL, nil)
	if err != nil {
		return false, fmt.Errorf("creating HEAD request for %s: %w", targetURL, err)
	}
	req.Header.Set("User-Agent", "Botla-Sitemap-Parser/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("probing sitemap %s: %w", targetURL, err)
	}
	// Always close and drain body
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	// If Method Not Allowed or Not Implemented, try GET
	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		// We can't reuse the same request since it was setup for HEAD
		return probeWithGET(ctx, client, targetURL)
	}

	return false, nil
}

func probeWithGET(ctx context.Context, client *http.Client, targetURL string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return false, fmt.Errorf("creating GET request for %s: %w", targetURL, err)
	}
	req.Header.Set("User-Agent", "Botla-Sitemap-Parser/1.0")
	// Use Range header to fetch only first byte to minimize data transfer if server supports it
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("probing sitemap with GET %s: %w", targetURL, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
	}()

	// 200 OK or 206 Partial Content are both valid signs existence
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
		return true, nil
	}

	return false, nil
}
