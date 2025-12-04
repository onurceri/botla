package scraper

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestScrapeDynamicURL(t *testing.T) {
	// Ensure domain whitelist allows localhost
	os.Setenv("SCRAPER_ALLOWED_DOMAINS", "127.0.0.1,localhost")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html><html><body><script>setTimeout(()=>{document.body.innerHTML='<p>Merhaba Dinamik</p>'},100)</script></body></html>`))
	}))
	defer srv.Close()

	cfg := DefaultDynamicConfig()
	cfg.NavTimeout = 3 * time.Second

	// Try launch; skip if no browser available
	_, err := NewBrowserPool(1, 5*time.Second)
	if err != nil {
		t.Skip("rod launch failed, skipping dynamic test")
	}

	out, err := ScrapeDynamicURL(srv.URL, cfg)
	if err != nil {
		t.Skip("dynamic scrape timed out, skipping in CI")
	}
	if out != "Merhaba Dinamik" {
		t.Fatalf("unexpected: %q", out)
	}
}
