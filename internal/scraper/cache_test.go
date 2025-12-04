package scraper

import (
    "os"
    "testing"
    "time"
)

func TestMemoryCache_SetGetAndTTL(t *testing.T) {
    mc := NewMemoryCache()
    if _, ok := mc.Get("k"); ok {
        t.Fatalf("expected miss on empty cache")
    }
    if err := mc.Set("k", "v", 50*time.Millisecond); err != nil {
        t.Fatalf("set error: %v", err)
    }
    v, ok := mc.Get("k")
    if !ok || v != "v" {
        t.Fatalf("expected hit=\"v\" got ok=%v v=%q", ok, v)
    }
    time.Sleep(60 * time.Millisecond)
    if _, ok := mc.Get("k"); ok {
        t.Fatalf("expected miss after TTL expiry")
    }
}

func TestNewCache_DefaultsToMemory(t *testing.T) {
    old := os.Getenv("REDIS_URL")
    _ = os.Unsetenv("REDIS_URL")
    defer func() { _ = os.Setenv("REDIS_URL", old) }()
    c := NewCache()
    if _, ok := c.(*MemoryCache); !ok {
        t.Fatalf("expected MemoryCache when REDIS_URL is empty")
    }
}
