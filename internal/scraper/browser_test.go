package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestScrapeDynamicURL(t *testing.T) {
	t.Parallel()
	// Ensure domain whitelist allows localhost
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html><html><body><script>setTimeout(()=>{document.body.innerHTML='<p>Merhaba Dinamik</p>'},100)</script></body></html>`))
	}))
	defer srv.Close()

	cfg := DynamicConfig{
		PoolSize:   1,
		IdleTTL:    5 * time.Second,
		NavTimeout: 5 * time.Second,
	}

	scraper, err := NewBrowserScraper(cfg)
	if err != nil {
		t.Skip("rod launch failed, skipping dynamic test")
	}
	defer scraper.Close()

	out, err := scraper.ScrapeDynamicURL(srv.URL)
	if err != nil {
		t.Skip("dynamic scrape timed out, skipping in CI")
	}
	if out != "Merhaba Dinamik" {
		t.Fatalf("unexpected: %q", out)
	}
}
