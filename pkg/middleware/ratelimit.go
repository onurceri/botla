package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/onurceri/botla-co/pkg/ratelimit"
)

// RateLimiter wraps multiple rate limiting implementations for different tiers
type RateLimiter struct {
	globalLimiter   ratelimit.Limiter // IP-based for unauthenticated
	userLimiter     ratelimit.Limiter // User-based for authenticated
	endpointLimiters map[string]ratelimit.Limiter // Endpoint-specific overrides
	tieredConfig    *ratelimit.TieredConfig
}

// NewRateLimiter creates a new tiered rate limiter with separate limiters for each tier
func NewRateLimiter(globalLimiter, userLimiter ratelimit.Limiter, config *ratelimit.TieredConfig) *RateLimiter {
	return &RateLimiter{
		globalLimiter:    globalLimiter,
		userLimiter:      userLimiter,
		endpointLimiters: make(map[string]ratelimit.Limiter),
		tieredConfig:     config,
	}
}

// extractIP extracts the client IP from the request
// Handles X-Forwarded-For, X-Real-IP headers for proxy support
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header (can contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// normalizeEndpoint normalizes endpoint path for matching
// /api/v1/chatbots/123/chat -> /api/v1/chat
func normalizeEndpoint(path string) string {
	// Simple normalization - you can make this more sophisticated
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 {
		// Pattern: /api/v1/{resource}
		return "/" + strings.Join(parts[:3], "/")
	}
	return path
}

// RateLimitMiddleware creates a rate limiting middleware using the tiered approach
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			// Determine rate limit tier, key, and limiter
			var key string
			var limiter ratelimit.Limiter
			
			// 1. Check for endpoint-specific limit
			endpoint := normalizeEndpoint(r.URL.Path)
			if _, exists := rl.tieredConfig.EndpointOverrides[endpoint]; exists {
				// Use endpoint-specific limiter (if we had created them)
				// For now, fall through to user/global since we don't create endpoint limiters
				// This is a placeholder for future enhancement
			}
			
			// 2. Authenticated user - use user-based limiter
			if uid, ok := UserIDFromContext(ctx); ok && uid != "" {
				key = ratelimit.Key(ratelimit.TierUser, uid)
				limiter = rl.userLimiter
			} else {
				// 3. Unauthenticated - use global IP-based limiter
				ip := extractIP(r)
				key = ratelimit.Key(ratelimit.TierGlobal, ip)
				limiter = rl.globalLimiter
			}
			
			// Check rate limit
			result, err := limiter.Allow(ctx, key)
			if err != nil {
				// Log error but don't block request on rate limit failure
				// This ensures service availability even if Redis fails
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			
			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
			
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}
