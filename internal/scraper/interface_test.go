package scraper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockScraper_Implementation(t *testing.T) {
	t.Run("Implements Scraper interface", func(t *testing.T) {
		var _ Scraper = NewMockScraper()
	})

	t.Run("DefaultResponse", func(t *testing.T) {
		mock := NewMockScraper()
		task := ScrapingTask{URL: "https://example.com"}

		content, err := mock.ScrapeURLWithFallback(task, CollectorConfig{}, false, nil)

		assert.NoError(t, err)
		assert.Contains(t, content, "Mock scraped content")
		assert.True(t, mock.AssertCalled("ScrapeURLWithFallback"))
		assert.Equal(t, 1, mock.CallCount("ScrapeURLWithFallback"))
	})

	t.Run("ConfiguredResponse", func(t *testing.T) {
		mock := NewMockScraper()
		mock.SetResponse("https://example.com", "Custom content")
		task := ScrapingTask{URL: "https://example.com"}

		content, err := mock.ScrapeURLWithFallback(task, CollectorConfig{}, false, nil)

		assert.NoError(t, err)
		assert.Equal(t, "Custom content", content)
	})

	t.Run("ConfiguredError", func(t *testing.T) {
		mock := NewMockScraper()
		testErr := errors.New("connection refused")
		mock.SetError("https://example.com", testErr)
		task := ScrapingTask{URL: "https://example.com"}

		_, err := mock.ScrapeURLWithFallback(task, CollectorConfig{}, false, nil)

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("CustomFunction", func(t *testing.T) {
		mock := NewMockScraper()
		mock.ScrapeURLWithFallbackFunc = func(task ScrapingTask, cfg CollectorConfig, allowDynamic bool, scrapeConfig *ScrapeConfig) (string, error) {
			return "Func content for " + task.URL, nil
		}
		task := ScrapingTask{URL: "https://test.com"}

		content, err := mock.ScrapeURLWithFallback(task, CollectorConfig{}, false, nil)

		assert.NoError(t, err)
		assert.Equal(t, "Func content for https://test.com", content)
	})
}

func TestMockScraper_FetchRawHTML(t *testing.T) {
	t.Run("DefaultResponse", func(t *testing.T) {
		mock := NewMockScraper()

		html, err := mock.FetchRawHTML("https://example.com", CollectorConfig{})

		assert.NoError(t, err)
		assert.Contains(t, html, "<html>")
		assert.True(t, mock.AssertCalled("FetchRawHTML"))
	})

	t.Run("ConfiguredHTMLResponse", func(t *testing.T) {
		mock := NewMockScraper()
		mock.SetHTMLResponse("https://example.com", "<html><body>Custom HTML</body></html>")

		html, err := mock.FetchRawHTML("https://example.com", CollectorConfig{})

		assert.NoError(t, err)
		assert.Equal(t, "<html><body>Custom HTML</body></html>", html)
	})
}

func TestMockScraper_ExtractLinks(t *testing.T) {
	t.Run("DefaultEmptyLinks", func(t *testing.T) {
		mock := NewMockScraper()

		links, err := mock.ExtractLinks("<html></html>", "https://example.com", nil)

		assert.NoError(t, err)
		assert.Empty(t, links)
	})

	t.Run("ConfiguredLinks", func(t *testing.T) {
		mock := NewMockScraper()
		mock.SetLinks("https://example.com", []string{
			"https://example.com/page1",
			"https://example.com/page2",
		})

		links, err := mock.ExtractLinks("<html></html>", "https://example.com", nil)

		assert.NoError(t, err)
		assert.Len(t, links, 2)
		assert.Contains(t, links, "https://example.com/page1")
	})
}

func TestMockScraper_Reset(t *testing.T) {
	mock := NewMockScraper()
	mock.SetResponse("https://example.com", "content")
	mock.SetError("https://error.com", errors.New("error"))
	task := ScrapingTask{URL: "https://example.com"}
	_, _ = mock.ScrapeURLWithFallback(task, CollectorConfig{}, false, nil)

	mock.Reset()

	assert.Empty(t, mock.Responses)
	assert.Empty(t, mock.Errors)
	assert.Empty(t, mock.Calls)
}

func TestDefaultScraper_Implementation(t *testing.T) {
	t.Run("Implements Scraper interface", func(t *testing.T) {
		var _ Scraper = NewDefaultScraper()
	})
}
