package processing

import "testing"

func TestDefaultLang(t *testing.T) {
	if defaultLang("") != "tr" {
		t.Fatal("empty must default to tr")
	}
	if defaultLang("en") != "en" {
		t.Fatal("non-empty must be preserved")
	}
}
