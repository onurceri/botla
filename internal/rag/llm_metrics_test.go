package rag

import (
	"sync"
	"testing"
	"time"
)

func TestLLMMetrics_RecordSuccess(t *testing.T) {
	m := NewLLMMetrics()

	m.RecordSuccess("openai", 100*time.Millisecond)
	m.RecordSuccess("openai", 200*time.Millisecond)

	metrics := m.GetProviderMetrics("openai")
	if metrics.TotalCalls != 2 {
		t.Errorf("expected 2 total calls, got %d", metrics.TotalCalls)
	}
	if metrics.Failures != 0 {
		t.Errorf("expected 0 failures, got %d", metrics.Failures)
	}
	if metrics.SuccessRate != 1.0 {
		t.Errorf("expected 1.0 success rate, got %f", metrics.SuccessRate)
	}
	if metrics.AvgLatencyMs != 150.0 {
		t.Errorf("expected 150ms avg latency, got %f", metrics.AvgLatencyMs)
	}
	if metrics.LastLatencyMs != 200.0 {
		t.Errorf("expected 200ms last latency, got %f", metrics.LastLatencyMs)
	}
}

func TestLLMMetrics_RecordFailure(t *testing.T) {
	m := NewLLMMetrics()

	m.RecordSuccess("openrouter", 100*time.Millisecond)
	m.RecordFailure("openrouter", 50*time.Millisecond)

	metrics := m.GetProviderMetrics("openrouter")
	if metrics.TotalCalls != 2 {
		t.Errorf("expected 2 total calls, got %d", metrics.TotalCalls)
	}
	if metrics.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", metrics.Failures)
	}
	if metrics.SuccessRate != 0.5 {
		t.Errorf("expected 0.5 success rate, got %f", metrics.SuccessRate)
	}
}

func TestLLMMetrics_LatencyTracking(t *testing.T) {
	m := NewLLMMetrics()

	m.RecordSuccess("openai", 100*time.Millisecond)
	m.RecordSuccess("openai", 200*time.Millisecond)
	m.RecordSuccess("openai", 300*time.Millisecond)

	metrics := m.GetProviderMetrics("openai")
	expectedAvg := (100.0 + 200.0 + 300.0) / 3.0
	if metrics.AvgLatencyMs != expectedAvg {
		t.Errorf("expected %f avg latency, got %f", expectedAvg, metrics.AvgLatencyMs)
	}
	if metrics.LastLatencyMs != 300.0 {
		t.Errorf("expected 300ms last latency, got %f", metrics.LastLatencyMs)
	}
}

func TestLLMMetrics_ConcurrentAccess(t *testing.T) {
	m := NewLLMMetrics()
	var wg sync.WaitGroup

	// Spawn multiple goroutines writing concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				m.RecordSuccess("openai", time.Duration(i)*time.Millisecond)
			} else {
				m.RecordFailure("openai", time.Duration(i)*time.Millisecond)
			}
		}(i)
	}

	// Spawn readers concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.GetProviderMetrics("openai")
		}()
	}

	wg.Wait()

	metrics := m.GetProviderMetrics("openai")
	if metrics.TotalCalls != 100 {
		t.Errorf("expected 100 total calls, got %d", metrics.TotalCalls)
	}
	if metrics.Failures != 50 {
		t.Errorf("expected 50 failures, got %d", metrics.Failures)
	}
}

func TestLLMMetrics_GetAllMetrics(t *testing.T) {
	m := NewLLMMetrics()

	m.RecordSuccess("openai", 100*time.Millisecond)
	m.RecordSuccess("openrouter", 200*time.Millisecond)
	m.RecordFailure("openrouter", 50*time.Millisecond)

	all := m.GetAllMetrics()
	if len(all) != 2 {
		t.Errorf("expected 2 providers, got %d", len(all))
	}

	openai := all["openai"]
	if openai.TotalCalls != 1 || openai.Failures != 0 {
		t.Errorf("unexpected openai metrics: %+v", openai)
	}

	openrouter := all["openrouter"]
	if openrouter.TotalCalls != 2 || openrouter.Failures != 1 {
		t.Errorf("unexpected openrouter metrics: %+v", openrouter)
	}
}

func TestLLMMetrics_Reset(t *testing.T) {
	m := NewLLMMetrics()

	m.RecordSuccess("openai", 100*time.Millisecond)
	m.RecordFailure("openai", 50*time.Millisecond)

	m.Reset()

	metrics := m.GetProviderMetrics("openai")
	if metrics.TotalCalls != 0 {
		t.Errorf("expected 0 total calls after reset, got %d", metrics.TotalCalls)
	}
}

func TestLLMMetrics_EmptyProvider(t *testing.T) {
	m := NewLLMMetrics()

	metrics := m.GetProviderMetrics("nonexistent")
	if metrics.TotalCalls != 0 || metrics.Failures != 0 || metrics.SuccessRate != 0 {
		t.Errorf("expected zero metrics for nonexistent provider, got %+v", metrics)
	}
}
