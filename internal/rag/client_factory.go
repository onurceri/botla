package rag

import (
	"errors"
	"os"
	"strings"
	"github.com/onurceri/botla-co/pkg/config"
)

// ClientFactory manages LLM client creation
type ClientFactory struct{}

// NewClientFactory creates a new ClientFactory
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// GetClient returns an LLMClient for the specified provider
// If provider is empty, it defaults to openai
func (f *ClientFactory) GetClient(provider string) (LLMClient, error) {
	if provider == "" {
		provider = "openai"
	}

	switch strings.ToLower(provider) {
	case "openai":
		return NewOpenAIClientFromEnv()
	case "anthropic":
		return NewAnthropicClientFromEnv()
	case "google":
		return NewGoogleAIClientFromEnv()
	case "openrouter":
		return NewOpenRouterClientFromEnv()
	default:
		return nil, errors.New("unsupported provider: " + provider)
	}
}

// GetClientForModel parses the model string (provider:model) and returns the appropriate client
func (f *ClientFactory) GetClientForModel(modelString string) (LLMClient, string, error) {
	if !config.IsModelSupported(modelString) {
		return nil, "", errors.New("unsupported model: " + modelString)
	}

	parts := strings.SplitN(modelString, ":", 2)
	var provider, modelName string

	if len(parts) == 2 {
		provider = parts[0]
		modelName = parts[1]
	} else {
		// Default to openai if no prefix
		provider = "openai"
		modelName = modelString
	}

	client, err := f.GetClient(provider)
	if err != nil {
		return nil, "", err
	}

	return client, modelName, nil
}

// GetAvailableProviders returns a list of configured providers
func (f *ClientFactory) GetAvailableProviders() []string {
	providers := []string{}

	if os.Getenv("OPENAI_API_KEY") != "" {
		providers = append(providers, "openai")
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		providers = append(providers, "anthropic")
	}
	if os.Getenv("GOOGLE_AI_API_KEY") != "" {
		providers = append(providers, "google")
	}
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		providers = append(providers, "openrouter")
	}

	return providers
}

// IsProviderConfigured checks if a provider is configured (via env vars)
func (f *ClientFactory) IsProviderConfigured(provider string) bool {
	switch strings.ToLower(provider) {
	case "openai":
		return os.Getenv("OPENAI_API_KEY") != ""
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY") != ""
	case "google":
		return os.Getenv("GOOGLE_AI_API_KEY") != ""
	case "openrouter":
		return os.Getenv("OPENROUTER_API_KEY") != ""
	default:
		return false
	}
}
