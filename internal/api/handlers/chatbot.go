package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// ChatbotHandlers handles chatbot-related HTTP endpoints
type ChatbotHandlers struct {
	DB  *sql.DB
	Cfg *config.Config
}

type createChatbotRequest struct {
	Name                 string    `json:"name"`
	Description          *string   `json:"description"`
	SystemPrompt         *string   `json:"system_prompt"`
	Language             *string   `json:"language"`
	Model                *string   `json:"model"`
	Temperature          *float32  `json:"temperature"`
	MaxTokens            *int      `json:"max_tokens"`
	ThemeColor           *string   `json:"theme_color"`
	WelcomeMessage       *string   `json:"welcome_message"`
	Position             *string   `json:"position"`
	BotMessageColor      *string   `json:"bot_message_color"`
	UserMessageColor     *string   `json:"user_message_color"`
	BotMessageTextColor  *string   `json:"bot_message_text_color"`
	UserMessageTextColor *string   `json:"user_message_text_color"`
	ChatFontFamily       *string   `json:"chat_font_family"`
	ChatHeaderColor      *string   `json:"chat_header_color"`
	ChatHeaderTextColor  *string   `json:"chat_header_text_color"`
	ChatBackgroundColor  *string   `json:"chat_background_color"`
	BotIcon              *string   `json:"bot_icon"`
	BotDisplayName       *string   `json:"bot_display_name"`
	SecureEmbedEnabled   *bool     `json:"secure_embed_enabled"`
	AllowedDomains       *string   `json:"allowed_domains"`
	EmbedSecret          *string   `json:"embed_secret"`
	SuggestedQuestions   *[]string `json:"suggested_questions"`
	SuggestionsEnabled   *bool     `json:"suggestions_enabled"`
}

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
		w.WriteHeader(http.StatusInternalServerError)
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
	}
}

