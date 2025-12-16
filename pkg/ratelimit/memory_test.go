package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLimiter_Allow(t *testing.T) {
	config := Config{
		RequestsPerWindow: 3,
		WindowSize:        1 * time.Second,
	}
	limiter := NewMemoryLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-user-123"

	// First 3 requests should be allowed
	for i := 1; i <= 3; i++ {
		result, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i)
		}
		if result.Limit != 3 {
			t.Errorf("expected limit 3, got %d", result.Limit)
		}
		expectedRemaining := 3 - i
		if result.Remaining != expectedRemaining {
			t.Errorf("request %d: expected remaining %d, got %d", i, expectedRemaining, result.Remaining)
		}
	}

	// 4th request should be denied
	result, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Fatal("request should be denied")
	}
	if result.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", result.Remaining)
	}
	if result.RetryAfter <= 0 {
		t.Errorf("expected positive retry after, got %d", result.RetryAfter)
	}
}

func TestMemoryLimiter_SlidingWindow(t *testing.T) {
	config := Config{
		RequestsPerWindow: 2,
		WindowSize:        500 * time.Millisecond,
	}
	limiter := NewMemoryLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-user-456"

	// Use 2 requests
	limiter.Allow(ctx, key)
	limiter.Allow(ctx, key)

	// 3rd request should be denied
	result, _ := limiter.Allow(ctx, key)
	if result.Allowed {
		t.Fatal("3rd request should be denied")
	}

	// Wait for window to pass
	time.Sleep(600 * time.Millisecond)

	// Now requests should be allowed again
	result, _ = limiter.Allow(ctx, key)
	if !result.Allowed {
		t.Fatal("request after window should be allowed")
	}
}

func TestMemoryLimiter_MultipleKeys(t *testing.T) {
	config := Config{
		RequestsPerWindow: 2,
		WindowSize:        1 * time.Second,
	}
	limiter := NewMemoryLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key1 := "user-1"
	key2 := "user-2"

	// User 1 uses 2 requests
	limiter.Allow(ctx, key1)
	limiter.Allow(ctx, key1)

	// User 1's 3rd request should be denied
	result, _ := limiter.Allow(ctx, key1)
	if result.Allowed {
		t.Fatal("user 1 should be rate limited")
	}

	// User 2 should still have quota
	result, _ = limiter.Allow(ctx, key2)
	if !result.Allowed {
		t.Fatal("user 2 should not be rate limited")
	}
}

func TestMemoryLimiter_Reset(t *testing.T) {
	config := Config{
		RequestsPerWindow: 2,
		WindowSize:        1 * time.Second,
	}
	limiter := NewMemoryLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-user"

	// Use up quota
	limiter.Allow(ctx, key)
	limiter.Allow(ctx, key)

	// Should be denied
	result, _ := limiter.Allow(ctx, key)
	if result.Allowed {
		t.Fatal("should be rate limited")
	}

	// Reset the key
	err := limiter.Reset(ctx, key)
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	// Should be allowed again
	result, _ = limiter.Allow(ctx, key)
	if !result.Allowed {
		t.Fatal("should be allowed after reset")
	}
}

func TestMemoryLimiter_Cleanup(t *testing.T) {
	config := Config{
		RequestsPerWindow: 10,
		WindowSize:        100 * time.Millisecond,
	}
	limiter := NewMemoryLimiter(config)
	defer limiter.Close()

	ctx := context.Background()

	// Create some entries
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, "user-"+string(rune(i)))
	}

	// Wait for cleanup to run (cleanup runs at WindowSize interval)
	time.Sleep(250 * time.Millisecond)

	// Limiter should still work
	result, err := limiter.Allow(ctx, "new-user")
	if err != nil {
		t.Fatalf("unexpected error after cleanup: %v", err)
	}
	if !result.Allowed {
		t.Fatal("request should be allowed")
	}
}

func TestMemoryLimiter_AllowN(t *testing.T) {
	config := Config{
		RequestsPerWindow: 10,
		WindowSize:        1 * time.Second,
	}
	limiter := NewMemoryLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-user"

	// Use 5 requests at once
	result, err := limiter.AllowN(ctx, key, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Fatal("5 requests should be allowed")
	}
	if result.Remaining != 5 {
		t.Errorf("expected remaining 5, got %d", result.Remaining)
	}

	// Try to use 6 more (should be denied)
	result, _ = limiter.AllowN(ctx, key, 6)
	if result.Allowed {
		t.Fatal("6 more requests should be denied")
	}

	// Use 5 more (should be allowed)
	result, _ = limiter.AllowN(ctx, key, 5)
	if !result.Allowed {
		t.Fatal("5 more requests should be allowed")
	}
	if result.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", result.Remaining)
	}
}
