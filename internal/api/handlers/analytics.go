package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"strconv"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type AnalyticsHandlers struct {
	DB               *sql.DB
	AnalyticsService *services.AnalyticsService
	OrgService       *services.OrganizationService
	WorkspaceService *services.WorkspaceService
}

func (h *AnalyticsHandlers) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Determine scope from context
	var orgIDPtr, wsIDPtr *string

	// Check for Workspace Context
	if wsID, ok := middleware.WorkspaceIDFromContext(r.Context()); ok && wsID != "" {
		// Validating access to workspace:
		// 1. Get workspace info (to get OrgID if needed, or if memberships are per workspace)
		// Since membership is per Organization, we need to know valid Org of this WS.
		ws, err := h.WorkspaceService.GetWorkspace(r.Context(), wsID)
		if err != nil || ws == nil {
			w.WriteHeader(http.StatusForbidden) // Or NotFound
			return
		}

		// 2. Check Org Membership
		mem, err := h.OrgService.CheckMembership(r.Context(), userID, ws.OrganizationID)
		if err != nil || mem == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		wsIDPtr = &wsID
	} else if orgID, ok := middleware.OrgIDFromContext(r.Context()); ok && orgID != "" {
		// Global Org Context
		mem, err := h.OrgService.CheckMembership(r.Context(), userID, orgID)
		if err != nil || mem == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		orgIDPtr = &orgID
	}

	// Fetch data with scoped logic
	data, err := db.GetGlobalAnalytics(r.Context(), h.DB, userID, orgIDPtr, wsIDPtr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		data = []db.AnalyticsPoint{}
	}

	api.WriteJSON(w, http.StatusOK, data)
}

// GetChatbotAnalyticsOverview returns aggregated analytics for a specific chatbot
func (h *AnalyticsHandlers) GetChatbotAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	_, botID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	stats, err := h.AnalyticsService.GetChatbotOverview(r.Context(), botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stats == nil {
		stats = &models.AnalyticsOverview{}
	}

	api.WriteJSON(w, http.StatusOK, stats)
}

// GetChatbotAnalyticsTrends returns daily trends for a chatbot
func (h *AnalyticsHandlers) GetChatbotAnalyticsTrends(w http.ResponseWriter, r *http.Request) {
	_, botID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Parse days param
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err2 := strconv.Atoi(d); err2 == nil && n > 0 && n <= 365 {
			days = n
		}
	}

	data, err := h.AnalyticsService.GetChatbotTrends(r.Context(), botID, days)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, data)
}

// GetSourceUsage returns source usage analytics for a chatbot
func (h *AnalyticsHandlers) GetSourceUsage(w http.ResponseWriter, r *http.Request) {
	_, botID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	// Parse days param
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err2 := strconv.Atoi(d); err2 == nil && n > 0 && n <= 365 {
			days = n
		}
	}

	stats, err := db.GetSourceUsageStats(r.Context(), h.DB, botID, days)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
