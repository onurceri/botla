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
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/pkg/langconfig"
)

type publicChatbot struct {
	ID                   string   `json:"id"`
	ThemeColor           string   `json:"theme_color"`
	WelcomeMessage       string   `json:"welcome_message"`
	Position             string   `json:"position"`
	BotMessageColor      string   `json:"bot_message_color"`
	UserMessageColor     string   `json:"user_message_color"`
	BotMessageTextColor  string   `json:"bot_message_text_color"`
	UserMessageTextColor string   `json:"user_message_text_color"`
	ChatFontFamily       string   `json:"chat_font_family"`
	ChatHeaderColor      string   `json:"chat_header_color"`
	ChatHeaderTextColor  string   `json:"chat_header_text_color"`
	ChatBackgroundColor  string   `json:"chat_background_color"`
	BotIcon              *string  `json:"bot_icon,omitempty"`
	BotDisplayName       *string  `json:"bot_display_name,omitempty"`
	SuggestedQuestions   []string `json:"suggested_questions,omitempty"`
}

func PublicChatbotConfig(dbpool *sql.DB) http.HandlerFunc {
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
		key := "public:chatbot:" + c.ID + ":suggestions:" + c.UpdatedAt.UTC().Format(time.RFC3339Nano)
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
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(out)
	}
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
	Response    string             `json:"response"`
	TokensUsed  int                `json:"tokens_used"`
	SourcesUsed []publicSourceUsed `json:"sources_used"`
}

func PublicChat(dbpool *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		cbot, err := db.GetChatbotByID(r.Context(), dbpool, botID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if cbot == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var req publicChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		req.Message = strings.TrimSpace(req.Message)
		if req.Message == "" || strings.TrimSpace(req.SessionID) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conv, err := db.GetOrCreateConversationBySessionID(r.Context(), dbpool, cbot.ID, req.SessionID)
		if err != nil || conv == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		um := &models.Message{ConversationID: conv.ID, Role: "user", Content: req.Message, TokensUsed: 0}
		if _, err := db.CreateMessage(r.Context(), dbpool, um); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = db.IncrementConversationMessageCount(r.Context(), dbpool, conv.ID)

		oai, err := rag.NewOpenAIClientFromEnv()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		qc, err := rag.NewQdrantClientFromEnv()
		if err != nil {
			qc = nil
		}

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
		var sources []publicSourceUsed
		if err == nil && qc != nil {
			ctxText, metas, _ := rag.SearchContext(embedding, cbot.ID)
			contextText = ctxText
			for _, m := range metas {
				sources = append(sources, publicSourceUsed{ChunkIndex: m.ChunkIndex, SourceType: m.SourceType})
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
		if _, err := db.CreateMessage(r.Context(), dbpool, am); err == nil {
			_ = db.IncrementConversationMessageCount(r.Context(), dbpool, conv.ID)
		}

		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = db.IncrementAnalytics(bgCtx, dbpool, cbot.ID, time.Now(), conv.MessageCount == 0, tokens)
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(publicChatResponse{Response: ans, TokensUsed: tokens, SourcesUsed: sources})
	}
}
