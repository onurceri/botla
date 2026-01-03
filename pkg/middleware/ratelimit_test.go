package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/pkg/ratelimit"
)

func TestRateLimitMiddleware_EndpointOverride(t *testing.T) {
	// Create a config with a strict limit for login endpoint
	config := &ratelimit.TieredConfig{
		Global: ratelimit.Config{
			RequestsPerWindow: 100, // High global limit
			WindowSize:        60 * time.Second,
		},
		User: ratelimit.Config{
			RequestsPerWindow: 200,
			WindowSize:        60 * time.Second,
		},
		EndpointOverrides: map[string]ratelimit.Config{
			"/api/v1/auth/login": {
				RequestsPerWindow: 3, // Strict limit: only 3 requests
				WindowSize:        60 * time.Second,
			},
		},
	}

	// Create memory-based global limiter
	globalLimiter := ratelimit.NewMemoryLimiter(config.Global)
	defer globalLimiter.Close()
	rl := NewRateLimiter(globalLimiter, nil, config)

	// Test handler just returns 200
	handler := RateLimitMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i+1, rec.Code)
		}

		// Check rate limit headers
		if rec.Header().Get("X-RateLimit-Limit") != "3" {
			t.Errorf("Expected X-RateLimit-Limit=3, got %s", rec.Header().Get("X-RateLimit-Limit"))
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: expected 429, got %d", rec.Code)
	}

	// Retry-After header should be set
	if rec.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header to be set")
	}
}

func TestRateLimitMiddleware_DifferentIPsNotAffected(t *testing.T) {
	config := &ratelimit.TieredConfig{
		Global: ratelimit.Config{
			RequestsPerWindow: 100,
			WindowSize:        60 * time.Second,
		},
		User: ratelimit.Config{
			RequestsPerWindow: 200,
			WindowSize:        60 * time.Second,
		},
		EndpointOverrides: map[string]ratelimit.Config{
			"/api/v1/auth/login": {
				RequestsPerWindow: 2,
				WindowSize:        60 * time.Second,
			},
		},
	}

	globalLimiter := ratelimit.NewMemoryLimiter(config.Global)
	defer globalLimiter.Close()
	rl := NewRateLimiter(globalLimiter, nil, config)

	handler := RateLimitMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// IP 1 uses up its quota
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("IP1 Request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// IP 1 should be rate limited now
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 3rd request: expected 429, got %d", rec.Code)
	}

	// IP 2 should still be able to make requests
	req = httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("IP2 request: expected 200, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_NoEndpointOverride(t *testing.T) {
	config := &ratelimit.TieredConfig{
		Global: ratelimit.Config{
			RequestsPerWindow: 5,
			WindowSize:        60 * time.Second,
		},
		User: ratelimit.Config{
			RequestsPerWindow: 10,
			WindowSize:        60 * time.Second,
		},
		EndpointOverrides: map[string]ratelimit.Config{
			"/api/v1/auth/login": {
				RequestsPerWindow: 2,
				WindowSize:        60 * time.Second,
			},
		},
	}

	globalLimiter := ratelimit.NewMemoryLimiter(config.Global)
	defer globalLimiter.Close()
	rl := NewRateLimiter(globalLimiter, nil, config)

	handler := RateLimitMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Regular endpoint should use global limiter (5 requests)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/v1/chatbots", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/api/v1/chatbots", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("6th request: expected 429, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_AuthEndpointsStrictLimits(t *testing.T) {
	// Use the actual default config to verify auth endpoints have strict limits
	config := ratelimit.DefaultConfig()

	globalLimiter := ratelimit.NewMemoryLimiter(config.Global)
	defer globalLimiter.Close()
	rl := NewRateLimiter(globalLimiter, nil, config)

	handler := RateLimitMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	testCases := []struct {
		endpoint      string
		expectedLimit string
	}{
		{"/api/v1/auth/login", "5"},
		{"/api/v1/auth/register", "3"},
		{"/api/v1/auth/refresh", "10"},
	}

	for _, tc := range testCases {
		t.Run(tc.endpoint, func(t *testing.T) {
			// Use unique IP per endpoint to avoid cross-contamination
			req := httptest.NewRequest("POST", tc.endpoint, nil)
			req.RemoteAddr = tc.endpoint + ":12345" // Unique per endpoint
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("First request to %s: expected 200, got %d", tc.endpoint, rec.Code)
			}

			gotLimit := rec.Header().Get("X-RateLimit-Limit")
			if gotLimit != tc.expectedLimit {
				t.Errorf("%s: expected X-RateLimit-Limit=%s, got %s", tc.endpoint, tc.expectedLimit, gotLimit)
			}
		})
	}
}
