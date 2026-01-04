package models

import (
	"errors"
	"fmt"
	"time"
)

// PlanLimits represents the normalized plan limits stored in the plan_limits table.
// This replaces the old PlanConfig JSONB column with typed, validated fields.
type PlanLimits struct {
	PlanID string `json:"-" db:"plan_id"`

	// Top-level limits
	MaxChatbots               int `json:"max_chatbots" db:"max_chatbots"`
	MaxMonthlyIngestions      int `json:"max_monthly_ingestions" db:"max_monthly_ingestions"`
	MaxMonthlyEmbeddingTokens int `json:"max_monthly_embedding_tokens" db:"max_monthly_embedding_tokens"`
	MinReAddCooldownMinutes   int `json:"min_readd_cooldown_minutes" db:"min_readd_cooldown_minutes"`

	// Scraping limits
	ScrapingDynamicEnabled   bool `json:"scraping_dynamic_enabled" db:"scraping_dynamic_enabled"`
	ScrapingMaxURLsPerBot    int  `json:"scraping_max_urls_per_bot" db:"scraping_max_urls_per_bot"`
	ScrapingMaxPagesPerCrawl int  `json:"scraping_max_pages_per_crawl" db:"scraping_max_pages_per_crawl"`

	// Files limits
	FilesMaxSizeMB      int `json:"files_max_size_mb" db:"files_max_size_mb"`
	FilesMaxFilesPerBot int `json:"files_max_files_per_bot" db:"files_max_files_per_bot"`
	FilesMaxFilesTotal  int `json:"files_max_files_total" db:"files_max_files_total"`
	FilesTotalStorageMB int `json:"files_total_storage_mb" db:"files_total_storage_mb"`
	FilesMaxTextLength  int `json:"files_max_text_length" db:"files_max_text_length"`

	// Chat limits
	ChatDefaultModel     string   `json:"chat_default_model" db:"chat_default_model"`
	ChatAllowedModels    []string `json:"chat_allowed_models" db:"chat_allowed_models"`
	ChatMaxMonthlyTokens int      `json:"chat_max_monthly_tokens" db:"chat_max_monthly_tokens"`

	ChatRAGTopK               int `json:"chat_rag_top_k" db:"chat_rag_top_k"`
	ChatRAGMaxContextTokens   int `json:"chat_rag_max_context_tokens" db:"chat_rag_max_context_tokens"`
	ChatMaxSuggestedQuestions int `json:"chat_max_suggested_questions" db:"chat_max_suggested_questions"`
	ChatMaxManualQuestions    int `json:"chat_max_manual_questions" db:"chat_max_manual_questions"`
	ChatMinResponseTokenLimit int `json:"chat_min_response_token_limit" db:"chat_min_response_token_limit"`
	ChatMaxResponseTokenLimit int `json:"chat_max_response_token_limit" db:"chat_max_response_token_limit"`

	// Refresh limits
	RefreshEnabled    bool `json:"refresh_enabled" db:"refresh_enabled"`
	RefreshMaxMonthly int  `json:"refresh_max_monthly" db:"refresh_max_monthly"`

	// Security features
	SecuritySecureEmbedEnabled bool `json:"security_secure_embed_enabled" db:"security_secure_embed_enabled"`

	// Guardrails features
	GuardrailsCanCustomizeThresholds bool `json:"guardrails_can_customize_thresholds" db:"guardrails_can_customize_thresholds"`
	GuardrailsCanUseSmartFallback    bool `json:"guardrails_can_use_smart_fallback" db:"guardrails_can_use_smart_fallback"`
	GuardrailsCanUseEscalateFallback bool `json:"guardrails_can_use_escalate_fallback" db:"guardrails_can_use_escalate_fallback"`
	GuardrailsCanManageTopics        bool `json:"guardrails_can_manage_topics" db:"guardrails_can_manage_topics"`
	GuardrailsCanCustomizeMessages   bool `json:"guardrails_can_customize_messages" db:"guardrails_can_customize_messages"`

	// Branding features
	BrandingCanHideBranding   bool `json:"branding_can_hide_branding" db:"branding_can_hide_branding"`
	BrandingCanCustomBranding bool `json:"branding_can_custom_branding" db:"branding_can_custom_branding"`

	// Rate limits
	RateLimitsRequestsPerMinute int `json:"rate_limits_requests_per_minute" db:"rate_limits_requests_per_minute"`
	RateLimitsWindowSeconds     int `json:"rate_limits_window_seconds" db:"rate_limits_window_seconds"`
	RateLimitsChatRPM           int `json:"rate_limits_chat_rpm" db:"rate_limits_chat_rpm"`
	RateLimitsChatWindow        int `json:"rate_limits_chat_window" db:"rate_limits_chat_window"`
	RateLimitsSourcesRPM        int `json:"rate_limits_sources_rpm" db:"rate_limits_sources_rpm"`
	RateLimitsSourcesWindow     int `json:"rate_limits_sources_window" db:"rate_limits_sources_window"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// Validate validates the plan limits configuration.
// Returns an error containing all validation failures joined together.
func (l *PlanLimits) Validate() error {
	var errs []error

	// Top-level limits
	if l.MaxChatbots < 1 {
		errs = append(errs, fmt.Errorf("max_chatbots must be >= 1, got %d", l.MaxChatbots))
	}
	if l.MaxMonthlyIngestions < 0 {
		errs = append(errs, fmt.Errorf("max_monthly_ingestions must be >= 0, got %d", l.MaxMonthlyIngestions))
	}
	if l.MaxMonthlyEmbeddingTokens < 0 {
		errs = append(errs, fmt.Errorf("max_monthly_embedding_tokens must be >= 0, got %d", l.MaxMonthlyEmbeddingTokens))
	}
	if l.MinReAddCooldownMinutes < 0 {
		errs = append(errs, fmt.Errorf("min_readd_cooldown_minutes must be >= 0, got %d", l.MinReAddCooldownMinutes))
	}

	// Scraping limits
	if l.ScrapingMaxURLsPerBot < 0 {
		errs = append(errs, fmt.Errorf("scraping_max_urls_per_bot must be >= 0, got %d", l.ScrapingMaxURLsPerBot))
	}
	if l.ScrapingMaxPagesPerCrawl < 0 {
		errs = append(errs, fmt.Errorf("scraping_max_pages_per_crawl must be >= 0, got %d", l.ScrapingMaxPagesPerCrawl))
	}

	// Files limits
	if l.FilesMaxSizeMB <= 0 {
		errs = append(errs, fmt.Errorf("files_max_size_mb must be > 0, got %d", l.FilesMaxSizeMB))
	}
	if l.FilesMaxFilesPerBot < 0 {
		errs = append(errs, fmt.Errorf("files_max_files_per_bot must be >= 0, got %d", l.FilesMaxFilesPerBot))
	}
	if l.FilesMaxFilesTotal < 0 {
		errs = append(errs, fmt.Errorf("files_max_files_total must be >= 0, got %d", l.FilesMaxFilesTotal))
	}
	if l.FilesTotalStorageMB <= 0 {
		errs = append(errs, fmt.Errorf("files_total_storage_mb must be > 0, got %d", l.FilesTotalStorageMB))
	}
	if l.FilesMaxTextLength < 0 {
		errs = append(errs, fmt.Errorf("files_max_text_length must be >= 0, got %d", l.FilesMaxTextLength))
	}

	// Chat limits
	if l.ChatMaxMonthlyTokens < 0 {
		errs = append(errs, fmt.Errorf("chat_max_monthly_tokens must be >= 0, got %d", l.ChatMaxMonthlyTokens))
	}
	if l.ChatRAGTopK < 1 {
		errs = append(errs, fmt.Errorf("chat_rag_top_k must be >= 1, got %d", l.ChatRAGTopK))
	}
	if l.ChatRAGMaxContextTokens < 1 {
		errs = append(errs, fmt.Errorf("chat_rag_max_context_tokens must be >= 1, got %d", l.ChatRAGMaxContextTokens))
	}
	if l.ChatMaxSuggestedQuestions < 0 {
		errs = append(errs, fmt.Errorf("chat_max_suggested_questions must be >= 0, got %d", l.ChatMaxSuggestedQuestions))
	}
	if l.ChatMaxManualQuestions < 0 {
		errs = append(errs, fmt.Errorf("chat_max_manual_questions must be >= 0, got %d", l.ChatMaxManualQuestions))
	}
	if l.ChatMinResponseTokenLimit < 1 {
		errs = append(errs, fmt.Errorf("chat_min_response_token_limit must be >= 1, got %d", l.ChatMinResponseTokenLimit))
	}
	if l.ChatMaxResponseTokenLimit < l.ChatMinResponseTokenLimit {
		errs = append(errs, fmt.Errorf("chat_max_response_token_limit (%d) must be >= chat_min_response_token_limit (%d)",
			l.ChatMaxResponseTokenLimit, l.ChatMinResponseTokenLimit))
	}

	// Refresh limits
	if l.RefreshMaxMonthly < 0 {
		errs = append(errs, fmt.Errorf("refresh_max_monthly must be >= 0, got %d", l.RefreshMaxMonthly))
	}

	// Rate limits
	if l.RateLimitsRequestsPerMinute < 1 {
		errs = append(errs, fmt.Errorf("rate_limits_requests_per_minute must be >= 1, got %d", l.RateLimitsRequestsPerMinute))
	}
	if l.RateLimitsWindowSeconds < 1 {
		errs = append(errs, fmt.Errorf("rate_limits_window_seconds must be >= 1, got %d", l.RateLimitsWindowSeconds))
	}
	if l.RateLimitsChatRPM < 1 {
		errs = append(errs, fmt.Errorf("rate_limits_chat_rpm must be >= 1, got %d", l.RateLimitsChatRPM))
	}
	if l.RateLimitsChatWindow < 1 {
		errs = append(errs, fmt.Errorf("rate_limits_chat_window must be >= 1, got %d", l.RateLimitsChatWindow))
	}
	if l.RateLimitsSourcesRPM < 1 {
		errs = append(errs, fmt.Errorf("rate_limits_sources_rpm must be >= 1, got %d", l.RateLimitsSourcesRPM))
	}
	if l.RateLimitsSourcesWindow < 1 {
		errs = append(errs, fmt.Errorf("rate_limits_sources_window must be >= 1, got %d", l.RateLimitsSourcesWindow))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// DefaultPlanLimits returns a PlanLimits with sensible defaults (matches Free plan).
func DefaultPlanLimits() PlanLimits {
	return PlanLimits{
		MaxChatbots:                      1,
		MaxMonthlyIngestions:             50,
		MaxMonthlyEmbeddingTokens:        250000,
		MinReAddCooldownMinutes:          60,
		ScrapingDynamicEnabled:           false,
		ScrapingMaxURLsPerBot:            1,
		ScrapingMaxPagesPerCrawl:         5,
		FilesMaxSizeMB:                   5,
		FilesMaxFilesPerBot:              1,
		FilesMaxFilesTotal:               5,
		FilesTotalStorageMB:              10,
		FilesMaxTextLength:               400000,
		ChatDefaultModel:                 "openai/gpt-4o-mini",
		ChatAllowedModels:                []string{"openai/gpt-4o-mini"},
		ChatMaxMonthlyTokens:             100000,
		ChatRAGTopK:                      3,
		ChatRAGMaxContextTokens:          2000,
		ChatMaxSuggestedQuestions:        3,
		ChatMaxManualQuestions:           3,
		ChatMinResponseTokenLimit:        1,
		ChatMaxResponseTokenLimit:        4096,
		RefreshEnabled:                   false,
		RefreshMaxMonthly:                0,
		SecuritySecureEmbedEnabled:       false,
		GuardrailsCanCustomizeThresholds: false,
		GuardrailsCanUseSmartFallback:    true,
		GuardrailsCanUseEscalateFallback: false,
		GuardrailsCanManageTopics:        false,
		GuardrailsCanCustomizeMessages:   false,
		BrandingCanHideBranding:          false,
		BrandingCanCustomBranding:        false,
		RateLimitsRequestsPerMinute:      100,
		RateLimitsWindowSeconds:          60,
		RateLimitsChatRPM:                30,
		RateLimitsChatWindow:             60,
		RateLimitsSourcesRPM:             10,
		RateLimitsSourcesWindow:          60,
	}
}

// ProPlanLimits returns a PlanLimits with Pro plan defaults.
func ProPlanLimits() PlanLimits {
	return PlanLimits{
		MaxChatbots:                      10,
		MaxMonthlyIngestions:             500,
		MaxMonthlyEmbeddingTokens:        2500000,
		MinReAddCooldownMinutes:          30,
		ScrapingDynamicEnabled:           true,
		ScrapingMaxURLsPerBot:            10,
		ScrapingMaxPagesPerCrawl:         50,
		FilesMaxSizeMB:                   20,
		FilesMaxFilesPerBot:              20,
		FilesMaxFilesTotal:               100,
		FilesTotalStorageMB:              500,
		FilesMaxTextLength:               400000,
		ChatDefaultModel:                 "openai/gpt-4o",
		ChatAllowedModels:                []string{"openai/gpt-4o-mini", "openai/gpt-4o"},
		ChatMaxMonthlyTokens:             1000000,
		ChatRAGTopK:                      5,
		ChatRAGMaxContextTokens:          4000,
		ChatMaxSuggestedQuestions:        6,
		ChatMaxManualQuestions:           6,
		ChatMinResponseTokenLimit:        1,
		ChatMaxResponseTokenLimit:        4096,
		RefreshEnabled:                   true,
		RefreshMaxMonthly:                5,
		SecuritySecureEmbedEnabled:       true,
		GuardrailsCanCustomizeThresholds: true,
		GuardrailsCanUseSmartFallback:    true,
		GuardrailsCanUseEscalateFallback: false,
		GuardrailsCanManageTopics:        true,
		GuardrailsCanCustomizeMessages:   true,
		BrandingCanHideBranding:          true,
		BrandingCanCustomBranding:        false,
		RateLimitsRequestsPerMinute:      500,
		RateLimitsWindowSeconds:          60,
		RateLimitsChatRPM:                100,
		RateLimitsChatWindow:             60,
		RateLimitsSourcesRPM:             30,
		RateLimitsSourcesWindow:          60,
	}
}

// UltraPlanLimits returns a PlanLimits with Ultra plan defaults.
func UltraPlanLimits() PlanLimits {
	return PlanLimits{
		MaxChatbots:                      100,
		MaxMonthlyIngestions:             10000,
		MaxMonthlyEmbeddingTokens:        100000000,
		MinReAddCooldownMinutes:          0,
		ScrapingDynamicEnabled:           true,
		ScrapingMaxURLsPerBot:            50,
		ScrapingMaxPagesPerCrawl:         200,
		FilesMaxSizeMB:                   50,
		FilesMaxFilesPerBot:              100,
		FilesMaxFilesTotal:               1000,
		FilesTotalStorageMB:              2000,
		FilesMaxTextLength:               400000,
		ChatDefaultModel:                 "openai/gpt-4o",
		ChatAllowedModels:                []string{"openai/gpt-4o-mini", "openai/gpt-4o", "openai/gpt-5"},
		ChatMaxMonthlyTokens:             5000000,
		ChatRAGTopK:                      10,
		ChatRAGMaxContextTokens:          8000,
		ChatMaxSuggestedQuestions:        10,
		ChatMaxManualQuestions:           10,
		ChatMinResponseTokenLimit:        1,
		ChatMaxResponseTokenLimit:        8192,
		RefreshEnabled:                   true,
		RefreshMaxMonthly:                100,
		SecuritySecureEmbedEnabled:       true,
		GuardrailsCanCustomizeThresholds: true,
		GuardrailsCanUseSmartFallback:    true,
		GuardrailsCanUseEscalateFallback: true,
		GuardrailsCanManageTopics:        true,
		GuardrailsCanCustomizeMessages:   true,
		BrandingCanHideBranding:          true,
		BrandingCanCustomBranding:        true,
		RateLimitsRequestsPerMinute:      2000,
		RateLimitsWindowSeconds:          60,
		RateLimitsChatRPM:                500,
		RateLimitsChatWindow:             60,
		RateLimitsSourcesRPM:             100,
		RateLimitsSourcesWindow:          60,
	}
}
