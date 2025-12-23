package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScrapeURL_Unit(t *testing.T) {
	t.Run("Successful Static Scrape", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body><h1>Hello World</h1><p>This is test content.</p></body></html>`))
		}))
		defer srv.Close()

		task := ScrapingTask{URL: srv.URL}
		content, err := ScrapeURLWithFallback(task, DefaultCollectorConfig(), false, nil)

		assert.NoError(t, err)
		assert.Contains(t, content, "Hello World")
		assert.Contains(t, content, "This is test content.")
	})

	t.Run("Handle 404 Error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		task := ScrapingTask{URL: srv.URL}
		_, err := ScrapeURLWithFallback(task, DefaultCollectorConfig(), false, nil)

		assert.Error(t, err)
	})
}
