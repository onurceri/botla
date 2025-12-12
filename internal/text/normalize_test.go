package text

import "testing"

func TestNormalizeTR_Basic(t *testing.T) {
	got := NormalizeTR("  merhaba\t dünya\n\n\u00A0\u200B ")
	if got != "merhaba dünya" {
		t.Fatalf("normalize: %q", got)
	}
}

func TestNormalizeTR_ControlsRemoved(t *testing.T) {
	s := string([]rune{'a', 0x0001, 0x0002, 'b'})
	got := NormalizeTR(s)
	if got != "ab" {
		t.Fatalf("controls not removed: %q", got)
	}
}

func TestNormalizeTR_Empty(t *testing.T) {
	if NormalizeTR("") != "" {
		t.Fatalf("empty not preserved")
	}
}
