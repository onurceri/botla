package ai

import (
	"os"
	"testing"
)

func TestConfigFromEnv_Defaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg := ConfigFromEnv()

	if cfg.VectorProvider != ProviderQdrant {
		t.Errorf("expected VectorProvider to be 'qdrant', got %s", cfg.VectorProvider)
	}
	if cfg.EmbedderProvider != ProviderOpenAI {
		t.Errorf("expected EmbedderProvider to be 'openai', got %s", cfg.EmbedderProvider)
	}
	if cfg.LLMProvider != ProviderOpenRouter {
		t.Errorf("expected LLMProvider to be 'openrouter', got %s", cfg.LLMProvider)
	}
}

func TestConfigFromEnv_CustomProviders(t *testing.T) {
	t.Setenv("VECTOR_PROVIDER", "pgvector")
	t.Setenv("EMBEDDER_PROVIDER", "openrouter")
	t.Setenv("LLM_PROVIDER", "openai")

	cfg := ConfigFromEnv()

	if cfg.VectorProvider != ProviderPGVector {
		t.Errorf("expected VectorProvider to be 'pgvector', got %s", cfg.VectorProvider)
	}
	if cfg.EmbedderProvider != ProviderOpenRouter {
		t.Errorf("expected EmbedderProvider to be 'openrouter', got %s", cfg.EmbedderProvider)
	}
	if cfg.LLMProvider != ProviderOpenAI {
		t.Errorf("expected LLMProvider to be 'openai', got %s", cfg.LLMProvider)
	}
}

func TestNewVectorStore_Qdrant(t *testing.T) {
	// Register a test factory
	RegisterVectorStore(ProviderQdrant, func() (VectorStore, error) {
		return &MockVectorStore{}, nil
	})
	defer delete(vectorStoreFactories, ProviderQdrant)

	cfg := Config{VectorProvider: ProviderQdrant}
	store, err := NewVectorStore(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Error("expected store to be created")
	}
}

func TestNewVectorStore_PGVector(t *testing.T) {
	cfg := Config{VectorProvider: ProviderPGVector}
	_, err := NewVectorStore(cfg)
	if err == nil {
		t.Error("expected error for unregistered provider")
	}
}

func TestNewVectorStore_Unknown(t *testing.T) {
	cfg := Config{VectorProvider: "invalid"}
	_, err := NewVectorStore(cfg)
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestNewEmbedder_OpenAI(t *testing.T) {
	// Register a test factory
	RegisterEmbedder(ProviderOpenAI, func() (Embedder, error) {
		return &MockEmbedder{}, nil
	})
	defer delete(embedderFactories, ProviderOpenAI)

	cfg := Config{EmbedderProvider: ProviderOpenAI}
	embedder, err := NewEmbedder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedder == nil {
		t.Error("expected embedder to be created")
	}
}

func TestNewEmbedder_OpenRouter(t *testing.T) {
	// Register a test factory
	RegisterEmbedder(ProviderOpenRouter, func() (Embedder, error) {
		return &MockEmbedder{}, nil
	})
	defer delete(embedderFactories, ProviderOpenRouter)

	cfg := Config{EmbedderProvider: ProviderOpenRouter}
	embedder, err := NewEmbedder(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedder == nil {
		t.Error("expected embedder to be created")
	}
}

func TestNewEmbedder_Unknown(t *testing.T) {
	cfg := Config{EmbedderProvider: "invalid"}
	_, err := NewEmbedder(cfg)
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "env not set",
			key:          "MISSING_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
