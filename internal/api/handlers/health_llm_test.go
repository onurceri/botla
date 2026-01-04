package handlers

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
)

// mockLLMClient implements rag.LLMClient for testing
type mockHealthLLMClient struct{}

func (m *mockHealthLLMClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	return &models.CompletionResult{Content: "test", UsageTokens: 10}, nil
}

func (m *mockHealthLLMClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{Name: "mock", Provider: "mock"}
}

func TestHealth_IncludesLLMStatus_WithFactory(t *testing.T) {
	// Test the circuit breaker client directly
	mockClient := &mockHealthLLMClient{}
	settings := rag.DefaultCircuitBreakerSettings()
	metrics := rag.NewLLMMetrics()
	cbClient := rag.NewCircuitBreakerClient("openai", mockClient, settings, metrics)

	// Test the circuit breaker client state
	if cbClient.State() != "closed" {
		t.Errorf("expected closed state, got %s", cbClient.State())
	}

	// Make a call and verify metrics
	_, err := cbClient.CreateCompletion(context.Background(), models.CompletionParams{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	providerMetrics := metrics.GetProviderMetrics("openai")
	if providerMetrics.TotalCalls != 1 {
		t.Errorf("expected 1 call, got %d", providerMetrics.TotalCalls)
	}
}

func TestHealth_CircuitBreakerStates(t *testing.T) {
	// Test circuit breaker state reporting
	mockClient := &mockHealthLLMClient{}
	settings := rag.CircuitBreakerSettings{
		MaxRequests:  1,
		Interval:     0,
		Timeout:      0,
		FailureRatio: 0.5,
		MinRequests:  2,
	}
	metrics := rag.NewLLMMetrics()

	cb := rag.NewCircuitBreakerClient("test-provider", mockClient, settings, metrics)

	// Initial state should be closed
	if cb.State() != "closed" {
		t.Errorf("expected closed state initially, got %s", cb.State())
	}

	// Make successful calls - state should remain closed
	for i := 0; i < 3; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	if cb.State() != "closed" {
		t.Errorf("expected closed state after successes, got %s", cb.State())
	}

	// Check metrics
	providerMetrics := metrics.GetProviderMetrics("test-provider")
	if providerMetrics.TotalCalls != 3 {
		t.Errorf("expected 3 calls, got %d", providerMetrics.TotalCalls)
	}
	if providerMetrics.SuccessRate != 1.0 {
		t.Errorf("expected 100%% success rate, got %f", providerMetrics.SuccessRate)
	}
}
