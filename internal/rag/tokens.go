package rag

import (
	"math"
	"unicode/utf8"

	"github.com/onurceri/botla-app/pkg/langconfig"
)

// CountTokens estimates token count for a text with language-specific multiplier.
// Approximation: tokens ≈ (characters / 4) * multiplier
func CountTokens(text string, langCode string) int {
	if text == "" {
		return 0
	}
	cfg := langconfig.Get(langCode)
	n := utf8.RuneCountInString(text)
	base := float64(n) / 4.0
	est := base * cfg.TokenMultiplier
	t := int(math.Round(est))
	if t < 1 {
		t = 1
	}
	return t
}
