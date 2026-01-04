package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

// validPlanLimitFields contains all valid field names for UpdatePlanLimitField.
// This prevents SQL injection by validating field names against this allowlist.
var validPlanLimitFields = map[string]bool{
	"max_chatbots":                         true,
	"max_monthly_ingestions":               true,
	"max_monthly_embedding_tokens":         true,
	"min_readd_cooldown_minutes":           true,
	"scraping_dynamic_enabled":             true,
	"scraping_max_urls_per_bot":            true,
	"scraping_max_pages_per_crawl":         true,
	"files_max_size_mb":                    true,
	"files_max_files_per_bot":              true,
	"files_max_files_total":                true,
	"files_total_storage_mb":               true,
	"files_max_text_length":                true,
	"chat_default_model":                   true,
	"chat_allowed_models":                  true,
	"chat_max_monthly_tokens":              true,
	"chat_rag_top_k":                       true,
	"chat_rag_max_context_tokens":          true,
	"chat_max_suggested_questions":         true,
	"chat_max_manual_questions":            true,
	"chat_min_response_token_limit":        true,
	"chat_max_response_token_limit":        true,
	"refresh_enabled":                      true,
	"refresh_max_monthly":                  true,
	"security_secure_embed_enabled":        true,
	"guardrails_can_customize_thresholds":  true,
	"guardrails_can_use_smart_fallback":    true,
	"guardrails_can_use_escalate_fallback": true,
	"guardrails_can_manage_topics":         true,
	"guardrails_can_customize_messages":    true,
	"branding_can_hide_branding":           true,
	"branding_can_custom_branding":         true,
	"rate_limits_requests_per_minute":      true,
	"rate_limits_window_seconds":           true,
	"rate_limits_chat_rpm":                 true,
	"rate_limits_chat_window":              true,
	"rate_limits_sources_rpm":              true,
	"rate_limits_sources_window":           true,
}

// GetPlanLimitsByPlanID fetches plan limits from the plan_limits table by plan ID.
// Returns nil, nil if no plan limits found for the given plan ID.
func GetPlanLimitsByPlanID(ctx context.Context, db *sql.DB, planID string) (*models.PlanLimits, error) {
	var l models.PlanLimits
	err := db.QueryRowContext(ctx, `
		SELECT plan_id, max_chatbots, max_monthly_ingestions, max_monthly_embedding_tokens,
		       min_readd_cooldown_minutes, scraping_dynamic_enabled, scraping_max_urls_per_bot,
		       scraping_max_pages_per_crawl, files_max_size_mb, files_max_files_per_bot,
		       files_max_files_total, files_total_storage_mb, files_max_text_length,
		       chat_default_model, chat_allowed_models, chat_max_monthly_tokens,
		       chat_rag_top_k, chat_rag_max_context_tokens, chat_max_suggested_questions,
		       chat_max_manual_questions, chat_min_response_token_limit, chat_max_response_token_limit,
		       refresh_enabled, refresh_max_monthly, security_secure_embed_enabled,
		       guardrails_can_customize_thresholds, guardrails_can_use_smart_fallback,
		       guardrails_can_use_escalate_fallback, guardrails_can_manage_topics,
		       guardrails_can_customize_messages, branding_can_hide_branding,
		       branding_can_custom_branding, rate_limits_requests_per_minute,
		       rate_limits_window_seconds, rate_limits_chat_rpm, rate_limits_chat_window,
		       rate_limits_sources_rpm, rate_limits_sources_window, created_at, updated_at
		FROM plan_limits
		WHERE plan_id = $1
	`, planID).Scan(
		&l.PlanID, &l.MaxChatbots, &l.MaxMonthlyIngestions, &l.MaxMonthlyEmbeddingTokens,
		&l.MinReAddCooldownMinutes, &l.ScrapingDynamicEnabled, &l.ScrapingMaxURLsPerBot,
		&l.ScrapingMaxPagesPerCrawl, &l.FilesMaxSizeMB, &l.FilesMaxFilesPerBot,
		&l.FilesMaxFilesTotal, &l.FilesTotalStorageMB, &l.FilesMaxTextLength,
		&l.ChatDefaultModel, pq.Array(&l.ChatAllowedModels), &l.ChatMaxMonthlyTokens,
		&l.ChatRAGTopK, &l.ChatRAGMaxContextTokens, &l.ChatMaxSuggestedQuestions,
		&l.ChatMaxManualQuestions, &l.ChatMinResponseTokenLimit, &l.ChatMaxResponseTokenLimit,
		&l.RefreshEnabled, &l.RefreshMaxMonthly, &l.SecuritySecureEmbedEnabled,
		&l.GuardrailsCanCustomizeThresholds, &l.GuardrailsCanUseSmartFallback,
		&l.GuardrailsCanUseEscalateFallback, &l.GuardrailsCanManageTopics,
		&l.GuardrailsCanCustomizeMessages, &l.BrandingCanHideBranding,
		&l.BrandingCanCustomBranding, &l.RateLimitsRequestsPerMinute,
		&l.RateLimitsWindowSeconds, &l.RateLimitsChatRPM, &l.RateLimitsChatWindow,
		&l.RateLimitsSourcesRPM, &l.RateLimitsSourcesWindow, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get plan limits by plan id")
	}
	return &l, nil
}

