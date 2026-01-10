package scraper

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestScrapeDynamicURL_Fixture(t *testing.T) {
	if os.Getenv("RUN_DYNAMIC_TESTS") != "1" {
		t.Skip("dynamic tests disabled")
	}
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><h1>Başlık</h1><p>Dinamik içerik testi</p></body></html>"))
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	cfg := DynamicConfig{
		PoolSize:   1,
		IdleTTL:    5 * time.Second,
		NavTimeout: 500 * time.Millisecond,
	}
	scraper, err := NewBrowserScraper(cfg)
	if err != nil {
		t.Fatalf("new browser scraper error: %v", err)
	}
	defer scraper.Close()

	out, err := scraper.ScrapeDynamicURL(srv.URL)
	if err != nil {
		t.Fatalf("dynamic scrape error: %v", err)
	}
	if out == "" {
		t.Fatalf("expected non-empty dynamic content")
	}
}
