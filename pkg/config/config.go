package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"github.com/onurceri/botla-app/pkg/logger"
)

type Config struct {
	DB_HOST                string
	DB_PORT                string
	DB_NAME                string
	DB_USER                string
	DB_PASSWORD            string
	DB_SCHEMA              string
	DB_SSLMODE             string
	REDIS_URL              string
	QDRANT_URL             string
	QDRANT_API_KEY         string
	OPENAI_API_KEY         string
	OPENAI_API_BASE        string
	OPENAI_TIMEOUT_MS      int
	OPENROUTER_API_KEY     string
	OPENROUTER_API_BASE    string
	OPENROUTER_TIMEOUT_MS  int
	IYZICO_API_KEY         string
	IYZICO_SECRET_KEY      string
	JWT_SECRET             string
	PORT                   string
	CORS_ALLOWED_ORIGINS   string
	WORKER_COUNT           int
	ANALYTICS_WORKER_COUNT int
	R2_ACCOUNT_ID          string
	R2_ACCESS_KEY_ID       string
	R2_SECRET_ACCESS_KEY   string
	R2_BUCKET_NAME         string
	DEFAULT_CHATBOT_MODEL  string
	ANTHROPIC_API_KEY      string
	ANTHROPIC_API_BASE     string
	GOOGLE_AI_API_KEY      string
	GOOGLE_AI_API_BASE     string
	QDRANT_TIMEOUT_MS      int
	SCRAPER_ALLOWED_DOMAINS string
	SCRAPER_BROWSER_POOL_SIZE int
	SCRAPER_DYNAMIC_IDLE_SECS int
	SCRAPER_NAV_TIMEOUT_MS int
	SCRAPER_BROWSER_PATH string

	// RAG Configuration
	RAG_TOPK               int
	RAG_MAX_CONTEXT_TOKENS int

	// Chat Configuration
	CHAT_TIMEOUT_MS int

	// Environment
	GO_ENV string

	// Cookie Configuration
	CookieSecure bool
	CookieDomain string

	// Rate Limit Configuration
	RateLimitGlobalRequestsPerMinute int
	RateLimitGlobalWindowSeconds     int
	RateLimitUserRequestsPerMinute   int
	RateLimitUserWindowSeconds       int
	RateLimitEndpointChat            int
	RateLimitEndpointSources         int
	RateLimitAuthLogin               int
	RateLimitAuthRegister            int
	RateLimitAuthRefresh             int
}

