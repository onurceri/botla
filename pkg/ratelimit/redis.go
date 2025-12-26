package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements the Limiter interface using Redis sorted sets
// and a sliding window log algorithm for accurate distributed rate limiting
type RedisLimiter struct {
	client *redis.Client
	config Config
}

// NewRedisLimiter creates a new Redis-backed rate limiter
func NewRedisLimiter(client *redis.Client, config Config) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		config: config,
	}
}

// Allow checks if a single request should be allowed for the given key
func (r *RedisLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if N requests should be allowed for the given key
// Uses a sliding window log algorithm with Redis sorted sets
func (r *RedisLimiter) AllowN(ctx context.Context, key string, n int) (*Result, error) {
	now := time.Now()
	windowStart := now.Add(-r.config.WindowSize)

	// Lua script for atomic sliding window check
	// This ensures race-free rate limiting across distributed instances
	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local max_requests = tonumber(ARGV[3])
		local window_size = tonumber(ARGV[4])
		local n = tonumber(ARGV[5])
		
		-- Remove old entries outside the sliding window
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
		
		-- Count current requests in the window
		local current = redis.call('ZCARD', key)
		
		-- Check if we can allow N more requests
		if current + n <= max_requests then
			-- Add N entries with current timestamp
			for i = 1, n do
				-- Use microsecond precision to handle concurrent requests
				local score = now + (i / 1000000)
				redis.call('ZADD', key, score, now .. ':' .. i)
			end
			
			-- Set expiration to window size + small buffer
			redis.call('EXPIRE', key, window_size + 10)
			
			return {1, max_requests - (current + n), window_size}
		else
			-- Get oldest entry to calculate reset time
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local reset = window_size
			if #oldest > 0 then
				local oldest_time = tonumber(oldest[2])
				reset = math.ceil(oldest_time + window_size - now)
				if reset < 0 then reset = 0 end
			end
			
			return {0, 0, reset}
		end
	`)

	// Execute the Lua script
	result, err := script.Run(
		ctx,
		r.client,
		[]string{key},
		now.Unix(),
		windowStart.Unix(),
		r.config.RequestsPerWindow,
		int(r.config.WindowSize.Seconds()),
		n,
	).Result()

	if err != nil {
		return nil, fmt.Errorf("redis rate limit check failed: %w", err)
	}

	// Parse result from Lua script
	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		return nil, fmt.Errorf("unexpected redis script result format")
	}

	allowed := values[0].(int64) == 1
	remaining := int(values[1].(int64))
	resetSeconds := int(values[2].(int64))

	resetAt := now.Add(time.Duration(resetSeconds) * time.Second)
	retryAfter := 0
	if !allowed {
		retryAfter = resetSeconds
	}

	return &Result{
		Allowed:    allowed,
		Limit:      r.config.RequestsPerWindow,
		Remaining:  remaining,
		ResetAt:    resetAt,
		RetryAfter: retryAfter,
	}, nil
}

// Reset clears the rate limit for the given key
func (r *RedisLimiter) Reset(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

// Close closes the Redis client connection
func (r *RedisLimiter) Close() error {
	err := r.client.Close()
	if err != nil {
		return fmt.Errorf("redis close: %w", err)
	}
	return nil
}
