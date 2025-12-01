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
