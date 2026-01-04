package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/pkg/middleware"
)

// RefreshSource handles POST /api/v1/sources/:id/refresh
func (h *SourcesHandlers) RefreshSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s, _, sourceID, ok := getSourceContext(w, r, h.SourceRepo, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())

	if s.SourceType != "url" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrOnlyURLRefresh)
		return
	}

	// Check if source is already processing
	if s.Status == "pending" || s.Status == "processing" {
		api.WriteErrorCode(w, http.StatusConflict, api.ErrSourceAlreadyProcessing)
		return
	}

	plan, err := h.PlanRepo.GetPlanWithLimits(r.Context(), userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if refresh is enabled for this plan
	if !plan.Limits.RefreshEnabled {
		api.WriteErrorCode(w, http.StatusForbidden, api.ErrPlanRefreshUnavailable)
		return
	}

	// Check monthly refresh quota
	usedRefreshes, _ := h.UsageRepo.GetMonthlyRefreshCount(r.Context(), userID, time.Now())
	if plan.Limits.RefreshMaxMonthly > 0 && usedRefreshes >= plan.Limits.RefreshMaxMonthly {
		api.WriteErrorCode(w, http.StatusPaymentRequired, api.ErrMonthlyRefreshExceeded)
		return
	}

	// Check cooldown
	if remaining, ok := h.checkCooldown(r, s.LastRefreshedAt, plan); !ok {
		w.Header().Set("Retry-After", strconv.Itoa(int(remaining.Seconds())))
		api.WriteErrorCode(w, http.StatusTooManyRequests, api.ErrRefreshCooldownActive)
		return
	}

	// Enqueue for processing
	if h.Queue != nil {
		_, enqErr := h.Queue.EnqueueSource(r.Context(), sourceID, s.ChatbotID)
		if enqErr != nil {
			h.logError("refresh_enqueue_failed", map[string]any{"error": enqErr.Error(), "source_id": sourceID})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Update source for refresh
	if err = h.SourceRepo.UpdateForRefresh(r.Context(), sourceID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Increment refresh count
	_ = h.UsageRepo.IncrementRefreshCount(r.Context(), userID, time.Now())

	api.WriteJSON(w, http.StatusAccepted, map[string]string{"id": sourceID})
}