// GetPlanLimitsByCode fetches plan limits by plan code.
// Returns nil, nil if no plan found with the given code.
func GetPlanLimitsByCode(ctx context.Context, db *sql.DB, code string) (*models.PlanLimits, error) {
	var planID string
	err := db.QueryRowContext(ctx, `SELECT id FROM plans WHERE code = $1 AND deleted_at IS NULL`, code).Scan(&planID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get plan id by code")
	}
	return GetPlanLimitsByPlanID(ctx, db, planID)
}

// UpdatePlanLimitField updates a single field in plan_limits table by plan code.
// The field parameter is validated against an allowlist to prevent SQL injection.
// Returns an error if the field is invalid or no plan is found with the given code.
func UpdatePlanLimitField(ctx context.Context, db *sql.DB, planCode string, field string, value any) error {
	// Validate field name to prevent SQL injection
	if !validPlanLimitFields[field] {
		return fmt.Errorf("invalid field name: %s", field)
	}

	query := fmt.Sprintf(`
		UPDATE plan_limits 
		SET %s = $1, updated_at = NOW()
		WHERE plan_id = (SELECT id FROM plans WHERE code = $2 AND deleted_at IS NULL)
	`, field)
	result, err := db.ExecContext(ctx, query, value, planCode)
	if err != nil {
		return pkgerrors.Wrapf(err, "update plan limit field %s", field)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrapf(err, "get rows affected for plan limit field %s", field)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no plan found with code: %s", planCode)
	}
	return nil
}

