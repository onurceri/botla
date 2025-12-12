package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// ListOrCreate handles GET (list) and POST (create) for chatbots
func (h *ChatbotHandlers) ListOrCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listChatbots(w, r, userID)
	case http.MethodPost:
		h.createChatbot(w, r, userID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// listChatbots returns all chatbots for a user or workspace
func (h *ChatbotHandlers) listChatbots(w http.ResponseWriter, r *http.Request, userID string) {
	var bots []models.Chatbot
	var err error

	// Check for workspace context from header
	if wsID, ok := middleware.WorkspaceIDFromContext(r.Context()); ok && wsID != "" {
		bots, err = db.GetChatbotsByWorkspace(r.Context(), h.DB, wsID)
	} else {
		// Fallback to user-based query for backward compatibility
		bots, err = db.GetChatbotsByUserID(r.Context(), h.DB, userID)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(bots); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// createChatbot creates a new chatbot
func (h *ChatbotHandlers) createChatbot(w http.ResponseWriter, r *http.Request, userID string) {
	var req createChatbotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	langCodeBCP := normalizeLocale(defaultString(req.Language, "tr-TR"))
	baseLang := baseLangCode(langCodeBCP)
	langCfg := langconfig.Get(baseLang)

	// Check plan limits
	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil {
		api.WriteLocalizedError(w, http.StatusInternalServerError, "get_plan_error: "+err.Error(), langCfg)
		return
	}

	if plan != nil && plan.Config.MaxChatbots > 0 {
		count, countErr := db.CountChatbotsByUserID(r.Context(), h.DB, userID)
		if countErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if count >= plan.Config.MaxChatbots {
			api.WriteLocalizedError(w, http.StatusForbidden, api.ErrMaxChatbotsExceeded, langCfg)
			return
		}
	}

	// Get workspace/org context from headers
	var wsID, orgID *string
	if ws, ok := middleware.WorkspaceIDFromContext(r.Context()); ok && ws != "" {
		wsID = &ws
	}
	if org, ok := middleware.OrgIDFromContext(r.Context()); ok && org != "" {
		orgID = &org
	}

	bot := h.buildNewChatbot(userID, wsID, orgID, req, langCodeBCP, langCfg)

	newID, err := db.CreateChatbot(r.Context(), h.DB, bot)
	if err != nil {
		api.WriteLocalizedError(w, http.StatusInternalServerError, "create_bot_error: "+err.Error(), langCfg)
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, newID)
	if err != nil || c == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(c); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// buildNewChatbot constructs a new chatbot with defaults
func (h *ChatbotHandlers) buildNewChatbot(userID string, wsID, orgID *string, req createChatbotRequest, langCode string, langCfg langconfig.LanguageConfig) *models.Chatbot {
	return &models.Chatbot{
		UserID:               userID,
		WorkspaceID:          wsID,
		OrganizationID:       orgID,
		Name:                 req.Name,
		Description:          req.Description,
		CustomInstruction:    defaultString(req.CustomInstruction, ""),
		LanguageCode:         langCode,
		Model:                defaultString(req.Model, config.ResolveChatbotModel(h.Cfg)),
		Temperature:          defaultFloat32(req.Temperature, 0.7),
		MaxTokens:            defaultInt(req.MaxTokens, 4096),
		ThemeColor:           defaultString(req.ThemeColor, "#3b82f6"),
		WelcomeMessage:       defaultString(req.WelcomeMessage, langCfg.ResponseTemplates.WelcomeMessage),
		Position:             defaultString(req.Position, "bottom-right"),
		BotMessageColor:      defaultString(req.BotMessageColor, "#fcfcfd"),
		UserMessageColor:     defaultString(req.UserMessageColor, "#2e408a"),
		BotMessageTextColor:  defaultString(req.BotMessageTextColor, "#030303"),
		UserMessageTextColor: defaultString(req.UserMessageTextColor, "#ffffff"),
		ChatFontFamily:       defaultString(req.ChatFontFamily, "Inter, sans-serif"),
		ChatHeaderColor:      defaultString(req.ChatHeaderColor, "#3b82f6"),
		ChatHeaderTextColor:  defaultString(req.ChatHeaderTextColor, "#ffffff"),
		ChatBackgroundColor:  defaultString(req.ChatBackgroundColor, "#fff5e6"),
		BotIcon:              req.BotIcon,
		BotDisplayName:       req.BotDisplayName,
		SecureEmbedEnabled:   boolValue(req.SecureEmbedEnabled, false),
		AllowedDomains:       req.AllowedDomains,
		EmbedSecret:          req.EmbedSecret,
		SuggestedQuestions:   suggestionsValue(req.SuggestedQuestions),
		SuggestionsEnabled:   boolValue(req.SuggestionsEnabled, false),
		IncludePaths:         pathsValue(req.IncludePaths),
		ExcludePaths:         pathsValue(req.ExcludePaths),
		SelectorWhitelist:    selectorsValue(req.SelectorWhitelist),
		DiscoveryMode:        defaultString(req.DiscoveryMode, "auto"),
		RefreshPolicy:        defaultString(req.RefreshPolicy, "manual"),
		RefreshFrequency:     req.RefreshFrequency,
		HideBranding:         boolValue(req.HideBranding, false),
		CustomBranding:       req.CustomBranding,
		ConfidenceThreshold:  defaultFloat64(req.ConfidenceThreshold, 0.7),
		FallbackMessages:     req.FallbackMessages,
		TopicRestrictions:    req.TopicRestrictions,
	}
}
