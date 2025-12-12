package services

import "testing"

func TestNormalizeLangCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "tr"},
		{"tr", "tr"},
		{"en", "en"},
		{"en-US", "en"},
		{"tr-TR", "tr"},
		{"  ", "tr"},
	}
	for _, tc := range tests {
		result := normalizeLangCode(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeLangCode(%q) = %q; want %q", tc.input, result, tc.expected)
		}
	}
}
