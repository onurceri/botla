package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_AreDefined(t *testing.T) {
	// Test that all sentinel errors are defined and are not nil
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrRateLimit", ErrRateLimit},
		{"ErrTimeout", ErrTimeout},
		{"ErrNetwork", ErrNetwork},
		{"ErrNotFound", ErrNotFound},
		{"ErrContextCancelled", ErrContextCancelled},
	}

	for _, s := range sentinels {
		t.Run(s.name+"_is_not_nil", func(t *testing.T) {
			if s.err == nil {
				t.Errorf("%s should not be nil", s.name)
			}
		})
	}
}

func TestSentinelErrors_HaveDistinctMessages(t *testing.T) {
	// Ensure no two sentinel errors have the same message
	sentinels := map[string]error{
		"ErrRateLimit":        ErrRateLimit,
		"ErrTimeout":          ErrTimeout,
		"ErrNetwork":          ErrNetwork,
		"ErrNotFound":         ErrNotFound,
		"ErrContextCancelled": ErrContextCancelled,
	}

	messages := make(map[string]string)
	for name, err := range sentinels {
		msg := err.Error()
		if existingName, exists := messages[msg]; exists {
			t.Errorf("duplicate error message %q used by %s and %s", msg, existingName, name)
		}
		messages[msg] = name
	}
}

func TestSentinelErrors_AreComparable(t *testing.T) {
	// Verify errors.Is works correctly for each sentinel
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrRateLimit", ErrRateLimit},
		{"ErrTimeout", ErrTimeout},
		{"ErrNetwork", ErrNetwork},
		{"ErrNotFound", ErrNotFound},
		{"ErrContextCancelled", ErrContextCancelled},
	}

	for _, s := range sentinels {
		t.Run(s.name+"_is_comparable", func(t *testing.T) {
			if !errors.Is(s.err, s.err) {
				t.Errorf("errors.Is(%s, %s) should return true", s.name, s.name)
			}
		})
	}
}

func TestSentinelErrors_CanBeWrapped(t *testing.T) {
	// Verify wrapped errors can be unwrapped using errors.Is
	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrRateLimit", ErrRateLimit},
		{"ErrTimeout", ErrTimeout},
		{"ErrNetwork", ErrNetwork},
		{"ErrNotFound", ErrNotFound},
		{"ErrContextCancelled", ErrContextCancelled},
	}

	for _, s := range sentinels {
		t.Run(s.name+"_wrapped", func(t *testing.T) {
			wrapped := fmt.Errorf("outer context: %w", s.err)
			if !errors.Is(wrapped, s.err) {
				t.Errorf("errors.Is(wrapped, %s) should return true", s.name)
			}
		})

		t.Run(s.name+"_double_wrapped", func(t *testing.T) {
			inner := fmt.Errorf("inner context: %w", s.err)
			outer := fmt.Errorf("outer context: %w", inner)
			if !errors.Is(outer, s.err) {
				t.Errorf("errors.Is(double_wrapped, %s) should return true", s.name)
			}
		})
	}
}

func TestSentinelErrors_AreNotConfused(t *testing.T) {
	// Ensure each sentinel is distinct from the others
	allSentinels := []error{
		ErrRateLimit,
		ErrTimeout,
		ErrNetwork,
		ErrNotFound,
		ErrContextCancelled,
	}

	for i, a := range allSentinels {
		for j, b := range allSentinels {
			if i != j && errors.Is(a, b) {
				t.Errorf("sentinel at index %d should not match sentinel at index %d", i, j)
			}
		}
	}
}
