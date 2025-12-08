package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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

// listChatbots returns all chatbots for a user
func (h *ChatbotHandlers) listChatbots(w http.ResponseWriter, r *http.Request, userID string) {
	bots, err := db.GetChatbotsByUserID(r.Context(), h.DB, userID)
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

	bot := h.buildNewChatbot(userID, req, langCodeBCP, langCfg)

	newID, err := db.CreateChatbot(r.Context(), h.DB, bot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func (h *ChatbotHandlers) buildNewChatbot(userID string, req createChatbotRequest, langCode string, langCfg langconfig.LanguageConfig) *models.Chatbot {
	return &models.Chatbot{
		UserID:               userID,
		Name:                 req.Name,
		Description:          req.Description,
		SystemPrompt:         defaultString(req.SystemPrompt, langCfg.ResponseTemplates.DefaultPersonaPrompt),
		LanguageCode:         langCode,
		Model:                defaultString(req.Model, config.ResolveChatbotModel(h.Cfg)),
		Temperature:          defaultFloat32(req.Temperature, 0.7),
		MaxTokens:            defaultInt(req.MaxTokens, 512),
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
	}
}
