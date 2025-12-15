package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
)

// RateLimiter wraps multiple rate limiting implementations for plan-based rate limiting
type RateLimiter struct {
	globalLimiter   ratelimit.Limiter                // IP-based for unauthenticated
	planLimiters    map[string]ratelimit.Limiter     // Plan code -> limiter
	planLimitersMu  sync.RWMutex                     // Protects planLimiters map
	redisClient     *redis.Client                    // For creating new limiters
	tieredConfig    *ratelimit.TieredConfig          // Legacy config
	useMemory       bool                             // Whether to use memory fallback
}

// NewRateLimiter creates a new tiered rate limiter with plan-based support
// The globalLimiter is used for unauthenticated requests
// Plan-based limiters are created dynamically as needed
func NewRateLimiter(globalLimiter ratelimit.Limiter, redisClient *redis.Client, config *ratelimit.TieredConfig) *RateLimiter {
	return &RateLimiter{
		globalLimiter: globalLimiter,
		planLimiters:  make(map[string]ratelimit.Limiter),
		redisClient:   redisClient,
		tieredConfig:  config,
		useMemory:     redisClient == nil,
	}
}

// getOrCreatePlanLimiter creates a limiter for the plan if it doesn't exist
// Uses double-checked locking pattern for thread-safe lazy initialization
func (rl *RateLimiter) getOrCreatePlanLimiter(plan *models.Plan) ratelimit.Limiter {
	// Fast path: check if limiter already exists (read lock)
	rl.planLimitersMu.RLock()
	limiter, exists := rl.planLimiters[plan.Code]
	rl.planLimitersMu.RUnlock()
	
	if exists {
		return limiter
	}
	
	// Slow path: create new limiter (write lock)
	rl.planLimitersMu.Lock()
	defer rl.planLimitersMu.Unlock()
	
	// Double-check after acquiring write lock (another goroutine might have created it)
	if limiter, exists := rl.planLimiters[plan.Code]; exists {
		return limiter
	}
	
	// Extract rate limit config from plan
	rateLimitsCfg := plan.Config.RateLimits
	
	// Validate config - use defaults if invalid
	requestsPerMinute := rateLimitsCfg.RequestsPerMinute
	windowSeconds := rateLimitsCfg.WindowSeconds
	
	if requestsPerMinute <= 0 {
		// Fallback to legacy user config if plan config is missing
		requestsPerMinute = rl.tieredConfig.User.RequestsPerWindow
	}
	if windowSeconds <= 0 {
		windowSeconds = int(rl.tieredConfig.User.WindowSize.Seconds())
	}
	
	cfg := ratelimit.Config{
		RequestsPerWindow: requestsPerMinute,
		WindowSize:        time.Duration(windowSeconds) * time.Second,
	}
	
	// Create limiter based on backend
	if !rl.useMemory {
		limiter = ratelimit.NewRedisLimiter(rl.redisClient, cfg)
	} else {
		limiter = ratelimit.NewMemoryLimiter(cfg)
	}
	
	rl.planLimiters[plan.Code] = limiter
	return limiter
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

// RateLimitMiddleware creates a rate limiting middleware using plan-based approach
// For authenticated users, uses their plan's rate limits
// For unauthenticated users, uses global IP-based limits
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			var key string
			var limiter ratelimit.Limiter
			
			// Check if user is authenticated and has a plan
			if plan, ok := PlanFromContext(ctx); ok && plan != nil {
				// Use plan-based limiter
				limiter = rl.getOrCreatePlanLimiter(plan)
				uid, _ := UserIDFromContext(ctx)
				key = ratelimit.Key(ratelimit.TierUser, uid)
			} else {
				// Use global IP-based limiter for unauthenticated
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
