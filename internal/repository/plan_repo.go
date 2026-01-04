// Package repository provides data access layer implementations for plans.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// Cache configuration constants.
const (
	// cacheKeyPrefix is the prefix for plan cache keys.
	cacheKeyPrefix = "plan:user:"
	// defaultCacheTTL is the default TTL for cached plans.
	defaultCacheTTL = 5 * time.Minute
)

// PostgresPlanRepo implements PlanRepository using PostgreSQL.
// It supports optional Redis caching for GetByUserID operations.
type PostgresPlanRepo struct {
	pool     *sql.DB
	redis    *redis.Client
	cacheTTL time.Duration
}

// Compile-time check that PostgresPlanRepo implements PlanRepository.
var _ PlanRepository = (*PostgresPlanRepo)(nil)

// NewPostgresPlanRepo creates a new PostgresPlanRepo instance.
// The redisClient parameter is optional - if provided, caching will be enabled.
func NewPostgresPlanRepo(pool *sql.DB, redisClient *redis.Client) *PostgresPlanRepo {
	return &PostgresPlanRepo{
		pool:     pool,
		redis:    redisClient,
		cacheTTL: defaultCacheTTL,
	}
}

// GetByUserID retrieves the active plan for a user.
// It first checks the Redis cache, then falls back to the database.
// Returns nil if the user has no plan or is not found.
func (r *PostgresPlanRepo) GetByUserID(ctx context.Context, userID string) (*models.Plan, error) {
	// Try cache first if Redis is available
	if r.redis != nil {
		cached, err := r.getFromCache(ctx, userID)
		// Only return cached plan if it exists (not nil) and no error
		if err == nil && cached != nil {
			return cached, nil
		}
		// Cache miss or error - proceed to database
	}

	// Query database
	plan, err := r.queryPlanByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate cache if Redis is available and plan exists
	if r.redis != nil && plan != nil {
		_ = r.setCache(ctx, userID, plan)
	}

	return plan, nil
}

// GetByCode retrieves a plan by its code (e.g., "free", "pro", "ultra").
// Returns nil, nil if no active plan with that code exists.
func (r *PostgresPlanRepo) GetByCode(ctx context.Context, code string) (*models.Plan, error) {
	query, args, err := psql.
		Select("p.id", "p.code", "p.status", "p.billing_cycle", "p.price", "p.currency", "p.trial_days",
			"p.created_at", "p.updated_at",
			"pl.max_chatbots", "pl.max_monthly_ingestions", "pl.max_monthly_embedding_tokens",
			"pl.min_readd_cooldown_minutes", "pl.scraping_dynamic_enabled", "pl.scraping_max_urls_per_bot",
			"pl.scraping_max_pages_per_crawl", "pl.files_max_size_mb", "pl.files_max_files_per_bot",
			"pl.files_max_files_total", "pl.files_total_storage_mb", "pl.files_max_text_length",
			"pl.chat_default_model", "pl.chat_allowed_models", "pl.chat_max_monthly_tokens",
			"pl.chat_rag_top_k", "pl.chat_rag_max_context_tokens", "pl.chat_max_suggested_questions",
			"pl.chat_max_manual_questions", "pl.chat_min_response_token_limit", "pl.chat_max_response_token_limit",
			"pl.refresh_enabled", "pl.refresh_max_monthly", "pl.security_secure_embed_enabled",
			"pl.guardrails_can_customize_thresholds", "pl.guardrails_can_use_smart_fallback",
			"pl.guardrails_can_use_escalate_fallback", "pl.guardrails_can_manage_topics",
			"pl.guardrails_can_customize_messages", "pl.branding_can_hide_branding",
			"pl.branding_can_custom_branding", "pl.rate_limits_requests_per_minute",
			"pl.rate_limits_window_seconds", "pl.rate_limits_chat_rpm", "pl.rate_limits_chat_window",
			"pl.rate_limits_sources_rpm", "pl.rate_limits_sources_window").
		From("plans p").
		LeftJoin("plan_limits pl ON pl.plan_id = p.id").
		Where(sq.Eq{"p.code": code}).
		Where(sq.Eq{"p.status": "active"}).
		Where(sq.Eq{"p.deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get plan by code query")
	}

	var p models.Plan
	var limits models.PlanLimits
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&p.ID, &p.Code, &p.Status, &p.BillingCycle, &p.Price, &p.Currency, &p.TrialDays,
		&p.CreatedAt, &p.UpdatedAt,
		&limits.MaxChatbots, &limits.MaxMonthlyIngestions, &limits.MaxMonthlyEmbeddingTokens,
		&limits.MinReAddCooldownMinutes, &limits.ScrapingDynamicEnabled, &limits.ScrapingMaxURLsPerBot,
		&limits.ScrapingMaxPagesPerCrawl, &limits.FilesMaxSizeMB, &limits.FilesMaxFilesPerBot,
		&limits.FilesMaxFilesTotal, &limits.FilesTotalStorageMB, &limits.FilesMaxTextLength,
		&limits.ChatDefaultModel, pq.Array(&limits.ChatAllowedModels), &limits.ChatMaxMonthlyTokens,
		&limits.ChatRAGTopK, &limits.ChatRAGMaxContextTokens, &limits.ChatMaxSuggestedQuestions,
		&limits.ChatMaxManualQuestions, &limits.ChatMinResponseTokenLimit, &limits.ChatMaxResponseTokenLimit,
		&limits.RefreshEnabled, &limits.RefreshMaxMonthly, &limits.SecuritySecureEmbedEnabled,
		&limits.GuardrailsCanCustomizeThresholds, &limits.GuardrailsCanUseSmartFallback,
		&limits.GuardrailsCanUseEscalateFallback, &limits.GuardrailsCanManageTopics,
		&limits.GuardrailsCanCustomizeMessages, &limits.BrandingCanHideBranding,
		&limits.BrandingCanCustomBranding, &limits.RateLimitsRequestsPerMinute,
		&limits.RateLimitsWindowSeconds, &limits.RateLimitsChatRPM, &limits.RateLimitsChatWindow,
		&limits.RateLimitsSourcesRPM, &limits.RateLimitsSourcesWindow,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get plan by code")
	}
	p.Limits = &limits
	return &p, nil
}

