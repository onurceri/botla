package services

import (
	"testing"

	"github.com/onurceri/botla-co/pkg/langconfig"
)

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

func TestResolveSystemPrompt(t *testing.T) {
	cfg := langconfig.Get("tr")
	if resolveSystemPrompt("", cfg) != cfg.ResponseTemplates.DefaultSystemPrompt {
		t.Fatalf("expected default system prompt fallback")
	}
	if resolveSystemPrompt("custom prompt", cfg) != "custom prompt" {
		t.Fatalf("expected custom prompt to be returned")
	}
	if resolveSystemPrompt("   ", cfg) != cfg.ResponseTemplates.DefaultSystemPrompt {
		t.Fatalf("expected whitespace-only to fallback to default")
	}
}
