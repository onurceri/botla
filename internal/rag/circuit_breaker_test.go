package rag

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
)

// mockLLMClient is a test implementation of LLMClient
type mockLLMClient struct {
	shouldFail   bool
	failureCount int
	callCount    int
	delay        time.Duration
}

func (m *mockLLMClient) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	m.callCount++
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.shouldFail {
		m.failureCount++
		return nil, errors.New("mock failure")
	}
	return &models.CompletionResult{
		Content:     "test response",
		UsageTokens: 10,
	}, nil
}

func (m *mockLLMClient) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{
		Name:     "mock-model",
		Provider: "mock",
	}
}

func TestCircuitBreaker_ClosedState(t *testing.T) {
	mock := &mockLLMClient{shouldFail: false}
	metrics := NewLLMMetrics()
	settings := DefaultCircuitBreakerSettings()
	cb := NewCircuitBreakerClient("test", mock, settings, metrics)

	// Should be closed initially
	if cb.State() != "closed" {
		t.Errorf("expected closed state, got %s", cb.State())
	}

	// Requests should pass through
	result, err := cb.CreateCompletion(context.Background(), models.CompletionParams{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Content != "test response" {
		t.Errorf("unexpected content: %s", result.Content)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
}

func TestCircuitBreaker_OpensOnFailures(t *testing.T) {
	mock := &mockLLMClient{shouldFail: true}
	metrics := NewLLMMetrics()
	settings := CircuitBreakerSettings{
		MaxRequests:  1,
		Interval:     time.Second,
		Timeout:      time.Second,
		FailureRatio: 0.5,
		MinRequests:  3, // Trip after 3 requests with 50% failure
	}
	cb := NewCircuitBreakerClient("test", mock, settings, metrics)

	// Make enough failing requests to trip the breaker
	for i := 0; i < 4; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	// Circuit should be open now
	if cb.State() != "open" {
		t.Errorf("expected open state, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpenStateRejectsQuickly(t *testing.T) {
	mock := &mockLLMClient{shouldFail: true}
	metrics := NewLLMMetrics()
	settings := CircuitBreakerSettings{
		MaxRequests:  1,
		Interval:     time.Minute,
		Timeout:      time.Minute, // Long timeout so it stays open
		FailureRatio: 0.5,
		MinRequests:  2,
	}
	cb := NewCircuitBreakerClient("test", mock, settings, metrics)

	// Trip the breaker
	for i := 0; i < 3; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	callsBefore := mock.callCount

	// Wait a tiny bit for state to settle
	time.Sleep(10 * time.Millisecond)

	// Now requests should be rejected without calling the underlying client
	_, err := cb.CreateCompletion(context.Background(), models.CompletionParams{})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}

	if mock.callCount != callsBefore {
		t.Errorf("expected no additional calls, got %d new calls", mock.callCount-callsBefore)
	}
}

func TestCircuitBreaker_HalfOpenAllowsProbe(t *testing.T) {
	mock := &mockLLMClient{shouldFail: true}
	metrics := NewLLMMetrics()
	settings := CircuitBreakerSettings{
		MaxRequests:  1,
		Interval:     time.Second,
		Timeout:      50 * time.Millisecond, // Short timeout for testing
		FailureRatio: 0.5,
		MinRequests:  2,
	}
	cb := NewCircuitBreakerClient("test", mock, settings, metrics)

	// Trip the breaker
	for i := 0; i < 3; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	// Wait for timeout to transition to half-open
	time.Sleep(100 * time.Millisecond)

	// Circuit should transition to half-open after timeout
	// A probe request should be allowed
	callsBefore := mock.callCount
	_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})

	// At least one probe call should have been made
	if mock.callCount <= callsBefore {
		t.Errorf("expected at least one probe call in half-open state")
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	mock := &mockLLMClient{shouldFail: true}
	metrics := NewLLMMetrics()
	settings := CircuitBreakerSettings{
		MaxRequests:  1,
		Interval:     time.Second,
		Timeout:      50 * time.Millisecond,
		FailureRatio: 0.5,
		MinRequests:  2,
	}
	cb := NewCircuitBreakerClient("test", mock, settings, metrics)

	// Trip the breaker
	for i := 0; i < 3; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Now make successful request
	mock.shouldFail = false
	_, err := cb.CreateCompletion(context.Background(), models.CompletionParams{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Circuit should be closed now
	if cb.State() != "closed" {
		t.Errorf("expected closed state after success, got %s", cb.State())
	}
}

func TestCircuitBreaker_MetricsRecorded(t *testing.T) {
	mock := &mockLLMClient{shouldFail: false}
	metrics := NewLLMMetrics()
	settings := DefaultCircuitBreakerSettings()
	cb := NewCircuitBreakerClient("test-provider", mock, settings, metrics)

	// Make some successful calls
	for i := 0; i < 3; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	providerMetrics := metrics.GetProviderMetrics("test-provider")
	if providerMetrics.TotalCalls != 3 {
		t.Errorf("expected 3 total calls, got %d", providerMetrics.TotalCalls)
	}
	if providerMetrics.Failures != 0 {
		t.Errorf("expected 0 failures, got %d", providerMetrics.Failures)
	}
	if providerMetrics.SuccessRate != 1.0 {
		t.Errorf("expected 1.0 success rate, got %f", providerMetrics.SuccessRate)
	}

	// Make some failing calls
	mock.shouldFail = true
	for i := 0; i < 2; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	providerMetrics = metrics.GetProviderMetrics("test-provider")
	if providerMetrics.TotalCalls != 5 {
		t.Errorf("expected 5 total calls, got %d", providerMetrics.TotalCalls)
	}
	if providerMetrics.Failures != 2 {
		t.Errorf("expected 2 failures, got %d", providerMetrics.Failures)
	}
}

func TestCircuitBreakerManager_RegisterAndGet(t *testing.T) {
	manager := NewCircuitBreakerManager()
	mock := &mockLLMClient{}

	manager.RegisterClient("openai", mock)
	manager.RegisterClient("openrouter", mock)

	client, ok := manager.GetClient("openai")
	if !ok || client == nil {
		t.Error("expected to get openai client")
	}

	_, ok = manager.GetClient("nonexistent")
	if ok {
		t.Error("expected nonexistent client to not be found")
	}
}

func TestCircuitBreakerManager_GetAllStates(t *testing.T) {
	manager := NewCircuitBreakerManager()
	mock := &mockLLMClient{}

	manager.RegisterClient("openai", mock)
	manager.RegisterClient("openrouter", mock)

	states := manager.GetAllStates()
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
	if states["openai"] != "closed" {
		t.Errorf("expected openai to be closed, got %s", states["openai"])
	}
	if states["openrouter"] != "closed" {
		t.Errorf("expected openrouter to be closed, got %s", states["openrouter"])
	}
}

func TestCircuitBreakerManager_GetAllProviderStatus(t *testing.T) {
	manager := NewCircuitBreakerManager()
	mock := &mockLLMClient{}

	cb := manager.RegisterClient("openai", mock)

	// Make some calls
	for i := 0; i < 3; i++ {
		_, _ = cb.CreateCompletion(context.Background(), models.CompletionParams{})
	}

	status := manager.GetAllProviderStatus()
	openaiStatus := status["openai"]

	if openaiStatus.Circuit != "closed" {
		t.Errorf("expected closed circuit, got %s", openaiStatus.Circuit)
	}
	if openaiStatus.Metrics.TotalCalls != 3 {
		t.Errorf("expected 3 total calls, got %d", openaiStatus.Metrics.TotalCalls)
	}
}

func TestDefaultCircuitBreakerSettings(t *testing.T) {
	settings := DefaultCircuitBreakerSettings()

	if settings.MaxRequests != 3 {
		t.Errorf("expected MaxRequests 3, got %d", settings.MaxRequests)
	}
	if settings.Interval != 60*time.Second {
		t.Errorf("expected Interval 60s, got %v", settings.Interval)
	}
	if settings.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", settings.Timeout)
	}
	if settings.FailureRatio != 0.5 {
		t.Errorf("expected FailureRatio 0.5, got %f", settings.FailureRatio)
	}
	if settings.MinRequests != 5 {
		t.Errorf("expected MinRequests 5, got %d", settings.MinRequests)
	}
}

func TestCircuitBreakerClient_Name(t *testing.T) {
	mock := &mockLLMClient{}
	metrics := NewLLMMetrics()
	cb := NewCircuitBreakerClient("test-name", mock, DefaultCircuitBreakerSettings(), metrics)

	if cb.Name() != "test-name" {
		t.Errorf("expected name 'test-name', got %s", cb.Name())
	}
}

func TestCircuitBreakerClient_UnwrappedClient(t *testing.T) {
	mock := &mockLLMClient{}
	metrics := NewLLMMetrics()
	cb := NewCircuitBreakerClient("test", mock, DefaultCircuitBreakerSettings(), metrics)

	unwrapped := cb.UnwrappedClient()
	if unwrapped != mock {
		t.Error("expected to get the original mock client")
	}
}

func TestCircuitBreakerClient_GetModelInfo(t *testing.T) {
	mock := &mockLLMClient{}
	metrics := NewLLMMetrics()
	cb := NewCircuitBreakerClient("test", mock, DefaultCircuitBreakerSettings(), metrics)

	info := cb.GetModelInfo()
	if info.Name != "mock-model" {
		t.Errorf("expected mock-model, got %s", info.Name)
	}
	if info.Provider != "mock" {
		t.Errorf("expected mock provider, got %s", info.Provider)
	}
}
