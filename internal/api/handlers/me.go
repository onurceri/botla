package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// MeResponse represents the /me endpoint response
type MeResponse struct {
	ID              string            `json:"id"`
	Email           string            `json:"email"`
	FullName        *string           `json:"full_name,omitempty"`
	AvatarURL       *string           `json:"avatar_url,omitempty"`
	PlanID          string            `json:"plan_id"`
	PlanCode        string            `json:"plan_code"`
	PlanName        *string           `json:"plan_name,omitempty"`
	PlanDescription *string           `json:"plan_description,omitempty"`
	PlanPrice       float64           `json:"plan_price"`
	PlanCurrency    string            `json:"plan_currency"`
	Config          models.PlanConfig `json:"config"`
	Usage           Usage             `json:"usage"`
}

// Usage represents user usage statistics
type Usage struct {
	FilesCount               int `json:"files_count"`
	MaxFilesCountInOneBot    int `json:"max_files_count_in_one_bot"`
	StorageUsedMB            int `json:"storage_used_mb"`
	URLsCount                int `json:"urls_count"`
	MaxURLsCountInOneBot     int `json:"max_urls_count_in_one_bot"`
	TokensUsed               int `json:"tokens_used"`
	IngestionsUsed           int `json:"ingestions_used"`
	IngestionEmbeddingTokens int `json:"ingestion_embedding_tokens"`
	RefreshCount             int `json:"refresh_count"`
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

// MeHandlers handles user profile endpoints
type MeHandlers struct {
	DB *sql.DB
}

// Me handles GET /me endpoint
func (h *MeHandlers) Me(w http.ResponseWriter, r *http.Request) {
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

	// Get plan info
	plan, err := h.getPlanInfo(r.Context(), u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get usage stats
	usage := h.getUserUsage(r.Context(), u.ID)

	// Build response
	res := h.buildMeResponse(u, plan, usage)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// getPlanInfo retrieves plan details with translations
func (h *MeHandlers) getPlanInfo(ctx context.Context, u *models.User) (*planInfo, error) {
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

// getUserUsage retrieves all usage statistics for a user
func (h *MeHandlers) getUserUsage(ctx context.Context, userID string) Usage {
	filesCount, _ := db.GetFileCountByUserID(ctx, h.DB, userID)
	urlsCount, _ := db.GetURLCountByUserID(ctx, h.DB, userID)
	tokensUsed, _ := db.GetMonthlyTokenUsage(ctx, h.DB, userID)
	storageUsedMB, _ := db.GetStorageUsedMBByUserID(ctx, h.DB, userID)
	usedIngestions, usedEmbedTokens, _ := db.GetMonthlyIngestionUsage(ctx, h.DB, userID, time.Now())
	maxFilesBot, _ := db.GetMaxFileCountInAnyBot(ctx, h.DB, userID)
	maxURLsBot, _ := db.GetMaxURLCountInAnyBot(ctx, h.DB, userID)
	refreshCount, _ := db.GetMonthlyRefreshCount(ctx, h.DB, userID, time.Now())

	return Usage{
		FilesCount:               filesCount,
		MaxFilesCountInOneBot:    maxFilesBot,
		StorageUsedMB:            storageUsedMB,
		URLsCount:                urlsCount,
		MaxURLsCountInOneBot:     maxURLsBot,
		TokensUsed:               tokensUsed,
		IngestionsUsed:           usedIngestions,
		IngestionEmbeddingTokens: usedEmbedTokens,
		RefreshCount:             refreshCount,
	}
}

// buildMeResponse constructs the response from user, plan, and usage data
func (h *MeHandlers) buildMeResponse(u *models.User, plan *planInfo, usage Usage) MeResponse {
	return MeResponse{
		ID:              u.ID,
		Email:           u.Email,
		FullName:        u.FullName,
		AvatarURL:       u.AvatarURL,
		PlanID:          plan.ID,
		PlanCode:        plan.Code,
		PlanName:        plan.Name,
		PlanDescription: plan.Description,
		PlanPrice:       plan.Price,
		PlanCurrency:    plan.Currency,
		Config:          plan.Config,
		Usage:           usage,
	}
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
}

// nullStringPtr converts sql.NullString to *string
func nullStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// nullStringValue returns the value or nil for sql.NullString
func nullStringValue(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

// planError represents a plan-related error
type planError struct {
	msg string
}

func (e *planError) Error() string { return e.msg }
