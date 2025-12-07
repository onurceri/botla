package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScrapeURLVisibleText(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html><html><head><style>body{}</style><script>var a=1</script></head><body><p>Merhaba</p><div style="display:none">Gizli</div></body></html>`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	out, err := ScrapeURL(ScrapingTask{URL: srv.URL + "/"}, CollectorConfig{})
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

	links, err := ExtractLinks(htmlContent, baseURL)
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
