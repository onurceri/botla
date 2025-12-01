package rag

import (
    "math"
    "unicode/utf8"
)

// CountTokens estimates token count for a text with Turkish-friendly multiplier.
// Approximation: tokens ≈ (characters / 4) * 1.3
func CountTokens(text string) int {
    if text == "" {
        return 0
    }
    n := utf8.RuneCountInString(text)
    base := float64(n) / 4.0
    est := base * 1.3
    t := int(math.Round(est))
    if t < 1 {
        t = 1
    }
    return t
}

