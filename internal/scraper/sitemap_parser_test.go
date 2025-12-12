package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseSitemap_StandardFormat(t *testing.T) {
	// Create mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>https://example.com/page1</loc>
        <lastmod>2024-01-15</lastmod>
        <changefreq>weekly</changefreq>
        <priority>0.8</priority>
    </url>
    <url>
        <loc>https://example.com/page2</loc>
        <lastmod>2024-01-10</lastmod>
        <priority>0.5</priority>
    </url>
</urlset>`
		w.Write([]byte(xml))
	}))
	defer srv.Close()

	parser := DefaultSitemapParser()
	ctx := context.Background()

	result, err := parser.ParseSitemap(ctx, srv.URL+"/sitemap.xml")
	if err != nil {
		t.Fatalf("ParseSitemap failed: %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("Expected 2 URLs, got %d", result.TotalCount)
	}

	if len(result.URLs) != 2 {
		t.Fatalf("Expected 2 URLs in slice, got %d", len(result.URLs))
	}

	if result.URLs[0].Loc != "https://example.com/page1" {
		t.Errorf("Expected first URL to be https://example.com/page1, got %s", result.URLs[0].Loc)
	}

	if result.URLs[0].Priority != 0.8 {
		t.Errorf("Expected priority 0.8, got %f", result.URLs[0].Priority)
	}

	if result.URLs[0].ChangeFreq != "weekly" {
		t.Errorf("Expected changefreq weekly, got %s", result.URLs[0].ChangeFreq)
	}

	if result.IsSitemapIndex {
		t.Error("Expected IsSitemapIndex to be false")
	}
}

func TestParseSitemap_IndexFormat(t *testing.T) {
	// Create mock servers - main server returns index, sub-server returns urlset
	subSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>https://example.com/blog/post1</loc>
        <lastmod>2024-01-15</lastmod>
    </url>
    <url>
        <loc>https://example.com/blog/post2</loc>
        <lastmod>2024-01-14</lastmod>
    </url>
</urlset>`
		w.Write([]byte(xml))
	}))
	defer subSrv.Close()

	mainSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <sitemap>
        <loc>` + subSrv.URL + `/sitemap-blog.xml</loc>
        <lastmod>2024-01-15</lastmod>
    </sitemap>
</sitemapindex>`
		w.Write([]byte(xml))
	}))
	defer mainSrv.Close()

	parser := DefaultSitemapParser()
	ctx := context.Background()

	result, err := parser.ParseSitemap(ctx, mainSrv.URL+"/sitemap_index.xml")
	if err != nil {
		t.Fatalf("ParseSitemap failed: %v", err)
	}

	if !result.IsSitemapIndex {
		t.Error("Expected IsSitemapIndex to be true")
	}

	if result.SubSitemaps != 1 {
		t.Errorf("Expected 1 sub-sitemap, got %d", result.SubSitemaps)
	}

	if result.TotalCount != 2 {
		t.Errorf("Expected 2 URLs, got %d", result.TotalCount)
	}
}

func TestParseSitemap_InvalidXML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte("not valid xml <<<<"))
	}))
	defer srv.Close()

	parser := DefaultSitemapParser()
	ctx := context.Background()

	_, err := parser.ParseSitemap(ctx, srv.URL+"/sitemap.xml")
	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}

func TestParseSitemap_EmptySitemap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
</urlset>`
		w.Write([]byte(xml))
	}))
	defer srv.Close()

	parser := DefaultSitemapParser()
	ctx := context.Background()

	result, err := parser.ParseSitemap(ctx, srv.URL+"/sitemap.xml")
	if err != nil {
		t.Fatalf("ParseSitemap failed: %v", err)
	}

	if result.TotalCount != 0 {
		t.Errorf("Expected 0 URLs, got %d", result.TotalCount)
	}
}

func TestParseSitemap_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	parser := DefaultSitemapParser()
	ctx := context.Background()

	_, err := parser.ParseSitemap(ctx, srv.URL+"/sitemap.xml")
	if err == nil {
		t.Error("Expected error for HTTP 404, got nil")
	}
}