// GetAll retrieves all active plans ordered by price ascending.
func (r *PostgresPlanRepo) GetAll(ctx context.Context) ([]models.Plan, error) {
	query, args, err := psql.
		Select("p.id", "p.code", "p.status", "p.billing_cycle", "p.price", "p.currency", "p.trial_days",
			"p.created_at", "p.updated_at",
			"pl.max_chatbots", "pl.max_monthly_ingestions", "pl.max_monthly_embedding_tokens",
			"pl.min_readd_cooldown_minutes", "pl.scraping_dynamic_enabled", "pl.scraping_max_urls_per_bot",
			"pl.scraping_max_pages_per_crawl", "pl.files_max_size_mb", "pl.files_max_files_per_bot",
			"pl.files_max_files_total", "pl.files_total_storage_mb", "pl.files_max_text_length",
			"pl.chat_default_model", "pl.chat_allowed_models", "pl.chat_max_monthly_tokens",
			"pl.chat_rag_top_k", "pl.chat_rag_max_context_tokens", "pl.chat_max_suggested_questions",
			"pl.chat_max_manual_questions", "pl.chat_min_response_token_limit", "pl.chat_max_response_token_limit",
			"pl.refresh_enabled", "pl.refresh_max_monthly", "pl.security_secure_embed_enabled",
			"pl.guardrails_can_customize_thresholds", "pl.guardrails_can_use_smart_fallback",
			"pl.guardrails_can_use_escalate_fallback", "pl.guardrails_can_manage_topics",
			"pl.guardrails_can_customize_messages", "pl.branding_can_hide_branding",
			"pl.branding_can_custom_branding", "pl.rate_limits_requests_per_minute",
			"pl.rate_limits_window_seconds", "pl.rate_limits_chat_rpm", "pl.rate_limits_chat_window",
			"pl.rate_limits_sources_rpm", "pl.rate_limits_sources_window").
		From("plans p").
		LeftJoin("plan_limits pl ON pl.plan_id = p.id").
		Where(sq.Eq{"p.status": "active"}).
		Where(sq.Eq{"p.deleted_at": nil}).
		OrderBy("p.price ASC").
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get all plans query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query all plans")
	}
	defer func() { _ = rows.Close() }()

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		var limits models.PlanLimits
		if err := rows.Scan(
			&p.ID, &p.Code, &p.Status, &p.BillingCycle, &p.Price, &p.Currency, &p.TrialDays,
			&p.CreatedAt, &p.UpdatedAt,
			&limits.MaxChatbots, &limits.MaxMonthlyIngestions, &limits.MaxMonthlyEmbeddingTokens,
			&limits.MinReAddCooldownMinutes, &limits.ScrapingDynamicEnabled, &limits.ScrapingMaxURLsPerBot,
			&limits.ScrapingMaxPagesPerCrawl, &limits.FilesMaxSizeMB, &limits.FilesMaxFilesPerBot,
			&limits.FilesMaxFilesTotal, &limits.FilesTotalStorageMB, &limits.FilesMaxTextLength,
			&limits.ChatDefaultModel, pq.Array(&limits.ChatAllowedModels), &limits.ChatMaxMonthlyTokens,
			&limits.ChatRAGTopK, &limits.ChatRAGMaxContextTokens, &limits.ChatMaxSuggestedQuestions,
			&limits.ChatMaxManualQuestions, &limits.ChatMinResponseTokenLimit, &limits.ChatMaxResponseTokenLimit,
			&limits.RefreshEnabled, &limits.RefreshMaxMonthly, &limits.SecuritySecureEmbedEnabled,
			&limits.GuardrailsCanCustomizeThresholds, &limits.GuardrailsCanUseSmartFallback,
			&limits.GuardrailsCanUseEscalateFallback, &limits.GuardrailsCanManageTopics,
			&limits.GuardrailsCanCustomizeMessages, &limits.BrandingCanHideBranding,
			&limits.BrandingCanCustomBranding, &limits.RateLimitsRequestsPerMinute,
			&limits.RateLimitsWindowSeconds, &limits.RateLimitsChatRPM, &limits.RateLimitsChatWindow,
			&limits.RateLimitsSourcesRPM, &limits.RateLimitsSourcesWindow,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan plan")
		}
		p.Limits = &limits
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "plans rows error")
	}
	return plans, nil
}

