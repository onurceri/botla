package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware_AllHeadersSet(t *testing.T) {
	mw := SecurityHeadersMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	tests := []struct {
		header string
		want   string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tc := range tests {
		got := rr.Header().Get(tc.header)
		if got != tc.want {
			t.Errorf("Header %s = %q, want %q", tc.header, got, tc.want)
		}
	}
}

func TestSecurityHeadersMiddleware_HSTS_OnHTTPS(t *testing.T) {
	mw := SecurityHeadersMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)
	req.TLS = &tls.ConnectionState{} // Simulate HTTPS connection
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	hsts := rr.Header().Get("Strict-Transport-Security")
	if hsts != "max-age=31536000; includeSubDomains" {
		t.Errorf("HSTS header = %q, want 'max-age=31536000; includeSubDomains'", hsts)
	}
}

func TestSecurityHeadersMiddleware_HSTS_OnXForwardedProto(t *testing.T) {
	mw := SecurityHeadersMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	hsts := rr.Header().Get("Strict-Transport-Security")
	if hsts != "max-age=31536000; includeSubDomains" {
		t.Errorf("HSTS header = %q, want 'max-age=31536000; includeSubDomains'", hsts)
	}
}

func TestSecurityHeadersMiddleware_NoHSTS_OnHTTP(t *testing.T) {
	mw := SecurityHeadersMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	hsts := rr.Header().Get("Strict-Transport-Security")
	if hsts != "" {
		t.Errorf("HSTS header should be empty for HTTP, got %q", hsts)
	}
}

func TestSecurityHeadersMiddleware_PassesThrough(t *testing.T) {
	mw := SecurityHeadersMiddleware()
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rr := httptest.NewRecorder()

	mw(handler).ServeHTTP(rr, req)

	if !called {
		t.Error("Handler was not called")
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("Status code = %d, want %d", rr.Code, http.StatusCreated)
	}
}
