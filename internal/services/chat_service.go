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
	DB       *sql.DB
	Factory  *rag.ClientFactory
	Embedder rag.EmbeddingClient
	QC       *rag.QdrantClient
	Log      *logger.Logger
}

// ...

// NewChatService creates a new ChatService instance
func NewChatService(db *sql.DB, factory *rag.ClientFactory, embedder rag.EmbeddingClient, qc *rag.QdrantClient, log *logger.Logger) *ChatService {
	if factory == nil {
		factory = rag.NewClientFactory()
	}
	if embedder == nil {
		embedder, _ = rag.NewOpenAIClientFromEnv()
	}
	return &ChatService{
		DB:       db,
		Factory:  factory,
		Embedder: embedder,
		QC:       qc,
		Log:      log,
	}
}

// ProcessChat handles the complete chat flow: embedding, context retrieval, completion, and storage
func (s *ChatService) ProcessChat(ctx context.Context, req models.ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig) (*models.ChatResult, error) {
	// Ensure embedder is available
	embedder := s.Embedder
	if embedder == nil {
		var err error
		embedder, err = rag.NewOpenAIClientFromEnv()
		if err != nil {
			return nil, err
		}
	}

	// Lazy initialization of Qdrant
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
	embedding, err := embedder.CreateEmbedding(ctx, req.Message)
	var contextText string
	var sources []models.SourceUsed

	if err == nil && qc != nil {
		ctxText, metas, _ := rag.SearchContext(embedding, bot.ID, ragConfig.TopK, ragConfig.MaxContextTokens)
		contextText = ctxText
		for _, m := range metas {
			sources = append(sources, models.SourceUsed{ChunkIndex: m.ChunkIndex, SourceType: m.SourceType})
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

		// Get client for the model
		client, modelName, err := s.Factory.GetClientForModel(bot.Model)
		if err != nil {
			// Fallback to OpenAI if factory fails
			client, _ = rag.NewOpenAIClientFromEnv()
			modelName = "gpt-4o-mini"
		}

		params := models.CompletionParams{
			SystemPrompt: sp,
			Context:      contextText,
			UserMessage:  req.Message,
			Model:        modelName,
			Temperature:  bot.Temperature,
			MaxTokens:    bot.MaxTokens,
		}

		res, err := client.CreateCompletion(ctx, params)
		if err != nil {
			ans = cfg.ResponseTemplates.ErrorMessage
			tokens = 0
			if s.Log != nil {
				s.Log.Error("completion_error", map[string]any{"error": err.Error(), "model": bot.Model})
			}
		} else {
			ans = res.Content
			tokens = res.UsageTokens
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

	return &models.ChatResult{
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
