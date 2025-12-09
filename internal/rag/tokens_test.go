package rag

import (
	"math"
	"testing"
	"unicode/utf8"
)

// TOK-001: Count tokens for empty string
func TestCountTokens_Empty(t *testing.T) {
	if got := CountTokens("", "en"); got != 0 {
		t.Errorf("TOK-001: expected 0 for empty string, got %d", got)
	}
}

// TOK-002: Count tokens for Turkish text (şğıöüç)
func TestCountTokens_Turkish(t *testing.T) {
	text := "şğıöüç" // 6 chars
	// Formula: round((6 / 4) * 1.3) = round(1.5 * 1.3) = round(1.95) = 2
	expected := 2
	got := CountTokens(text, "tr")
	if got != expected {
		t.Errorf("TOK-002: expected %d for '%s', got %d", expected, text, got)
	}
}

// TOK-003: Count tokens for English text
func TestCountTokens_English(t *testing.T) {
	text := "hello" // 5 chars
	// Formula: round((5 / 4) * 1.0) = round(1.25) = 1
	expected := 1
	got := CountTokens(text, "en")
	if got != expected {
		t.Errorf("TOK-003: expected %d for '%s', got %d", expected, text, got)
	}
}

// TOK-004: Token count minimum is 1
func TestCountTokens_Minimum(t *testing.T) {
	text := "a"
	// Formula: round((1 / 4) * 1.0) = round(0.25) = 0 -> adjusted to 1
	expected := 1
	got := CountTokens(text, "en")
	if got != expected {
		t.Errorf("TOK-004: expected %d for '%s', got %d", expected, text, got)
	}
}

// TOK-005: Token formula verification
func TestCountTokens_Formula(t *testing.T) {
	// "Türkiye'de yaşıyorum." (21 runes)
	// Expected: round((21/4) * 1.3) = round(5.25 * 1.3) = round(6.825) = 7 tokens
	text := "Türkiye'de yaşıyorum."
	runes := utf8.RuneCountInString(text)
	if runes != 21 {
		t.Fatalf("sanity check failed: expected 21 runes, got %d", runes)
	}
	
	expected := int(math.Round((float64(runes) / 4.0) * 1.3))
	got := CountTokens(text, "tr")
	
	if got != expected {
		t.Errorf("TOK-005: expected %d, got %d", expected, got)
	}
}
