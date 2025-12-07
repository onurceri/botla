package middleware

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type limiterState struct {
	start time.Time
	count int
}

type RateLimiter struct {
	mu     sync.Mutex
	m      map[string]*limiterState
	max    int
	window time.Duration
}

func NewRateLimiterFromEnv() *RateLimiter {
	max := 10
	if v := os.Getenv("RATE_LIMIT_REQUESTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			max = n
		}
	}
	win := 60 * time.Second
	if v := os.Getenv("RATE_LIMIT_WINDOW_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			win = time.Duration(n) * time.Second
		}
	}
	return &RateLimiter{m: make(map[string]*limiterState), max: max, window: win}
}

// NewRateLimiterFromEnvWithPrefix constructs a limiter using env variables with a prefix
// e.g., prefix="SOURCES" reads SOURCES_RATE_LIMIT_REQUESTS and SOURCES_RATE_LIMIT_WINDOW_SECONDS
func NewRateLimiterFromEnvWithPrefix(prefix string) *RateLimiter {
    max := 10
    if v := os.Getenv(prefix + "_RATE_LIMIT_REQUESTS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            max = n
        }
    }
    win := 60 * time.Second
    if v := os.Getenv(prefix + "_RATE_LIMIT_WINDOW_SECONDS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            win = time.Duration(n) * time.Second
        }
    }
    return &RateLimiter{m: make(map[string]*limiterState), max: max, window: win}
}

func (rl *RateLimiter) allow(key string) (bool, int, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	s := rl.m[key]
	now := time.Now()
	if s == nil || now.Sub(s.start) > rl.window {
		rl.m[key] = &limiterState{start: now, count: 1}
		return true, rl.max - 1, int(rl.window.Seconds())
	}
	if s.count < rl.max {
		s.count++
		remaining := rl.max - s.count
		reset := int((rl.window - now.Sub(s.start)).Seconds())
		if reset < 0 {
			reset = 0
		}
		return true, remaining, reset
	}
	reset := int((rl.window - now.Sub(s.start)).Seconds())
	if reset < 0 {
		reset = 0
	}
	return false, 0, reset
}

func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ""
			if uid, ok := UserIDFromContext(r.Context()); ok && uid != "" {
				key = uid
			} else {
				host, _, _ := net.SplitHostPort(r.RemoteAddr)
				key = host
			}
			allowed, remaining, reset := rl.allow(key)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.max))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(reset))
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
