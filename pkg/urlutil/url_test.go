package urlutil

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "trailing slash on root",
			input:    "https://example.com/",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "no trailing slash on root",
			input:    "https://example.com",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "trailing slash on path",
			input:    "https://example.com/page/",
			expected: "https://example.com/page",
			wantErr:  false,
		},
		{
			name:     "no trailing slash on path",
			input:    "https://example.com/page",
			expected: "https://example.com/page",
			wantErr:  false,
		},
		{
			name:     "uppercase scheme and host",
			input:    "HTTPS://EXAMPLE.COM/Page/",
			expected: "https://example.com/Page",
			wantErr:  false,
		},
		{
			name:     "with query parameters",
			input:    "https://example.com/search/?q=test",
			expected: "https://example.com/search?q=test",
			wantErr:  false,
		},
		{
			name:     "with fragment",
			input:    "https://example.com/page/#section",
			expected: "https://example.com/page#section",
			wantErr:  false,
		},
		{
			name:     "with port",
			input:    "https://example.com:8080/api/",
			expected: "https://example.com:8080/api",
			wantErr:  false,
		},
		{
			name:     "leading and trailing whitespace",
			input:    "  https://example.com/  ",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "http scheme",
			input:    "HTTP://Example.Com/Path/",
			expected: "http://example.com/Path",
			wantErr:  false,
		},
		{
			name:     "deep nested path",
			input:    "https://example.com/a/b/c/d/",
			expected: "https://example.com/a/b/c/d",
			wantErr:  false,
		},
		{
			name:     "real world example - espacefarmento",
			input:    "https://espacefarmento.fr/",
			expected: "https://espacefarmento.fr",
			wantErr:  false,
		},
		{
			name:     "real world example - espacefarmento without slash",
			input:    "https://espacefarmento.fr",
			expected: "https://espacefarmento.fr",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("NormalizeURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNormalizeURL_Consistency(t *testing.T) {
	// These pairs should normalize to the same value
	pairs := [][]string{
		{"https://example.com/", "https://example.com"},
		{"https://example.com/page/", "https://example.com/page"},
		{"HTTPS://EXAMPLE.COM/", "https://example.com"},
		{"https://espacefarmento.fr/", "https://espacefarmento.fr"},
	}

	for _, pair := range pairs {
		norm1, err1 := NormalizeURL(pair[0])
		norm2, err2 := NormalizeURL(pair[1])

		if err1 != nil || err2 != nil {
			t.Errorf("Unexpected error: %v, %v", err1, err2)
			continue
		}

		if norm1 != norm2 {
			t.Errorf("URLs should normalize to same value:\n  %q -> %q\n  %q -> %q",
				pair[0], norm1, pair[1], norm2)
		}
	}
}
