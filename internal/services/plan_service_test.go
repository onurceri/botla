package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPlanService(t *testing.T) {
	svc := NewPlanService(nil, nil)
	assert.NotNil(t, svc, "NewPlanService should return non-nil service")
}

func TestPlanService_GetPlanByCode_EmptyCode(t *testing.T) {
	svc := NewPlanService(nil, nil)

	plan, err := svc.GetPlanByCode(context.Background(), "")
	assert.Error(t, err, "expected error for empty code")
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "required")
}

func TestPlanService_MemoryCache_StoresAndRetrieves(t *testing.T) {
	svc := NewPlanService(nil, nil)
	ctx := context.Background()

	testKey := "test:key"
	testData := []byte(`{"test": "data"}`)

	// Store in cache
	svc.setInCache(ctx, testKey, testData)

	// Retrieve from cache
	retrieved, err := svc.getFromCache(ctx, testKey)
	assert.NoError(t, err)
	assert.Equal(t, testData, retrieved)
}

func TestPlanService_MemoryCache_ExpiresAfterTTL(t *testing.T) {
	svc := NewPlanService(nil, nil)
	ctx := context.Background()

	testKey := "test:expiring"
	testData := []byte(`{"test": "data"}`)

	// Manually set with immediate expiration
	svc.memoryCache.Store(testKey, memoryCacheEntry{
		data:      testData,
		expiresAt: time.Now().Add(-1 * time.Second), // Already expired
	})

	// Should return nil for expired entry
	retrieved, err := svc.getFromCache(ctx, testKey)
	assert.NoError(t, err)
	assert.Nil(t, retrieved, "expired cache entry should return nil")

	// Entry should be deleted
	_, exists := svc.memoryCache.Load(testKey)
	assert.False(t, exists, "expired entry should be deleted")
}

func TestPlanService_InvalidateCache_ClearsMemory(t *testing.T) {
	svc := NewPlanService(nil, nil)
	ctx := context.Background()

	// Store some data
	svc.setInCache(ctx, "plan:free", []byte(`{"code": "free"}`))
	svc.setInCache(ctx, "plans:all", []byte(`[{"code": "free"}]`))

	// Verify data exists
	data, _ := svc.getFromCache(ctx, "plan:free")
	assert.NotNil(t, data)

	// Invalidate specific plan
	err := svc.InvalidateCache(ctx, "free")
	assert.NoError(t, err)

	// Plan cache should be cleared
	data, _ = svc.getFromCache(ctx, "plan:free")
	assert.Nil(t, data)

	// All plans cache should also be cleared
	data, _ = svc.getFromCache(ctx, "plans:all")
	assert.Nil(t, data)
}

func TestPlanService_InvalidateAllCache_ClearsEverything(t *testing.T) {
	svc := NewPlanService(nil, nil)
	ctx := context.Background()

	// Store multiple entries
	svc.setInCache(ctx, "plan:free", []byte(`{"code": "free"}`))
	svc.setInCache(ctx, "plan:pro", []byte(`{"code": "pro"}`))
	svc.setInCache(ctx, "plans:all", []byte(`[]`))

	// Invalidate all
	err := svc.InvalidateAllCache(ctx)
	assert.NoError(t, err)

	// All caches should be cleared
	var count int
	svc.memoryCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count, "all cache entries should be cleared")
}

func TestPlanService_ConcurrentAccess(t *testing.T) {
	svc := NewPlanService(nil, nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			svc.setInCache(ctx, "plan:test", []byte(`{"concurrent": true}`))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			svc.getFromCache(ctx, "plan:test")
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions occurred
}

func TestPlanService_CacheKeyFormat(t *testing.T) {
	assert.Equal(t, "plan:", planCacheKeyPrefix)
	assert.Equal(t, "plans:all", allPlansCacheKey)
	assert.Equal(t, 15*time.Minute, planCacheTTL)
}
