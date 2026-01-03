package ratelimit

import (
	"time"
)

// Settings contains raw rate limit settings usually loaded from environment variables
type Settings struct {
	GlobalRequestsPerMinute int
	GlobalWindowSeconds     int
	UserRequestsPerMinute   int
	UserWindowSeconds       int
	EndpointChat            int
	EndpointSources         int
	AuthLogin               int
	AuthRegister            int
	AuthRefresh             int
}

// NewConfig creates a TieredConfig from the provided settings
func NewConfig(s Settings) *TieredConfig {
	cfg := DefaultConfig()

	// Global (IP-based) limits
	if s.GlobalRequestsPerMinute > 0 {
		cfg.Global.RequestsPerWindow = s.GlobalRequestsPerMinute
	}

	if s.GlobalWindowSeconds > 0 {
		cfg.Global.WindowSize = time.Duration(s.GlobalWindowSeconds) * time.Second
	}

	// User (authenticated) limits
	if s.UserRequestsPerMinute > 0 {
		cfg.User.RequestsPerWindow = s.UserRequestsPerMinute
	}

	if s.UserWindowSeconds > 0 {
		cfg.User.WindowSize = time.Duration(s.UserWindowSeconds) * time.Second
	}

	// Endpoint-specific overrides
	if s.EndpointChat > 0 {
		cfg.EndpointOverrides["/api/v1/chat"] = Config{
			RequestsPerWindow: s.EndpointChat,
			WindowSize:        60 * time.Second,
		}
	}

	if s.EndpointSources > 0 {
		cfg.EndpointOverrides["/api/v1/sources"] = Config{
			RequestsPerWindow: s.EndpointSources,
			WindowSize:        60 * time.Second,
		}
	}

	// Auth endpoint overrides (strict limits for brute-force protection)
	if s.AuthLogin > 0 {
		cfg.EndpointOverrides["/api/v1/auth/login"] = Config{
			RequestsPerWindow: s.AuthLogin,
			WindowSize:        60 * time.Second,
		}
	}

	if s.AuthRegister > 0 {
		cfg.EndpointOverrides["/api/v1/auth/register"] = Config{
			RequestsPerWindow: s.AuthRegister,
			WindowSize:        60 * time.Second,
		}
	}

	if s.AuthRefresh > 0 {
		cfg.EndpointOverrides["/api/v1/auth/refresh"] = Config{
			RequestsPerWindow: s.AuthRefresh,
			WindowSize:        60 * time.Second,
		}
	}

	return cfg
}
