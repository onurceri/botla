package urlutil

import (
	"testing"
)

func TestSSRFValidator_BlockedSchemes(t *testing.T) {
	v := NewSSRFValidator(false)

	blocked := []string{
		"file:///etc/passwd",
		"ftp://example.com/file",
		"gopher://example.com",
		"data:text/html,<script>alert(1)</script>",
		"javascript:alert(1)",
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_BlockedHosts(t *testing.T) {
	v := NewSSRFValidator(false)

	blocked := []string{
		"http://localhost/path",
		"http://127.0.0.1/path",
		"http://0.0.0.0/path",
		"http://[::1]/path",
		"http://metadata.google.internal/path",
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_BlockedIPs(t *testing.T) {
	v := NewSSRFValidator(false)

	blocked := []string{
		"http://10.0.0.1/internal",
		"http://172.16.0.1/internal",
		"http://192.168.1.1/internal",
		"http://169.254.169.254/latest/meta-data/", // AWS metadata
		"http://100.64.0.1/cgnat",                  // Carrier-grade NAT
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_BlockedIPv6(t *testing.T) {
	v := NewSSRFValidator(false)

	blocked := []string{
		"http://[fc00::1]/private",    // IPv6 private
		"http://[fe80::1]/link-local", // IPv6 link-local
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_AllowedURLs(t *testing.T) {
	v := NewSSRFValidator(false)

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
	v := NewSSRFValidator(false)

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

func TestSSRFValidator_BoundaryConditions(t *testing.T) {
	v := NewSSRFValidator(false)

	tests := []struct {
		url     string
		blocked bool
		desc    string
	}{
		// Private Class B boundary
		{"http://172.15.255.255/", false, "just before private Class B range"},
		{"http://172.16.0.0/", true, "start of private Class B range"},
		{"http://172.31.255.255/", true, "end of private Class B range"},
		{"http://172.32.0.0/", false, "just after private Class B range"},
	}

	for _, tt := range tests {
		err := v.ValidateURL(tt.url)
		if tt.blocked && err == nil {
			t.Errorf("%s: expected %s to be blocked", tt.desc, tt.url)
		}
		if !tt.blocked && err != nil {
			t.Errorf("%s: expected %s to be allowed, got error: %v", tt.desc, tt.url, err)
		}
	}
}

func TestSSRFValidator_ValidateURLStrict(t *testing.T) {
	v := NewSSRFValidator(false)

	// Test with a public URL
	ips, err := v.ValidateURLStrict("https://example.com")
	if err != nil {
		t.Errorf("expected example.com to pass, got error: %v", err)
	}
	if len(ips) == 0 {
		t.Error("expected ValidateURLStrict to return resolved IPs")
	}

	// Test with direct IP URL
	ips, err = v.ValidateURLStrict("http://8.8.8.8/")
	if err != nil {
		t.Errorf("expected 8.8.8.8 to pass, got error: %v", err)
	}
	if len(ips) != 1 {
		t.Errorf("expected 1 IP for direct IP URL, got %d", len(ips))
	}

	// Test with blocked URL
	_, err = v.ValidateURLStrict("http://127.0.0.1/")
	if err == nil {
		t.Error("expected 127.0.0.1 to be blocked")
	}
}

func TestSSRFValidator_AllowPrivate(t *testing.T) {
	// Test with allowPrivate=true
	v := NewSSRFValidator(true)

	// Private IPs should be allowed when allowPrivate is true
	allowed := []string{
		"http://localhost/path",
		"http://127.0.0.1/path",
		"http://192.168.1.1/path",
	}

	for _, url := range allowed {
		err := v.ValidateURL(url)
		if err != nil {
			t.Errorf("with allowPrivate=true, expected %s to be allowed, got error: %v", url, err)
		}
	}

	// SetAllowPrivate should work
	v.SetAllowPrivate(false)
	err := v.ValidateURL("http://127.0.0.1/path")
	if err == nil {
		t.Error("after SetAllowPrivate(false), expected 127.0.0.1 to be blocked")
	}
}
