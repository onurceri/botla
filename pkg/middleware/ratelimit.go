package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/httputil"
	"github.com/onurceri/botla-co/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
)

// RateLimiter wraps multiple rate limiting implementations for plan-based rate limiting
type RateLimiter struct {
	globalLimiter     ratelimit.Limiter            // IP-based for unauthenticated
	planLimiters      map[string]ratelimit.Limiter // Plan code -> limiter
	endpointLimiters  map[string]ratelimit.Limiter // Endpoint path -> limiter (for strict limits on auth, etc.)
	planLimitersMu    sync.RWMutex                 // Protects planLimiters map
	endpointLimiterMu sync.RWMutex                 // Protects endpointLimiters map
	redisClient       *redis.Client                // For creating new limiters
	tieredConfig      *ratelimit.TieredConfig      // Legacy config
	useMemory         bool                         // Whether to use memory fallback
}

// NewRateLimiter creates a new tiered rate limiter with plan-based support
// The globalLimiter is used for unauthenticated requests
// Plan-based limiters are created dynamically as needed
func NewRateLimiter(globalLimiter ratelimit.Limiter, redisClient *redis.Client, config *ratelimit.TieredConfig) *RateLimiter {
	return &RateLimiter{
		globalLimiter:    globalLimiter,
		planLimiters:     make(map[string]ratelimit.Limiter),
		endpointLimiters: make(map[string]ratelimit.Limiter),
		redisClient:      redisClient,
		tieredConfig:     config,
		useMemory:        redisClient == nil,
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
	if existingLimiter, exists := rl.planLimiters[plan.Code]; exists {
		return existingLimiter
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

// getOrCreateEndpointLimiter creates a limiter for the endpoint if it doesn't exist
func (rl *RateLimiter) getOrCreateEndpointLimiter(endpoint string) ratelimit.Limiter {
	// Fast path: check if limiter already exists (read lock)
	rl.endpointLimiterMu.RLock()
	limiter, exists := rl.endpointLimiters[endpoint]
	rl.endpointLimiterMu.RUnlock()

	if exists {
		return limiter
	}

	// Slow path: create new limiter (write lock)
	rl.endpointLimiterMu.Lock()
	defer rl.endpointLimiterMu.Unlock()

	// Double-check after acquiring write lock
	if existingLimiter, exists := rl.endpointLimiters[endpoint]; exists {
		return existingLimiter
	}

	// Get config from endpoint overrides
	cfg, ok := rl.tieredConfig.EndpointOverrides[endpoint]
	if !ok {
		return nil // No override for this endpoint
	}

	// Create limiter based on backend
	if !rl.useMemory {
		limiter = ratelimit.NewRedisLimiter(rl.redisClient, cfg)
	} else {
		limiter = ratelimit.NewMemoryLimiter(cfg)
	}

	rl.endpointLimiters[endpoint] = limiter
	return limiter
}

// RateLimitMiddleware creates a rate limiting middleware using plan-based approach
// For authenticated users, uses their plan's rate limits
// For unauthenticated users, uses global IP-based limits
// Endpoint-specific overrides (like auth endpoints) take precedence for stricter limits
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ip := httputil.ExtractIP(r)

			// Check for endpoint-specific rate limits FIRST (for auth endpoints, etc.)
			// These are stricter and IP-based to prevent brute-force attacks
			endpointLimiter := rl.getOrCreateEndpointLimiter(r.URL.Path)
			if endpointLimiter != nil {
				key := ratelimit.Key(ratelimit.TierEndpoint, ip+":"+r.URL.Path)
				result, err := endpointLimiter.Allow(ctx, key)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				// Set rate limit headers for endpoint limits
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

				if !result.Allowed {
					w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}

				// For endpoint-specific limits, we still proceed but skip general limiter
				// Endpoint limits are sufficient for auth endpoints
				next.ServeHTTP(w, r)
				return
			}

			// Apply general rate limits (plan-based or global) only if no endpoint override
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

// Close cleans up all rate limiters and their background goroutines
// This is critical for preventing memory leaks in tests
func (rl *RateLimiter) Close() error {
	// Close global limiter
	if rl.globalLimiter != nil {
		if err := rl.globalLimiter.Close(); err != nil {
			return err
		}
	}

	// Close all plan limiters
	rl.planLimitersMu.Lock()
	for _, limiter := range rl.planLimiters {
		if err := limiter.Close(); err != nil {
			rl.planLimitersMu.Unlock()
			return err
		}
	}
	rl.planLimitersMu.Unlock()

	// Close all endpoint limiters
	rl.endpointLimiterMu.Lock()
	for _, limiter := range rl.endpointLimiters {
		if err := limiter.Close(); err != nil {
			rl.endpointLimiterMu.Unlock()
			return err
		}
	}
	rl.endpointLimiterMu.Unlock()

	return nil
}
