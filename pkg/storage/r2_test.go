package storage

import (
	"strings"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	k := GenerateKey("sources", "a/b\\c.pdf")
	if !strings.HasPrefix(k, "sources/") {
		t.Fatalf("prefix missing in key: %q", k)
	}
	parts := strings.SplitN(k, "_", 2)
	if len(parts) != 2 {
		t.Fatalf("expected timestamp separator '_' in key: %q", k)
	}
	if !strings.HasSuffix(k, "c.pdf") {
		t.Fatalf("basename not applied: %q", k)
	}
}

func TestGenerateSourceKey(t *testing.T) {
	k := GenerateSourceKey("org-123", "ws-456", "bot-789", "test.pdf")
	expected := "org/org-123/ws/ws-456/bot/bot-789/sources/"
	if !strings.HasPrefix(k, expected) {
		t.Fatalf("expected prefix %q, got %q", expected, k)
	}
	if !strings.HasSuffix(k, "test.pdf") {
		t.Fatalf("expected suffix test.pdf, got %q", k)
	}
}

func TestSystemKey(t *testing.T) {
	tests := []struct {
		parts    []string
		expected string
	}{
		{[]string{"tokenizer", "tr.json"}, "system/tokenizer/tr.json"},
		{[]string{"config"}, "system/config"},
	}
	for _, tt := range tests {
		got := SystemKey(tt.parts...)
		if got != tt.expected {
			t.Errorf("SystemKey(%v) = %q, want %q", tt.parts, got, tt.expected)
		}
	}
}

