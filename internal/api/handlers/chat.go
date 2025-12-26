package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type ChatHandlers struct {
	DB               *sql.DB
	ChatService      *services.ChatService
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
}

type chatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

type sourceUsed struct {
	ChunkIndex int    `json:"chunk_index"`
	SourceType string `json:"source_type"`
}

type chatResponse struct {
	Response    string       `json:"response"`
	TokensUsed  int          `json:"tokens_used"`
	SourcesUsed []sourceUsed `json:"sources_used"`
}

func (h *ChatHandlers) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cbot, _, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	if !ok {
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	// Get plan and check limits (keep in handler for early rejection)
	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var ragConfig models.RAGConfig
	if plan != nil {
		ragConfig = plan.Config.Chat.RAG
		if len(plan.Config.Chat.AllowedModels) > 0 {
			allowed := false
			for _, m := range plan.Config.Chat.AllowedModels {
				if m == cbot.Model {
					allowed = true
					break
				}
			}
			if !allowed {
				cbot.Model = plan.Config.Chat.AllowedModels[0]
			}
		}
		// Check token limits
		if plan.Config.Chat.MaxMonthlyTokens > 0 {
			used, errUsage := db.GetMonthlyTokenUsage(r.Context(), h.DB, userID)
			if errUsage == nil && used >= plan.Config.Chat.MaxMonthlyTokens {
				base := api.BaseLang(cbot.LanguageCode)
				cfg := api.ConfigFromBase(base)
				api.WriteLocalizedError(w, http.StatusPaymentRequired, api.ErrMonthlyTokensExceeded, cfg)
				return
			}
		}
	}

	var req chatRequest
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
	var sources []sourceUsed
	for _, s := range result.Sources {
		sources = append(sources, sourceUsed{ChunkIndex: s.ChunkIndex, SourceType: s.SourceType})
	}

	api.WriteJSON(w, http.StatusOK, chatResponse{Response: result.Response, TokensUsed: result.TokensUsed, SourcesUsed: sources})
}

type feedbackRequest struct {
	ThumbsUp bool `json:"thumbs_up"`
}

func (h *ChatHandlers) FeedbackHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// /api/v1/messages/{id}/feedback
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[len(parts)-1] != "feedback" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	msgID := parts[len(parts)-2]

	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chatbotID, oldThumbsUp, err := db.UpdateMessageFeedback(r.Context(), h.DB, msgID, req.ThumbsUp)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update Analytics
	go func() {
		// CR-002: Recover from panics to prevent server crash
		defer func() {
			if r := recover(); r != nil {
				// Log panic for debugging - use fmt since we don't have logger access
				fmt.Printf("feedback_analytics_panic: %v\n", r)
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.IncrementFeedback(bgCtx, h.DB, chatbotID, time.Now(), oldThumbsUp, req.ThumbsUp)
	}()

	w.WriteHeader(http.StatusOK)
}

func chatTimeout() time.Duration {
	d := 20 * time.Second
	if v := os.Getenv("CHAT_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			d = time.Duration(n) * time.Millisecond
		}
	}
	return d
}
