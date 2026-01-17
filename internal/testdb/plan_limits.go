package testdb

import (
	"context"
	"database/sql"
	"fmt"
)

// ValidPlanFields defines the allowed columns for plan_limits updates to prevent SQL injection.
// This whitelist ensures that dynamic SQL construction in UpdatePlanLimit is safe.
var ValidPlanFields = map[string]bool{
	"max_chatbots":                       true,
	"max_monthly_ingestions":             true,
	"max_monthly_embedding_tokens":       true,
	"min_readd_cooldown_minutes":         true,
	"scraping_dynamic_enabled":           true,
	"scraping_max_urls_per_bot":          true,
	"scraping_max_pages_per_crawl":       true,
	"files_max_size_mb":                  true,
	"files_max_files_per_bot":            true,
	"files_max_files_total":              true,
	"files_total_storage_mb":             true,
	"files_max_text_length":              true,
	"chat_default_model":                 true,
	"chat_max_monthly_tokens":            true,
	"chat_rag_top_k":                     true,
	"chat_rag_max_context_tokens":        true,
	"chat_max_suggested_questions":       true,
	"chat_max_manual_questions":          true,
	"chat_min_response_token_limit":      true,
	"chat_max_response_token_limit":      true,
	"refresh_enabled":                    true,
	"refresh_max_monthly":                true,
	"security_secure_embed_enabled":      true,
	"guardrails_can_customize_thresholds": true,
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

// UpdatePlanLimit safely updates a single field in the plan_limits table for a given plan code.
// It uses a whitelist to validate the 'field' parameter, preventing SQL injection.
func UpdatePlanLimit(ctx context.Context, db *sql.DB, planCode, field string, value any) error {
	if !ValidPlanFields[field] {
		return fmt.Errorf("invalid plan limit field: %s", field)
	}

	// Build dynamic SET clause using the validated field name
	//nolint:gosec // field is whitelisted in ValidPlanFields
	query := fmt.Sprintf(`
		UPDATE plan_limits
		SET %s = $1, updated_at = NOW()
		WHERE plan_id = (SELECT id FROM plans WHERE code = $2)
	`, field)
	_, err := db.ExecContext(ctx, query, value, planCode)
	if err != nil {
		return fmt.Errorf("update plan limit: %w", err)
	}
	return nil
}
