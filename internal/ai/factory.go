package ai

import (
	"fmt"
	"os"
)

// ProviderType represents a provider implementation
type ProviderType string

const (
	// ProviderOpenAI represents OpenAI provider
	ProviderOpenAI ProviderType = "openai"

	// ProviderOpenRouter represents OpenRouter provider
	ProviderOpenRouter ProviderType = "openrouter"

	// ProviderQdrant represents Qdrant vector database
	ProviderQdrant ProviderType = "qdrant"

	// ProviderPGVector represents PostgreSQL pgvector extension (future)
	ProviderPGVector ProviderType = "pgvector"
)

// Config holds provider configuration
type Config struct {
	VectorProvider   ProviderType
	EmbedderProvider ProviderType
	LLMProvider      ProviderType
}

// ConfigFromEnv creates a Config from environment variables
func ConfigFromEnv() Config {
	return Config{
		VectorProvider:   ProviderType(getEnvOrDefault("VECTOR_PROVIDER", "qdrant")),
		EmbedderProvider: ProviderType(getEnvOrDefault("EMBEDDER_PROVIDER", "openai")),
		LLMProvider:      ProviderType(getEnvOrDefault("LLM_PROVIDER", "openrouter")),
	}
}

// getEnvOrDefault retrieves an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// VectorStoreFactory is a function that creates a VectorStore
type VectorStoreFactory func() (VectorStore, error)

// EmbedderFactory is a function that creates an Embedder
type EmbedderFactory func() (Embedder, error)

var (
	vectorStoreFactories = make(map[ProviderType]VectorStoreFactory)
	embedderFactories    = make(map[ProviderType]EmbedderFactory)
)

// RegisterVectorStore registers a factory for a vector store provider
func RegisterVectorStore(provider ProviderType, factory VectorStoreFactory) {
	vectorStoreFactories[provider] = factory
}

// RegisterEmbedder registers a factory for an embedder provider
func RegisterEmbedder(provider ProviderType, factory EmbedderFactory) {
	embedderFactories[provider] = factory
}

// NewVectorStore creates a VectorStore based on the configuration
func NewVectorStore(cfg Config) (VectorStore, error) {
	factory, ok := vectorStoreFactories[cfg.VectorProvider]
	if !ok {
		return nil, fmt.Errorf("unknown vector provider: %s", cfg.VectorProvider)
	}
	return factory()
}

// NewEmbedder creates an Embedder based on the configuration
func NewEmbedder(cfg Config) (Embedder, error) {
	factory, ok := embedderFactories[cfg.EmbedderProvider]
	if !ok {
		return nil, fmt.Errorf("unknown embedder provider: %s", cfg.EmbedderProvider)
	}
	return factory()
}
