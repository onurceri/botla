package rag

import "testing"

func TestCountTokens_TurkishApprox(t *testing.T) {
	s := "Merhaba dünya! Türkçe metin örneği."
	got := CountTokens(s, "tr")
	if got <= 0 {
		t.Fatalf("expected >0 tokens, got %d", got)
	}
	// Rough bounds check based on heuristic
	chars := len([]rune(s))
	est := int(float64(chars)/4.0*1.3 + 0.5)
	if got < est-2 || got > est+2 {
		t.Fatalf("unexpected token estimate: got=%d est=%d", got, est)
	}
}
