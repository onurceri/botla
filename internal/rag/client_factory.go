package rag

import (
	"errors"
	"os"
	"strings"

	"github.com/onurceri/botla-app/pkg/config"
)

// ClientFactory manages LLM client creation
// Architecture: OpenRouter (primary LLM) + OpenAI (embeddings + fallback)
type ClientFactory struct {
	customClients map[string]LLMClient
	cfg           *config.Config
	cbManager     *CircuitBreakerManager
}

// NewClientFactory creates a new ClientFactory with config
func NewClientFactory(cfg *config.Config) *ClientFactory {
	return &ClientFactory{
		customClients: make(map[string]LLMClient),
		cfg:           cfg,
		cbManager:     nil, // Lazily initialized
	}
}

// InitCircuitBreakers initializes circuit breakers for all configured providers
func (f *ClientFactory) InitCircuitBreakers() error {
	f.cbManager = NewCircuitBreakerManager()

	// Register OpenRouter if configured
	if f.IsProviderConfigured("openrouter") {
		client, err := f.GetClient("openrouter")
		if err == nil {
			f.cbManager.RegisterClient("openrouter", client)
		}
	}

	// Register OpenAI if configured
	if f.IsProviderConfigured("openai") {
		client, err := f.GetClient("openai")
		if err == nil {
			f.cbManager.RegisterClient("openai", client)
		}
	}

	return nil
}

// InitCircuitBreakersWithSettings initializes circuit breakers with custom settings
func (f *ClientFactory) InitCircuitBreakersWithSettings(settings CircuitBreakerSettings) error {
	f.cbManager = NewCircuitBreakerManagerWithSettings(settings)

	if f.IsProviderConfigured("openrouter") {
		client, err := f.GetClient("openrouter")
		if err == nil {
			f.cbManager.RegisterClient("openrouter", client)
		}
	}

	if f.IsProviderConfigured("openai") {
		client, err := f.GetClient("openai")
		if err == nil {
			f.cbManager.RegisterClient("openai", client)
		}
	}

	return nil
}

// GetCircuitBreakerClient returns a circuit breaker wrapped client for the provider
func (f *ClientFactory) GetCircuitBreakerClient(provider string) (*CircuitBreakerClient, error) {
	if f.cbManager == nil {
		if err := f.InitCircuitBreakers(); err != nil {
			return nil, err
		}
	}

	client, ok := f.cbManager.GetClient(provider)
	if !ok {
		return nil, errors.New("no circuit breaker registered for provider: " + provider)
	}
	return client, nil
}

// GetCircuitBreakerStatus returns the current status of all circuit breakers
func (f *ClientFactory) GetCircuitBreakerStatus() map[string]ProviderStatus {
	if f.cbManager == nil {
		return make(map[string]ProviderStatus)
	}
	return f.cbManager.GetAllProviderStatus()
}

// GetLLMMetrics returns the LLM metrics if circuit breakers are initialized
func (f *ClientFactory) GetLLMMetrics() *LLMMetrics {
	if f.cbManager == nil {
		return nil
	}
	return f.cbManager.GetMetrics()
}

// RegisterClient registers a custom client (useful for mocking)
func (f *ClientFactory) RegisterClient(provider string, client LLMClient) {
	if f.customClients == nil {
		f.customClients = make(map[string]LLMClient)
	}
	f.customClients[strings.ToLower(provider)] = client
}

// GetClient returns an LLMClient for the specified provider
// Supported providers: openrouter (primary), openai (fallback + embeddings)
func (f *ClientFactory) GetClient(provider string) (LLMClient, error) {
	if provider == "" {
		provider = "openrouter" // Default to OpenRouter for LLM calls
	}

	switch strings.ToLower(provider) {
	case "openrouter":
		if c, ok := f.customClients["openrouter"]; ok {
			return c, nil
		}
		// Pass current config, NewOpenRouterClient will re-check env for base URL
		return NewOpenRouterClient(f.cfg)
	case "openai":
		if c, ok := f.customClients["openai"]; ok {
			return c, nil
		}
		// Pass current config, NewOpenAIClient will re-check env for base URL
		return NewOpenAIClient(f.cfg)
	default:
		return nil, errors.New("unsupported provider: " + provider + " (use openrouter or openai)")
	}
}

// GetClientForModel parses the model string (provider:model) and returns the appropriate client
// Default provider is openrouter for LLM operations
func (f *ClientFactory) GetClientForModel(modelString string) (LLMClient, string, error) {
	parts := strings.SplitN(modelString, ":", 2)
	var provider, modelName string

	if len(parts) == 2 {
		provider = parts[0]
		modelName = parts[1]

		// Map other providers to OpenRouter (except OpenAI which has native client)
		if provider != "openai" && provider != "openrouter" {
			// Construct OpenRouter model format: provider/model
			// e.g. anthropic:claude -> anthropic/claude
			modelName = provider + "/" + modelName
			provider = "openrouter"
		}
	} else {
		// Default behavior: models with "/" are OpenRouter, others are OpenAI
		if strings.Contains(modelString, "/") {
			provider = "openrouter"
			modelName = modelString
		} else {
			provider = "openai"
			modelName = modelString
		}
	}

	// Force redirect to OpenAI if OPENAI_API_BASE is local (for tests)
	if os.Getenv("OPENAI_API_BASE") != "" && strings.Contains(os.Getenv("OPENAI_API_BASE"), "127.0.0.1") {
		if strings.HasPrefix(modelName, "gpt-") {
			provider = "openai"
		}
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

	if f.cfg != nil && f.cfg.OPENROUTER_API_KEY != "" {
		providers = append(providers, "openrouter")
	}
	if f.cfg != nil && f.cfg.OPENAI_API_KEY != "" {
		providers = append(providers, "openai")
	}

	return providers
}

// IsProviderConfigured checks if a provider is configured
func (f *ClientFactory) IsProviderConfigured(provider string) bool {
	if f.cfg == nil {
		return false
	}
	switch strings.ToLower(provider) {
	case "openrouter":
		return f.cfg.OPENROUTER_API_KEY != ""
	case "openai":
		return f.cfg.OPENAI_API_KEY != ""
	default:
		return false
	}
}
