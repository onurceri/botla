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
	startTime := time.Now()
	// Check if we should use tool-enabled flow
	// For now, let's check if there are any enabled actions for this bot
	actions, err := db.GetEnabledActions(ctx, s.DB, bot.ID)
	if err == nil && len(actions) > 0 {
		return s.ProcessChatWithTools(ctx, req, bot, ragConfig, actions)
	}

	// Ensure embedder is available
	embedder := s.Embedder
	if embedder == nil {
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

	// Get language config
	langCode := normalizeLangCode(bot.LanguageCode)
	cfg := langconfig.Get(langCode)

	// Create embedding and perform tiered search
	embedding, err := embedder.CreateEmbedding(ctx, req.Message)
	var searchResult *rag.TieredSearchResult
	var sources []models.SourceUsed
	var chunkMetas []models.ChunkMetadata

	if err == nil && qc != nil {
		// Use tiered search with ThresholdConfig
		thresholdCfg := bot.ThresholdConfig
		if thresholdCfg == nil {
			thresholdCfg = models.DefaultThresholdConfig()
		}
		searchResult, err = rag.SearchContextTiered(embedding, bot.ID, ragConfig.TopK, ragConfig.MaxContextTokens, thresholdCfg)
		if err != nil && s.Log != nil {
			s.Log.Warn("tiered_search_error", map[string]any{"error": err.Error(), "chatbot_id": bot.ID})
		}
		if searchResult != nil {
			chunkMetas = searchResult.AllChunks
			for _, m := range searchResult.AllChunks {
				sources = append(sources, models.SourceUsed{ChunkIndex: m.ChunkIndex, SourceType: m.SourceType})
			}
		}
	}

	// Handle nil searchResult
	if searchResult == nil {
		searchResult = &rag.TieredSearchResult{Tier: rag.TierLow}
	}

	// Generate response based on tier
	var ans string
	var tokens int
	var confidenceTier string = string(searchResult.Tier)

	switch searchResult.Tier {
	case rag.TierHigh:
		// 🟢 Strong match - normal RAG flow
		ans, tokens, err = s.generateWithContext(ctx, bot, searchResult.ContextText, req.Message, cfg)
		if err != nil {
			ans = cfg.ResponseTemplates.ErrorMessage
			if bot.FallbackMessages != nil && bot.FallbackMessages.ErrorMessage != "" {
				ans = bot.FallbackMessages.ErrorMessage
			}
		}

	case rag.TierMedium:
		// 🟡 Weak match - RAG with confidence warning
		ans, tokens, err = s.generateWithContext(ctx, bot, searchResult.ContextText, req.Message, cfg)
		if err != nil {
			ans = cfg.ResponseTemplates.ErrorMessage
			if bot.FallbackMessages != nil && bot.FallbackMessages.ErrorMessage != "" {
				ans = bot.FallbackMessages.ErrorMessage
			}
		} else {
			// Add confidence warning if enabled
			thresholdCfg := bot.ThresholdConfig
			if thresholdCfg == nil {
				thresholdCfg = models.DefaultThresholdConfig()
			}
			if thresholdCfg.ShowConfidenceWarning {
				ans = ans + cfg.ResponseTemplates.ConfidenceWarning
			}
		}

	case rag.TierLow:
		// 🔴 No match - fallback mode
		thresholdCfg := bot.ThresholdConfig
		if thresholdCfg == nil {
			thresholdCfg = models.DefaultThresholdConfig()
		}

		switch thresholdCfg.FallbackMode {
		case "smart":
			// Use LLM with smart fallback prompt
			ans, tokens, err = s.smartFallback(ctx, bot, req.Message, cfg)
			if err != nil {
				ans = cfg.ResponseTemplates.NoInfoFound
				if bot.FallbackMessages != nil && bot.FallbackMessages.NoInfoFound != "" {
					ans = bot.FallbackMessages.NoInfoFound
				}
			}
		case "escalate":
			// Suggest human handoff
			ans = cfg.ResponseTemplates.HandoffSuggestion
			tokens = 0
		default: // "static"
			ans = cfg.ResponseTemplates.NoInfoFound
			if bot.FallbackMessages != nil && bot.FallbackMessages.NoInfoFound != "" {
				ans = bot.FallbackMessages.NoInfoFound
			}
			tokens = 0
		}
	}

	// Save assistant message
	am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: ans, TokensUsed: tokens}
	var amID string
	if id, err := db.CreateMessage(ctx, s.DB, am); err == nil {
		amID = id
		_ = db.IncrementConversationMessageCount(ctx, s.DB, conv.ID)

		// Save source usage
		if len(chunkMetas) > 0 {
			if err := db.SaveMessageSources(ctx, s.DB, amID, chunkMetas); err != nil && s.Log != nil {
				s.Log.Warn("save_message_sources_error", map[string]any{"message_id": amID, "error": err.Error()})
			}
		}
	}

	// Update analytics asynchronously
	isUnanswered := searchResult.Tier == rag.TierLow
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		responseTime := int(time.Since(startTime).Milliseconds())
		if err := db.IncrementAnalytics(bgCtx, s.DB, bot.ID, time.Now(), isNewConv, tokens, false, responseTime); err != nil && s.Log != nil {
			s.Log.Warn("analytics_error", map[string]any{"chatbot_id": bot.ID, "error": err.Error()})
		}

		if isUnanswered {
			_ = db.TrackUnansweredQuery(bgCtx, s.DB, bot.ID, req.Message)
		}
	}()

	return &models.ChatResult{
		Response:       ans,
		TokensUsed:     tokens,
		Sources:        sources,
		ConversationID: conv.ID,
		MessageID:      amID,
		IsNewConv:      isNewConv,
		ConfidenceTier: confidenceTier,
	}, nil
}

