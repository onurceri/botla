package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/policy"
)

// PlanResponse represents the /me/plan endpoint response
type PlanResponse struct {
	ID              string               `json:"id"`
	Code            string               `json:"code"`
	Name            *string              `json:"name,omitempty"`
	Description     *string              `json:"description,omitempty"`
	Price           float64              `json:"price"`
	Currency        string               `json:"currency"`
	Limits          PlanLimitsResponse   `json:"limits"`
	Features        PlanFeaturesResponse `json:"features"`
	AvailableModels []models.ModelInfo   `json:"available_models"`
}

// PlanLimitsResponse holds top-level plan limits
type PlanLimitsResponse struct {
	MaxChatbots               int `json:"max_chatbots"`
	MaxMonthlyIngestions      int `json:"max_monthly_ingestions"`
	MaxMonthlyEmbeddingTokens int `json:"max_monthly_embedding_tokens"`
	MinReAddCooldownMinutes   int `json:"min_readd_cooldown_minutes"`
}

// PlanFeaturesResponse holds feature-specific configurations
type PlanFeaturesResponse struct {
	Scraping   models.ScrapingConfig   `json:"scraping"`
	Files      models.FilesConfig      `json:"files"`
	Chat       models.ChatConfig       `json:"chat"`
	Refresh    models.RefreshConfig    `json:"refresh"`
	Security   models.SecurityConfig   `json:"security"`
	Guardrails models.GuardrailsConfig `json:"guardrails"`
	Branding   models.BrandingConfig   `json:"branding"`
	RateLimits models.RateLimitsConfig `json:"rate_limits"`
}

// planInfo holds plan-related data
type planInfo struct {
	ID          string
	Code        string
	Name        *string
	Description *string
	Price       float64
	Currency    string
	Limits      *models.PlanLimits
}

// PlanHandlers handles plan-related endpoints
type PlanHandlers struct {
	UserRepo repository.UserRepository
	PlanRepo repository.PlanRepository
	DB       *sql.DB // Kept for ModelService which still needs it
}

// NewPlanHandlers creates a new PlanHandlers instance
func NewPlanHandlers(userRepo repository.UserRepository, planRepo repository.PlanRepository, db *sql.DB) *PlanHandlers {
	return &PlanHandlers{
		UserRepo: userRepo,
		PlanRepo: planRepo,
		DB:       db,
	}
}

