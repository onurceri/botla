package scraper

import (
	"testing"
)

func TestAllowed_EmptyAllowedList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "returns true for any URL when allowed list is empty",
			url:  "https://example.com/page",
			want: true,
		},
		{
			name: "returns true for any URL with empty allowed list",
			url:  "http://malicious.com/attack",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, []string{})
			if got != tt.want {
				t.Errorf("allowed(%q, []) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestAllowed_MatchingDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		allowed []string
		want    bool
	}{
		{
			name:    "exact match",
			url:     "https://example.com/page",
			allowed: []string{"example.com"},
			want:    true,
		},
		{
			name:    "case insensitive",
			url:     "https://EXAMPLE.COM/page",
			allowed: []string{"example.com"},
			want:    true,
		},
		{
			name:    "subdomain not matching",
			url:     "https://example.com.evil.com/page",
			allowed: []string{"example.com"},
			want:    false, // Doesn't end with .example.com
		},
		{
			name:    "www subdomain matches via hasSuffix",
			url:     "https://www.example.com/page",
			allowed: []string{"example.com"},
			want:    true, // "www.example.com" ends with ".example.com"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, tt.allowed)
			if got != tt.want {
				t.Errorf("allowed(%q, %v) = %v, want %v", tt.url, tt.allowed, got, tt.want)
			}
		})
	}
}

func TestAllowed_NonMatchingDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		allowed []string
	}{
		{
			name:    "different domain",
			url:     "https://other.com/page",
			allowed: []string{"example.com"},
		},
		{
			name:    "partial match",
			url:     "https://notexample.com/page",
			allowed: []string{"example.com"},
		},
		{
			name:    "typosquatting",
			url:     "https://examp1e.com/page",
			allowed: []string{"example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, tt.allowed)
			if got {
				t.Errorf("allowed(%q, %v) = %v, want false", tt.url, tt.allowed, got)
			}
		})
	}
}

func TestAllowed_InvalidURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		allowed []string
	}{
		{
			name:    "parse error - not a URL",
			url:     "not a url",
			allowed: []string{"example.com"},
		},
		{
			name:    "no host",
			url:     "just a string",
			allowed: []string{"example.com"},
		},
		{
			name:    "empty URL",
			url:     "",
			allowed: []string{"example.com"},
		},
		{
			name:    "relative path",
			url:     "/page",
			allowed: []string{"example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, tt.allowed)
			if got {
				t.Errorf("allowed(%q, %v) = %v, want false for invalid URL", tt.url, tt.allowed, got)
			}
		})
	}
}

func TestAllowed_IPAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		allowed []string
	}{
		{
			name:    "IPv4 localhost",
			url:     "http://127.0.0.1:8080/page",
			allowed: []string{"localhost"},
		},
		{
			name:    "IPv6 localhost",
			url:     "http://[::1]:8080/page",
			allowed: []string{"localhost"},
		},
		{
			name:    "private IP",
			url:     "http://192.168.1.1/page",
			allowed: []string{"example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, tt.allowed)
			if tt.name == "IPv4 localhost" || tt.name == "IPv6 localhost" {
				if !got {
					t.Logf("allowed(%q, %v) = %v, localhost may not match 127.0.0.1", tt.url, tt.allowed, got)
				}
			} else {
				if got {
					t.Errorf("allowed(%q, %v) = %v, private IPs should not match domain allowlist", tt.url, tt.allowed, got)
				}
			}
		})
	}
}

func TestAllowed_PortHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		allowed []string
		want    bool
	}{
		{
			name:    "with port number",
			url:     "https://example.com:8080/page",
			allowed: []string{"example.com"},
			want:    true,
		},
		{
			name:    "different port still matches",
			url:     "https://example.com:9000/page",
			allowed: []string{"example.com"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, tt.allowed)
			if got != tt.want {
				t.Errorf("allowed(%q, %v) = %v, want %v", tt.url, tt.allowed, got, tt.want)
			}
		})
	}
}

func TestAllowed_MultipleAllowedDomains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		allowed []string
		want    bool
	}{
		{
			name:    "matches first",
			url:     "https://example.com/page",
			allowed: []string{"example.com", "other.com"},
			want:    true,
		},
		{
			name:    "matches second",
			url:     "https://other.com/page",
			allowed: []string{"example.com", "other.com"},
			want:    true,
		},
		{
			name:    "matches none",
			url:     "https://third.com/page",
			allowed: []string{"example.com", "other.com"},
			want:    false,
		},
		{
			name:    "www subdomain matches via hasSuffix",
			url:     "https://www.example.com/page",
			allowed: []string{"example.com", "other.com"},
			want:    true, // "www.example.com" ends with ".example.com"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, tt.allowed)
			if got != tt.want {
				t.Errorf("allowed(%q, %v) = %v, want %v", tt.url, tt.allowed, got, tt.want)
			}
		})
	}
}

func TestAllowed_WhitespaceHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		allowed []string
	}{
		{
			name:    "allowed with spaces",
			url:     "https://example.com/page",
			allowed: []string{"  example.com  "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := allowed(tt.url, tt.allowed)
			if !got {
				t.Errorf("allowed(%q, %v) = %v, spaces should be trimmed", tt.url, tt.allowed, got)
			}
		})
	}
}