// GetByID retrieves a plan by its unique identifier.
func (r *PostgresPlanRepo) GetByID(ctx context.Context, id string) (*models.Plan, error) {
	query, args, err := psql.
		Select("p.id", "p.code", "p.status", "p.billing_cycle", "p.price", "p.currency", "p.trial_days",
			"p.created_at", "p.updated_at",
			"pl.max_chatbots", "pl.max_monthly_ingestions", "pl.max_monthly_embedding_tokens",
			"pl.min_readd_cooldown_minutes", "pl.scraping_dynamic_enabled", "pl.scraping_max_urls_per_bot",
			"pl.scraping_max_pages_per_crawl", "pl.files_max_size_mb", "pl.files_max_files_per_bot",
			"pl.files_max_files_total", "pl.files_total_storage_mb", "pl.files_max_text_length",
			"pl.chat_default_model", "pl.chat_allowed_models", "pl.chat_max_monthly_tokens",
			"pl.chat_rag_top_k", "pl.chat_rag_max_context_tokens", "pl.chat_max_suggested_questions",
			"pl.chat_max_manual_questions", "pl.chat_min_response_token_limit", "pl.chat_max_response_token_limit",
			"pl.refresh_enabled", "pl.refresh_max_monthly", "pl.security_secure_embed_enabled",
			"pl.guardrails_can_customize_thresholds", "pl.guardrails_can_use_smart_fallback",
			"pl.guardrails_can_use_escalate_fallback", "pl.guardrails_can_manage_topics",
			"pl.guardrails_can_customize_messages", "pl.branding_can_hide_branding",
			"pl.branding_can_custom_branding", "pl.rate_limits_requests_per_minute",
			"pl.rate_limits_window_seconds", "pl.rate_limits_chat_rpm", "pl.rate_limits_chat_window",
			"pl.rate_limits_sources_rpm", "pl.rate_limits_sources_window").
		From("plans p").
		LeftJoin("plan_limits pl ON pl.plan_id = p.id").
		Where(sq.Eq{"p.id": id}).
		Where(sq.Eq{"p.deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get plan by id query")
	}

	var p models.Plan
	var limits models.PlanLimits
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&p.ID, &p.Code, &p.Status, &p.BillingCycle, &p.Price, &p.Currency, &p.TrialDays,
		&p.CreatedAt, &p.UpdatedAt,
		&limits.MaxChatbots, &limits.MaxMonthlyIngestions, &limits.MaxMonthlyEmbeddingTokens,
		&limits.MinReAddCooldownMinutes, &limits.ScrapingDynamicEnabled, &limits.ScrapingMaxURLsPerBot,
		&limits.ScrapingMaxPagesPerCrawl, &limits.FilesMaxSizeMB, &limits.FilesMaxFilesPerBot,
		&limits.FilesMaxFilesTotal, &limits.FilesTotalStorageMB, &limits.FilesMaxTextLength,
		&limits.ChatDefaultModel, pq.Array(&limits.ChatAllowedModels), &limits.ChatMaxMonthlyTokens,
		&limits.ChatRAGTopK, &limits.ChatRAGMaxContextTokens, &limits.ChatMaxSuggestedQuestions,
		&limits.ChatMaxManualQuestions, &limits.ChatMinResponseTokenLimit, &limits.ChatMaxResponseTokenLimit,
		&limits.RefreshEnabled, &limits.RefreshMaxMonthly, &limits.SecuritySecureEmbedEnabled,
		&limits.GuardrailsCanCustomizeThresholds, &limits.GuardrailsCanUseSmartFallback,
		&limits.GuardrailsCanUseEscalateFallback, &limits.GuardrailsCanManageTopics,
		&limits.GuardrailsCanCustomizeMessages, &limits.BrandingCanHideBranding,
		&limits.BrandingCanCustomBranding, &limits.RateLimitsRequestsPerMinute,
		&limits.RateLimitsWindowSeconds, &limits.RateLimitsChatRPM, &limits.RateLimitsChatWindow,
		&limits.RateLimitsSourcesRPM, &limits.RateLimitsSourcesWindow,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get plan by id")
	}
	p.Limits = &limits
	return &p, nil
}

