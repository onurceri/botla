package scraper

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScrapeURL_Error(t *testing.T) {
	// Setup server that returns 403
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}))
	defer srv.Close()

	task := ScrapingTask{URL: srv.URL}
	cfg := DefaultCollectorConfig()

	_, err := ScrapeURL(task, cfg, nil)

	if err == nil {
		t.Fatal("expected error")
	}

	var se *ScrapeError
	if errors.As(err, &se) {
		if se.StatusCode != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", se.StatusCode)
		}
	} else {
		t.Logf("Got error: %T %v", err, err)
		t.Fatalf("expected ScrapeError, got %T", err)
	}
}
