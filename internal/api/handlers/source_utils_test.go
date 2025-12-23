package handlers

import (
	"testing"
)

// TestParseBotIDFromPath tests bot ID extraction from path
func TestParseBotIDFromPath(t *testing.T) {
	tests := []struct {
		path   string
		wantID string
		wantOK bool
	}{
		{"/api/v1/chatbots/abc/sources", "abc", true},
		{"/api/v1/chatbots/uuid-123/sources", "uuid-123", true},
		{"/api/v1/chatbots/abc/handoffs", "abc", true},
		{"/api/v1/chatbots/abc", "abc", true},
		{"/api/v1/chatbots/", "", false},
		{"/wrong/path", "", false},
	}
	for _, tc := range tests {
		id, ok := parseBotIDFromPath(tc.path)
		if ok != tc.wantOK || id != tc.wantID {
			t.Errorf("parseBotIDFromPath(%q) = (%q, %v), want (%q, %v)",
				tc.path, id, ok, tc.wantID, tc.wantOK)
		}
	}
}

// TestParseSourceIDFromPath tests source ID extraction from path
func TestParseSourceIDFromPath(t *testing.T) {
	tests := []struct {
		path   string
		wantID string
		wantOK bool
	}{
		{"/api/v1/sources/abc", "abc", true},
		{"/api/v1/sources/uuid-123", "uuid-123", true},
		{"/api/v1/sources/", "", false},
		{"/api/v1/sources/abc/refresh", "", false},
		{"/api/v1/sources/abc/extra", "", false},
		{"/wrong/path", "", false},
	}
	for _, tc := range tests {
		id, ok := parseSourceIDFromPath(tc.path)
		if ok != tc.wantOK || id != tc.wantID {
			t.Errorf("parseSourceIDFromPath(%q) = (%q, %v), want (%q, %v)",
				tc.path, id, ok, tc.wantID, tc.wantOK)
		}
	}
}

// TestParseRefreshSourceIDFromPath tests refresh source ID extraction
func TestParseRefreshSourceIDFromPath(t *testing.T) {
	tests := []struct {
		path   string
		wantID string
		wantOK bool
	}{
		{"/api/v1/sources/abc/refresh", "abc", true},
		{"/api/v1/sources/uuid-123/refresh", "uuid-123", true},
		{"/api/v1/sources//refresh", "", false},
		{"/api/v1/sources/abc", "", false},
		{"/api/v1/sources/abc/other", "", false},
		{"/wrong/path/refresh", "", false},
	}
	for _, tc := range tests {
		id, ok := parseRefreshSourceIDFromPath(tc.path)
		if ok != tc.wantOK || id != tc.wantID {
			t.Errorf("parseRefreshSourceIDFromPath(%q) = (%q, %v), want (%q, %v)",
				tc.path, id, ok, tc.wantID, tc.wantOK)
		}
	}
}

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
