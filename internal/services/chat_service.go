package services

import (
	"context"
	"database/sql"
	"fmt"
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
	// Check if we should use tool-enabled flow
	// For now, let's check if there are any enabled actions for this bot
	actions, err := db.GetEnabledActions(ctx, s.DB, bot.ID)
	if err == nil && len(actions) > 0 {
		return s.ProcessChatWithTools(ctx, req, bot, ragConfig, actions)
	}

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
		ctxText, metas, _ := rag.SearchContext(embedding, bot.ID, ragConfig.TopK, ragConfig.MaxContextTokens, bot.ConfidenceThreshold)
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
		if bot.FallbackMessages != nil && bot.FallbackMessages.NoInfoFound != "" {
			ans = bot.FallbackMessages.NoInfoFound
		}
		tokens = 0
	} else {
		sp := resolveSystemPrompt(bot.SystemPrompt, cfg)

		// Get client for the model
		client, modelName, err := s.Factory.GetClientForModel(bot.Model)
		if err != nil {
			c, e := rag.NewOpenAIClientFromEnv()
			if e != nil || c == nil {
				return nil, fmt.Errorf("openai client not configured: %w", e)
			}
			client = c
			modelName = "gpt-4o-mini"
		}

		if client == nil {
			return nil, fmt.Errorf("model client unavailable")
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
			if bot.FallbackMessages != nil && bot.FallbackMessages.ErrorMessage != "" {
				ans = bot.FallbackMessages.ErrorMessage
			}
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

// ProcessChatWithTools handles the chat flow with tool support
func (s *ChatService) ProcessChatWithTools(ctx context.Context, req models.ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig, actions []*models.ChatbotAction) (*models.ChatResult, error) {
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
		qc, _ = rag.NewQdrantClientFromEnv()
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

	// Context retrieval
	embedding, err := embedder.CreateEmbedding(ctx, req.Message)
	var contextText string
	var sources []models.SourceUsed

	if err == nil && qc != nil {
		ctxText, metas, _ := rag.SearchContext(embedding, bot.ID, ragConfig.TopK, ragConfig.MaxContextTokens, bot.ConfidenceThreshold)
		contextText = ctxText
		for _, m := range metas {
			sources = append(sources, models.SourceUsed{ChunkIndex: m.ChunkIndex, SourceType: m.SourceType})
		}
	}

	// Prepare system prompt and messages
	langCode := normalizeLangCode(bot.LanguageCode)
	cfg := langconfig.Get(langCode)
	systemPrompt := resolveSystemPrompt(bot.SystemPrompt, cfg)

	// If no context found, normally we might just return "no info".
	// But with tools, maybe the tools can answer?
	// For now, let's include the context if found.

	userMsgContent := "Context:\n" + contextText + "\n\nQuestion:\n" + req.Message
	messages := []rag.ChatMessage{
		{Role: "system", Content: &systemPrompt},
		{Role: "user", Content: &userMsgContent},
	}

	// Tools
	tools := rag.ConvertActionsToTools(actions)
	tools = append(tools, rag.GetBuiltinTools()...)

	// Get OpenAI Client
	// Currently only OpenAI supports this implementation
	client, modelName, err := s.Factory.GetClientForModel(bot.Model)
	if err != nil {
		client, _ = rag.NewOpenAIClientFromEnv()
		modelName = "gpt-4o-mini"
	}

	openaiClient, ok := client.(*rag.OpenAIClient)
	if !ok {
		// Fallback to standard OpenAI client if the factory returned something else but we need OpenAI features
		// Or just return error if other providers don't support tools yet
		var err error
		openaiClient, err = rag.NewOpenAIClientFromEnv()
		if err != nil {
			return nil, fmt.Errorf("tool support requires OpenAI client: %w", err)
		}
		modelName = "gpt-4o-mini" // Force OpenAI model
	}

	if openaiClient == nil {
		return nil, fmt.Errorf("openai client unavailable")
	}

	executor := &rag.ToolExecutor{DB: s.DB, Log: s.Log}
	var finalResponse string
	var totalTokens int

	// Agentic loop
	for i := 0; i < 5; i++ {
		response, err := openaiClient.CreateCompletionWithTools(ctx, messages, tools, modelName, bot.Temperature, bot.MaxTokens)
		if err != nil {
			if s.Log != nil {
				s.Log.Error("completion_with_tools_error", map[string]any{"error": err.Error()})
			}
			return nil, err
		}

		totalTokens += response.Usage.TotalTokens
		choice := response.Choices[0]

		// Add assistant message to history
		messages = append(messages, choice.Message)

		if len(choice.Message.ToolCalls) == 0 {
			if choice.Message.Content != nil {
				finalResponse = *choice.Message.Content
			}
			break
		}

		// Execute tools
		for _, tc := range choice.Message.ToolCalls {
			action := findActionByName(actions, tc.Function.Name)
			result, err := executor.Execute(ctx, tc, action)
			content := ""
			if err != nil {
				content = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			} else {
				content = result.Result
			}

			messages = append(messages, rag.ChatMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    &content,
			})
		}
	}

	if finalResponse == "" {
		finalResponse = "İşlem tamamlanamadı veya çok uzun sürdü."
	}

	// Save assistant message
	am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: finalResponse, TokensUsed: totalTokens}
	if _, err = db.CreateMessage(ctx, s.DB, am); err == nil {
		_ = db.IncrementConversationMessageCount(ctx, s.DB, conv.ID)
	}

	// Update analytics asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.IncrementAnalytics(bgCtx, s.DB, bot.ID, time.Now(), isNewConv, totalTokens); err != nil && s.Log != nil {
			s.Log.Warn("analytics_error", map[string]any{"chatbot_id": bot.ID, "error": err.Error()})
		}
	}()

	return &models.ChatResult{
		Response:       finalResponse,
		TokensUsed:     totalTokens,
		Sources:        sources,
		ConversationID: conv.ID,
		IsNewConv:      isNewConv,
	}, nil
}

func findActionByName(actions []*models.ChatbotAction, name string) *models.ChatbotAction {
	for _, a := range actions {
		if a.Name == name {
			return a
		}
	}
	return nil
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
