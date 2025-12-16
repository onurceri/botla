package ratelimit

import (
	"context"
	"time"
)

// Result contains the outcome of a rate limit check
type Result struct {
	Allowed    bool      // Whether the request is allowed
	Limit      int       // Maximum requests allowed in the window
	Remaining  int       // Requests remaining in current window
	ResetAt    time.Time // When the rate limit window resets
	RetryAfter int       // Seconds to wait before retrying (0 if allowed)
}

// Limiter defines the interface for rate limiting implementations
type Limiter interface {
	// Allow checks if a request should be allowed for the given key
	// key is typically a user ID, IP address, or combination
	Allow(ctx context.Context, key string) (*Result, error)

	// AllowN checks if N requests should be allowed for the given key
	AllowN(ctx context.Context, key string, n int) (*Result, error)

	// Reset clears the rate limit for the given key (useful for testing)
	Reset(ctx context.Context, key string) error

	// Close releases any resources held by the limiter
	Close() error
}

// Tier represents a rate limit tier (global, user, endpoint-specific)
type Tier string

const (
	TierGlobal   Tier = "global"   // IP-based for unauthenticated requests
	TierUser     Tier = "user"     // User-based for authenticated requests
	TierEndpoint Tier = "endpoint" // Endpoint-specific overrides
)

// Config holds rate limit configuration for a specific tier
type Config struct {
	RequestsPerWindow int           // Maximum requests allowed
	WindowSize        time.Duration // Time window duration
}

// TieredConfig holds configuration for all rate limit tiers
type TieredConfig struct {
	Global            Config            // Global IP-based limits
	User              Config            // Authenticated user limits
	EndpointOverrides map[string]Config // Per-endpoint limit overrides
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() *TieredConfig {
	return &TieredConfig{
		Global: Config{
			RequestsPerWindow: 60,
			WindowSize:        60 * time.Second,
		},
		User: Config{
			RequestsPerWindow: 120,
			WindowSize:        60 * time.Second,
		},
		EndpointOverrides: map[string]Config{
			"/api/v1/chat": {
				RequestsPerWindow: 20,
				WindowSize:        60 * time.Second,
			},
			"/api/v1/sources": {
				RequestsPerWindow: 10,
				WindowSize:        60 * time.Second,
			},
		},
	}
}

// Key builds a rate limit key for the given tier and identifier
func Key(tier Tier, identifier string) string {
	return "ratelimit:" + string(tier) + ":" + identifier
}
