package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/internal/services"
)

type publicChatbot struct {
	ID                   string                 `json:"id"`
	ThemeColor           string                 `json:"theme_color"`
	WelcomeMessage       string                 `json:"welcome_message"`
	Position             string                 `json:"position"`
	BotMessageColor      string                 `json:"bot_message_color"`
	UserMessageColor     string                 `json:"user_message_color"`
	BotMessageTextColor  string                 `json:"bot_message_text_color"`
	UserMessageTextColor string                 `json:"user_message_text_color"`
	ChatFontFamily       string                 `json:"chat_font_family"`
	ChatHeaderColor      string                 `json:"chat_header_color"`
	ChatHeaderTextColor  string                 `json:"chat_header_text_color"`
	ChatBackgroundColor  string                 `json:"chat_background_color"`
	BotIcon              *string                `json:"bot_icon,omitempty"`
	BotDisplayName       *string                `json:"bot_display_name,omitempty"`
	SuggestedQuestions   []string               `json:"suggested_questions,omitempty"`
	MaxChars             int                    `json:"max_chars"`
	HideBranding         bool                   `json:"hide_branding"`
	CustomBranding       *models.CustomBranding `json:"custom_branding,omitempty"`
	HandoffEnabled       bool                   `json:"handoff_enabled"`
}

