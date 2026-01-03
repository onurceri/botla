package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/urlutil"
	"github.com/stretchr/testify/assert"
)

func TestScrapeURL_Unit(t *testing.T) {
	// Allow localhost for testing by overriding the package-level validator
	ssrfValidator = urlutil.NewSSRFValidator(true)
	defer func() { ssrfValidator = urlutil.NewSSRFValidator(false) }()

	t.Run("Successful Static Scrape", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body><h1>Hello World</h1><p>This is test content.</p></body></html>`))
		}))
		defer srv.Close()

		task := ScrapingTask{URL: srv.URL}
		ccfg := CollectorConfig{}
		content, err := ScrapeURLWithFallback(task, ccfg, nil, false, nil)

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
		ccfg := CollectorConfig{}
		_, err := ScrapeURLWithFallback(task, ccfg, nil, false, nil)

		assert.Error(t, err)
	})
}