func TestParseSitemap_InvalidURL(t *testing.T) {
	parser := DefaultSitemapParser()
	ctx := context.Background()

	// Test invalid scheme
	_, err := parser.ParseSitemap(ctx, "ftp://example.com/sitemap.xml")
	if err == nil {
		t.Error("Expected error for invalid scheme, got nil")
	}

	// Test invalid URL
	_, err = parser.ParseSitemap(ctx, "not-a-url")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestParseSitemap_MaxURLsLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url><loc>https://example.com/page1</loc></url>
    <url><loc>https://example.com/page2</loc></url>
    <url><loc>https://example.com/page3</loc></url>
    <url><loc>https://example.com/page4</loc></url>
    <url><loc>https://example.com/page5</loc></url>
</urlset>`
		w.Write([]byte(xml))
	}))
	defer srv.Close()

	// Create parser with max 3 URLs
	parser := NewSitemapParser(nil, 3, 30*time.Second, 3)
	ctx := context.Background()

	result, err := parser.ParseSitemap(ctx, srv.URL+"/sitemap.xml")
	if err != nil {
		t.Fatalf("ParseSitemap failed: %v", err)
	}

	if result.TotalCount != 3 {
		t.Errorf("Expected 3 URLs (max limit), got %d", result.TotalCount)
	}
}

func TestFilterURLsByPath(t *testing.T) {
	urls := []SitemapURL{
		{Loc: "https://example.com/blog/post1"},
		{Loc: "https://example.com/blog/post2"},
		{Loc: "https://example.com/about"},
		{Loc: "https://example.com/contact"},
		{Loc: "https://example.com/docs/api"},
	}

	tests := []struct {
		name         string
		includePaths []string
		excludePaths []string
		expected     int
	}{
		{
			name:         "include all (nil filter)",
			includePaths: nil,
			excludePaths: nil,
			expected:     5,
		},
		{
			name:         "include only blog",
			includePaths: []string{"/blog/*"},
			excludePaths: nil,
			expected:     2,
		},
		{
			name:         "exclude blog",
			includePaths: nil,
			excludePaths: []string{"/blog/*"},
			expected:     3,
		},
		{
			name:         "include blog, exclude post1",
			includePaths: []string{"/blog/*"},
			excludePaths: []string{"/blog/post1"},
			expected:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filter *PathFilter
			if tt.includePaths != nil || tt.excludePaths != nil {
				var err error
				filter, err = NewPathFilter(tt.includePaths, tt.excludePaths)
				if err != nil {
					t.Fatalf("NewPathFilter failed: %v", err)
				}
			}

			filtered := FilterURLsByPath(urls, filter)
			if len(filtered) != tt.expected {
				t.Errorf("Expected %d URLs, got %d", tt.expected, len(filtered))
			}
		})
	}
}

func TestDeduplicateURLs(t *testing.T) {
	urls := []SitemapURL{
		{Loc: "https://example.com/page1"},
		{Loc: "https://example.com/page1/"}, // Duplicate with trailing slash
		{Loc: "https://example.com/page2"},
		{Loc: "https://example.com/page2"}, // Exact duplicate
		{Loc: "https://example.com/page3/"},
	}

	result := DeduplicateURLs(urls)

	if len(result) != 3 {
		t.Errorf("Expected 3 unique URLs, got %d", len(result))
	}
}

func TestValidateSitemapURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com/sitemap.xml", false},
		{"http://example.com/sitemap.xml", false},
		{"ftp://example.com/sitemap.xml", true},
		{"not-a-url", true},
		{"://missing-scheme", true},
		{"https://", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := ValidateSitemapURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSitemapURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestIsSitemapIndex(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "standard sitemapindex",
			content: `<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`,
			want:    true,
		},
		{
			name:    "namespaced sitemapindex",
			content: `<sitemap:sitemapindex xmlns:sitemap="http://www.sitemaps.org/schemas/sitemap/0.9">`,
			want:    true,
		},
		{
			name:    "urlset not index",
			content: `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSitemapIndex([]byte(tt.content))
			if got != tt.want {
				t.Errorf("isSitemapIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSitemap_WithLastmodFormats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>https://example.com/page1</loc>
        <lastmod>2024-01-15T10:30:00+00:00</lastmod>
    </url>
    <url>
        <loc>https://example.com/page2</loc>
        <lastmod>2024-01-15</lastmod>
    </url>
    <url>
        <loc>https://example.com/page3</loc>
    </url>
</urlset>`
		w.Write([]byte(xml))
	}))
	defer srv.Close()

	parser := DefaultSitemapParser()
	ctx := context.Background()

	result, err := parser.ParseSitemap(ctx, srv.URL+"/sitemap.xml")
	if err != nil {
		t.Fatalf("ParseSitemap failed: %v", err)
	}

	if result.TotalCount != 3 {
		t.Errorf("Expected 3 URLs, got %d", result.TotalCount)
	}

	// Check that lastmod is preserved as string
	if result.URLs[0].LastMod != "2024-01-15T10:30:00+00:00" {
		t.Errorf("Expected full datetime, got %s", result.URLs[0].LastMod)
	}
	if result.URLs[1].LastMod != "2024-01-15" {
		t.Errorf("Expected date only, got %s", result.URLs[1].LastMod)
	}
	if result.URLs[2].LastMod != "" {
		t.Errorf("Expected empty lastmod, got %s", result.URLs[2].LastMod)
	}
}

func TestParseSitemap_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	parser := DefaultSitemapParser()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := parser.ParseSitemap(ctx, srv.URL+"/sitemap.xml")
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestNewSitemapParser_Defaults(t *testing.T) {
	// Test with all zero/nil values - should use defaults
	parser := NewSitemapParser(nil, 0, 0, 0)

	if parser.maxDepth != 3 {
		t.Errorf("Expected default maxDepth 3, got %d", parser.maxDepth)
	}
	if parser.timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", parser.timeout)
	}
	if parser.maxURLs != 10000 {
		t.Errorf("Expected default maxURLs 10000, got %d", parser.maxURLs)
	}
	if parser.client == nil {
		t.Error("Expected client to be initialized")
	}
}

func TestDiscoverSitemapURL(t *testing.T) {
	// Create a test server that responds to /sitemap.xml
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sitemap.xml" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	ctx := context.Background()
	found, err := DiscoverSitemapURL(ctx, srv.URL)
	if err != nil {
		t.Fatalf("DiscoverSitemapURL failed: %v", err)
	}

	expected := srv.URL + "/sitemap.xml"
	if found != expected {
		t.Errorf("Expected %s, got %s", expected, found)
	}
}

func TestDiscoverSitemapURL_NotFound(t *testing.T) {
	// Create a test server that returns 404 for all paths
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	ctx := context.Background()
	_, err := DiscoverSitemapURL(ctx, srv.URL)
	if err == nil {
		t.Error("Expected error when no sitemap found, got nil")
	}
}
