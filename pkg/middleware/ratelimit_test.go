package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewRateLimiterFromEnv_Defaults(t *testing.T) {
	os.Unsetenv("RATE_LIMIT_REQUESTS")
	os.Unsetenv("RATE_LIMIT_WINDOW_SECONDS")
	rl := NewRateLimiterFromEnv()
	if rl == nil {
		t.Fatalf("nil rl")
	}
}

func TestRateLimitMiddleware_HeadersAndThrottle(t *testing.T) {
	os.Setenv("RATE_LIMIT_REQUESTS", "2")
	os.Setenv("RATE_LIMIT_WINDOW_SECONDS", "1")
	rl := NewRateLimiterFromEnv()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mh := RateLimitMiddleware(rl)(h)
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		mh.ServeHTTP(rec, req)
		if i < 2 && rec.Code != http.StatusOK {
			t.Fatalf("want 200 early, got %d", rec.Code)
		}
		if i == 2 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("want 429 late, got %d", rec.Code)
		}
		if rec.Header().Get("X-RateLimit-Limit") == "" {
			t.Fatalf("limit header missing")
		}
		if rec.Header().Get("X-RateLimit-Remaining") == "" {
			t.Fatalf("remaining header missing")
		}
	}
}
