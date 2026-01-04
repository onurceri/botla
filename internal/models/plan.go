package models

import (
	"time"
)

type Plan struct {
	ID           string      `json:"id"`
	Code         string      `json:"code"`
	Status       string      `json:"status"`
	BillingCycle string      `json:"billing_cycle"`
	Price        float64     `json:"price"`
	Currency     string      `json:"currency"`
	TrialDays    int         `json:"trial_days"`
	Limits       *PlanLimits `json:"limits,omitempty"` // From plan_limits table
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    *time.Time  `json:"updated_at"`
}

// RateLimitsConfig defines rate limit configuration per plan
type RateLimitsConfig struct {
	RequestsPerMinute int                       `json:"requests_per_minute"`
	WindowSeconds     int                       `json:"window_seconds"`
	Endpoints         map[string]EndpointLimits `json:"endpoints"`
}

// EndpointLimits defines rate limits for specific endpoints
type EndpointLimits struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	WindowSeconds     int `json:"window_seconds"`
}

type SecurityConfig struct {
	SecureEmbedEnabled bool `json:"secure_embed_enabled"`
}

type RefreshConfig struct {
	Enabled    bool `json:"enabled"`
	MaxMonthly int  `json:"max_monthly"`
}

// BrandingConfig defines branding customization options per plan
type BrandingConfig struct {
	CanHideBranding   bool `json:"can_hide_branding"`   // Pro+ plans can hide "Powered by Botla"
	CanCustomBranding bool `json:"can_custom_branding"` // Enterprise can use custom branding
}

type GuardrailsConfig struct {
	CanCustomizeThresholds bool `json:"can_customize_thresholds"`  // Can adjust high/medium thresholds
	CanUseSmartFallback    bool `json:"can_use_smart_fallback"`    // Can use AI-powered fallback
	CanUseEscalateFallback bool `json:"can_use_escalate_fallback"` // Can use escalate to human mode
	CanManageTopics        bool `json:"can_manage_topics"`         // Can use whitelist/blacklist
	CanCustomizeMessages   bool `json:"can_customize_messages"`    // Can edit fallback messages
}

type ScrapingConfig struct {
	DynamicEnabled   bool `json:"dynamic_enabled"`
	MaxURLsPerBot    int  `json:"max_urls_per_bot"`
	MaxPagesPerCrawl int  `json:"max_pages_per_crawl"` // New: Limit sub-pages per URL
}

type FilesConfig struct {
	MaxSizeMB      int `json:"max_size_mb"`
	MaxFilesPerBot int `json:"max_files_per_bot"`
	MaxFilesTotal  int `json:"max_files_total"`
	TotalStorageMB int `json:"total_storage_mb"`
	MaxTextLength  int `json:"max_text_length"`
}

type ChatConfig struct {
	DefaultModel          string    `json:"default_model,omitempty"` // e.g., "openai/gpt-4o-mini" for OpenRouter
	AllowedModels         []string  `json:"allowed_models"`
	MaxMonthlyTokens      int       `json:"max_monthly_tokens"`
	RAG                   RAGConfig `json:"rag"`
	MaxSuggestedQuestions int       `json:"max_suggested_questions"`  // Plan-based limit: Free=3, Pro=6, Ultra=10
	MaxManualQuestions    int       `json:"max_manual_questions"`     // Plan-based limit: Free=3, Pro=6, Ultra=10
	MinResponseTokenLimit int       `json:"min_response_token_limit"` // Min valid value for max_tokens (e.g., 1)
	MaxResponseTokenLimit int       `json:"max_response_token_limit"` // Max valid value for max_tokens (e.g., 4096 or 8192)
}

type RAGConfig struct {
	TopK             int `json:"top_k"`
	MaxContextTokens int `json:"max_context_tokens"`
}
