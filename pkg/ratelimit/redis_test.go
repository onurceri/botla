package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestRedisLimiter spins up an in-memory Redis (miniredis) and a RedisLimiter
// against it. The limiter's clock is left as time.Now; tests that need a fixed
// time override limiter.now.
func newTestRedisLimiter(t *testing.T, cfg Config) (*RedisLimiter, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewRedisLimiter(client, cfg), mr
}

// TestRedisLimiter_SameSecondBurst is the regression test for the member-collision
// bug: with second-precision timestamps and a per-call loop index, concurrent
// requests in the same second used to produce identical sorted-set members
// (e.g. "<epoch>:1"), so ZADD overwrote them and the window under-counted,
// letting more than RequestsPerWindow through. The per-call nonce fixes this.
// The clock is frozen so every request shares the same "now" (same second).
func TestRedisLimiter_SameSecondBurst(t *testing.T) {
	const limit = 5
	limiter, _ := newTestRedisLimiter(t, Config{RequestsPerWindow: limit, WindowSize: 60 * time.Second})
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return fixed }

	ctx := context.Background()
	allowed := 0
	for i := 0; i < limit+1; i++ {
		res, err := limiter.Allow(ctx, "burst-key")
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if res.Allowed {
			allowed++
		}
	}
	if allowed != limit {
		t.Fatalf("expected exactly %d allowed within one second, got %d (member-collision regression)", limit, allowed)
	}
}

func TestRedisLimiter_BasicAllowDeny(t *testing.T) {
	limiter, _ := newTestRedisLimiter(t, Config{RequestsPerWindow: 3, WindowSize: 60 * time.Second})
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return fixed }
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		res, err := limiter.Allow(ctx, "basic-key")
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !res.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
		if want := 3 - (i + 1); res.Remaining != want {
			t.Fatalf("remaining = %d, want %d", res.Remaining, want)
		}
	}

	res, err := limiter.Allow(ctx, "basic-key")
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if res.Allowed {
		t.Fatal("4th request should be denied")
	}
	if res.RetryAfter <= 0 {
		t.Fatalf("RetryAfter should be > 0 when denied, got %d", res.RetryAfter)
	}
}

func TestRedisLimiter_WindowExpiry(t *testing.T) {
	limiter, _ := newTestRedisLimiter(t, Config{RequestsPerWindow: 2, WindowSize: 60 * time.Second})
	current := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return current }
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		res, err := limiter.Allow(ctx, "expiry-key")
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !res.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if res, _ := limiter.Allow(ctx, "expiry-key"); res.Allowed {
		t.Fatal("3rd request should be denied within the window")
	}

	// Advance past the window so the old entries fall out of the sliding window.
	current = current.Add(61 * time.Second)
	res, err := limiter.Allow(ctx, "expiry-key")
	if err != nil {
		t.Fatalf("Allow after expiry: %v", err)
	}
	if !res.Allowed {
		t.Fatal("request after window expiry should be allowed")
	}
}

func TestRedisLimiter_Reset(t *testing.T) {
	limiter, _ := newTestRedisLimiter(t, Config{RequestsPerWindow: 1, WindowSize: 60 * time.Second})
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return fixed }
	ctx := context.Background()

	if res, _ := limiter.Allow(ctx, "reset-key"); !res.Allowed {
		t.Fatal("first request should be allowed")
	}
	if res, _ := limiter.Allow(ctx, "reset-key"); res.Allowed {
		t.Fatal("second request should be denied")
	}

	if err := limiter.Reset(ctx, "reset-key"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if res, _ := limiter.Allow(ctx, "reset-key"); !res.Allowed {
		t.Fatal("request after Reset should be allowed")
	}
}
