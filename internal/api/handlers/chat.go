package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/internal/workers"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type ChatHandlers struct {
	DB               *sql.DB
	ChatService      *services.ChatService
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
	WorkerPool       *workers.WorkerPool
	Logger           *logger.Logger
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

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// Delegate all business logic to service
	result, err := h.ChatService.ProcessChatWithValidation(ctx, services.ChatRequestWithUser{
		UserID:      userID,
		Chatbot:     cbot,
		ChatRequest: models.ChatRequest{Message: req.Message, SessionID: req.SessionID},
	})

	if err != nil {
		if errors.Is(err, services.ErrTokenQuotaExceeded) {
			base := api.BaseLang(cbot.LanguageCode)
			cfg := api.ConfigFromBase(base)
			api.WriteLocalizedError(w, http.StatusPaymentRequired, api.ErrMonthlyTokensExceeded, cfg)
			return
		}

		h.Logger.Error("chat_processing_failed", map[string]any{
			"error":      err.Error(),
			"chatbot_id": cbot.ID,
			"user_id":    userID,
		})
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
	h.WorkerPool.Submit(func(ctx context.Context) {
		if err := db.IncrementFeedback(ctx, h.DB, chatbotID, time.Now(), oldThumbsUp, req.ThumbsUp); err != nil {
			h.Logger.Error("feedback_increment_failed", map[string]any{
				"chatbot_id": chatbotID,
				"error":      err.Error(),
			})
		}
	})

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
