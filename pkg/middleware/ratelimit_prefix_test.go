package middleware

import (
	"os"
	"testing"
)

func TestNewRateLimiterFromEnvWithPrefix(t *testing.T) {
	os.Setenv("SOURCES_RATE_LIMIT_REQUESTS", "7")
	os.Setenv("SOURCES_RATE_LIMIT_WINDOW_SECONDS", "10")
	rl := NewRateLimiterFromEnvWithPrefix("SOURCES")
	if rl == nil {
		t.Fatal("nil rl")
	}
	// allow 7 within window
	for i := 0; i < 7; i++ {
		ok, _, _ := rl.allow("k")
		if !ok {
			t.Fatalf("unexpected block at %d", i)
		}
	}
	ok, _, _ := rl.allow("k")
	if ok {
		t.Fatalf("expected block after limit")
	}
}
