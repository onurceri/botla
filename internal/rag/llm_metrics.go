package rag

import (
	"sync"
	"time"
)

// LLMMetrics collects metrics for LLM provider calls
type LLMMetrics struct {
	mu          sync.RWMutex
	totalCalls  map[string]int64
	failures    map[string]int64
	latencySum  map[string]float64
	lastLatency map[string]float64
}

// ProviderMetrics contains metrics for a single provider
type ProviderMetrics struct {
	TotalCalls    int64   `json:"total_calls"`
	Failures      int64   `json:"failures"`
	SuccessRate   float64 `json:"success_rate"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	LastLatencyMs float64 `json:"last_latency_ms"`
}

// NewLLMMetrics creates a new LLMMetrics instance
func NewLLMMetrics() *LLMMetrics {
	return &LLMMetrics{
		totalCalls:  make(map[string]int64),
		failures:    make(map[string]int64),
		latencySum:  make(map[string]float64),
		lastLatency: make(map[string]float64),
	}
}

// RecordSuccess records a successful LLM call
func (m *LLMMetrics) RecordSuccess(provider string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalCalls[provider]++
	latencyMs := float64(latency.Milliseconds())
	m.latencySum[provider] += latencyMs
	m.lastLatency[provider] = latencyMs
}

// RecordFailure records a failed LLM call
func (m *LLMMetrics) RecordFailure(provider string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalCalls[provider]++
	m.failures[provider]++
	latencyMs := float64(latency.Milliseconds())
	m.latencySum[provider] += latencyMs
	m.lastLatency[provider] = latencyMs
}

// GetProviderMetrics returns metrics for a specific provider
func (m *LLMMetrics) GetProviderMetrics(provider string) ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.totalCalls[provider]
	failures := m.failures[provider]

	var successRate float64
	if total > 0 {
		successRate = float64(total-failures) / float64(total)
	}

	var avgLatency float64
	if total > 0 {
		avgLatency = m.latencySum[provider] / float64(total)
	}

	return ProviderMetrics{
		TotalCalls:    total,
		Failures:      failures,
		SuccessRate:   successRate,
		AvgLatencyMs:  avgLatency,
		LastLatencyMs: m.lastLatency[provider],
	}
}

// GetAllMetrics returns metrics for all providers
func (m *LLMMetrics) GetAllMetrics() map[string]ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]ProviderMetrics)

	// Collect all unique providers
	providers := make(map[string]struct{})
	for p := range m.totalCalls {
		providers[p] = struct{}{}
	}

	for provider := range providers {
		total := m.totalCalls[provider]
		failures := m.failures[provider]

		var successRate float64
		if total > 0 {
			successRate = float64(total-failures) / float64(total)
		}

		var avgLatency float64
		if total > 0 {
			avgLatency = m.latencySum[provider] / float64(total)
		}

		result[provider] = ProviderMetrics{
			TotalCalls:    total,
			Failures:      failures,
			SuccessRate:   successRate,
			AvgLatencyMs:  avgLatency,
			LastLatencyMs: m.lastLatency[provider],
		}
	}

	return result
}

// Reset clears all metrics (useful for testing)
func (m *LLMMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalCalls = make(map[string]int64)
	m.failures = make(map[string]int64)
	m.latencySum = make(map[string]float64)
	m.lastLatency = make(map[string]float64)
}
