package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	planCacheKeyPrefix = "plan:"
	allPlansCacheKey   = "plans:all"
	planCacheTTL       = 24 * time.Hour
)

// PlanService provides cached access to plan configurations.
// It uses Redis for distributed caching when available, with in-memory fallback.
type PlanService struct {
	db          *sql.DB
	redis       *redis.Client
	memoryCache sync.Map
}

// NewPlanService creates a new PlanService instance.
// redis can be nil for in-memory only caching (tests, development).
func NewPlanService(db *sql.DB, redis *redis.Client) *PlanService {
	return &PlanService{
		db:    db,
		redis: redis,
	}
}

// memoryCacheEntry holds cached data with expiration
type memoryCacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// GetPlanByCode returns plan config for a given plan code.
// Results are cached for performance.
func (s *PlanService) GetPlanByCode(ctx context.Context, code string) (*models.Plan, error) {
	if code == "" {
		return nil, fmt.Errorf("plan code is required")
	}

	cacheKey := planCacheKeyPrefix + code

	// Try cache first
	if cached, err := s.getFromCache(ctx, cacheKey); err == nil && cached != nil {
		var plan models.Plan
		if err := json.Unmarshal(cached, &plan); err == nil {
			return &plan, nil
		}
	}

	// Fetch from database
	plan, err := s.fetchPlanByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, nil
	}

	// Store in cache
	if data, err := json.Marshal(plan); err == nil {
		s.setInCache(ctx, cacheKey, data)
	}

	return plan, nil
}

// GetAllPlans returns all active plans.
// Results are cached for performance.
func (s *PlanService) GetAllPlans(ctx context.Context) ([]models.Plan, error) {
	// Try cache first
	if cached, err := s.getFromCache(ctx, allPlansCacheKey); err == nil && cached != nil {
		var plans []models.Plan
		if err := json.Unmarshal(cached, &plans); err == nil {
			return plans, nil
		}
	}

	// Fetch from database
	plans, err := s.fetchAllPlans(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(plans); err == nil {
		s.setInCache(ctx, allPlansCacheKey, data)
	}

	return plans, nil
}

// InvalidateCache clears plan cache.
// Should be called when plan configurations are updated.
func (s *PlanService) InvalidateCache(ctx context.Context, planCode string) error {
	if planCode != "" {
		cacheKey := planCacheKeyPrefix + planCode
		if s.redis != nil {
			s.redis.Del(ctx, cacheKey)
		}
		s.memoryCache.Delete(cacheKey)
	}

	// Always invalidate the all-plans cache
	if s.redis != nil {
		s.redis.Del(ctx, allPlansCacheKey)
	}
	s.memoryCache.Delete(allPlansCacheKey)

	return nil
}

// InvalidateAllCache clears all plan caches.
func (s *PlanService) InvalidateAllCache(ctx context.Context) error {
	if s.redis != nil {
		// Delete all plan keys using pattern matching
		iter := s.redis.Scan(ctx, 0, planCacheKeyPrefix+"*", 100).Iterator()
		for iter.Next(ctx) {
			s.redis.Del(ctx, iter.Val())
		}
		s.redis.Del(ctx, allPlansCacheKey)
	}

	// Clear in-memory cache
	s.memoryCache.Range(func(key, _ interface{}) bool {
		s.memoryCache.Delete(key)
		return true
	})

	return nil
}

// fetchPlanByCode retrieves a plan from the database by code using the plan_limits table.
func (s *PlanService) fetchPlanByCode(ctx context.Context, code string) (*models.Plan, error) {
	return db.GetPlanWithLimits(ctx, s.db, code)
}

// fetchAllPlans retrieves all active plans from the database using the plan_limits table.
func (s *PlanService) fetchAllPlans(ctx context.Context) ([]models.Plan, error) {
	return db.GetAllPlansWithLimits(ctx, s.db)
}

// getFromCache retrieves data from cache (Redis first, then memory).
func (s *PlanService) getFromCache(ctx context.Context, key string) ([]byte, error) {
	// Try Redis first
	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Bytes()
		if err == nil {
			return val, nil
		}
		if err != redis.Nil {
			// Log error but continue to memory cache
			fmt.Printf("redis get error for key %s: %v\n", key, err)
		}
	}

	// Try memory cache
	if entry, ok := s.memoryCache.Load(key); ok {
		if e, ok := entry.(memoryCacheEntry); ok {
			if time.Now().Before(e.expiresAt) {
				return e.data, nil
			}
			// Expired, remove from cache
			s.memoryCache.Delete(key)
		}
	}

	return nil, nil
}

// setInCache stores data in cache (both Redis and memory).
func (s *PlanService) setInCache(ctx context.Context, key string, data []byte) {
	// Store in Redis
	if s.redis != nil {
		if err := s.redis.Set(ctx, key, data, planCacheTTL).Err(); err != nil {
			// Log error but continue to memory cache
			fmt.Printf("redis set error for key %s: %v\n", key, err)
		}
	}

	// Store in memory cache
	s.memoryCache.Store(key, memoryCacheEntry{
		data:      data,
		expiresAt: time.Now().Add(planCacheTTL),
	})
}

// ValidateAllPlans fetches all active plans from the database and validates each plan's limits.
// Returns a combined error if any plan fails validation.
// Returns nil if all plans are valid or if no plans exist.
func (s *PlanService) ValidateAllPlans(ctx context.Context) error {
	plans, err := s.fetchAllPlans(ctx)
	if err != nil {
		return pkgerrors.Wrapf(err, "fetching all plans for validation")
	}

	if len(plans) == 0 {
		return nil
	}

	var validationErrors []error
	for _, plan := range plans {
		if plan.Limits == nil {
			validationErrors = append(validationErrors, fmt.Errorf("plan %q: missing limits", plan.Code))
			continue
		}
		if err := plan.Limits.Validate(); err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("plan %q: %w", plan.Code, err))
		}
	}

	if len(validationErrors) > 0 {
		return errors.Join(validationErrors...)
	}
	return nil
}
