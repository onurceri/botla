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
		http.Error(w, "Forbidden", http.StatusForbidden)
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
		// Currently ScrapeURL returns plain error, so this will fail until I update it.
		// For TDD, I expect this assertion to fail or I accept it fails now.
		// But wait, ScrapeURL uses Colly, which returns a plain error.
		// My goal is to make it return ScrapeError.
		t.Logf("Got error: %T %v", err, err)
		// Mark as expected failure for TDD red phase? 
		// Actually, I can assert that it IS a ScrapeError, and if not, fail.
		t.Fatalf("expected ScrapeError, got %T", err)
	}
}
