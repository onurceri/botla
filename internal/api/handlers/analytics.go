package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"strconv"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type AnalyticsHandlers struct {
	DB               *sql.DB
	AnalyticsService *services.AnalyticsService
	OrgService       *services.OrganizationService
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
		ws, err := h.OrgService.GetWorkspace(r.Context(), wsID)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// GetChatbotAnalyticsOverview returns aggregated analytics for a specific chatbot
func (h *AnalyticsHandlers) GetChatbotAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[4]

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Access Check
	allowed := false
	switch {
	case bot.WorkspaceID != nil && *bot.WorkspaceID != "":
		// Check workspace access via Org membership
		ws, err2 := h.OrgService.GetWorkspace(r.Context(), *bot.WorkspaceID)
		if err2 == nil && ws != nil {
			mem, err3 := h.OrgService.CheckMembership(r.Context(), userID, ws.OrganizationID)
			if err3 == nil && mem != nil {
				allowed = true
			}
		}
	case bot.OrganizationID != nil && *bot.OrganizationID != "":
		// Check Org membership
		mem, err2 := h.OrgService.CheckMembership(r.Context(), userID, *bot.OrganizationID)
		if err2 == nil && mem != nil {
			allowed = true
		}
	default:
		// Personal bot
		if bot.UserID == userID {
			allowed = true
		}
	}

	if !allowed {
		w.WriteHeader(http.StatusForbidden)
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetChatbotAnalyticsTrends returns daily trends for a chatbot
func (h *AnalyticsHandlers) GetChatbotAnalyticsTrends(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[4]

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Access Check
	allowed := false
	switch {
	case bot.WorkspaceID != nil && *bot.WorkspaceID != "":
		// Check workspace access via Org membership
		ws, err2 := h.OrgService.GetWorkspace(r.Context(), *bot.WorkspaceID)
		if err2 == nil && ws != nil {
			mem, err3 := h.OrgService.CheckMembership(r.Context(), userID, ws.OrganizationID)
			if err3 == nil && mem != nil {
				allowed = true
			}
		}
	case bot.OrganizationID != nil && *bot.OrganizationID != "":
		// Check Org membership
		mem, err2 := h.OrgService.CheckMembership(r.Context(), userID, *bot.OrganizationID)
		if err2 == nil && mem != nil {
			allowed = true
		}
	default:
		// Personal bot
		if bot.UserID == userID {
			allowed = true
		}
	}

	if !allowed {
		w.WriteHeader(http.StatusForbidden)
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetSourceUsage returns source usage analytics for a chatbot
func (h *AnalyticsHandlers) GetSourceUsage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[4]

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Access Check
	allowed := false
	switch {
	case bot.WorkspaceID != nil && *bot.WorkspaceID != "":
		// Check workspace access via Org membership
		ws, err2 := h.OrgService.GetWorkspace(r.Context(), *bot.WorkspaceID)
		if err2 == nil && ws != nil {
			mem, err3 := h.OrgService.CheckMembership(r.Context(), userID, ws.OrganizationID)
			if err3 == nil && mem != nil {
				allowed = true
			}
		}
	case bot.OrganizationID != nil && *bot.OrganizationID != "":
		// Check Org membership
		mem, err2 := h.OrgService.CheckMembership(r.Context(), userID, *bot.OrganizationID)
		if err2 == nil && mem != nil {
			allowed = true
		}
	default:
		// Personal bot
		if bot.UserID == userID {
			allowed = true
		}
	}

	if !allowed {
		w.WriteHeader(http.StatusForbidden)
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
