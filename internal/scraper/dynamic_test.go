package scraper

import (
    "net/http/httptest"
    "os"
    "testing"
)

func TestScrapeDynamicURL_Fixture(t *testing.T) {
    if os.Getenv("RUN_DYNAMIC_TESTS") != "1" { t.Skip("dynamic tests disabled") }
    h := http.NewServeMux()
    h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("<html><body><h1>Başlık</h1><p>Dinamik içerik testi</p></body></html>"))
    })
    srv := httptest.NewServer(h)
    defer srv.Close()
    cfg := DefaultDynamicConfig()
    cfg.Allowed = []string{"localhost"}
    cfg.NavTimeout = 500 * 1e6 // 500ms
    out, err := ScrapeDynamicURL(srv.URL, cfg)
    if err != nil { t.Fatalf("dynamic scrape error: %v", err) }
    if out == "" { t.Fatalf("expected non-empty dynamic content") }
}

