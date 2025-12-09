package scraper

import (
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// SMP-003: Parse gzipped sitemap
func TestParseSitemapGzip(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Encoding", "gzip")

		gw := gzip.NewWriter(w)
		defer gw.Close()

		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
		<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
			<url>
				<loc>https://example.com/page1</loc>
			</url>
		</urlset>`

		gw.Write([]byte(xmlContent))
	}))
	defer ts.Close()

	parser := NewSitemapParser(nil, 3, 0, 0)
	result, err := parser.ParseSitemap(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("ParseSitemap failed: %v", err)
	}

	if len(result.URLs) != 1 {
		t.Errorf("expected 1 URL, got %d", len(result.URLs))
	}
	if result.URLs[0].Loc != "https://example.com/page1" {
		t.Errorf("expected URL https://example.com/page1, got %s", result.URLs[0].Loc)
	}
}
