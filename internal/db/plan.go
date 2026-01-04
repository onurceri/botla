package db

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

func GetPlanByUserID(ctx context.Context, pool *sql.DB, userID string) (*models.Plan, error) {
	var p models.Plan
	var limits models.PlanLimits
	err := pool.QueryRowContext(ctx, `
		SELECT p.id, p.code, p.status, p.billing_cycle, p.price, p.currency, p.trial_days, 
		       p.created_at, p.updated_at,
		       pl.max_chatbots, pl.max_monthly_ingestions, pl.max_monthly_embedding_tokens,
		       pl.min_readd_cooldown_minutes, pl.scraping_dynamic_enabled, pl.scraping_max_urls_per_bot,
		       pl.scraping_max_pages_per_crawl, pl.files_max_size_mb, pl.files_max_files_per_bot,
		       pl.files_max_files_total, pl.files_total_storage_mb, pl.files_max_text_length,
		       pl.chat_default_model, pl.chat_allowed_models, pl.chat_max_monthly_tokens,
		       pl.chat_rag_top_k, pl.chat_rag_max_context_tokens, pl.chat_max_suggested_questions,
		       pl.chat_max_manual_questions, pl.chat_min_response_token_limit, pl.chat_max_response_token_limit,
		       pl.refresh_enabled, pl.refresh_max_monthly, pl.security_secure_embed_enabled,
		       pl.guardrails_can_customize_thresholds, pl.guardrails_can_use_smart_fallback,
		       pl.guardrails_can_use_escalate_fallback, pl.guardrails_can_manage_topics,
		       pl.guardrails_can_customize_messages, pl.branding_can_hide_branding,
		       pl.branding_can_custom_branding, pl.rate_limits_requests_per_minute,
		       pl.rate_limits_window_seconds, pl.rate_limits_chat_rpm, pl.rate_limits_chat_window,
		       pl.rate_limits_sources_rpm, pl.rate_limits_sources_window
		FROM plans p
		JOIN users u ON u.plan_id = p.id
		LEFT JOIN plan_limits pl ON pl.plan_id = p.id
		WHERE u.id = $1 AND u.deleted_at IS NULL AND p.deleted_at IS NULL
	`, userID).Scan(
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
