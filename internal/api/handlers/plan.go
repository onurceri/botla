package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
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
	Config      models.PlanConfig
}

// PlanHandlers handles plan-related endpoints
type PlanHandlers struct {
	DB *sql.DB
}

// GetPlan handles GET /me/plan endpoint
func (h *PlanHandlers) GetPlan(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u, err := db.GetUserByID(r.Context(), h.DB, uid)
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

	// Initialize model service
	modelService := services.NewModelService(h.DB)
	availableModels, err := modelService.GetAvailableModels(r.Context(), plan.Config.Chat.AllowedModels)
	if err != nil {
		// Log error but continue with empty/default models to avoid blocking plan info
		// In production, you might want to log this properly
		availableModels = []models.ModelInfo{}
	}

	res := PlanResponse{
		ID:          plan.ID,
		Code:        plan.Code,
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Currency:    plan.Currency,
		Limits: PlanLimitsResponse{
			MaxChatbots:               plan.Config.MaxChatbots,
			MaxMonthlyIngestions:      plan.Config.MaxMonthlyIngestions,
			MaxMonthlyEmbeddingTokens: plan.Config.MaxMonthlyEmbeddingTokens,
			MinReAddCooldownMinutes:   plan.Config.MinReAddCooldownMinutes,
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
		AvailableModels: availableModels,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// getPlanInfo retrieves plan details with translations
func (h *PlanHandlers) getPlanInfo(ctx context.Context, u *models.User) (*planInfo, error) {
	if u.PlanID == nil || *u.PlanID == "" {
		return nil, &planError{msg: "no plan assigned"}
	}

	langID := u.PreferredLanguageID
	if langID == nil {
		var defaultID string
		_ = h.DB.QueryRow(`SELECT id FROM languages WHERE code='tr-TR'`).Scan(&defaultID)
		if defaultID != "" {
			langID = &defaultID
		}
	}

	var planCode string
	var name sql.NullString
	var desc sql.NullString
	var planPrice float64
	var planCurrency string
	var config models.PlanConfig

	err := h.DB.QueryRowContext(ctx, `
		SELECT p.code, pt.name, pt.description, p.price, p.currency, p.config
		FROM plans p
		LEFT JOIN plan_translations pt ON pt.plan_id=p.id AND pt.language_id=$2
		WHERE p.id=$1
	`, u.PlanID, langID).Scan(&planCode, &name, &desc, &planPrice, &planCurrency, &config)

	if err != nil {
		planCode = "free"
	}

	// Apply config defaults
	applyConfigDefaults(&config, planCode)

	return &planInfo{
		ID:          *u.PlanID,
		Code:        planCode,
		Name:        nullStringPtr(name),
		Description: nullStringPtr(desc),
		Price:       planPrice,
		Currency:    planCurrency,
		Config:      config,
	}, nil
}

// --- Helper functions ---

// applyConfigDefaults sets fallback values for plan config
func applyConfigDefaults(config *models.PlanConfig, planCode string) {
	if config.Files.MaxFilesTotal == 0 {
		switch planCode {
		case "free":
			config.Files.MaxFilesTotal = 5
		case "pro":
			config.Files.MaxFilesTotal = 100
		case "ultra":
			config.Files.MaxFilesTotal = 1000
		default:
			config.Files.MaxFilesTotal = 5
		}
	}

	if config.Files.MaxTextLength == 0 {
		config.Files.MaxTextLength = 400000
	}

	if config.Chat.MinResponseTokenLimit == 0 {
		config.Chat.MinResponseTokenLimit = 20
	}
	if config.Chat.MaxResponseTokenLimit == 0 {
		config.Chat.MaxResponseTokenLimit = 8192
	}
}

// nullStringPtr converts sql.NullString to *string
func nullStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// planError represents a plan-related error
type planError struct {
	msg string
}

func (e *planError) Error() string { return e.msg }
