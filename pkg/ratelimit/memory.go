package ratelimit

import (
	"context"
	"sync"
	"time"
)

// memoryEntry represents a single request timestamp in the sliding window
type memoryEntry struct {
	timestamp time.Time
}

// memoryState holds the sliding window state for a key
type memoryState struct {
	mu      sync.Mutex
	entries []memoryEntry
}

// MemoryLimiter implements the Limiter interface using in-memory storage
// This is a fallback for when Redis is unavailable (development/testing)
// WARNING: Not suitable for distributed deployments
type MemoryLimiter struct {
	mu     sync.RWMutex
	states map[string]*memoryState
	config Config

	// Cleanup ticker
	stopCleanup chan struct{}
}

// NewMemoryLimiter creates a new in-memory rate limiter
func NewMemoryLimiter(config Config) *MemoryLimiter {
	limiter := &MemoryLimiter{
		states:      make(map[string]*memoryState),
		config:      config,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go limiter.cleanupLoop()

	return limiter
}

// Allow checks if a single request should be allowed for the given key
func (m *MemoryLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	return m.AllowN(ctx, key, 1)
}

// AllowN checks if N requests should be allowed for the given key
// Uses a sliding window log algorithm similar to Redis implementation
func (m *MemoryLimiter) AllowN(ctx context.Context, key string, n int) (*Result, error) {
	now := time.Now()
	windowStart := now.Add(-m.config.WindowSize)

	// Get or create state for this key
	m.mu.Lock()
	state, exists := m.states[key]
	if !exists {
		state = &memoryState{
			entries: make([]memoryEntry, 0, m.config.RequestsPerWindow),
		}
		m.states[key] = state
	}
	m.mu.Unlock()

	// Lock the state for this specific key
	state.mu.Lock()
	defer state.mu.Unlock()

	// Remove entries outside the sliding window
	validEntries := make([]memoryEntry, 0, len(state.entries))
	for _, entry := range state.entries {
		if entry.timestamp.After(windowStart) {
			validEntries = append(validEntries, entry)
		}
	}
	state.entries = validEntries

	current := len(state.entries)

	// Check if we can allow N more requests
	if current+n <= m.config.RequestsPerWindow {
		// Add N new entries
		for i := 0; i < n; i++ {
			state.entries = append(state.entries, memoryEntry{
				timestamp: now.Add(time.Duration(i) * time.Microsecond),
			})
		}

		remaining := m.config.RequestsPerWindow - (current + n)
		resetAt := now.Add(m.config.WindowSize)

		return &Result{
			Allowed:    true,
			Limit:      m.config.RequestsPerWindow,
			Remaining:  remaining,
			ResetAt:    resetAt,
			RetryAfter: 0,
		}, nil
	}

	// Request denied - calculate reset time
	resetSeconds := int(m.config.WindowSize.Seconds())
	if len(state.entries) > 0 {
		// Calculate when the oldest entry will expire
		oldestTime := state.entries[0].timestamp
		resetTime := oldestTime.Add(m.config.WindowSize)
		resetSeconds = int(time.Until(resetTime).Seconds())
		if resetSeconds <= 0 {
			resetSeconds = 1 // Ensure at least 1 second
		}
	}

	resetAt := now.Add(time.Duration(resetSeconds) * time.Second)

	return &Result{
		Allowed:    false,
		Limit:      m.config.RequestsPerWindow,
		Remaining:  0,
		ResetAt:    resetAt,
		RetryAfter: resetSeconds,
	}, nil
}

// Reset clears the rate limit for the given key
func (m *MemoryLimiter) Reset(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, key)
	return nil
}

// Close stops the cleanup goroutine and releases resources
func (m *MemoryLimiter) Close() error {
	close(m.stopCleanup)
	return nil
}

// cleanupLoop periodically removes expired keys to prevent memory leaks
func (m *MemoryLimiter) cleanupLoop() {
	ticker := time.NewTicker(m.config.WindowSize)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanup removes keys with no entries in the current window
func (m *MemoryLimiter) cleanup() {
	now := time.Now()
	windowStart := now.Add(-m.config.WindowSize * 2) // Use 2x window for safety

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, state := range m.states {
		state.mu.Lock()
		hasRecent := false
		for _, entry := range state.entries {
			if entry.timestamp.After(windowStart) {
				hasRecent = true
				break
			}
		}
		state.mu.Unlock()

		if !hasRecent {
			delete(m.states, key)
		}
	}
}