// GetPlanWithLimits retrieves a plan by user ID with all limits populated.
func (r *PostgresPlanRepo) GetPlanWithLimits(ctx context.Context, userID string) (*models.Plan, error) {
	return r.GetByUserID(ctx, userID)
}

// GetAllPlansWithLimits retrieves all active plans with their limits.
func (r *PostgresPlanRepo) GetAllPlansWithLimits(ctx context.Context) ([]models.Plan, error) {
	return r.GetAll(ctx)
}

// InvalidateCache removes the cached plan for a user.
// Call this when a user's plan changes (upgrade/downgrade).
func (r *PostgresPlanRepo) InvalidateCache(ctx context.Context, userID string) error {
	if r.redis == nil {
		return nil
	}
	key := cacheKeyPrefix + userID
	return r.redis.Del(ctx, key).Err()
}

// queryPlanByUserID performs the database query to get a user's plan.
// This is the core query that the caching layer wraps.
func (r *PostgresPlanRepo) queryPlanByUserID(ctx context.Context, userID string) (*models.Plan, error) {
	query, args, err := psql.
		Select("p.id", "p.code", "p.status", "p.billing_cycle", "p.price", "p.currency", "p.trial_days",
			"p.created_at", "p.updated_at",
			"pl.max_chatbots", "pl.max_monthly_ingestions", "pl.max_monthly_embedding_tokens",
			"pl.min_readd_cooldown_minutes", "pl.scraping_dynamic_enabled", "pl.scraping_max_urls_per_bot",
			"pl.scraping_max_pages_per_crawl", "pl.files_max_size_mb", "pl.files_max_files_per_bot",
			"pl.files_max_files_total", "pl.files_total_storage_mb", "pl.files_max_text_length",
			"pl.chat_default_model", "pl.chat_allowed_models", "pl.chat_max_monthly_tokens",
			"pl.chat_rag_top_k", "pl.chat_rag_max_context_tokens", "pl.chat_max_suggested_questions",
			"pl.chat_max_manual_questions", "pl.chat_min_response_token_limit", "pl.chat_max_response_token_limit",
			"pl.refresh_enabled", "pl.refresh_max_monthly", "pl.security_secure_embed_enabled",
			"pl.guardrails_can_customize_thresholds", "pl.guardrails_can_use_smart_fallback",
			"pl.guardrails_can_use_escalate_fallback", "pl.guardrails_can_manage_topics",
			"pl.guardrails_can_customize_messages", "pl.branding_can_hide_branding",
			"pl.branding_can_custom_branding", "pl.rate_limits_requests_per_minute",
			"pl.rate_limits_window_seconds", "pl.rate_limits_chat_rpm", "pl.rate_limits_chat_window",
			"pl.rate_limits_sources_rpm", "pl.rate_limits_sources_window").
		From("plans p").
		Join("users u ON u.plan_id = p.id").
		LeftJoin("plan_limits pl ON pl.plan_id = p.id").
		Where(sq.Eq{"u.id": userID}).
		Where(sq.Eq{"u.deleted_at": nil}).
		Where(sq.Eq{"p.deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get plan by user id query")
	}

	var p models.Plan
	var limits models.PlanLimits
	err = r.pool.QueryRowContext(ctx, query, args...).Scan(
		&p.ID, &p.Code, &p.Status, &p.BillingCycle, &p.Price, &p.Currency, &p.TrialDays,
		&p.CreatedAt, &p.UpdatedAt,
		&limits.MaxChatbots, &limits.MaxMonthlyIngestions, &limits.MaxMonthlyEmbeddingTokens,
		&limits.MinReAddCooldownMinutes, &limits.ScrapingDynamicEnabled, &limits.ScrapingMaxURLsPerBot,
		&limits.ScrapingMaxPagesPerCrawl, &limits.FilesMaxSizeMB, &limits.FilesMaxFilesPerBot,
		&limits.FilesMaxFilesTotal, &limits.FilesTotalStorageMB, &limits.FilesMaxTextLength,
		&limits.ChatDefaultModel, pq.Array(&limits.ChatAllowedModels), &limits.ChatMaxMonthlyTokens,
		&limits.ChatRAGTopK, &limits.ChatRAGMaxContextTokens, &limits.ChatMaxSuggestedQuestions,
		&limits.ChatMaxManualQuestions, &limits.ChatMinResponseTokenLimit, &limits.ChatMaxResponseTokenLimit,
		&limits.RefreshEnabled, &limits.RefreshMaxMonthly, &limits.SecuritySecureEmbedEnabled,
		&limits.GuardrailsCanCustomizeThresholds, &limits.GuardrailsCanUseSmartFallback,
		&limits.GuardrailsCanUseEscalateFallback, &limits.GuardrailsCanManageTopics,
		&limits.GuardrailsCanCustomizeMessages, &limits.BrandingCanHideBranding,
		&limits.BrandingCanCustomBranding, &limits.RateLimitsRequestsPerMinute,
		&limits.RateLimitsWindowSeconds, &limits.RateLimitsChatRPM, &limits.RateLimitsChatWindow,
		&limits.RateLimitsSourcesRPM, &limits.RateLimitsSourcesWindow,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get plan by user id")
	}
	p.Limits = &limits
	return &p, nil
}

// getFromCache retrieves a plan from Redis cache.
// Returns nil if cache miss or error.
func (r *PostgresPlanRepo) getFromCache(ctx context.Context, userID string) (*models.Plan, error) {
	key := cacheKeyPrefix + userID
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, pkgerrors.Wrapf(err, "get plan from cache")
	}

	var plan models.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, pkgerrors.Wrapf(err, "unmarshal cached plan")
	}

	return &plan, nil
}

// setCache stores a plan in Redis cache.
func (r *PostgresPlanRepo) setCache(ctx context.Context, userID string, plan *models.Plan) error {
	key := cacheKeyPrefix + userID
	data, err := json.Marshal(plan)
	if err != nil {
		return pkgerrors.Wrapf(err, "marshal plan for cache")
	}

	if err := r.redis.Set(ctx, key, data, r.cacheTTL).Err(); err != nil {
		return pkgerrors.Wrapf(err, "set plan cache")
	}

	return nil
}
