package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Plan struct {
	ID           string     `json:"id"`
	Code         string     `json:"code"`
	Status       string     `json:"status"`
	BillingCycle string     `json:"billing_cycle"`
	Price        float64    `json:"price"`
	Currency     string     `json:"currency"`
	TrialDays    int        `json:"trial_days"`
	Config       PlanConfig `json:"config"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type PlanConfig struct {
	Scraping                  ScrapingConfig `json:"scraping"`
	Files                     FilesConfig    `json:"files"`
	Chat                      ChatConfig     `json:"chat"`
	Refresh                   RefreshConfig  `json:"refresh"`
	Security                  SecurityConfig `json:"security"`
	Branding                  BrandingConfig `json:"branding"`
	MaxChatbots               int            `json:"max_chatbots"`
	MaxMonthlyIngestions      int            `json:"max_monthly_ingestions"`
	MaxMonthlyEmbeddingTokens int            `json:"max_monthly_embedding_tokens"`
	MinReAddCooldownMinutes   int            `json:"min_readd_cooldown_minutes"`
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

type ScrapingConfig struct {
	DynamicEnabled   bool `json:"dynamic_enabled"`
	MaxURLsPerBot    int  `json:"max_urls_per_bot"`
	MaxPagesPerCrawl int  `json:"max_pages_per_crawl"` // New: Limit sub-pages per URL
}

type FilesConfig struct {
	OCREnabled     bool `json:"ocr_enabled"`
	MaxSizeMB      int  `json:"max_size_mb"`
	MaxFilesPerBot int  `json:"max_files_per_bot"`
	MaxFilesTotal  int  `json:"max_files_total"`
	TotalStorageMB int  `json:"total_storage_mb"`
}

type ChatConfig struct {
	AllowedModels    []string  `json:"allowed_models"`
	MaxMonthlyTokens int       `json:"max_monthly_tokens"`
	RAG              RAGConfig `json:"rag"`
}

type RAGConfig struct {
	TopK             int `json:"top_k"`
	MaxContextTokens int `json:"max_context_tokens"`
}

// Value implements the driver.Valuer interface for PlanConfig
func (p PlanConfig) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for PlanConfig
func (p *PlanConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &p)
}
