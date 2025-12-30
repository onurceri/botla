package handlers

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
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
	return PublicPlanResponse{
		Code:     plan.Code,
		Price:    plan.Price,
		Currency: plan.Currency,
		Limits: PublicPlanLimitsResponse{
			MaxChatbots:               plan.Config.MaxChatbots,
			MaxMonthlyIngestions:      plan.Config.MaxMonthlyIngestions,
			MaxMonthlyEmbeddingTokens: plan.Config.MaxMonthlyEmbeddingTokens,
		},
		Features: PlanFeaturesResponse{
			Scraping:   plan.Config.Scraping,
			Files:      plan.Config.Files,
			Chat:       plan.Config.Chat,
			Refresh:    plan.Config.Refresh,
			Security:   plan.Config.Security,
			Guardrails: plan.Config.Guardrails,
			Branding:   plan.Config.Branding,
			RateLimits: plan.Config.RateLimits,
		},
	}
}
