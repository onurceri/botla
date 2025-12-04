package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type ChatHandlers struct {
	DB *sql.DB
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
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	const prefix = "/api/v1/chatbots/"
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
	if cbot.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}
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

	conv, err := db.GetOrCreateConversationBySessionID(r.Context(), h.DB, cbot.ID, req.SessionID)
	if err != nil || conv == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Save user message
	um := &models.Message{ConversationID: conv.ID, Role: "user", Content: req.Message, TokensUsed: 0}
	if _, err := db.CreateMessage(r.Context(), h.DB, um); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = db.IncrementConversationMessageCount(r.Context(), h.DB, conv.ID)

	// Clients
	oai, err := rag.NewOpenAIClientFromEnv()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	qc, err := rag.NewQdrantClientFromEnv()
	if err != nil {
		// proceed without context
		qc = nil
	}

	// Embedding
	to := 20 * time.Second
	if v := os.Getenv("CHAT_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			to = time.Duration(n) * time.Millisecond
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), to)
	defer cancel()
	embedding, err := oai.CreateEmbedding(ctx, req.Message)
	var contextText string
	var sources []sourceUsed
	if err == nil && qc != nil {
		ctxText, metas, _ := rag.SearchContext(embedding, cbot.ID)
		contextText = ctxText
		for _, m := range metas {
			sources = append(sources, sourceUsed{ChunkIndex: m.ChunkIndex, SourceType: m.SourceType})
		}
	}

	// Completion
	var ans string
	var tokens int

	// Get language config
	langCode := cbot.Language
	if langCode == "" {
		langCode = "tr"
	}
	cfg := langconfig.Get(langCode)

	if strings.TrimSpace(contextText) == "" {
		ans = cfg.ResponseTemplates.NoInfoFound
		tokens = 0
	} else {
		sp := strings.TrimSpace(cbot.SystemPrompt)
		if sp == "" {
			sp = cfg.ResponseTemplates.DefaultSystemPrompt
		}
		ans, tokens, err = oai.CreateCompletion(ctx, sp, contextText, req.Message, cbot.Model, cbot.Temperature, cbot.MaxTokens)
		if err != nil {
			ans = cfg.ResponseTemplates.ErrorMessage
			tokens = 0
		}
	}
	am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: ans, TokensUsed: tokens}
	if _, err := db.CreateMessage(r.Context(), h.DB, am); err == nil {
		_ = db.IncrementConversationMessageCount(r.Context(), h.DB, conv.ID)
	}

	// Update Analytics
	// Check if this was a new conversation (0 messages before this interaction)
	isNew := conv.MessageCount == 0
	// We added 2 messages (User + Assistant)
	// Note: IncrementConversationMessageCount was called twice (once for user, once for assistant)
	// But conv.MessageCount holds the value *before* these increments because we haven't re-fetched it.
	// So checking conv.MessageCount == 0 is correct for "was it new when we started".

	go func() {
		// Use a background context or the request context if it's not cancelled immediately
		// Better to use a detached context to ensure it runs even if request finishes
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.IncrementAnalytics(bgCtx, h.DB, cbot.ID, time.Now(), isNew, tokens)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatResponse{Response: ans, TokensUsed: tokens, SourcesUsed: sources})
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

	chatbotID, err := db.UpdateMessageFeedback(r.Context(), h.DB, msgID, req.ThumbsUp)
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
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.IncrementFeedback(bgCtx, h.DB, chatbotID, time.Now(), req.ThumbsUp)
	}()

	w.WriteHeader(http.StatusOK)
}
