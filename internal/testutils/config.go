package testutils

import (
	"time"

	"github.com/onurceri/botla-app/pkg/config"
)

// TestConfig returns a Config struct with test values, avoiding t.Setenv() calls.
// This enables parallel test execution by eliminating process-global state changes.
//
// Usage:
//
//	func TestExample(t *testing.T) {
//		cfg := testutils.TestConfig()
//		// cfg has all fields set to test values, no env vars needed
//	}
func TestConfig() *config.Config {
	return &config.Config{
		DB_HOST:                "localhost",
		DB_PORT:                "5432",
		DB_NAME:                "botla_test",
		DB_USER:                "botla",
		DB_PASSWORD:            "botla",
		DB_SCHEMA:              "public",
		DB_SSLMODE:             "disable",
		REDIS_URL:              "redis://localhost:6379",
		QDRANT_URL:             "http://localhost:6333",
		QDRANT_API_KEY:         "",
		OPENAI_API_KEY:         "test-key",
		OPENAI_API_BASE:        "https://api.openai.com",
		OPENAI_TIMEOUT_MS:      30000,
		OPENROUTER_API_KEY:     "test-key",
		OPENROUTER_API_BASE:    "https://openrouter.ai/api/v1",
		OPENROUTER_TIMEOUT_MS:  30000,
		IYZICO_API_KEY:         "",
		IYZICO_SECRET_KEY:      "",
		JWT_SECRET:             "test-secret-for-testing-only",
		PORT:                   "8080",
		CORS_ALLOWED_ORIGINS:   "http://localhost:5173",
		WORKER_COUNT:           4,
		ANALYTICS_WORKER_COUNT: 10,
		R2_ACCOUNT_ID:          "",
		R2_ACCESS_KEY_ID:       "",
		R2_SECRET_ACCESS_KEY:   "",
		R2_BUCKET_NAME:         "",
		DEFAULT_CHATBOT_MODEL:  "gpt-4o-mini",
		RAG_TOPK:               5,
		RAG_MAX_CONTEXT_TOKENS: 2000,
		CHAT_TIMEOUT_MS:        60000,
		GO_ENV:                 "test",
		CookieSecure:           false,
	}
}

// TestConfigWith overrides specific fields of the default test config.
// Use this to customize config for specific test scenarios.
//
// Usage:
//
//	func TestSpecificConfig(t *testing.T) {
//		cfg := testutils.TestConfigWith(func(c *config.Config) {
//			c.JWT_SECRET = "custom-secret"
//			c.OPENAI_API_KEY = "custom-key"
//		})
//	}
func TestConfigWith(override func(*config.Config)) *config.Config {
	cfg := TestConfig()
	if override != nil {
		override(cfg)
	}
	return cfg
}

// TestRAGConfig returns a Config optimized for RAG testing.
func TestRAGConfig() *config.Config {
	return TestConfigWith(func(c *config.Config) {
		c.RAG_TOPK = 10
		c.RAG_MAX_CONTEXT_TOKENS = 4000
	})
}

// TestRateLimitConfig returns a Config with relaxed rate limits for testing.
func TestRateLimitConfig() *config.Config {
	return TestConfigWith(func(c *config.Config) {
		// Rate limits are set via pkg/ratelimit.Config, not main config
	})
}

// RateLimitTestConfig returns a ratelimit.Config for testing.
type RateLimitTestConfig struct {
	Global struct {
		RequestsPerWindow int
		WindowSize        time.Duration
	}
	User struct {
		RequestsPerWindow int
		WindowSize        time.Duration
	}
	EndpointOverrides map[string]struct {
		RequestsPerWindow int
		WindowSize        time.Duration
	}
}

// DefaultRateLimitTestConfig returns a Config for rate limit testing.
func DefaultRateLimitTestConfig() RateLimitTestConfig {
	cfg := RateLimitTestConfig{}
	cfg.Global.RequestsPerWindow = 100
	cfg.Global.WindowSize = 1 * time.Minute
	cfg.User.RequestsPerWindow = 50
	cfg.User.WindowSize = 1 * time.Minute
	cfg.EndpointOverrides = make(map[string]struct {
		RequestsPerWindow int
		WindowSize        time.Duration
	})
	return cfg
}

// FastRateLimitTestConfig returns a Config optimized for fast rate limit tests.
// Uses shorter windows to speed up test execution.
func FastRateLimitTestConfig() RateLimitTestConfig {
	cfg := RateLimitTestConfig{}
	cfg.Global.RequestsPerWindow = 10
	cfg.Global.WindowSize = 100 * time.Millisecond
	cfg.User.RequestsPerWindow = 5
	cfg.User.WindowSize = 100 * time.Millisecond
	cfg.EndpointOverrides = make(map[string]struct {
		RequestsPerWindow int
		WindowSize        time.Duration
	})
	return cfg
}
