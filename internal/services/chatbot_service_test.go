package services

import (
	"encoding/json"
	"testing"
)

func TestFlexibleStringSlice_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "array format",
			input:    `["example.com", "localhost", "127.0.0.1"]`,
			expected: []string{"example.com", "localhost", "127.0.0.1"},
		},
		{
			name:     "string comma-separated",
			input:    `"example.com, localhost, 127.0.0.1"`,
			expected: []string{"example.com", "localhost", "127.0.0.1"},
		},
		{
			name:     "string comma-separated no spaces",
			input:    `"example.com,localhost,127.0.0.1"`,
			expected: []string{"example.com", "localhost", "127.0.0.1"},
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: []string{},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: []string{},
		},
		{
			name:     "single domain string",
			input:    `"example.com"`,
			expected: []string{"example.com"},
		},
		{
			name:     "single domain array",
			input:    `["example.com"]`,
			expected: []string{"example.com"},
		},
		{
			name:     "string with extra spaces",
			input:    `"  example.com  ,  localhost  "`,
			expected: []string{"example.com", "localhost"},
		},
		{
			name:     "string with empty segments",
			input:    `"example.com,,localhost"`,
			expected: []string{"example.com", "localhost"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result FlexibleStringSlice
			err := json.Unmarshal([]byte(tt.input), &result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d elements, got %d: %v", len(tt.expected), len(result), result)
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("element %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestSecuritySettingsRequest_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectedDomain []string
	}{
		{
			name:           "allowed_domains as string",
			input:          `{"secure_embed_enabled": true, "allowed_domains": "botla.app, localhost", "embed_secret": "secret123"}`,
			expectedDomain: []string{"botla.app", "localhost"},
		},
		{
			name:           "allowed_domains as array",
			input:          `{"secure_embed_enabled": true, "allowed_domains": ["botla.app", "localhost"], "embed_secret": "secret123"}`,
			expectedDomain: []string{"botla.app", "localhost"},
		},
		{
			name:           "allowed_domains empty string",
			input:          `{"secure_embed_enabled": false, "allowed_domains": ""}`,
			expectedDomain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req SecuritySettingsRequest
			err := json.Unmarshal([]byte(tt.input), &req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(req.AllowedDomains) != len(tt.expectedDomain) {
				t.Fatalf("expected %d domains, got %d: %v", len(tt.expectedDomain), len(req.AllowedDomains), req.AllowedDomains)
			}

			for i, v := range req.AllowedDomains {
				if v != tt.expectedDomain[i] {
					t.Errorf("domain %d: expected %q, got %q", i, tt.expectedDomain[i], v)
				}
			}
		})
	}
}