// GetPlan handles GET /me/plan endpoint
func (h *PlanHandlers) GetPlan(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u, err := h.UserRepo.GetByID(r.Context(), uid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if u == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	plan, err := h.getPlanInfo(r.Context(), u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if plan == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Initialize model service
	modelService := services.NewModelService(h.DB)
	availableModels, err := modelService.GetAvailableModels(r.Context(), plan.Limits.ChatAllowedModels)
	if err != nil {
		// Log error but continue with empty/default models to avoid blocking plan info
		// In production, you might want to log this properly
		availableModels = []models.ModelInfo{}
	}

	// Convert PlanLimits to PlanFeaturesResponse
	features := PlanFeaturesResponse{
		Scraping: models.ScrapingConfig{
			DynamicEnabled:   plan.Limits.ScrapingDynamicEnabled,
			MaxURLsPerBot:    plan.Limits.ScrapingMaxURLsPerBot,
			MaxPagesPerCrawl: plan.Limits.ScrapingMaxPagesPerCrawl,
		},
		Files: models.FilesConfig{
			MaxSizeMB:      plan.Limits.FilesMaxSizeMB,
			MaxFilesPerBot: plan.Limits.FilesMaxFilesPerBot,
			MaxFilesTotal:  plan.Limits.FilesMaxFilesTotal,
			TotalStorageMB: plan.Limits.FilesTotalStorageMB,
			MaxTextLength:  plan.Limits.FilesMaxTextLength,
		},
		Chat: models.ChatConfig{
			DefaultModel:     plan.Limits.ChatDefaultModel,
			AllowedModels:    plan.Limits.ChatAllowedModels,
			MaxMonthlyTokens: plan.Limits.ChatMaxMonthlyTokens,
			RAG: models.RAGConfig{
				TopK:             plan.Limits.ChatRAGTopK,
				MaxContextTokens: plan.Limits.ChatRAGMaxContextTokens,
			},
			MaxSuggestedQuestions: plan.Limits.ChatMaxSuggestedQuestions,
			MaxManualQuestions:    plan.Limits.ChatMaxManualQuestions,
			MinResponseTokenLimit: plan.Limits.ChatMinResponseTokenLimit,
			MaxResponseTokenLimit: plan.Limits.ChatMaxResponseTokenLimit,
		},
		Refresh: models.RefreshConfig{
			Enabled:    plan.Limits.RefreshEnabled,
			MaxMonthly: plan.Limits.RefreshMaxMonthly,
		},
		Security: models.SecurityConfig{
			SecureEmbedEnabled: plan.Limits.SecuritySecureEmbedEnabled,
		},
		Guardrails: models.GuardrailsConfig{
			CanCustomizeThresholds: plan.Limits.GuardrailsCanCustomizeThresholds,
			CanUseSmartFallback:    plan.Limits.GuardrailsCanUseSmartFallback,
			CanUseEscalateFallback: plan.Limits.GuardrailsCanUseEscalateFallback,
			CanManageTopics:        plan.Limits.GuardrailsCanManageTopics,
			CanCustomizeMessages:   plan.Limits.GuardrailsCanCustomizeMessages,
		},
		Branding: models.BrandingConfig{
			CanHideBranding:   plan.Limits.BrandingCanHideBranding,
			CanCustomBranding: plan.Limits.BrandingCanCustomBranding,
		},
		RateLimits: models.RateLimitsConfig{
			RequestsPerMinute: plan.Limits.RateLimitsRequestsPerMinute,
			WindowSeconds:     plan.Limits.RateLimitsWindowSeconds,
			Endpoints: map[string]models.EndpointLimits{
				"chat": {
					RequestsPerMinute: plan.Limits.RateLimitsChatRPM,
					WindowSeconds:     plan.Limits.RateLimitsChatWindow,
				},
				"sources": {
					RequestsPerMinute: plan.Limits.RateLimitsSourcesRPM,
					WindowSeconds:     plan.Limits.RateLimitsSourcesWindow,
				},
			},
		},
	}

	res := PlanResponse{
		ID:          plan.ID,
		Code:        plan.Code,
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Currency:    plan.Currency,
		Limits: PlanLimitsResponse{
			MaxChatbots:               plan.Limits.MaxChatbots,
			MaxMonthlyIngestions:      plan.Limits.MaxMonthlyIngestions,
			MaxMonthlyEmbeddingTokens: plan.Limits.MaxMonthlyEmbeddingTokens,
			MinReAddCooldownMinutes:   plan.Limits.MinReAddCooldownMinutes,
		},
		Features:        features,
		AvailableModels: availableModels,
	}

	api.WriteJSON(w, http.StatusOK, res)
}

// getPlanInfo retrieves plan details using repositories
func (h *PlanHandlers) getPlanInfo(ctx context.Context, u *models.User) (*planInfo, error) {
	if u.PlanID == nil || *u.PlanID == "" {
		return nil, &planError{msg: "no plan assigned"}
	}

	// Use GetPlanWithLimits to get complete plan with limits
	plan, err := h.PlanRepo.GetPlanWithLimits(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, &planError{msg: "plan not found"}
	}

	// Apply config defaults
	applyLimitsDefaults(plan.Limits, plan.Code)

	return &planInfo{
		ID:          plan.ID,
		Code:        plan.Code,
		Name:        nil, // Name comes from translations, not available in Plan model
		Description: nil, // Description comes from translations
		Price:       plan.Price,
		Currency:    plan.Currency,
		Limits:      plan.Limits,
	}, nil
}

// --- Helper functions ---

// applyLimitsDefaults sets fallback values for plan limits
func applyLimitsDefaults(limits *models.PlanLimits, planCode string) {
	if limits.FilesMaxFilesTotal == 0 {
		switch planCode {
		case policy.PlanFree.String():
			limits.FilesMaxFilesTotal = 5
		case policy.PlanPro.String():
			limits.FilesMaxFilesTotal = 100
		case policy.PlanUltra.String():
			limits.FilesMaxFilesTotal = 1000
		default:
			limits.FilesMaxFilesTotal = 5
		}
	}

	if limits.FilesMaxTextLength == 0 {
		limits.FilesMaxTextLength = 400000
	}

	if limits.ChatMinResponseTokenLimit == 0 {
		limits.ChatMinResponseTokenLimit = 20
	}
	if limits.ChatMaxResponseTokenLimit == 0 {
		limits.ChatMaxResponseTokenLimit = 8192
	}
}

// planError represents a plan-related error
type planError struct {
	msg string
}

func (e *planError) Error() string { return e.msg }
