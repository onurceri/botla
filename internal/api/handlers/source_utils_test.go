package handlers

import (
	"testing"
)

// TestIsPDFContentType tests PDF content type detection
func TestIsPDFContentType(t *testing.T) {
	tests := []struct {
		ct   string
		name string
		want bool
	}{
		{"application/pdf", "x.txt", true},
		{"", "x.pdf", true},
		{"application/pdf", "x.pdf", true},
		{"text/plain", "x.txt", false},
		{"application/octet-stream", "x.doc", false},
		{"", "", false},
	}
	for _, tc := range tests {
		got := isPDFContentType(tc.ct, tc.name)
		if got != tc.want {
			t.Errorf("isPDFContentType(%q, %q) = %v, want %v",
				tc.ct, tc.name, got, tc.want)
		}
	}
}

// TestComputeHash tests hash computation
func TestComputeHash(t *testing.T) {
	data := []byte("test data")
	hash := computeHash(data)
	if len(hash) != 64 { // SHA256 produces 64 hex chars
		t.Errorf("computeHash() returned hash of length %d, want 64", len(hash))
	}
	// Same input should produce same hash
	hash2 := computeHash(data)
	if hash != hash2 {
		t.Error("computeHash() is not deterministic")
	}
	// Different input should produce different hash
	hash3 := computeHash([]byte("different data"))
	if hash == hash3 {
		t.Error("computeHash() produced same hash for different inputs")
	}
}

// TestQuotaError tests quota error interface
func TestQuotaError(t *testing.T) {
	err := &quotaError{msg: "test error"}
	if err.Error() != "test error" {
		t.Errorf("quotaError.Error() = %q, want %q", err.Error(), "test error")
	}
}