// ProcessChatWithTools handles the chat flow with tool support
func (s *ChatService) ProcessChatWithTools(ctx context.Context, req models.ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig, actions []*models.ChatbotAction) (*models.ChatResult, error) {
	startTime := time.Now()
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
	var chunkMetas []models.ChunkMetadata

	if err == nil && qc != nil {
		ctxText, metas, _ := rag.SearchContext(embedding, bot.ID, ragConfig.TopK, ragConfig.MaxContextTokens, bot.ConfidenceThreshold)
		contextText = ctxText
		chunkMetas = metas
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
		fr := cfg.ResponseTemplates.Errors["CHAT_TIMEOUT_OR_INCOMPLETE"]
		if fr == "" {
			fr = cfg.ResponseTemplates.ErrorMessage
		}
		finalResponse = fr
	}

	// Save assistant message
	am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: finalResponse, TokensUsed: totalTokens}
	var amID string
	if id, err := db.CreateMessage(ctx, s.DB, am); err == nil {
		amID = id
		_ = db.IncrementConversationMessageCount(ctx, s.DB, conv.ID)

		// Save source usage
		if len(chunkMetas) > 0 {
			if err := db.SaveMessageSources(ctx, s.DB, amID, chunkMetas); err != nil && s.Log != nil {
				s.Log.Warn("save_message_sources_error", map[string]any{"message_id": amID, "error": err.Error()})
			}
		}
	}

	// Update analytics asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		responseTime := int(time.Since(startTime).Milliseconds())
		if err := db.IncrementAnalytics(bgCtx, s.DB, bot.ID, time.Now(), isNewConv, totalTokens, false, responseTime); err != nil && s.Log != nil {
			s.Log.Warn("analytics_error", map[string]any{"chatbot_id": bot.ID, "error": err.Error()})
		}
	}()

	return &models.ChatResult{
		Response:       finalResponse,
		TokensUsed:     totalTokens,
		Sources:        sources,
		ConversationID: conv.ID,
		MessageID:      amID,
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

// generateWithContext generates a response using LLM with the provided context
func (s *ChatService) generateWithContext(ctx context.Context, bot *models.Chatbot, contextText string, userMessage string, cfg langconfig.LanguageConfig) (string, int, error) {
	sp := resolveSystemPrompt(bot.SystemPrompt, cfg)

	// Get client for the model
	client, modelName, err := s.Factory.GetClientForModel(bot.Model)
	if err != nil {
		c, e := rag.NewOpenAIClientFromEnv()
		if e != nil || c == nil {
			return "", 0, fmt.Errorf("openai client not configured: %w", e)
		}
		client = c
		modelName = "gpt-4o-mini"
	}

	if client == nil {
		return "", 0, fmt.Errorf("model client unavailable")
	}

	// Add localized context intro prefix
	formattedContext := contextText
	if strings.TrimSpace(contextText) != "" {
		formattedContext = cfg.ResponseTemplates.RAGContextIntro + contextText
	}

	// Prepare completion params
	params := models.CompletionParams{
		SystemPrompt: sp,
		Context:      formattedContext,
		UserMessage:  userMessage,
		Model:        modelName,
		Temperature:  bot.Temperature,
		MaxTokens:    bot.MaxTokens,
	}

	res, err := client.CreateCompletion(ctx, params)
	if err != nil {
		if s.Log != nil {
			s.Log.Error("completion_error", map[string]any{"error": err.Error(), "model": bot.Model})
		}
		return "", 0, err
	}

	return res.Content, res.UsageTokens, nil
}

// smartFallback generates a helpful response when no context is available
func (s *ChatService) smartFallback(ctx context.Context, bot *models.Chatbot, userMessage string, cfg langconfig.LanguageConfig) (string, int, error) {
	// Get capability summaries from sources
	capabilities := s.getCapabilitySummaries(ctx, bot.ID)

	// Build the smart fallback prompt
	capabilityText := ""
	if capabilities != "" {
		capabilityText = cfg.ResponseTemplates.CapabilityIntro + "\n" + capabilities
	}
	systemPrompt := fmt.Sprintf(cfg.ResponseTemplates.SmartFallbackPrompt, capabilityText)

	// Get client for the model
	client, modelName, err := s.Factory.GetClientForModel(bot.Model)
	if err != nil {
		c, e := rag.NewOpenAIClientFromEnv()
		if e != nil || c == nil {
			return "", 0, fmt.Errorf("openai client not configured: %w", e)
		}
		client = c
		modelName = "gpt-4o-mini"
	}

	if client == nil {
		return "", 0, fmt.Errorf("model client unavailable")
	}

	// Use lower temperature for more controlled response
	params := models.CompletionParams{
		SystemPrompt: systemPrompt,
		Context:      "",
		UserMessage:  userMessage,
		Model:        modelName,
		Temperature:  0.3, // Lower temperature for controlled fallback
		MaxTokens:    200, // Short response for fallback
	}

	res, err := client.CreateCompletion(ctx, params)
	if err != nil {
		if s.Log != nil {
			s.Log.Error("smart_fallback_error", map[string]any{"error": err.Error(), "model": bot.Model})
		}
		return "", 0, err
	}

	return res.Content, res.UsageTokens, nil
}

// getCapabilitySummaries retrieves capability summaries from data sources
func (s *ChatService) getCapabilitySummaries(ctx context.Context, chatbotID string) string {
	sources, err := db.ListSourcesByChatbotID(ctx, s.DB, chatbotID)
	if err != nil || len(sources) == 0 {
		return ""
	}

	var summaries []string
	for _, src := range sources {
		if src.CapabilitySummary != nil && *src.CapabilitySummary != "" {
			summaries = append(summaries, "- "+*src.CapabilitySummary)
		}
	}

	if len(summaries) == 0 {
		return ""
	}

	// Limit to first 5 to avoid too long prompt
	if len(summaries) > 5 {
		summaries = summaries[:5]
	}

	return strings.Join(summaries, "\n")
}

