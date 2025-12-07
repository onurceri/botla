package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/logger"
)

// ChatService handles core chat logic, shared between authenticated and public endpoints
type ChatService struct {
	DB  *sql.DB
	OAI *rag.OpenAIClient
	QC  *rag.QdrantClient
	Log *logger.Logger
}

// ChatRequest contains the input for a chat interaction
type ChatRequest struct {
	Message   string
	SessionID string
	BotID     string
	UserID    *string // nil for public/anonymous
}

// SourceUsed represents a source that contributed to the response
type SourceUsed struct {
	ChunkIndex int    `json:"chunk_index"`
	SourceType string `json:"source_type"`
}

// ChatResult contains the output of a chat interaction
type ChatResult struct {
	Response       string
	TokensUsed     int
	Sources        []SourceUsed
	ConversationID string
	IsNewConv      bool
}

// NewChatService creates a new ChatService instance
func NewChatService(db *sql.DB, oai *rag.OpenAIClient, qc *rag.QdrantClient, log *logger.Logger) *ChatService {
	return &ChatService{
		DB:  db,
		OAI: oai,
		QC:  qc,
		Log: log,
	}
}

// ProcessChat handles the complete chat flow: embedding, context retrieval, completion, and storage
func (s *ChatService) ProcessChat(ctx context.Context, req ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig) (*ChatResult, error) {
	// Lazy initialization of clients if not provided
	oai := s.OAI
	if oai == nil {
		var err error
		oai, err = rag.NewOpenAIClientFromEnv()
		if err != nil {
			return nil, err
		}
	}
	qc := s.QC
	if qc == nil {
		qc, _ = rag.NewQdrantClientFromEnv() // Proceed without context if Qdrant is unavailable
	}

	// Get or create conversation
	conv, err := db.GetOrCreateConversationBySessionID(ctx, s.DB, bot.ID, req.SessionID)
	if err != nil || conv == nil {
		return nil, err
	}

	isNewConv := conv.MessageCount == 0

	// Save user message
	um := &models.Message{ConversationID: conv.ID, Role: "user", Content: req.Message, TokensUsed: 0}
	if _, err = db.CreateMessage(ctx, s.DB, um); err != nil {
		return nil, err
	}
	_ = db.IncrementConversationMessageCount(ctx, s.DB, conv.ID)

	// Create embedding and search for context
	embedding, err := oai.CreateEmbedding(ctx, req.Message)
	var contextText string
	var sources []SourceUsed

	if err == nil && qc != nil {
		ctxText, metas, _ := rag.SearchContext(embedding, bot.ID, ragConfig.TopK, ragConfig.MaxContextTokens)
		contextText = ctxText
		for _, m := range metas {
			sources = append(sources, SourceUsed{ChunkIndex: m.ChunkIndex, SourceType: m.SourceType})
		}
	}

	// Get language config
	langCode := normalizeLangCode(bot.LanguageCode)
	cfg := langconfig.Get(langCode)

	// Generate response
	var ans string
	var tokens int

	if strings.TrimSpace(contextText) == "" {
		ans = cfg.ResponseTemplates.NoInfoFound
		tokens = 0
	} else {
		sp := resolveSystemPrompt(bot.SystemPrompt, cfg)
		ans, tokens, err = oai.CreateCompletion(ctx, sp, contextText, req.Message, bot.Model, bot.Temperature, bot.MaxTokens)
		if err != nil {
			ans = cfg.ResponseTemplates.ErrorMessage
			tokens = 0
		}
	}

	// Save assistant message
	am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: ans, TokensUsed: tokens}
	if _, err = db.CreateMessage(ctx, s.DB, am); err == nil {
		_ = db.IncrementConversationMessageCount(ctx, s.DB, conv.ID)
	}

	// Update analytics asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.IncrementAnalytics(bgCtx, s.DB, bot.ID, time.Now(), isNewConv, tokens); err != nil && s.Log != nil {
			s.Log.Warn("analytics_error", map[string]any{"chatbot_id": bot.ID, "error": err.Error()})
		}
	}()

	return &ChatResult{
		Response:       ans,
		TokensUsed:     tokens,
		Sources:        sources,
		ConversationID: conv.ID,
		IsNewConv:      isNewConv,
	}, nil
}

// normalizeLangCode extracts the language code prefix (e.g., "tr" from "tr-TR")
func normalizeLangCode(code string) string {
	s := strings.TrimSpace(code)
	if s == "" {
		return "tr"
	}
	if i := strings.Index(s, "-"); i > 0 {
		s = s[:i]
	}
	return s
}

// resolveSystemPrompt returns the custom prompt or falls back to the language default
func resolveSystemPrompt(customPrompt string, cfg langconfig.LanguageConfig) string {
	if strings.TrimSpace(customPrompt) == "" {
		return cfg.ResponseTemplates.DefaultSystemPrompt
	}
	return customPrompt
}
