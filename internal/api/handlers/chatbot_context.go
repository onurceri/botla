package handlers

import (
	"net/http"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/httputil"
	"github.com/onurceri/botla-app/pkg/middleware"
)

// getChatbotContext helper to avoid code duplication across handlers.
// It handles authentication check, path parsing, database fetching, and access control.
func getChatbotContext(w http.ResponseWriter, r *http.Request, repo repository.ChatbotRepository, wsService *services.WorkspaceService, orgService *services.OrganizationService) (*models.Chatbot, string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, "", false
	}

	botID := r.PathValue("id")
	if botID == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, "", false
	}
	if botID == "new" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, "", false
	}
	if !httputil.IsValidUUID(botID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidIDFormat)
		return nil, "", false
	}

	if repo == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, "", false
	}

	c, err := repo.GetByID(r.Context(), botID)
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

// getChatbotContextWithRepo is like getChatbotContext but uses a ChatbotRepository interface.
// This enables handlers to use the repository pattern for better testability.
func getChatbotContextWithRepo(w http.ResponseWriter, r *http.Request, repo repository.ChatbotRepository, wsService *services.WorkspaceService, orgService *services.OrganizationService) (*models.Chatbot, string, bool) {
	return getChatbotContext(w, r, repo, wsService, orgService)
}

// getSourceContext helper to avoid code duplication across source handlers.
func getSourceContext(w http.ResponseWriter, r *http.Request, sourceRepo repository.SourceRepository, chatbotRepo repository.ChatbotRepository, wsService *services.WorkspaceService, orgService *services.OrganizationService) (*models.DataSource, *models.Chatbot, string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, nil, "", false
	}

	sourceID := r.PathValue("id")
	if sourceID == "" {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, "", false
	}
	if !httputil.IsValidUUID(sourceID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidIDFormat)
		return nil, nil, "", false
	}

	if sourceRepo == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, "", false
	}

	s, err := sourceRepo.GetByID(r.Context(), sourceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, "", false
	}
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, "", false
	}

	c, err := chatbotRepo.GetByID(r.Context(), s.ChatbotID)
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

// getSourceContextWithRepos is like getSourceContext but uses SourceRepository and ChatbotRepository interfaces.
// This enables handlers to use the repository pattern for better testability.
func getSourceContextWithRepos(w http.ResponseWriter, r *http.Request, sourceRepo repository.SourceRepository, chatbotRepo repository.ChatbotRepository, wsService *services.WorkspaceService, orgService *services.OrganizationService) (*models.DataSource, *models.Chatbot, string, bool) {
	return getSourceContext(w, r, sourceRepo, chatbotRepo, wsService, orgService)
}
