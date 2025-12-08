package handlers

import "testing"

func TestIsValidHexColor(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"#fff", true},
		{"#ffffff", true},
		{"#FFF", true},
		{"#GGG", false},
		{"fff", false},
		{"#12345", false},
	}
	for _, c := range cases {
		if isValidHexColor(c.in) != c.ok {
			t.Fatalf("color %q expected %v", c.in, c.ok)
		}
	}
}

func TestNormalizeSuggestions(t *testing.T) {
	in := []string{"  Merhaba  ", "merhaba", "", "çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun çok uzun"}
	out := normalizeSuggestions(in)
	if len(out) == 0 {
		t.Fatalf("expected non-empty suggestions")
	}
	for _, s := range out {
		if len(s) > 120 {
			t.Fatalf("suggestion length should be <=120, got %d", len(s))
		}
	}
}

func TestDefaultHelpers(t *testing.T) {
	if defaultString(nil, "d") != "d" {
		t.Fatal("defaultString failed")
	}
	s := " x "
	if defaultString(&s, "d") != "x" {
		t.Fatal("defaultString trim failed")
	}
	if defaultInt(nil, 3) != 3 {
		t.Fatal("defaultInt failed")
	}
	i := 7
	if defaultInt(&i, 3) != 7 {
		t.Fatal("defaultInt override failed")
	}
	if defaultFloat32(nil, 0.5) != 0.5 {
		t.Fatal("defaultFloat32 failed")
	}
	f := float32(0.9)
	if defaultFloat32(&f, 0.5) != 0.9 {
		t.Fatal("defaultFloat32 override failed")
	}
}
