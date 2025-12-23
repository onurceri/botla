package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// parseBotIDFromPath extracts bot ID from paths like /api/v1/chatbots/:id/...
func parseBotIDFromPath(path string) (string, bool) {
	const prefix = "/api/v1/chatbots/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	botID := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(botID, "/"); idx != -1 {
		botID = botID[:idx]
	}
	if botID == "" {
		return "", false
	}
	return botID, true
}

// parseSourceIDFromPath extracts source ID from /api/v1/sources/:id
func parseSourceIDFromPath(p string) (string, bool) {
	const prefix = "/api/v1/sources/"
	if !strings.HasPrefix(p, prefix) {
		return "", false
	}
	sourceID := strings.TrimPrefix(p, prefix)
	// Ensure no trailing paths like /refresh
	if strings.Contains(sourceID, "/") || sourceID == "" {
		return "", false
	}
	return sourceID, true
}

// parseRefreshSourceIDFromPath extracts source ID from /api/v1/sources/:id/refresh
func parseRefreshSourceIDFromPath(p string) (string, bool) {
	const prefix = "/api/v1/sources/"
	const suffix = "/refresh"
	if !strings.HasPrefix(p, prefix) || !strings.HasSuffix(p, suffix) {
		return "", false
	}
	sourceID := strings.TrimSuffix(strings.TrimPrefix(p, prefix), suffix)
	if sourceID == "" {
		return "", false
	}
	return sourceID, true
}

// getChatbotContext helper to avoid code duplication across handlers.
// It handles authentication check, path parsing, database fetching, and access control.
func getChatbotContext(w http.ResponseWriter, r *http.Request, dbConn *sql.DB, wsService *services.WorkspaceService, orgService *services.OrganizationService) (*models.Chatbot, string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, "", false
	}

	botID := r.PathValue("id")
	if botID == "" {
		// Fallback to manual parsing for routes that might not use {id} or during tests
		botID, ok = parseBotIDFromPath(r.URL.Path)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return nil, "", false
		}
	}
	if botID == "new" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, "", false
	}

	if dbConn == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, "", false
	}

	c, err := db.GetChatbotByID(r.Context(), dbConn, botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, "", false
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, "", false
	}

	allowed, err := checkChatbotAccess(r.Context(), c, userID, wsService, orgService)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, "", false
	}
	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return nil, "", false
	}

	return c, botID, true
}

// getSourceContext helper to avoid code duplication across source handlers.
func getSourceContext(w http.ResponseWriter, r *http.Request, dbConn *sql.DB, wsService *services.WorkspaceService, orgService *services.OrganizationService, suffix string) (*models.DataSource, *models.Chatbot, string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, nil, "", false
	}

	const prefix = "/api/v1/sources/"
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, "", false
	}

	sourceID := strings.TrimPrefix(path, prefix)
	if suffix != "" {
		if !strings.HasSuffix(sourceID, suffix) {
			w.WriteHeader(http.StatusNotFound)
			return nil, nil, "", false
		}
		sourceID = strings.TrimSuffix(sourceID, suffix)
	}

	// Ensure no remaining slashes (e.g. /api/v1/sources/id/something/else)
	if strings.Contains(sourceID, "/") || sourceID == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, "", false
	}

	if dbConn == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, "", false
	}

	s, err := db.GetSourceByID(r.Context(), dbConn, sourceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, "", false
	}
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, "", false
	}

	c, err := db.GetChatbotByID(r.Context(), dbConn, s.ChatbotID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, "", false
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, "", false
	}

	allowed, err := checkChatbotAccess(r.Context(), c, userID, wsService, orgService)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, "", false
	}
	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return nil, nil, "", false
	}

	return s, c, sourceID, true
}
