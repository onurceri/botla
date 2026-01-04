package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/scraper"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/httputil"
	"github.com/onurceri/botla-app/pkg/logger"
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
	BubbleRadius         string                 `json:"bubble_radius"`
	InputBackgroundColor string                 `json:"input_background_color"`
	InputTextColor       string                 `json:"input_text_color"`
	SendButtonColor      string                 `json:"send_button_color"`
	BotIcon              *string                `json:"bot_icon,omitempty"`
	BotDisplayName       *string                `json:"bot_display_name,omitempty"`
	SuggestedQuestions   []string               `json:"suggested_questions,omitempty"`
	MaxChars             int                    `json:"max_chars"`
	HideBranding         bool                   `json:"hide_branding"`
	CustomBranding       *models.CustomBranding `json:"custom_branding,omitempty"`
	HandoffEnabled       bool                   `json:"handoff_enabled"`
}

func PublicChatbotConfig(chatbotRepo repository.ChatbotRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const prefix = "/api/v1/public/chatbots/"
		path := r.URL.Path
		if !strings.HasPrefix(path, prefix) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		botID := strings.TrimPrefix(path, prefix)
		if botID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if !httputil.IsValidUUID(botID) {
			api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidIDFormat)
			return
		}

		c, err := chatbotRepo.GetByID(r.Context(), botID)
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
			BubbleRadius:         c.BubbleRadius,
			InputBackgroundColor: c.InputBackgroundColor,
			InputTextColor:       c.InputTextColor,
			SendButtonColor:      c.SendButtonColor,
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
	Response         string             `json:"response"`
	MessageID        string             `json:"message_id"`
	TokensUsed       int                `json:"tokens_used"`
	SourcesUsed      []publicSourceUsed `json:"sources_used"`
	ConfidenceTier   string             `json:"confidence_tier,omitempty"`
	HandoffRequestID string             `json:"handoff_request_id,omitempty"`
}

// PublicHandlers contains handlers for public (unauthenticated) endpoints
type PublicHandlers struct {
	ChatService   *services.ChatService
	Log           *logger.Logger
	ChatbotRepo   repository.ChatbotRepository
	PlanRepo      repository.PlanRepository
	UsageRepo     repository.UsageRepository
	AnalyticsRepo repository.AnalyticsRepository
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
	if !httputil.IsValidUUID(botID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidIDFormat)
		return
	}
	cbot, err := h.ChatbotRepo.GetByID(r.Context(), botID)
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
		// 1. Origin/Domain Check (independent of token)
		if cbot.AllowedDomains != nil && *cbot.AllowedDomains != "" {
			origin := r.Header.Get("Origin")
			// If no origin provided, block if restriction is enabled
			// Browsers always send Origin. If called from non-browser (e.g. curl), it might be empty.
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
				// CR-003: Secure origin validation using proper URL parsing
				// Prevents bypass via origins like "https://example.com.evil.com"
				parsed, pErr := url.Parse(origin)
				if pErr == nil {
					hostname := parsed.Hostname()
					if hostname == d || strings.HasSuffix(hostname, "."+d) {
						allowed = true
						break
					}
				}
			}
			if !allowed {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// 2. Token Check (only if embed_secret is configured)
		if cbot.EmbedSecret != nil && *cbot.EmbedSecret != "" {
			tokenStr := r.Header.Get("X-Embed-Token")
			if tokenStr == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			secret := *cbot.EmbedSecret

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
	}

	// Plan and limits
	plan, err := h.PlanRepo.GetPlanWithLimits(r.Context(), cbot.UserID)
	var ragConfig models.RAGConfig
	var maxMonthlyTokens int
	var estimatedTokens int
	if err == nil && plan != nil {
		ragConfig = models.RAGConfig{
			TopK:             plan.Limits.ChatRAGTopK,
			MaxContextTokens: plan.Limits.ChatRAGMaxContextTokens,
		}
		maxMonthlyTokens = plan.Limits.ChatMaxMonthlyTokens

		// Enforce allowed model if set
		if len(plan.Limits.ChatAllowedModels) > 0 {
			allowed := slices.Contains(plan.Limits.ChatAllowedModels, cbot.Model)
			if !allowed {
				cbot.Model = plan.Limits.ChatAllowedModels[0]
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

	// Atomic token reservation to prevent TOCTOU race condition (Issue #002)
	// Reserve estimated tokens before processing.
	if maxMonthlyTokens > 0 {
		estimatedTokens = cbot.MaxTokens
		if estimatedTokens <= 0 {
			estimatedTokens = 512 // Default
		}
		err = h.UsageRepo.ReserveChatTokens(r.Context(), cbot.UserID, estimatedTokens, maxMonthlyTokens)
		if errors.Is(err, repository.ErrTokenQuotaExceeded) {
			api.WriteErrorCode(w, http.StatusPaymentRequired, api.ErrMonthlyTokensExceeded)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
		// On error, refund the reserved tokens
		if maxMonthlyTokens > 0 && estimatedTokens > 0 {
			_ = h.UsageRepo.AdjustChatTokens(context.Background(), cbot.UserID, -estimatedTokens)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Adjust token count: correct the reservation to actual usage
	if maxMonthlyTokens > 0 && estimatedTokens > 0 {
		delta := result.TokensUsed - estimatedTokens
		if delta != 0 {
			_ = h.UsageRepo.AdjustChatTokens(context.Background(), cbot.UserID, delta)
		}
	}

	// Convert sources
	var sources []publicSourceUsed
	for _, s := range result.Sources {
		sources = append(sources, publicSourceUsed{ChunkIndex: s.ChunkIndex, SourceType: s.SourceType})
	}

	api.WriteJSON(w, http.StatusOK, publicChatResponse{
		Response:         result.Response,
		MessageID:        result.MessageID,
		TokensUsed:       result.TokensUsed,
		SourcesUsed:      sources,
		ConfidenceTier:   result.ConfidenceTier,
		HandoffRequestID: result.HandoffRequestID,
	})
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
	if !httputil.IsValidUUID(botID) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidIDFormat)
		return
	}

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
	chatbotID, oldThumbsUp, err := h.AnalyticsRepo.UpdateMessageFeedback(r.Context(), req.MessageID, req.ThumbsUp)
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
		defer func() {
			if r := recover(); r != nil {
				if h.Log != nil {
					h.Log.Error("public_feedback_analytics_panic", map[string]any{"panic": r})
				} else {
					fmt.Printf("public_feedback_analytics_panic: %v\n", r)
				}
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		oldThumbsUpPtr := oldThumbsUp
		_ = h.AnalyticsRepo.IncrementFeedback(bgCtx, chatbotID, &oldThumbsUpPtr, req.ThumbsUp)
	}()

	w.WriteHeader(http.StatusOK)
}
