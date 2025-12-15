package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_HOST               string
	DB_PORT               string
	DB_NAME               string
	DB_USER               string
	DB_PASSWORD           string
	DB_SCHEMA             string
	DB_SSLMODE            string
	QDRANT_URL            string
	QDRANT_API_KEY        string
	OPENAI_API_KEY        string
	OPENAI_API_BASE       string
	OPENAI_TIMEOUT_MS     int
	OPENROUTER_API_KEY    string
	OPENROUTER_API_BASE   string
	OPENROUTER_TIMEOUT_MS int
	IYZICO_API_KEY        string
	IYZICO_SECRET_KEY     string
	JWT_SECRET            string
	PORT                  string
	CORS_ALLOWED_ORIGINS  string
	R2_ACCOUNT_ID         string
	R2_ACCESS_KEY_ID      string
	R2_SECRET_ACCESS_KEY  string
	R2_BUCKET_NAME        string
	DEFAULT_CHATBOT_MODEL string

	// RAG Configuration
	RAG_TOPK               int
	RAG_MAX_CONTEXT_TOKENS int

	// Chat Configuration
	CHAT_TIMEOUT_MS int

	// Environment
	GO_ENV string
}

var fatalf = func(msg string) { log.Fatal(msg) }

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
		log.Println("WARNING: OPENAI_API_KEY is missing. Embeddings will fail.")
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
		R2_ACCOUNT_ID:          os.Getenv("R2_ACCOUNT_ID"),
		R2_ACCESS_KEY_ID:       os.Getenv("R2_ACCESS_KEY_ID"),
		R2_SECRET_ACCESS_KEY:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2_BUCKET_NAME:         os.Getenv("R2_BUCKET_NAME"),
		DEFAULT_CHATBOT_MODEL:  DefaultChatbotModel(),
		RAG_TOPK:               parseIntEnv("RAG_TOPK", 5),
		RAG_MAX_CONTEXT_TOKENS: parseIntEnv("RAG_MAX_CONTEXT_TOKENS", 2000),
		CHAT_TIMEOUT_MS:        parseIntEnv("CHAT_TIMEOUT_MS", 60000),
		GO_ENV:                 os.Getenv("GO_ENV"),
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
		return ModelGPT4oMini
	}
	return v
}

func ResolveChatbotModel(cfg *Config) string {
	v := DefaultChatbotModel()
	if cfg != nil && strings.TrimSpace(cfg.DEFAULT_CHATBOT_MODEL) != "" {
		v = cfg.DEFAULT_CHATBOT_MODEL
	}
	return v
}

func GetDefaultModelForProvider(provider string) string {
	switch strings.ToLower(provider) {
	case "openai":
		return ModelGPT4oMini
	case "anthropic":
		return ModelClaude35Sonnet
	case "google":
		return ModelGemini15Flash
	default:
		return ModelGPT4oMini
	}
}

// NormalizeModelName extracts the bare model name from a provider-prefixed string.
// Handles formats: "openai:gpt-4o-mini", "openai/gpt-4o-mini", "gpt-4o-mini"
// Returns the bare model name (e.g., "gpt-4o-mini") for unified comparison.
func NormalizeModelName(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	// Handle colon prefix (openai:gpt-4o-mini)
	if parts := strings.SplitN(model, ":", 2); len(parts) == 2 {
		model = parts[1]
	}
	// Handle slash prefix (openai/gpt-4o-mini)
	if parts := strings.SplitN(model, "/", 2); len(parts) == 2 {
		model = parts[1]
	}
	return model
}

// IsModelAllowed checks if a model (possibly with provider prefix) matches any allowed model.
// The allowed models list can contain either bare names or prefixed names.
func IsModelAllowed(model string, allowedModels []string) bool {
	normalizedInput := NormalizeModelName(model)
	for _, allowed := range allowedModels {
		normalizedAllowed := NormalizeModelName(allowed)
		if normalizedInput == normalizedAllowed {
			return true
		}
	}
	return false
}

// IsModelSupported checks if a model is supported by the system
// Architecture: OpenRouter (primary LLM) + OpenAI (embeddings + fallback)
func IsModelSupported(model string) bool {
	// Handle provider prefixes
	parts := strings.SplitN(model, ":", 2)
	provider := "openrouter" // Default to OpenRouter
	modelName := model
	if len(parts) == 2 {
		provider = strings.ToLower(parts[0])
		modelName = parts[1]
	}

	switch provider {
	case "openai":
		// Known OpenAI models (used directly or via fallback)
		valid := []string{
			"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo", "o1", "o1-mini", "o1-preview", "o3-mini",
		}
		for _, v := range valid {
			if strings.HasPrefix(modelName, v) {
				return true
			}
		}
		return false
	case "openrouter":
		// OpenRouter allows any model via its gateway
		// Common format: "openai/gpt-4o-mini", "anthropic/claude-3.5-sonnet"
		return true
	default:
		// For backwards compatibility, also allow bare model names (mapped to OpenRouter)
		return true
	}
}