var fatalf = func(msg string) {
	logger.New("ERROR").Error(msg, nil)
	os.Exit(1)
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	if os.Getenv("DB_HOST") == "" ||
		os.Getenv("DB_PORT") == "" ||
		os.Getenv("DB_NAME") == "" ||
		os.Getenv("DB_USER") == "" ||
		os.Getenv("DB_PASSWORD") == "" {
		fatalf("DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD cannot be empty")
	}

	if os.Getenv("QDRANT_URL") == "" {
		fatalf("QDRANT_URL cannot be empty")
	}

	// Check for required LLM providers
	// OpenAI is required for embeddings, OpenRouter is preferred for LLM calls
	if os.Getenv("OPENAI_API_KEY") == "" {
		// Just warn instead of fatal to allow starting up without key (useful for tests or limited functionality)
		logger.New("WARN").Warn("OPENAI_API_KEY is missing. Embeddings will fail.", nil)
	}

	if os.Getenv("JWT_SECRET") == "" {
		fatalf("JWT_SECRET cannot be empty")
	}

	if os.Getenv("PORT") == "" {
		fatalf("PORT cannot be empty")
	}

	return &Config{
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_NAME:     os.Getenv("DB_NAME"),
		DB_USER:     os.Getenv("DB_USER"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_SCHEMA: func() string {
			v := os.Getenv("DB_SCHEMA")
			if strings.TrimSpace(v) == "" {
				return "public"
			}
			return v
		}(),
		DB_SSLMODE: func() string {
			v := os.Getenv("DB_SSLMODE")
			if v != "" {
				return v
			}
			if os.Getenv("GO_ENV") == "production" {
				return "require"
			}
			return "disable"
		}(),
		REDIS_URL:      os.Getenv("REDIS_URL"),
		QDRANT_URL:     os.Getenv("QDRANT_URL"),
		QDRANT_API_KEY: os.Getenv("QDRANT_API_KEY"),
		OPENAI_API_KEY: os.Getenv("OPENAI_API_KEY"),
		OPENAI_API_BASE: func() string {
			v := os.Getenv("OPENAI_API_BASE")
			if v == "" {
				return "https://api.openai.com"
			}
			return v
		}(),
		OPENAI_TIMEOUT_MS:  parseIntEnv("OPENAI_TIMEOUT_MS", 30000),
		OPENROUTER_API_KEY: os.Getenv("OPENROUTER_API_KEY"),
		OPENROUTER_API_BASE: func() string {
			v := os.Getenv("OPENROUTER_API_BASE")
			if v == "" {
				return "https://openrouter.ai/api/v1"
			}
			return v
		}(),
		OPENROUTER_TIMEOUT_MS: parseIntEnv("OPENROUTER_TIMEOUT_MS", 30000),
		IYZICO_API_KEY:        os.Getenv("IYZICO_API_KEY"),
		IYZICO_SECRET_KEY:     os.Getenv("IYZICO_SECRET_KEY"),
		JWT_SECRET:            os.Getenv("JWT_SECRET"),
		PORT:                  os.Getenv("PORT"),
		CORS_ALLOWED_ORIGINS: func() string {
			v := os.Getenv("CORS_ALLOWED_ORIGINS")
			if v == "" {
				return "http://localhost:5173"
			}
			return v
		}(),
		WORKER_COUNT:           parseIntEnv("WORKER_COUNT", 4),
		ANALYTICS_WORKER_COUNT: parseIntEnv("ANALYTICS_WORKER_COUNT", 10),
		R2_ACCOUNT_ID:          os.Getenv("R2_ACCOUNT_ID"),
		R2_ACCESS_KEY_ID:       os.Getenv("R2_ACCESS_KEY_ID"),
		R2_SECRET_ACCESS_KEY:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2_BUCKET_NAME:         os.Getenv("R2_BUCKET_NAME"),
		DEFAULT_CHATBOT_MODEL:  DefaultChatbotModel(),
		ANTHROPIC_API_KEY:      os.Getenv("ANTHROPIC_API_KEY"),
		ANTHROPIC_API_BASE:     os.Getenv("ANTHROPIC_API_BASE"),
		GOOGLE_AI_API_KEY:      os.Getenv("GOOGLE_AI_API_KEY"),
		GOOGLE_AI_API_BASE:     os.Getenv("GOOGLE_AI_API_BASE"),
		QDRANT_TIMEOUT_MS:      parseIntEnv("QDRANT_TIMEOUT_MS", 30000),
		SCRAPER_ALLOWED_DOMAINS: os.Getenv("SCRAPER_ALLOWED_DOMAINS"),
		SCRAPER_BROWSER_POOL_SIZE: parseIntEnv("SCRAPER_BROWSER_POOL_SIZE", 2),
		SCRAPER_DYNAMIC_IDLE_SECS: parseIntEnv("SCRAPER_DYNAMIC_IDLE_SECS", 60),
		SCRAPER_NAV_TIMEOUT_MS:    parseIntEnv("SCRAPER_NAV_TIMEOUT_MS", 10000),
		SCRAPER_BROWSER_PATH: func() string {
			v := os.Getenv("SCRAPER_BROWSER_PATH")
			return v
		}(),
		RAG_TOPK:               parseIntEnv("RAG_TOPK", 5),
		RAG_MAX_CONTEXT_TOKENS: parseIntEnv("RAG_MAX_CONTEXT_TOKENS", 2000),
		CHAT_TIMEOUT_MS:        parseIntEnv("CHAT_TIMEOUT_MS", 60000),
		GO_ENV:                 os.Getenv("GO_ENV"),
		CookieSecure:           os.Getenv("GO_ENV") == "production",
		CookieDomain:           os.Getenv("COOKIE_DOMAIN"), // e.g., ".botla.app" for cross-subdomain cookies

		// Rate Limit Configuration
		RateLimitGlobalRequestsPerMinute: parseIntEnv("RATE_LIMIT_GLOBAL_REQUESTS_PER_MINUTE", 0),
		RateLimitGlobalWindowSeconds:     parseIntEnv("RATE_LIMIT_GLOBAL_WINDOW_SECONDS", 0),
		RateLimitUserRequestsPerMinute:   parseIntEnv("RATE_LIMIT_USER_REQUESTS_PER_MINUTE", 0),
		RateLimitUserWindowSeconds:       parseIntEnv("RATE_LIMIT_USER_WINDOW_SECONDS", 0),
		RateLimitEndpointChat:            parseIntEnv("RATE_LIMIT_ENDPOINT_CHAT", 0),
		RateLimitEndpointSources:         parseIntEnv("RATE_LIMIT_ENDPOINT_SOURCES", 0),
		RateLimitAuthLogin:               parseIntEnv("RATE_LIMIT_AUTH_LOGIN", 0),
		RateLimitAuthRegister:            parseIntEnv("RATE_LIMIT_AUTH_REGISTER", 0),
		RateLimitAuthRefresh:             parseIntEnv("RATE_LIMIT_AUTH_REFRESH", 0),
	}
}

// parseIntEnv parses an environment variable as int, returning defaultVal if empty or invalid
func parseIntEnv(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

func DefaultChatbotModel() string {
	v := os.Getenv("DEFAULT_CHATBOT_MODEL")
	if strings.TrimSpace(v) == "" {
		return DefaultModelName
	}
	return v
}
