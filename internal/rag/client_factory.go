package rag

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/onurceri/botla-app/pkg/config"
)

// ClientFactory manages LLM client creation
// Architecture: OpenRouter (primary LLM) + OpenAI (embeddings + fallback)
type ClientFactory struct {
	customClients map[string]LLMClient
	cfg           *config.Config
	cbMu          sync.Mutex // guards cbManager against concurrent lazy-init and reads
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
	m := NewCircuitBreakerManager()
	f.registerConfiguredProviders(m)
	f.setCBManager(m)
	return nil
}

// InitCircuitBreakersWithSettings initializes circuit breakers with custom settings
func (f *ClientFactory) InitCircuitBreakersWithSettings(settings CircuitBreakerSettings) error {
	m := NewCircuitBreakerManagerWithSettings(settings)
	f.registerConfiguredProviders(m)
	f.setCBManager(m)
	return nil
}

// registerConfiguredProviders registers all configured providers on the given manager.
// It does not touch f.cbManager and is safe to call with or without f.cbMu held.
func (f *ClientFactory) registerConfiguredProviders(m *CircuitBreakerManager) {
	// Register OpenRouter if configured
	if f.IsProviderConfigured("openrouter") {
		if client, err := f.GetClient("openrouter"); err == nil {
			m.RegisterClient("openrouter", client)
		}
	}

	// Register OpenAI if configured
	if f.IsProviderConfigured("openai") {
		if client, err := f.GetClient("openai"); err == nil {
			m.RegisterClient("openai", client)
		}
	}
}

// setCBManager publishes the manager, guarding the pointer write.
func (f *ClientFactory) setCBManager(m *CircuitBreakerManager) {
	f.cbMu.Lock()
	f.cbManager = m
	f.cbMu.Unlock()
}

// getCBManager returns the manager, lazily initializing it under the lock so that
// concurrent first-callers cannot trigger a double init (or read a half-written pointer).
func (f *ClientFactory) getCBManager() *CircuitBreakerManager {
	f.cbMu.Lock()
	defer f.cbMu.Unlock()
	if f.cbManager == nil {
		m := NewCircuitBreakerManager()
		f.registerConfiguredProviders(m)
		f.cbManager = m
	}
	return f.cbManager
}

// GetCircuitBreakerClient returns a circuit breaker wrapped client for the provider
func (f *ClientFactory) GetCircuitBreakerClient(provider string) (*CircuitBreakerClient, error) {
	m := f.getCBManager()

	client, ok := m.GetClient(provider)
	if !ok {
		return nil, errors.New("no circuit breaker registered for provider: " + provider)
	}
	return client, nil
}

// GetCircuitBreakerStatus returns the current status of all circuit breakers
func (f *ClientFactory) GetCircuitBreakerStatus() map[string]ProviderStatus {
	f.cbMu.Lock()
	m := f.cbManager
	f.cbMu.Unlock()
	if m == nil {
		return make(map[string]ProviderStatus)
	}
	return m.GetAllProviderStatus()
}

// GetLLMMetrics returns the LLM metrics if circuit breakers are initialized
func (f *ClientFactory) GetLLMMetrics() *LLMMetrics {
	f.cbMu.Lock()
	m := f.cbManager
	f.cbMu.Unlock()
	if m == nil {
		return nil
	}
	return m.GetMetrics()
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