// ByID handles GET/PUT/DELETE for a specific chatbot
func (h *ChatbotHandlers) ByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	botID, ok := parseBotIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if botID == "new" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, botID)
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

	switch r.Method {
	case http.MethodGet:
		h.getChatbot(w, c)
	case http.MethodPut:
		h.updateChatbot(w, r, c, botID)
	case http.MethodDelete:
		h.deleteChatbot(w, r, botID, userID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// getChatbot returns a single chatbot
func (h *ChatbotHandlers) getChatbot(w http.ResponseWriter, c *models.Chatbot) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(c); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// updateChatbot handles PUT request to update a chatbot
func (h *ChatbotHandlers) updateChatbot(w http.ResponseWriter, r *http.Request, c *models.Chatbot, botID string) {
	var req createChatbotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate hex color if provided
	if req.ChatBackgroundColor != nil {
		s := strings.TrimSpace(*req.ChatBackgroundColor)
		if s != "" && !isValidHexColor(s) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Apply updates
	applyChatbotUpdates(c, req)

	if err := db.UpdateChatbot(r.Context(), h.DB, c); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Re-read for updated_at change
	c2, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil || c2 == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(c2); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// deleteChatbot handles DELETE request
func (h *ChatbotHandlers) deleteChatbot(w http.ResponseWriter, r *http.Request, botID, userID string) {
	if err := db.SoftDeleteChatbot(r.Context(), h.DB, botID, userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// applyChatbotUpdates applies request fields to chatbot model
func applyChatbotUpdates(c *models.Chatbot, req createChatbotRequest) {
	if req.Name != "" {
		c.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != nil {
		c.Description = req.Description
	}
	if req.SystemPrompt != nil {
		c.SystemPrompt = *req.SystemPrompt
	}
	if req.Language != nil {
		c.LanguageCode = normalizeLocale(*req.Language)
	}
	if req.Model != nil {
		c.Model = *req.Model
	}
	if req.Temperature != nil {
		c.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		c.MaxTokens = *req.MaxTokens
	}
	if req.ThemeColor != nil {
		c.ThemeColor = *req.ThemeColor
	}
	if req.WelcomeMessage != nil {
		c.WelcomeMessage = *req.WelcomeMessage
	}
	if req.Position != nil {
		c.Position = *req.Position
	}
	if req.BotMessageColor != nil {
		c.BotMessageColor = *req.BotMessageColor
	}
	if req.UserMessageColor != nil {
		c.UserMessageColor = *req.UserMessageColor
	}
	if req.BotMessageTextColor != nil {
		c.BotMessageTextColor = *req.BotMessageTextColor
	}
	if req.UserMessageTextColor != nil {
		c.UserMessageTextColor = *req.UserMessageTextColor
	}
	if req.ChatFontFamily != nil {
		c.ChatFontFamily = *req.ChatFontFamily
	}
	if req.ChatHeaderColor != nil {
		c.ChatHeaderColor = *req.ChatHeaderColor
	}
	if req.ChatHeaderTextColor != nil {
		c.ChatHeaderTextColor = *req.ChatHeaderTextColor
	}
	if req.ChatBackgroundColor != nil {
		c.ChatBackgroundColor = *req.ChatBackgroundColor
	}
	if req.BotIcon != nil {
		c.BotIcon = req.BotIcon
	}
	if req.BotDisplayName != nil {
		c.BotDisplayName = req.BotDisplayName
	}
	if req.SecureEmbedEnabled != nil {
		c.SecureEmbedEnabled = *req.SecureEmbedEnabled
	}
	if req.AllowedDomains != nil {
		c.AllowedDomains = req.AllowedDomains
	}
	if req.EmbedSecret != nil {
		c.EmbedSecret = req.EmbedSecret
	}
	if req.SuggestedQuestions != nil {
		c.SuggestedQuestions = normalizeSuggestions(*req.SuggestedQuestions)
	}
	if req.SuggestionsEnabled != nil {
		c.SuggestionsEnabled = *req.SuggestionsEnabled
	}
}

// --- Helper functions ---

// parseBotIDFromPath extracts bot ID from /api/v1/chatbots/:id
func parseBotIDFromPath(path string) (string, bool) {
	const prefix = "/api/v1/chatbots/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	botID := strings.TrimPrefix(path, prefix)
	if botID == "" {
		return "", false
	}
	return botID, true
}

// normalizeSuggestions deduplicates and truncates suggestions
func normalizeSuggestions(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, q := range in {
		t := strings.TrimSpace(q)
		if t == "" {
			continue
		}
		if len(t) > 120 {
			t = t[:120]
		}
		k := strings.ToLower(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, t)
		if len(out) >= 6 {
			break
		}
	}
	return out
}

var hexColorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func isValidHexColor(s string) bool { return hexColorRe.MatchString(s) }

func defaultString(p *string, d string) string {
	if p != nil {
		s := strings.TrimSpace(*p)
		if s != "" {
			return s
		}
	}
	return d
}

func defaultInt(p *int, d int) int {
	if p != nil {
		return *p
	}
	return d
}

func defaultFloat32(p *float32, d float32) float32 {
	if p != nil {
		return *p
	}
	return d
}

func boolValue(p *bool, d bool) bool {
	if p != nil {
		return *p
	}
	return d
}

func suggestionsValue(p *[]string) []string {
	if p != nil {
		return normalizeSuggestions(*p)
	}
	return nil
}

func normalizeLocale(code string) string {
	if code == "" {
		return "tr-TR"
	}
	s := strings.TrimSpace(code)
	switch s {
	case "tr":
		return "tr-TR"
	case "en":
		return "en-US"
	}
	return s
}

func baseLangCode(code string) string {
	s := strings.TrimSpace(code)
	if s == "" {
		return "tr"
	}
	if i := strings.Index(s, "-"); i > 0 {
		s = s[:i]
	}
	return s
}
