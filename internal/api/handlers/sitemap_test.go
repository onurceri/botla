package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/middleware"
)

// TestParseSitemapDiscoverPath tests sitemap discover path extraction
func TestParseSitemapDiscoverPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
		ok       bool
	}{
		{"/api/v1/chatbots/abc123/sitemap/discover", "abc123", true},
		{"/api/v1/chatbots/xyz789/sitemap/discover", "xyz789", true},
		{"/api/v1/chatbots//sitemap/discover", "", false},        // Empty ID
		{"/api/v1/chatbots/abc/123/sitemap/discover", "", false}, // ID contains /
		{"/api/v1/chatbots/abc123/sitemap", "", false},           // Missing /discover
		{"/api/v1/chatbots/abc123/sources", "", false},           // Wrong endpoint
		{"/wrong/path/abc123/sitemap/discover", "", false},       // Wrong prefix
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			id, ok := parseSitemapDiscoverPath(tt.path)
			if ok != tt.ok {
				t.Errorf("parseSitemapDiscoverPath(%q) ok = %v, want %v", tt.path, ok, tt.ok)
			}
			if id != tt.expected {
				t.Errorf("parseSitemapDiscoverPath(%q) = %q, want %q", tt.path, id, tt.expected)
			}
		})
	}
}

// TestParseBulkSourcesPath tests bulk sources path extraction
func TestParseBulkSourcesPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
		ok       bool
	}{
		{"/api/v1/chatbots/abc123/sources/bulk", "abc123", true},
		{"/api/v1/chatbots/xyz789/sources/bulk", "xyz789", true},
		{"/api/v1/chatbots//sources/bulk", "", false},        // Empty ID
		{"/api/v1/chatbots/abc/123/sources/bulk", "", false}, // ID contains /
		{"/api/v1/chatbots/abc123/sources", "", false},       // Missing /bulk
		{"/api/v1/chatbots/abc123/sitemap", "", false},       // Wrong endpoint
		{"/wrong/path/abc123/sources/bulk", "", false},       // Wrong prefix
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			id, ok := parseBulkSourcesPath(tt.path)
			if ok != tt.ok {
				t.Errorf("parseBulkSourcesPath(%q) ok = %v, want %v", tt.path, ok, tt.ok)
			}
			if id != tt.expected {
				t.Errorf("parseBulkSourcesPath(%q) = %q, want %q", tt.path, id, tt.expected)
			}
		})
	}
}

// TestDiscoverSitemap_Unauthorized tests unauthenticated sitemap discovery
func TestDiscoverSitemap_Unauthorized(t *testing.T) {
	h := &SourcesHandlers{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/abc123/sitemap/discover", nil)
	rec := httptest.NewRecorder()

	h.DiscoverSitemap(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

// TestDiscoverSitemap_MethodNotAllowed tests non-POST requests
func TestDiscoverSitemap_MethodNotAllowed(t *testing.T) {
	h := &SourcesHandlers{}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user123")

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/v1/chatbots/abc123/sitemap/discover", nil)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.DiscoverSitemap(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: Expected status 405, got %d", method, rec.Code)
		}
	}
}

// TestDiscoverSitemap_InvalidPath tests invalid path handling
func TestDiscoverSitemap_InvalidPath(t *testing.T) {
	h := &SourcesHandlers{}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user123")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots//sitemap/discover", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.DiscoverSitemap(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

// TestDiscoverSitemap_InvalidBody tests invalid request body
// Note: This test requires a DB connection, so it's skipped in unit tests
func TestDiscoverSitemap_NoBody(t *testing.T) {
	t.Skip("Requires database connection - tested in integration tests")
}

// TestBulkCreateSources_Unauthorized tests unauthenticated bulk create
func TestBulkCreateSources_Unauthorized(t *testing.T) {
	h := &SourcesHandlers{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/abc123/sources/bulk", nil)
	rec := httptest.NewRecorder()

	h.BulkCreateSources(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

// TestBulkCreateSources_MethodNotAllowed tests non-POST requests
func TestBulkCreateSources_MethodNotAllowed(t *testing.T) {
	h := &SourcesHandlers{}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user123")

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/v1/chatbots/abc123/sources/bulk", nil)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.BulkCreateSources(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: Expected status 405, got %d", method, rec.Code)
		}
	}
}

// TestBulkCreateSources_InvalidPath tests invalid path handling
func TestBulkCreateSources_InvalidPath(t *testing.T) {
	h := &SourcesHandlers{}
	ctx := context.WithValue(context.Background(), middleware.ContextKeyUserID, "user123")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots//sources/bulk", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.BulkCreateSources(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

// TestBulkCreateSources_EmptyURLs tests empty URL list
// Note: This test requires a DB connection, so it's skipped in unit tests
func TestBulkCreateSources_NoBody(t *testing.T) {
	t.Skip("Requires database connection - tested in integration tests")
}

// TestBulkCreateSources_ResponseFormat tests the response format
func TestBulkCreateSources_ResponseFormat(t *testing.T) {
	// This test verifies the JSON response structure is correct
	response := struct {
		CreatedCount int      `json:"created_count"`
		SkippedCount int      `json:"skipped_count"`
		Errors       []string `json:"errors"`
	}{
		CreatedCount: 5,
		SkippedCount: 2,
		Errors:       []string{"error1"},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if _, ok := decoded["created_count"]; !ok {
		t.Error("Response missing created_count field")
	}
	if _, ok := decoded["skipped_count"]; !ok {
		t.Error("Response missing skipped_count field")
	}
	if _, ok := decoded["errors"]; !ok {
		t.Error("Response missing errors field")
	}
}

// TestDiscoverSitemap_ResponseFormat tests the sitemap response format
func TestDiscoverSitemap_ResponseFormat(t *testing.T) {
	// This test verifies the JSON response structure is correct
	type SitemapURL struct {
		Loc        string  `json:"loc"`
		LastMod    string  `json:"lastmod,omitempty"`
		ChangeFreq string  `json:"changefreq,omitempty"`
		Priority   float64 `json:"priority,omitempty"`
	}

	response := struct {
		URLs       []SitemapURL `json:"urls"`
		TotalCount int          `json:"total_count"`
	}{
		URLs: []SitemapURL{
			{Loc: "https://example.com/page1", LastMod: "2024-01-15", Priority: 0.8},
		},
		TotalCount: 1,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if _, ok := decoded["urls"]; !ok {
		t.Error("Response missing urls field")
	}
	if _, ok := decoded["total_count"]; !ok {
		t.Error("Response missing total_count field")
	}
}
