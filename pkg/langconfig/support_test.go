package langconfig

import "testing"

func TestIsSupported_Positive(t *testing.T) {
	if !IsSupported("en") || !IsSupported("tr") {
		t.Fatalf("expected en and tr supported")
	}
}

func TestIsSupported_Negative(t *testing.T) {
	if IsSupported("xx") {
		t.Fatalf("expected xx unsupported")
	}
}
