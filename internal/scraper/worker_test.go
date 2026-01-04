package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/pkg/urlutil"
)

func TestScrapeURLVisibleText(t *testing.T) {
	// Allow localhost for testing by overriding the package-level validator
	ssrfValidator = urlutil.NewSSRFValidator(true)
	defer func() { ssrfValidator = urlutil.NewSSRFValidator(false) }()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html><html><head><style>body{}</style><script>var a=1</script></head><body><p>Merhaba</p><div style="display:none">Gizli</div></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	out, err := ScrapeURL(ScrapingTask{URL: srv.URL + "/"}, CollectorConfig{}, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out == "" {
		t.Fatalf("empty")
	}
	if out != "Merhaba" {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestExtractLinks(t *testing.T) {
	htmlContent := `
<html>
<body>
  <a href="/page1">Page 1</a>
  <a href="sub/page2">Page 2</a>
  <a href="http://example.com/page3">Page 3 (Absolute Same Domain)</a>
  <a href="https://example.com/page3-secure">Page 3 Secure (Same Domain, diff scheme?)</a>
  <a href="http://other.com/page4">Page 4 (Different Domain)</a>
  <a href="#section">Anchor</a>
  <a href="javascript:void(0)">JS</a>
  <a href="mailto:test@example.com">Mail</a>
</body>
</html>
`
	baseURL := "http://example.com/base/"

	// Note: Our ExtractLinks implementation checks u.Host == base.Host.
	// So http vs https might be treated as different if host is strictly checked including port?
	// url.URL.Host usually includes port if present.
	// "example.com" == "example.com".

	links, err := ExtractLinks(htmlContent, baseURL, nil) // nil filter returns all links

	if err != nil {
		t.Fatalf("ExtractLinks failed: %v", err)
	}

	expected := map[string]bool{
		"http://example.com/page1":          true,
		"http://example.com/base/sub/page2": true,
		"http://example.com/page3":          true,
		// "https://example.com/page3-secure": true, // Host matches, but scheme diff? Implementation checks u.Host == base.Host.
		// If base is http, base.Host is example.com.
		// If link is https://example.com, u.Host is example.com.
		// So it should be included.
		"https://example.com/page3-secure": true,
	}

	if len(links) != len(expected) {
		t.Errorf("expected %d links, got %d: %v", len(expected), len(links), links)
	}

	for _, link := range links {
		if !expected[link] {
			t.Errorf("unexpected link found: %s", link)
		}
	}
}

func TestExtractLinks_WithFilter(t *testing.T) {
	htmlContent := `
<html>
<body>
  <a href="/blog/post-1">Blog Post 1</a>
  <a href="/blog/post-2">Blog Post 2</a>
  <a href="/docs/intro">Docs Intro</a>
  <a href="/admin/users">Admin Users</a>
  <a href="/tag/golang">Tag Golang</a>
  <a href="/about">About</a>
</body>
</html>
`
	baseURL := "http://example.com/"

	testCases := []struct {
		name     string
		include  []string
		exclude  []string
		expected map[string]bool
	}{
		{
			name:    "include blog only",
			include: []string{"/blog/*"},
			exclude: nil,
			expected: map[string]bool{
				"http://example.com/blog/post-1": true,
				"http://example.com/blog/post-2": true,
			},
		},
		{
			name:    "exclude admin and tag",
			include: nil,
			exclude: []string{"/admin/*", "/tag/*"},
			expected: map[string]bool{
				"http://example.com/blog/post-1": true,
				"http://example.com/blog/post-2": true,
				"http://example.com/docs/intro":  true,
				"http://example.com/about":       true,
			},
		},
		{
			name:    "include blog and docs, exclude nothing",
			include: []string{"/blog/*", "/docs/*"},
			exclude: nil,
			expected: map[string]bool{
				"http://example.com/blog/post-1": true,
				"http://example.com/blog/post-2": true,
				"http://example.com/docs/intro":  true,
			},
		},
		{
			name:    "include all, exclude admin",
			include: nil,
			exclude: []string{"/admin/*"},
			expected: map[string]bool{
				"http://example.com/blog/post-1": true,
				"http://example.com/blog/post-2": true,
				"http://example.com/docs/intro":  true,
				"http://example.com/tag/golang":  true,
				"http://example.com/about":       true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter, err := NewPathFilter(tc.include, tc.exclude)
			if err != nil {
				t.Fatalf("Failed to create filter: %v", err)
			}

			links, err := ExtractLinks(htmlContent, baseURL, filter)
			if err != nil {
				t.Fatalf("ExtractLinks failed: %v", err)
			}

			if len(links) != len(tc.expected) {
				t.Errorf("expected %d links, got %d: %v", len(tc.expected), len(links), links)
			}

			for _, link := range links {
				if !tc.expected[link] {
					t.Errorf("unexpected link found: %s", link)
				}
			}

			// Also check that expected links are present
			linkMap := make(map[string]bool)
			for _, link := range links {
				linkMap[link] = true
			}
			for expectedLink := range tc.expected {
				if !linkMap[expectedLink] {
					t.Errorf("expected link not found: %s", expectedLink)
				}
			}
		})
	}
}
