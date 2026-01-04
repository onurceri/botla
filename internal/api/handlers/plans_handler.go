package handlers

import (
	"net/http"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/services"
)

// PublicPlanResponse represents a plan for public API endpoints
type PublicPlanResponse struct {
	Code     string                   `json:"code"`
	Name     string                   `json:"name,omitempty"`
	Price    float64                  `json:"price"`
	Currency string                   `json:"currency"`
	Limits   PublicPlanLimitsResponse `json:"limits"`
	Features PlanFeaturesResponse     `json:"features"`
}

// PublicPlanLimitsResponse holds top-level plan limits for public API
type PublicPlanLimitsResponse struct {
	MaxChatbots               int `json:"max_chatbots"`
	MaxMonthlyIngestions      int `json:"max_monthly_ingestions"`
	MaxMonthlyEmbeddingTokens int `json:"max_monthly_embedding_tokens"`
}

// PlansHandlers handles public plan endpoints (no auth required)
type PlansHandlers struct {
	planService *services.PlanService
}

// NewPlansHandlers creates a new PlansHandlers instance
func NewPlansHandlers(planService *services.PlanService) *PlansHandlers {
	return &PlansHandlers{
		planService: planService,
	}
}

// GetAllPlans handles GET /api/v1/plans
// Returns all active plans with their limits and features
func (h *PlansHandlers) GetAllPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.planService.GetAllPlans(r.Context())
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, "Failed to retrieve plans")
		return
	}

	response := make([]PublicPlanResponse, 0, len(plans))
	for _, plan := range plans {
		response = append(response, h.toPublicPlanResponse(plan))
	}

	api.WriteJSON(w, http.StatusOK, response)
}

// GetPlanByCode handles GET /api/v1/plans/{code}
// Returns a specific plan by its code
func (h *PlansHandlers) GetPlanByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, "Plan code is required")
		return
	}

	plan, err := h.planService.GetPlanByCode(r.Context(), code)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, "Failed to retrieve plan")
		return
	}
	if plan == nil {
		api.WriteErrorCode(w, http.StatusNotFound, "Plan not found")
		return
	}

	api.WriteJSON(w, http.StatusOK, h.toPublicPlanResponse(*plan))
}

// toPublicPlanResponse converts a Plan model to a public API response
func (h *PlansHandlers) toPublicPlanResponse(plan models.Plan) PublicPlanResponse {
	limits := plan.Limits
	if limits == nil {
		defaults := models.DefaultPlanLimits()
		limits = &defaults
	}

	// Convert PlanLimits to PlanFeaturesResponse
	features := PlanFeaturesResponse{
		Scraping: models.ScrapingConfig{
			DynamicEnabled:   limits.ScrapingDynamicEnabled,
			MaxURLsPerBot:    limits.ScrapingMaxURLsPerBot,
			MaxPagesPerCrawl: limits.ScrapingMaxPagesPerCrawl,
		},
		Files: models.FilesConfig{
			MaxSizeMB:      limits.FilesMaxSizeMB,
			MaxFilesPerBot: limits.FilesMaxFilesPerBot,
			MaxFilesTotal:  limits.FilesMaxFilesTotal,
			TotalStorageMB: limits.FilesTotalStorageMB,
			MaxTextLength:  limits.FilesMaxTextLength,
		},
		Chat: models.ChatConfig{
			DefaultModel:     limits.ChatDefaultModel,
			AllowedModels:    limits.ChatAllowedModels,
			MaxMonthlyTokens: limits.ChatMaxMonthlyTokens,
			RAG: models.RAGConfig{
				TopK:             limits.ChatRAGTopK,
				MaxContextTokens: limits.ChatRAGMaxContextTokens,
			},
			MaxSuggestedQuestions: limits.ChatMaxSuggestedQuestions,
			MaxManualQuestions:    limits.ChatMaxManualQuestions,
			MinResponseTokenLimit: limits.ChatMinResponseTokenLimit,
			MaxResponseTokenLimit: limits.ChatMaxResponseTokenLimit,
		},
		Refresh: models.RefreshConfig{
			Enabled:    limits.RefreshEnabled,
			MaxMonthly: limits.RefreshMaxMonthly,
		},
		Security: models.SecurityConfig{
			SecureEmbedEnabled: limits.SecuritySecureEmbedEnabled,
		},
		Guardrails: models.GuardrailsConfig{
			CanCustomizeThresholds: limits.GuardrailsCanCustomizeThresholds,
			CanUseSmartFallback:    limits.GuardrailsCanUseSmartFallback,
			CanUseEscalateFallback: limits.GuardrailsCanUseEscalateFallback,
			CanManageTopics:        limits.GuardrailsCanManageTopics,
			CanCustomizeMessages:   limits.GuardrailsCanCustomizeMessages,
		},
		Branding: models.BrandingConfig{
			CanHideBranding:   limits.BrandingCanHideBranding,
			CanCustomBranding: limits.BrandingCanCustomBranding,
		},
		RateLimits: models.RateLimitsConfig{
			RequestsPerMinute: limits.RateLimitsRequestsPerMinute,
			WindowSeconds:     limits.RateLimitsWindowSeconds,
			Endpoints: map[string]models.EndpointLimits{
				"chat": {
					RequestsPerMinute: limits.RateLimitsChatRPM,
					WindowSeconds:     limits.RateLimitsChatWindow,
				},
				"sources": {
					RequestsPerMinute: limits.RateLimitsSourcesRPM,
					WindowSeconds:     limits.RateLimitsSourcesWindow,
				},
			},
		},
	}

	return PublicPlanResponse{
		Code:     plan.Code,
		Price:    plan.Price,
		Currency: plan.Currency,
		Limits: PublicPlanLimitsResponse{
			MaxChatbots:               limits.MaxChatbots,
			MaxMonthlyIngestions:      limits.MaxMonthlyIngestions,
			MaxMonthlyEmbeddingTokens: limits.MaxMonthlyEmbeddingTokens,
		},
		Features: features,
	}
}
