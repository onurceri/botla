package urlutil

import (
	"testing"
)

func TestSSRFValidator_BlockedSchemes(t *testing.T) {
	v := NewSSRFValidator()

	blocked := []string{
		"file:///etc/passwd",
		"ftp://example.com/file",
		"gopher://example.com",
		"data:text/html,<script>alert(1)</script>",
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_BlockedHosts(t *testing.T) {
	v := NewSSRFValidator()

	blocked := []string{
		"http://localhost/path",
		"http://127.0.0.1/path",
		"http://0.0.0.0/path",
		"http://[::1]/path",
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_BlockedIPs(t *testing.T) {
	v := NewSSRFValidator()

	blocked := []string{
		"http://10.0.0.1/internal",
		"http://172.16.0.1/internal",
		"http://192.168.1.1/internal",
		"http://169.254.169.254/latest/meta-data/", // AWS metadata
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_AllowedURLs(t *testing.T) {
	v := NewSSRFValidator()

	allowed := []string{
		"https://example.com",
		"https://www.google.com/search",
		"http://github.com",
	}

	for _, url := range allowed {
		err := v.ValidateURL(url)
		if err != nil {
			t.Errorf("expected %s to be allowed, got error: %v", url, err)
		}
	}
}

func TestSSRFValidator_InvalidURLs(t *testing.T) {
	v := NewSSRFValidator()

	invalid := []string{
		"not-a-url",
		"://missing-scheme",
		"http://", // empty host
	}

	for _, url := range invalid {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to fail validation", url)
		}
	}
}
