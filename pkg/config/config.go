package config

import (
	"log"
	"os"
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
	QDRANT_URL            string
	OPENAI_API_KEY        string
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

	// Check for at least one LLM provider
	if os.Getenv("OPENAI_API_KEY") == "" &&
		os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("GOOGLE_AI_API_KEY") == "" {
		fatalf("At least one LLM API key (OPENAI_API_KEY, ANTHROPIC_API_KEY, GOOGLE_AI_API_KEY) must be provided")
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
		QDRANT_URL:        os.Getenv("QDRANT_URL"),
		OPENAI_API_KEY:    os.Getenv("OPENAI_API_KEY"),
		IYZICO_API_KEY:    os.Getenv("IYZICO_API_KEY"),
		IYZICO_SECRET_KEY: os.Getenv("IYZICO_SECRET_KEY"),
		JWT_SECRET:        os.Getenv("JWT_SECRET"),
		PORT:              os.Getenv("PORT"),
		CORS_ALLOWED_ORIGINS: func() string {
			v := os.Getenv("CORS_ALLOWED_ORIGINS")
			if v == "" {
				return "http://localhost:5173"
			}
			return v
		}(),
		R2_ACCOUNT_ID:         os.Getenv("R2_ACCOUNT_ID"),
		R2_ACCESS_KEY_ID:      os.Getenv("R2_ACCESS_KEY_ID"),
		R2_SECRET_ACCESS_KEY:  os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2_BUCKET_NAME:        os.Getenv("R2_BUCKET_NAME"),
		DEFAULT_CHATBOT_MODEL: DefaultChatbotModel(),
	}
}

func DefaultChatbotModel() string {
	v := os.Getenv("DEFAULT_CHATBOT_MODEL")
	if strings.TrimSpace(v) == "" {
		return "gpt-4o-mini"
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
		return "gpt-4o-mini"
	case "anthropic":
		return "claude-3-5-sonnet-20241022"
	case "google":
		return "gemini-1.5-flash"
	default:
		return "gpt-4o-mini"
	}
}

// IsModelSupported checks if a model is supported by the system
// This is a basic validation to prevent invalid model names from being passed to providers
// For OpenRouter, we allow all models as it's a gateway.
func IsModelSupported(model string) bool {
	// Handle provider prefixes
	parts := strings.SplitN(model, ":", 2)
	provider := "openai"
	modelName := model
	if len(parts) == 2 {
		provider = strings.ToLower(parts[0])
		modelName = parts[1]
	}

	switch provider {
	case "openai":
		// Known OpenAI models
		valid := []string{
			"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo",
		}
		for _, v := range valid {
			if strings.HasPrefix(modelName, v) {
				return true
			}
		}
		return false
	case "anthropic":
		// Known Anthropic models
		return strings.Contains(modelName, "claude")
	case "google":
		// Known Google models
		return strings.Contains(modelName, "gemini")
	case "openrouter":
		// OpenRouter allows many models, trust the user/config
		return true
	default:
		return false
	}
}
