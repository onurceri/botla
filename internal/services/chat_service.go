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
	"github.com/onurceri/botla-co/pkg/config"
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
// The factory must be provided with config - use rag.NewClientFactory(cfg)
func NewChatService(db *sql.DB, factory *rag.ClientFactory, embedder rag.EmbeddingClient, qc *rag.QdrantClient, log *logger.Logger) *ChatService {
	if embedder == nil {
		if client, err := rag.NewOpenAIClientFromEnv(); err == nil {
			embedder = client
		}
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
	// Use tool flow if actions exist OR handoff is enabled (for built-in handoff tool)
	if (err == nil && len(actions) > 0) || bot.HandoffEnabled {
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

	// Prepare system prompt with language enforcement
	langCode := normalizeLangCode(bot.LanguageCode)
	cfg := langconfig.Get(langCode)
	systemPrompt := buildSystemPrompt(bot, cfg)

	// Load conversation history for context
	historyLimit := calculateHistoryLimit(ragConfig.MaxContextTokens)
	historyMsgs, _ := db.ListRecentMessages(ctx, s.DB, conv.ID, historyLimit)

	// Build messages array with system prompt first
	messages := []rag.ChatMessage{
		{Role: "system", Content: &systemPrompt},
	}

	// Add conversation history (excluding the current message we just saved)
	for _, m := range historyMsgs {
		// Skip the current user message (already saved above, will be added with context)
		if m.Content == req.Message && m.Role == "user" {
			continue
		}
		content := m.Content
		messages = append(messages, rag.ChatMessage{Role: m.Role, Content: &content})
	}

	// Add current user message with RAG context
	userMsgContent := "Context:\n" + contextText + "\n\nQuestion:\n" + req.Message
	messages = append(messages, rag.ChatMessage{Role: "user", Content: &userMsgContent})

	// Tools
	tools := rag.ConvertActionsToTools(actions)

	includeHandoffTool := bot.HandoffEnabled
	if includeHandoffTool {
		plan, planErr := db.GetPlanByUserID(ctx, s.DB, bot.UserID)
		if planErr == nil && plan != nil && !plan.Config.Guardrails.CanUseEscalateFallback {
			includeHandoffTool = false
		}
	}

	tools = append(tools, rag.GetBuiltinTools(includeHandoffTool)...)

	// Get LLM Client with tool support
	// Try the bot's configured model first via the factory
	useOpenAIOnly := !s.Factory.IsProviderConfigured("openrouter")
	modelString := bot.Model
	if useOpenAIOnly && !strings.HasPrefix(modelString, "openai:") {
		modelString = "openai:" + modelString
	}
	client, modelName, err := s.Factory.GetClientForModel(modelString)
	if err != nil {
		// Fallback: try OpenRouter first (preferred for tool support with any model)
		orClient, orErr := s.Factory.GetClient("openrouter")
		if orErr == nil && orClient != nil {
			client = orClient
			modelName = config.ModelOpenRouterGPT4oMini // Default model for OpenRouter
		} else {
			// Last resort: try OpenAI directly
			oaiClient, oaiErr := rag.NewOpenAIClientFromEnv()
			if oaiErr == nil && oaiClient != nil {
				client = oaiClient
				modelName = config.ModelGPT4oMini // OpenAI model format (no prefix)
			}
		}
		if client == nil {
			return nil, fmt.Errorf("no LLM client available: %w", err)
		}
	}

	toolsClient, ok := client.(rag.ToolsLLMClient)
	if !ok {
		// Client doesn't support tools - try OpenRouter which now supports ToolsLLMClient
		orClient, orErr := s.Factory.GetClient("openrouter")
		if orErr == nil {
			if tc, tcOk := orClient.(rag.ToolsLLMClient); tcOk {
				toolsClient = tc
				// Use original model name if it looks like an OpenRouter model format
				if !strings.Contains(modelName, "/") {
					modelName = config.ModelOpenRouterGPT4oMini // Default to a known working model
				}
			}
		}
		// Still no tools client? Fall back to OpenAI
		if toolsClient == nil {
			c, err := rag.NewOpenAIClientFromEnv()
			if err != nil {
				return nil, fmt.Errorf("tool support requires OpenAI or OpenRouter client: %w", err)
			}
			toolsClient = c
			modelName = config.ModelGPT4oMini
		}
	}

	if toolsClient == nil {
		return nil, fmt.Errorf("tools client unavailable")
	}

	executor := &rag.ToolExecutor{DB: s.DB, Log: s.Log}
	var finalResponse string
	var totalTokens int
	var isHandoff bool
	var handoffRequestID string

	// Agentic loop
AgentLoop:
	for i := 0; i < 5; i++ {
		response, err := toolsClient.CreateCompletionWithTools(ctx, messages, tools, modelName, bot.Temperature, bot.MaxTokens)
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
			// Pass IDs to Executor (needed for handoff)
			result, err := executor.Execute(ctx, tc, action, bot.ID, conv.ID)
			content := ""
			if err != nil {
				content = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			} else {
				content = result.Result
			}

			// Special handling for handoff tool success
			if tc.Function.Name == "request_human_handoff" && err == nil {
				// Parse request_id from tool result
				// Result format: {"status": "handoff_requested", "request_id": "uuid"}
				if strings.Contains(result.Result, "request_id") {
					// Simple extraction - find the request_id value
					start := strings.Index(result.Result, `"request_id": "`)
					if start != -1 {
						start += len(`"request_id": "`)
						end := strings.Index(result.Result[start:], `"`)
						if end != -1 {
							handoffRequestID = result.Result[start : start+end]
						}
					}
				}

				// Use the configured handoff message
				msg := cfg.ResponseTemplates.HandoffSuggestion // Default
				if bot.FallbackMessages != nil && bot.FallbackMessages.HandoffMessage != "" {
					msg = bot.FallbackMessages.HandoffMessage
				}
				finalResponse = msg
				isHandoff = true
				break AgentLoop // Stop agent loop, return handoff message immediately
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
	msgType := "normal"
	if isHandoff {
		msgType = "handoff"
	}
	am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: finalResponse, TokensUsed: totalTokens, Type: msgType}
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
		Response:         finalResponse,
		TokensUsed:       totalTokens,
		Sources:          sources,
		ConversationID:   conv.ID,
		MessageID:        amID,
		IsNewConv:        isNewConv,
		HandoffRequestID: handoffRequestID,
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
// Deprecated: Use buildSystemPrompt for new code
func resolveSystemPrompt(customPrompt string, cfg langconfig.LanguageConfig) string {
	if strings.TrimSpace(customPrompt) == "" {
		return cfg.ResponseTemplates.DefaultSystemPrompt
	}
	return customPrompt
}

// buildSystemPrompt creates a complete system prompt with base rules, custom instructions, and language enforcement
func buildSystemPrompt(bot *models.Chatbot, cfg langconfig.LanguageConfig) string {
	// Start with base system prompt containing core rules
	base := cfg.ResponseTemplates.DefaultSystemPrompt

	// Add custom instructions if provided
	if strings.TrimSpace(bot.CustomInstruction) != "" {
		base = base + "\n\n### Ek Talimatlar:\n" + bot.CustomInstruction
	}

	// Append language enforcement directive
	if cfg.ResponseTemplates.LanguageDirective != "" {
		base = base + "\n\n" + cfg.ResponseTemplates.LanguageDirective
	}

	return base
}

// calculateHistoryLimit returns optimal message history count based on context budget
func calculateHistoryLimit(maxContextTokens int) int {
	// Reserve ~40% of context for history, assume ~150 tokens per message
	historyBudget := int(float64(maxContextTokens) * 0.4)
	limit := historyBudget / 150
	if limit < 4 {
		limit = 4 // minimum
	}
	if limit > 20 {
		limit = 20 // maximum
	}
	return limit
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
		modelName = config.ModelGPT4oMini
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
		modelName = config.ModelGPT4oMini
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
