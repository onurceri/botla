package rag

import (
	"context"
	"errors"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/sony/gobreaker/v2"
)

// ErrCircuitOpen is returned when the circuit breaker is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreakerSettings configures circuit breaker behavior
type CircuitBreakerSettings struct {
	MaxRequests  uint32        // Max requests in half-open state (default: 3)
	Interval     time.Duration // Reset interval for closed state counters (default: 60s)
	Timeout      time.Duration // Time to stay open before half-open (default: 30s)
	FailureRatio float64       // Failure ratio to trip breaker (default: 0.5)
	MinRequests  uint32        // Min requests before tripping (default: 5)
}

// DefaultCircuitBreakerSettings returns sensible defaults
func DefaultCircuitBreakerSettings() CircuitBreakerSettings {
	return CircuitBreakerSettings{
		MaxRequests:  3,
		Interval:     60 * time.Second,
		Timeout:      30 * time.Second,
		FailureRatio: 0.5,
		MinRequests:  5,
	}
}

// CircuitBreakerClient wraps an LLMClient with circuit breaker functionality
type CircuitBreakerClient struct {
	name    string
	client  LLMClient
	breaker *gobreaker.CircuitBreaker[*models.CompletionResult]
	metrics *LLMMetrics
}

// NewCircuitBreakerClient creates a new circuit breaker wrapped client
func NewCircuitBreakerClient(name string, client LLMClient, settings CircuitBreakerSettings, metrics *LLMMetrics) *CircuitBreakerClient {
	cbSettings := gobreaker.Settings{
		Name:        name,
		MaxRequests: settings.MaxRequests,
		Interval:    settings.Interval,
		Timeout:     settings.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < settings.MinRequests {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= settings.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Could add logging here if needed
		},
	}

	return &CircuitBreakerClient{
		name:    name,
		client:  client,
		breaker: gobreaker.NewCircuitBreaker[*models.CompletionResult](cbSettings),
		metrics: metrics,
	}
}

// CreateCompletion calls the underlying client through the circuit breaker
func (c *CircuitBreakerClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	start := time.Now()

	result, err := c.breaker.Execute(func() (*models.CompletionResult, error) {
		return c.client.CreateCompletion(ctx, params)
	})

	latency := time.Since(start)
	if c.metrics != nil {
		if err != nil {
			c.metrics.RecordFailure(c.name, latency)
		} else {
			c.metrics.RecordSuccess(c.name, latency)
		}
	}

	if err != nil {
		// Check if it's a circuit breaker error
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, ErrCircuitOpen
		}
		return nil, err
	}

	return result, nil
}

// GetModelInfo delegates to the underlying client
func (c *CircuitBreakerClient) GetModelInfo() models.ModelInfo {
	return c.client.GetModelInfo()
}

// State returns the current circuit breaker state as a string
func (c *CircuitBreakerClient) State() string {
	state := c.breaker.State()
	switch state {
	case gobreaker.StateClosed:
		return "closed"
	case gobreaker.StateHalfOpen:
		return "half-open"
	case gobreaker.StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// Counts returns the current circuit breaker counts
func (c *CircuitBreakerClient) Counts() gobreaker.Counts {
	return c.breaker.Counts()
}

// Name returns the circuit breaker name
func (c *CircuitBreakerClient) Name() string {
	return c.name
}

// UnwrappedClient returns the underlying LLMClient
func (c *CircuitBreakerClient) UnwrappedClient() LLMClient {
	return c.client
}

// CircuitBreakerManager manages multiple circuit breakers for different providers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreakerClient
	metrics  *LLMMetrics
	settings CircuitBreakerSettings
}

// NewCircuitBreakerManager creates a new manager with default settings
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreakerClient),
		metrics:  NewLLMMetrics(),
		settings: DefaultCircuitBreakerSettings(),
	}
}

// NewCircuitBreakerManagerWithSettings creates a new manager with custom settings
func NewCircuitBreakerManagerWithSettings(settings CircuitBreakerSettings) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreakerClient),
		metrics:  NewLLMMetrics(),
		settings: settings,
	}
}

// RegisterClient registers a client with circuit breaker protection
func (m *CircuitBreakerManager) RegisterClient(name string, client LLMClient) *CircuitBreakerClient {
	cbClient := NewCircuitBreakerClient(name, client, m.settings, m.metrics)
	m.breakers[name] = cbClient
	return cbClient
}

// GetClient returns the circuit breaker client for a provider
func (m *CircuitBreakerManager) GetClient(name string) (*CircuitBreakerClient, bool) {
	client, ok := m.breakers[name]
	return client, ok
}

// GetMetrics returns the shared metrics instance
func (m *CircuitBreakerManager) GetMetrics() *LLMMetrics {
	return m.metrics
}

// GetAllStates returns the state of all circuit breakers
func (m *CircuitBreakerManager) GetAllStates() map[string]string {
	states := make(map[string]string)
	for name, cb := range m.breakers {
		states[name] = cb.State()
	}
	return states
}

// GetProviderStatus returns detailed status for a provider
type ProviderStatus struct {
	Circuit string          `json:"circuit"`
	Metrics ProviderMetrics `json:"metrics"`
}

// GetAllProviderStatus returns status for all registered providers
func (m *CircuitBreakerManager) GetAllProviderStatus() map[string]ProviderStatus {
	status := make(map[string]ProviderStatus)
	for name, cb := range m.breakers {
		status[name] = ProviderStatus{
			Circuit: cb.State(),
			Metrics: m.metrics.GetProviderMetrics(name),
		}
	}
	return status
}
