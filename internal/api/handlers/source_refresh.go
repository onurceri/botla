package handlers

import (
	"encoding/json"
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

	s, c, sourceID, ok := getSourceContext(w, r, h.DB, h.WorkspaceService, h.OrgService, "/refresh")
	if !ok {
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())

	base := api.BaseLang(c.LanguageCode)
	cfg := api.ConfigFromBase(base)

	if s.SourceType != "url" {
		api.WriteLocalizedError(w, http.StatusBadRequest, api.ErrOnlyURLRefresh, cfg)
		return
	}

	// Check if source is already processing
	if s.Status == "pending" || s.Status == "processing" {
		api.WriteLocalizedError(w, http.StatusConflict, api.ErrSourceAlreadyProcessing, cfg)
		return
	}

	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if refresh is enabled for this plan
	if !plan.Config.Refresh.Enabled {
		api.WriteLocalizedError(w, http.StatusForbidden, api.ErrPlanRefreshUnavailable, cfg)
		return
	}

	// Check monthly refresh quota
	usedRefreshes, _ := db.GetMonthlyRefreshCount(r.Context(), h.DB, userID, time.Now())
	if plan.Config.Refresh.MaxMonthly > 0 && usedRefreshes >= plan.Config.Refresh.MaxMonthly {
		api.WriteLocalizedError(w, http.StatusPaymentRequired, api.ErrMonthlyRefreshExceeded, cfg)
		return
	}

	// Check cooldown
	cooldownMin := plan.Config.MinReAddCooldownMinutes
	if cooldownMin > 0 && s.LastRefreshedAt != nil {
		elapsed := time.Since(*s.LastRefreshedAt)
		if elapsed < time.Duration(cooldownMin)*time.Minute {
			remaining := time.Duration(cooldownMin)*time.Minute - elapsed
			w.Header().Set("Retry-After", strconv.Itoa(int(remaining.Seconds())))
			api.WriteLocalizedError(w, http.StatusTooManyRequests, api.ErrRefreshCooldownActive, cfg)
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

	// Enqueue for processing
	if h.Queue != nil {
		h.Queue.Enqueue(sourceID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": sourceID})
}
