package ratelimit

import (
	"os"
	"strconv"
	"time"
)

// NewConfigFromEnv creates a TieredConfig from environment variables
func NewConfigFromEnv() *TieredConfig {
	cfg := DefaultConfig()

	// Global (IP-based) limits
	if v := getEnvInt("RATE_LIMIT_GLOBAL_REQUESTS_PER_MINUTE", 0); v > 0 {
		cfg.Global.RequestsPerWindow = v
	}

	if v := getEnvInt("RATE_LIMIT_GLOBAL_WINDOW_SECONDS", 0); v > 0 {
		cfg.Global.WindowSize = time.Duration(v) * time.Second
	}

	// User (authenticated) limits
	if v := getEnvInt("RATE_LIMIT_USER_REQUESTS_PER_MINUTE", 0); v > 0 {
		cfg.User.RequestsPerWindow = v
	}

	if v := getEnvInt("RATE_LIMIT_USER_WINDOW_SECONDS", 0); v > 0 {
		cfg.User.WindowSize = time.Duration(v) * time.Second
	}

	// Endpoint-specific overrides
	if v := getEnvInt("RATE_LIMIT_ENDPOINT_CHAT", 0); v > 0 {
		cfg.EndpointOverrides["/api/v1/chat"] = Config{
			RequestsPerWindow: v,
			WindowSize:        60 * time.Second,
		}
	}

	if v := getEnvInt("RATE_LIMIT_ENDPOINT_SOURCES", 0); v > 0 {
		cfg.EndpointOverrides["/api/v1/sources"] = Config{
			RequestsPerWindow: v,
			WindowSize:        60 * time.Second,
		}
	}

	return cfg
}

// getEnvInt retrieves an integer from environment variable, returns defaultVal if not found or invalid
func getEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
