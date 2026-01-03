package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// RefreshSource handles POST /api/v1/sources/:id/refresh
func (h *SourcesHandlers) RefreshSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s, _, sourceID, ok := getSourceContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
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

	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if refresh is enabled for this plan
	if !plan.Config.Refresh.Enabled {
		api.WriteErrorCode(w, http.StatusForbidden, api.ErrPlanRefreshUnavailable)
		return
	}

	// Check monthly refresh quota
	usedRefreshes, _ := db.GetMonthlyRefreshCount(r.Context(), h.DB, userID, time.Now())
	if plan.Config.Refresh.MaxMonthly > 0 && usedRefreshes >= plan.Config.Refresh.MaxMonthly {
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
	if err = db.UpdateSourceForRefresh(r.Context(), h.DB, sourceID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Increment refresh count
	_ = db.IncrementRefreshCount(r.Context(), h.DB, userID, time.Now())

	api.WriteJSON(w, http.StatusAccepted, map[string]string{"id": sourceID})
}
