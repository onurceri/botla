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
    mu sync.Mutex
    m  map[string]*limiterState
    max int
    window time.Duration
}

func NewRateLimiterFromEnv() *RateLimiter {
    max := 10
    if v := os.Getenv("RATE_LIMIT_REQUESTS"); v != "" { if n, err := strconv.Atoi(v); err == nil { max = n } }
    win := 60 * time.Second
    if v := os.Getenv("RATE_LIMIT_WINDOW_SECONDS"); v != "" { if n, err := strconv.Atoi(v); err == nil { win = time.Duration(n) * time.Second } }
    return &RateLimiter{m: make(map[string]*limiterState), max: max, window: win}
}

func (rl *RateLimiter) allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    s := rl.m[key]
    now := time.Now()
    if s == nil || now.Sub(s.start) > rl.window {
        rl.m[key] = &limiterState{start: now, count: 1}
        return true
    }
    if s.count < rl.max {
        s.count++
        return true
    }
    return false
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
            if !rl.allow(key) {
                w.WriteHeader(http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

