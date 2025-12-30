package scraper

import (
	"errors"
	"testing"

	pkgErrors "github.com/onurceri/botla-co/pkg/errors"
)

// TestScrapeError_RateLimit verifies that HTTP 429 produces an error wrapping ErrRateLimit
func TestScrapeError_RateLimit(t *testing.T) {
	originalErr := errors.New("rate limit exceeded")
	scrapeErr := &ScrapeError{
		StatusCode: 429,
		URL:        "https://example.com",
		Err:        originalErr,
	}

	if !errors.Is(scrapeErr, pkgErrors.ErrRateLimit) {
		t.Errorf("expected error to wrap ErrRateLimit, got: %v", scrapeErr)
	}
}

// TestScrapeError_NotFound verifies that HTTP 404 produces an error wrapping ErrNotFound
func TestScrapeError_NotFound(t *testing.T) {
	originalErr := errors.New("page not found")
	scrapeErr := &ScrapeError{
		StatusCode: 404,
		URL:        "https://example.com/missing",
		Err:        originalErr,
	}

	if !errors.Is(scrapeErr, pkgErrors.ErrNotFound) {
		t.Errorf("expected error to wrap ErrNotFound, got: %v", scrapeErr)
	}
}

// TestScrapeError_InternalServerError verifies that HTTP 500 produces an error wrapping ErrNetwork
func TestScrapeError_InternalServerError(t *testing.T) {
	originalErr := errors.New("internal server error")
	scrapeErr := &ScrapeError{
		StatusCode: 500,
		URL:        "https://example.com",
		Err:        originalErr,
	}

	if !errors.Is(scrapeErr, pkgErrors.ErrNetwork) {
		t.Errorf("expected error to wrap ErrNetwork for 500, got: %v", scrapeErr)
	}
}

// TestScrapeError_BadGateway verifies that HTTP 502 produces an error wrapping ErrNetwork
func TestScrapeError_BadGateway(t *testing.T) {
	originalErr := errors.New("bad gateway")
	scrapeErr := &ScrapeError{
		StatusCode: 502,
		URL:        "https://example.com",
		Err:        originalErr,
	}

	if !errors.Is(scrapeErr, pkgErrors.ErrNetwork) {
		t.Errorf("expected error to wrap ErrNetwork for 502, got: %v", scrapeErr)
	}
}

// TestScrapeError_ServiceUnavailable verifies that HTTP 503 produces an error wrapping ErrNetwork
func TestScrapeError_ServiceUnavailable(t *testing.T) {
	originalErr := errors.New("service unavailable")
	scrapeErr := &ScrapeError{
		StatusCode: 503,
		URL:        "https://example.com",
		Err:        originalErr,
	}

	if !errors.Is(scrapeErr, pkgErrors.ErrNetwork) {
		t.Errorf("expected error to wrap ErrNetwork for 503, got: %v", scrapeErr)
	}
}

// TestScrapeError_GatewayTimeout verifies that HTTP 504 produces an error wrapping ErrNetwork
func TestScrapeError_GatewayTimeout(t *testing.T) {
	originalErr := errors.New("gateway timeout")
	scrapeErr := &ScrapeError{
		StatusCode: 504,
		URL:        "https://example.com",
		Err:        originalErr,
	}

	if !errors.Is(scrapeErr, pkgErrors.ErrNetwork) {
		t.Errorf("expected error to wrap ErrNetwork for 504, got: %v", scrapeErr)
	}
}

// TestScrapeError_Forbidden verifies that HTTP 403 does NOT wrap a sentinel error
func TestScrapeError_Forbidden(t *testing.T) {
	originalErr := errors.New("forbidden")
	scrapeErr := &ScrapeError{
		StatusCode: 403,
		URL:        "https://example.com",
		Err:        originalErr,
	}

	// 403 should not wrap any specific sentinel error
	if errors.Is(scrapeErr, pkgErrors.ErrRateLimit) {
		t.Error("403 should not wrap ErrRateLimit")
	}
	if errors.Is(scrapeErr, pkgErrors.ErrNotFound) {
		t.Error("403 should not wrap ErrNotFound")
	}
	if errors.Is(scrapeErr, pkgErrors.ErrNetwork) {
		t.Error("403 should not wrap ErrNetwork")
	}

	// But original error should still be unwrappable
	if !errors.Is(scrapeErr, originalErr) {
		t.Error("403 should still wrap original error")
	}
}

// TestScrapeError_NilError verifies Unwrap handles nil error
func TestScrapeError_NilError(t *testing.T) {
	scrapeErr := &ScrapeError{
		StatusCode: 500,
		URL:        "https://example.com",
		Err:        nil,
	}

	unwrapped := scrapeErr.Unwrap()
	if unwrapped != nil {
		t.Errorf("expected nil unwrap for nil error, got: %v", unwrapped)
	}
}

// TestScrapeError_ErrorMessage verifies the error message format
func TestScrapeError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name       string
		err        *ScrapeError
		wantPrefix string
	}{
		{
			name:       "with status code",
			err:        &ScrapeError{StatusCode: 429, URL: "https://example.com", Err: errors.New("rate limit")},
			wantPrefix: "HTTP 429:",
		},
		{
			name:       "without status code",
			err:        &ScrapeError{StatusCode: 0, URL: "https://example.com", Err: errors.New("network error")},
			wantPrefix: "Scraping failed:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if !contains(msg, tt.wantPrefix) {
				t.Errorf("expected message to contain %q, got: %s", tt.wantPrefix, msg)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))
}
