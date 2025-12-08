package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// RefreshSource handles POST /api/v1/sources/:id/refresh
func (h *SourcesHandlers) RefreshSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sourceID, ok := parseRefreshSourceIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s, err := db.GetSourceByID(r.Context(), h.DB, sourceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Verify ownership
	c, err := db.GetChatbotByID(r.Context(), h.DB, s.ChatbotID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if c.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Only URL sources can be refreshed
	if s.SourceType != "url" {
		http.Error(w, "Only URL sources can be refreshed", http.StatusBadRequest)
		return
	}

	// Check if source is already processing
	if s.Status == "pending" || s.Status == "processing" {
		http.Error(w, "Source is already being processed", http.StatusConflict)
		return
	}

	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if refresh is enabled for this plan
	if !plan.Config.Refresh.Enabled {
		http.Error(w, "Refresh feature is not available on your plan", http.StatusForbidden)
		return
	}

	// Check monthly refresh quota
	usedRefreshes, _ := db.GetMonthlyRefreshCount(r.Context(), h.DB, userID, time.Now())
	if plan.Config.Refresh.MaxMonthly > 0 && usedRefreshes >= plan.Config.Refresh.MaxMonthly {
		http.Error(w, "Monthly refresh limit exceeded", http.StatusPaymentRequired)
		return
	}

	// Check cooldown
	cooldownMin := plan.Config.MinReAddCooldownMinutes
	if cooldownMin > 0 && s.LastRefreshedAt != nil {
		elapsed := time.Since(*s.LastRefreshedAt)
		if elapsed < time.Duration(cooldownMin)*time.Minute {
			remaining := time.Duration(cooldownMin)*time.Minute - elapsed
			w.Header().Set("Retry-After", strconv.Itoa(int(remaining.Seconds())))
			http.Error(w, "Refresh cooldown active", http.StatusTooManyRequests)
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
