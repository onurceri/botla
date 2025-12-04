package scraper

import "testing"

func TestNormalizeText(t *testing.T) {
	s := string([]byte{0xEF, 0xBB, 0xBF}) + "Merhaba dünya\xff"
	out, err := NormalizeText(s)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out == "" {
		t.Fatalf("empty")
	}
	if !IsValidUTF8([]byte(out)) {
		t.Fatalf("not utf8")
	}
}
