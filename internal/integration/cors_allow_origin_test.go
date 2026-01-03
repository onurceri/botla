package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestCORS_AllowConfiguredOrigin(t *testing.T) {
	t.Parallel() // Now safe - no t.Setenv()

	// Create handler with specific CORS origins (no env vars needed)
	origins := []string{"http://localhost:5173"}
	cors := middleware.CORSMiddlewareAllowOrigins(origins)

	// Simple test handler
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	// Execute
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Verify
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("missing allow origin header, got: %s", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_DisallowUnconfiguredOrigin(t *testing.T) {
	t.Parallel() // Now safe - no t.Setenv()

	origins := []string{"http://localhost:5173"}
	cors := middleware.CORSMiddlewareAllowOrigins(origins)

	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request from unconfigured origin
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://evil-site.com")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should still succeed (CORS doesn't block, just doesn't add headers)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// Unconfigured origin should NOT have the header
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("should not have allow origin header for unconfigured origin")
	}
}

func TestCORS_MultipleOrigins(t *testing.T) {
	t.Parallel() // Now safe - no t.Setenv()

	origins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"https://app.example.com",
	}
	cors := middleware.CORSMiddlewareAllowOrigins(origins)

	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	testCases := []struct {
		origin      string
		shouldAllow bool
	}{
		{"http://localhost:5173", true},
		{"http://localhost:3000", true},
		{"https://app.example.com", true},
		{"http://evil.com", false},
		{"", false},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		if tc.origin != "" {
			req.Header.Set("Origin", tc.origin)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		got := rr.Header().Get("Access-Control-Allow-Origin")
		want := ""
		if tc.shouldAllow {
			want = tc.origin
		}
		if got != want {
			t.Errorf("origin %q: expected %q, got %q", tc.origin, want, got)
		}
	}
}
