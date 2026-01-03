package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealRedis_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	t.Run("connection is healthy", func(t *testing.T) {
		err := env.Redis.Ping(context.Background()).Err()
		assert.NoError(t, err)
	})

	t.Run("basic operations", func(t *testing.T) {
		ctx := context.Background()
		key := "test:basic:12345678-1234-1234-1234-1234-12345678"

		err := env.Redis.Set(ctx, key, "value", time.Hour).Err()
		assert.NoError(t, err)

		val, err := env.Redis.Get(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, "value", val)

		err = env.Redis.Del(ctx, key).Err()
		assert.NoError(t, err)

		_, err = env.Redis.Get(ctx, key).Result()
		assert.Error(t, err)
	})
}

func TestRealRedis_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := context.Background()

	t.Run("rate limit enforcement", func(t *testing.T) {
		key := "ratelimit:test:87654321-9876-5432-1098-7654-32109876"
		limit := 5
		window := time.Minute

		count := 0
		for i := 0; i < limit+1; i++ {
			pipe := env.Redis.Pipeline()

			incr := pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, window)

			_, err := pipe.Exec(ctx)
			require.NoError(t, err)

			count = int(incr.Val())

			if i < limit {
				assert.Less(t, count, limit+1)
			} else {
				assert.Equal(t, limit+1, count)
			}
		}

		env.Redis.Del(ctx, key)
	})

	t.Run("concurrent rate limiting", func(t *testing.T) {
		key := "ratelimit:concurrent:98765432-1234-5678-9012-3456-789012345678"
		limit := 10
		window := time.Minute

		done := make(chan int, 100)

		for i := 0; i < 100; i++ {
			go func() {
				ctx := context.Background()
				pipe := env.Redis.Pipeline()

				incr := pipe.Incr(ctx, key)
				pipe.Expire(ctx, key, window)

				pipe.Exec(ctx)
				done <- int(incr.Val())
			}()
		}

		rateLimited := 0
		for i := 0; i < 100; i++ {
			count := <-done
			if count > limit {
				rateLimited++
			}
		}

		assert.Equal(t, 90, rateLimited)
		env.Redis.Del(ctx, key)
	})
}

func TestRealRedis_SessionManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := context.Background()

	t.Run("session creation and retrieval", func(t *testing.T) {
		sessionID := "session:12345678-1234-1234-1234-1234-12345678"
		userData := map[string]interface{}{
			"user_id":    "98765432-1234-5678-9012-3456-789012345678",
			"email":      "test@example.com",
			"created_at": time.Now().Unix(),
		}

		sessionJSON, err := json.Marshal(userData)
		require.NoError(t, err)

		err = env.Redis.Set(ctx, sessionID, sessionJSON, 24*time.Hour).Err()
		assert.NoError(t, err)

		retrieved, err := env.Redis.Get(ctx, sessionID).Result()
		assert.NoError(t, err)
		assert.Equal(t, string(sessionJSON), retrieved)

		env.Redis.Del(ctx, sessionID)
	})

	t.Run("session expiration", func(t *testing.T) {
		sessionID := "session:expiring:87654321-2345-6789-0123-4567-89012345678901"

		err := env.Redis.Set(ctx, sessionID, "data", time.Second).Err()
		assert.NoError(t, err)

		time.Sleep(1100 * time.Millisecond)

		_, err = env.Redis.Get(ctx, sessionID).Result()
		assert.Error(t, err)
	})
}

func TestRealRedis_TTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := context.Background()

	t.Run("TTL expiration", func(t *testing.T) {
		key := "test:ttl:12345678-1234-1234-1234-1234-12345678"

		err := env.Redis.Set(ctx, key, "value", 2*time.Second).Err()
		assert.NoError(t, err)

		ttl, err := env.Redis.TTL(ctx, key).Result()
		assert.NoError(t, err)
		assert.LessOrEqual(t, ttl, time.Duration(2*time.Second))
		assert.Greater(t, ttl, time.Duration(time.Second))

		time.Sleep(2100 * time.Millisecond)

		// When key doesn't exist, TTL returns -2 and Result() returns (-2, nil)
		ttl, err = env.Redis.TTL(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-2), ttl, "TTL should be -2 for expired key")
	})

	t.Run("TTL update", func(t *testing.T) {
		key := "test:ttlupdate:87654321-2345-6789-0123-4567-89012345678901"

		err := env.Redis.Set(ctx, key, "value", 5*time.Minute).Err()
		assert.NoError(t, err)

		err = env.Redis.Expire(ctx, key, 10*time.Minute).Err()
		assert.NoError(t, err)

		ttl, err := env.Redis.TTL(ctx, key).Result()
		assert.NoError(t, err)
		assert.LessOrEqual(t, ttl, time.Duration(10*time.Minute))
		assert.Greater(t, ttl, time.Duration(9*time.Minute))

		env.Redis.Del(ctx, key)
	})
}