func PublicChatbotConfig(dbpool *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/api/v1/public/chatbots/"
		path := r.URL.Path
		if !strings.HasPrefix(path, prefix) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var err error
		botID := strings.TrimPrefix(path, prefix)
		if botID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		c, err := db.GetChatbotByID(r.Context(), dbpool, botID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if c == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Cache final suggestions keyed by updated_at to auto-invalidate
		cache := scraper.NewCache()
		key := publicSuggestionsCacheKey(c)
		var final []string
		if c.SuggestionsEnabled {
			if v, ok := cache.Get(key); ok {
				// parse cached JSON array
				var arr []string
				_ = json.Unmarshal([]byte(v), &arr)
				final = arr
			} else {
				final = c.SuggestedQuestions
				b, _ := json.Marshal(final)
				_ = cache.Set(key, string(b), 10*time.Minute)
			}
		} else {
			final = []string{}
		}
		var cb *models.CustomBranding
		if c.HideBranding {
			cb = c.CustomBranding
		} else {
			cb = nil
		}
		out := publicChatbot{
			ID:                   c.ID,
			ThemeColor:           c.ThemeColor,
			WelcomeMessage:       c.WelcomeMessage,
			Position:             c.Position,
			BotMessageColor:      c.BotMessageColor,
			UserMessageColor:     c.UserMessageColor,
			BotMessageTextColor:  c.BotMessageTextColor,
			UserMessageTextColor: c.UserMessageTextColor,
			ChatFontFamily:       c.ChatFontFamily,
			ChatHeaderColor:      c.ChatHeaderColor,
			ChatHeaderTextColor:  c.ChatHeaderTextColor,
			ChatBackgroundColor:  c.ChatBackgroundColor,
			BotIcon:              c.BotIcon,
			BotDisplayName:       c.BotDisplayName,
			SuggestedQuestions:   final,
			MaxChars:             1000,
			HideBranding:         c.HideBranding,
			CustomBranding:       cb,
			HandoffEnabled:       c.HandoffEnabled,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(out); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func publicSuggestionsCacheKey(c *models.Chatbot) string {
	return "public:chatbot:" + c.ID + ":suggestions:" + c.UpdatedAt.UTC().Format(time.RFC3339Nano)
}

type publicChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

type publicSourceUsed struct {
	ChunkIndex int    `json:"chunk_index"`
	SourceType string `json:"source_type"`
}

type publicChatResponse struct {
	Response       string             `json:"response"`
	MessageID      string             `json:"message_id"`
	TokensUsed     int                `json:"tokens_used"`
	SourcesUsed    []publicSourceUsed `json:"sources_used"`
	ConfidenceTier string             `json:"confidence_tier,omitempty"`
}

// PublicHandlers contains handlers for public (unauthenticated) endpoints
type PublicHandlers struct {
	DB          *sql.DB
	ChatService *services.ChatService
}

// PublicChat handles public chat requests using the ChatService
func (h *PublicHandlers) PublicChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	const prefix = "/api/v1/public/chatbots/"
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, "/chat") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	botID := strings.TrimSuffix(strings.TrimPrefix(path, prefix), "/chat")
	cbot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if cbot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Secure Embed Enforcement
	if cbot.SecureEmbedEnabled {
		// 1. Origin Check
		if cbot.AllowedDomains != nil && *cbot.AllowedDomains != "" {
			origin := r.Header.Get("Origin")
			// If no origin provided, block if restriction is enabled?
			// Usually browsers send Origin. If called from non-browser, it might be empty.
			// Let's require it if domains are restricted.
			if origin == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			allowed := false
			domains := strings.Split(*cbot.AllowedDomains, ",")
			for _, d := range domains {
				d = strings.TrimSpace(d)
				if d == "" {
					continue
				}
				// Simple check: origin contains the allowed domain
				// e.g. "example.com" matches "https://example.com" and "https://sub.example.com"
				if strings.Contains(origin, d) {
					allowed = true
					break
				}
			}
			if !allowed {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// 2. Token Check
		tokenStr := r.Header.Get("X-Embed-Token")
		if tokenStr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		secret := ""
		if cbot.EmbedSecret != nil {
			secret = *cbot.EmbedSecret
		}

		token, errParse := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if errParse != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Check chatbot_id matches
			if cid, ok := claims["chatbot_id"].(string); !ok || cid != cbot.ID {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// Plan and limits
	plan, err := db.GetPlanByUserID(r.Context(), h.DB, cbot.UserID)
	var ragConfig models.RAGConfig
	if err == nil && plan != nil {
		ragConfig = plan.Config.Chat.RAG
		// Check monthly token limit
		if plan.Config.Chat.MaxMonthlyTokens > 0 {
			used, uerr := db.GetMonthlyTokenUsage(r.Context(), h.DB, cbot.UserID)
			if uerr == nil && used >= plan.Config.Chat.MaxMonthlyTokens {
				base := api.BaseLang(cbot.LanguageCode)
				cfg := api.ConfigFromBase(base)
				api.WriteLocalizedError(w, http.StatusPaymentRequired, api.ErrMonthlyTokensExceeded, cfg)
				return
			}
		}
		// Enforce allowed model if set
		if len(plan.Config.Chat.AllowedModels) > 0 {
			allowed := slices.Contains(plan.Config.Chat.AllowedModels, cbot.Model)
			if !allowed {
				cbot.Model = plan.Config.Chat.AllowedModels[0]
			}
		}
	}

	var req publicChatRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" || strings.TrimSpace(req.SessionID) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create context with timeout
	to := chatTimeout()
	ctx, cancel := context.WithTimeout(r.Context(), to)
	defer cancel()

	// Delegate to chat service
	chatReq := models.ChatRequest{
		Message:   req.Message,
		SessionID: req.SessionID,
	}
	result, err := h.ChatService.ProcessChat(ctx, chatReq, cbot, ragConfig)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Convert sources
	var sources []publicSourceUsed
	for _, s := range result.Sources {
		sources = append(sources, publicSourceUsed{ChunkIndex: s.ChunkIndex, SourceType: s.SourceType})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(publicChatResponse{
		Response:       result.Response,
		MessageID:      result.MessageID,
		TokensUsed:     result.TokensUsed,
		SourcesUsed:    sources,
		ConfidenceTier: result.ConfidenceTier,
	}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// PublicChatFunc returns a http.HandlerFunc for backwards compatibility
//
// Deprecated: Use PublicHandlers.PublicChat instead
func PublicChat(dbpool *sql.DB) http.HandlerFunc {
	// Create a ChatService for backwards compatibility
	// Note: This is a transitional pattern; in production, ChatService should be properly initialized
	svc := services.NewChatService(dbpool, nil, nil, nil, nil)
	h := &PublicHandlers{DB: dbpool, ChatService: svc}
	return h.PublicChat
}

// SubmitFeedback handles POST /api/v1/public/chatbots/:id/feedback
func (h *PublicHandlers) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract bot ID from path
	// Path: /api/v1/public/chatbots/:id/feedback
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 { // ["", "api", "v1", "public", "chatbots", "{id}", "feedback"]
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(parts) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	botID := parts[5]

	var req struct {
		MessageID string `json:"message_id"`
		ThumbsUp  bool   `json:"thumbs_up"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify message belongs to bot (optional but good)
	// For now, simpler: UpdateMessageFeedback checks ID.
	// We need to return the chatbotID for analytics increment, but UpdateMessageFeedback returns it.
	chatbotID, oldThumbsUp, err := db.UpdateMessageFeedback(r.Context(), h.DB, req.MessageID, req.ThumbsUp)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Verify the message belongs to the bot in the path
	if chatbotID != botID {
		w.WriteHeader(http.StatusBadRequest) // mismatch
		return
	}

	// Update analytics asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.IncrementFeedback(bgCtx, h.DB, chatbotID, time.Now(), oldThumbsUp, req.ThumbsUp)
	}()

	w.WriteHeader(http.StatusOK)
}