// GetPlanWithLimits fetches a plan with its limits joined.
// This is more efficient than fetching plan and limits separately.
func GetPlanWithLimits(ctx context.Context, db *sql.DB, planCode string) (*models.Plan, error) {
	var p models.Plan
	var l models.PlanLimits

	err := db.QueryRowContext(ctx, `
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
		JOIN plan_limits pl ON pl.plan_id = p.id
		WHERE p.code = $1 AND p.deleted_at IS NULL AND p.status = 'active'
	`, planCode).Scan(
		&p.ID, &p.Code, &p.Status, &p.BillingCycle, &p.Price, &p.Currency, &p.TrialDays,
		&p.CreatedAt, &p.UpdatedAt,
		&l.MaxChatbots, &l.MaxMonthlyIngestions, &l.MaxMonthlyEmbeddingTokens,
		&l.MinReAddCooldownMinutes, &l.ScrapingDynamicEnabled, &l.ScrapingMaxURLsPerBot,
		&l.ScrapingMaxPagesPerCrawl, &l.FilesMaxSizeMB, &l.FilesMaxFilesPerBot,
		&l.FilesMaxFilesTotal, &l.FilesTotalStorageMB, &l.FilesMaxTextLength,
		&l.ChatDefaultModel, pq.Array(&l.ChatAllowedModels), &l.ChatMaxMonthlyTokens,
		&l.ChatRAGTopK, &l.ChatRAGMaxContextTokens, &l.ChatMaxSuggestedQuestions,
		&l.ChatMaxManualQuestions, &l.ChatMinResponseTokenLimit, &l.ChatMaxResponseTokenLimit,
		&l.RefreshEnabled, &l.RefreshMaxMonthly, &l.SecuritySecureEmbedEnabled,
		&l.GuardrailsCanCustomizeThresholds, &l.GuardrailsCanUseSmartFallback,
		&l.GuardrailsCanUseEscalateFallback, &l.GuardrailsCanManageTopics,
		&l.GuardrailsCanCustomizeMessages, &l.BrandingCanHideBranding,
		&l.BrandingCanCustomBranding, &l.RateLimitsRequestsPerMinute,
		&l.RateLimitsWindowSeconds, &l.RateLimitsChatRPM, &l.RateLimitsChatWindow,
		&l.RateLimitsSourcesRPM, &l.RateLimitsSourcesWindow,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "get plan with limits by code %s", planCode)
	}

	l.PlanID = p.ID
	p.Limits = &l
	return &p, nil
}

// GetAllPlansWithLimits fetches all active plans with their limits.
func GetAllPlansWithLimits(ctx context.Context, db *sql.DB) ([]models.Plan, error) {
	rows, err := db.QueryContext(ctx, `
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
		JOIN plan_limits pl ON pl.plan_id = p.id
		WHERE p.deleted_at IS NULL AND p.status = 'active'
		ORDER BY p.price ASC
	`)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get all plans with limits")
	}
	defer func() { _ = rows.Close() }()

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		var l models.PlanLimits

		if err := rows.Scan(
			&p.ID, &p.Code, &p.Status, &p.BillingCycle, &p.Price, &p.Currency, &p.TrialDays,
			&p.CreatedAt, &p.UpdatedAt,
			&l.MaxChatbots, &l.MaxMonthlyIngestions, &l.MaxMonthlyEmbeddingTokens,
			&l.MinReAddCooldownMinutes, &l.ScrapingDynamicEnabled, &l.ScrapingMaxURLsPerBot,
			&l.ScrapingMaxPagesPerCrawl, &l.FilesMaxSizeMB, &l.FilesMaxFilesPerBot,
			&l.FilesMaxFilesTotal, &l.FilesTotalStorageMB, &l.FilesMaxTextLength,
			&l.ChatDefaultModel, pq.Array(&l.ChatAllowedModels), &l.ChatMaxMonthlyTokens,
			&l.ChatRAGTopK, &l.ChatRAGMaxContextTokens, &l.ChatMaxSuggestedQuestions,
			&l.ChatMaxManualQuestions, &l.ChatMinResponseTokenLimit, &l.ChatMaxResponseTokenLimit,
			&l.RefreshEnabled, &l.RefreshMaxMonthly, &l.SecuritySecureEmbedEnabled,
			&l.GuardrailsCanCustomizeThresholds, &l.GuardrailsCanUseSmartFallback,
			&l.GuardrailsCanUseEscalateFallback, &l.GuardrailsCanManageTopics,
			&l.GuardrailsCanCustomizeMessages, &l.BrandingCanHideBranding,
			&l.BrandingCanCustomBranding, &l.RateLimitsRequestsPerMinute,
			&l.RateLimitsWindowSeconds, &l.RateLimitsChatRPM, &l.RateLimitsChatWindow,
			&l.RateLimitsSourcesRPM, &l.RateLimitsSourcesWindow,
		); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan plan row")
		}

		l.PlanID = p.ID
		p.Limits = &l
		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterate plan rows")
	}

	return plans, nil
}
