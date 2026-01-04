package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/middleware"
)

type OnboardingHandlers struct {
	DB       *sql.DB
	UserRepo repository.UserRepository
}

// GetOnboardingState handles GET /api/v1/me/onboarding
func (h *OnboardingHandlers) GetOnboardingState(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.UserRepo.GetByID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"completed": user.OnboardingCompleted,
		"skipped":   user.OnboardingSkipped,
		"step":      user.OnboardingStep,
	}
	if user.OnboardingData != nil {
		response["data"] = user.OnboardingData
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// UpdateOnboardingState handles PUT /api/v1/me/onboarding
func (h *OnboardingHandlers) UpdateOnboardingState(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req struct {
		Step int                    `json:"step"`
		Data *models.OnboardingData `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		req.Data = &models.OnboardingData{}
	}

	if err := h.UserRepo.UpdateOnboardingState(r.Context(), userID, req.Step, req.Data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SkipOnboarding handles POST /api/v1/me/onboarding/skip
func (h *OnboardingHandlers) SkipOnboarding(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.UserRepo.SkipOnboarding(r.Context(), userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CompleteOnboarding handles POST /api/v1/me/onboarding/complete
func (h *OnboardingHandlers) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req struct {
		BotID string `json:"bot_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BotID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.CompleteOnboarding(r.Context(), userID, req.BotID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
