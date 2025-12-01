package scraper

import (
    "os"
    "sync"
    "time"

    "github.com/redis/go-redis/v9"
    "context"
)

type Cache interface {
    Get(key string) (string, bool)
    Set(key, val string, ttl time.Duration) error
}

type MemoryCache struct {
    mu    sync.RWMutex
    items map[string]memItem
}

type memItem struct {
    v   string
    exp time.Time
}

func NewMemoryCache() *MemoryCache {
    return &MemoryCache{items: make(map[string]memItem)}
}

func (m *MemoryCache) Get(key string) (string, bool) {
    m.mu.RLock()
    it, ok := m.items[key]
    m.mu.RUnlock()
    if !ok {
        return "", false
    }
    if time.Now().After(it.exp) {
        m.mu.Lock()
        delete(m.items, key)
        m.mu.Unlock()
        return "", false
    }
    return it.v, true
}

func (m *MemoryCache) Set(key, val string, ttl time.Duration) error {
    m.mu.Lock()
    m.items[key] = memItem{v: val, exp: time.Now().Add(ttl)}
    m.mu.Unlock()
    return nil
}

type RedisCache struct {
    cli *redis.Client
}

func NewRedisCache(url string) *RedisCache {
    opts, _ := redis.ParseURL(url)
    return &RedisCache{cli: redis.NewClient(opts)}
}

func (r *RedisCache) Get(key string) (string, bool) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    s, err := r.cli.Get(ctx, key).Result()
    if err != nil {
        return "", false
    }
    return s, true
}

func (r *RedisCache) Set(key, val string, ttl time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return r.cli.Set(ctx, key, val, ttl).Err()
}

func NewCache() Cache {
    u := os.Getenv("REDIS_URL")
    if u == "" {
        return NewMemoryCache()
    }
    return NewRedisCache(u)
}

