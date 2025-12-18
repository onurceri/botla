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

// =============================================================================
// CHAT SERVICE - Core chat processing with RAG and tool support
// =============================================================================

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

// =============================================================================
// CHAT CONTEXT - Shared state across pipeline steps
// =============================================================================

// chatContext holds all state during chat processing pipeline
type chatContext struct {
	// Input
	Request   models.ChatRequest
	Bot       *models.Chatbot
	RAGConfig models.RAGConfig

	// Derived config
	LangConfig   langconfig.LanguageConfig
	ThresholdCfg *models.ThresholdConfig
	BotName      string

	// Conversation state
	Conversation *models.Conversation
	IsNewConv    bool

	// RAG results
	SearchResult *rag.TieredSearchResult
	ChunkMetas   []models.ChunkMetadata
	Sources      []models.SourceUsed

	// Messages for LLM
	Messages []rag.ChatMessage
	Tools    []rag.Tool
	Actions  []*models.ChatbotAction

	// Response state
	Response     string
	TotalTokens  int
	IsHandoff    bool
	HandoffReqID string

	// Timing
	StartTime time.Time
}

// =============================================================================
// MAIN ENTRY POINT - ProcessChat orchestrates the pipeline
// =============================================================================

// ProcessChat handles the complete chat flow with unified tool support.
// It orchestrates a pipeline of steps: context init → RAG search → LLM call → save response.
func (s *ChatService) ProcessChat(ctx context.Context, req models.ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig) (*models.ChatResult, error) {
	// Step 1: Initialize chat context
	cc := s.initChatContext(req, bot, ragConfig)

	// Step 2: Get or create conversation
	if err := s.getOrCreateConversation(ctx, cc); err != nil {
		return nil, err
	}

	// Step 3: Save user message
	if err := s.saveUserMessage(ctx, cc); err != nil {
		return nil, err
	}

	// Step 4: Perform RAG search
	s.performRAGSearch(ctx, cc)

	// Step 5: Build messages for LLM
	s.buildMessages(ctx, cc)

	// Step 6: Execute agentic loop (LLM + tools)
	if err := s.executeAgenticLoop(ctx, cc); err != nil {
		return nil, err
	}

	// Step 7: Apply fallback if needed
	s.applyFallback(ctx, cc)

	// Step 8: Save assistant message
	messageID := s.saveAssistantMessage(ctx, cc)

	// Step 9: Track analytics (async)
	s.trackAnalyticsAsync(cc, messageID)

	// Step 10: Build and return result
	return s.buildChatResult(cc, messageID), nil
}

// =============================================================================
// PIPELINE STEP 1: Initialize Context
// =============================================================================

func (s *ChatService) initChatContext(req models.ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig) *chatContext {
	cc := &chatContext{
		Request:   req,
		Bot:       bot,
		RAGConfig: ragConfig,
		StartTime: time.Now(),
	}

	// Language config
	langCode := normalizeLangCode(bot.LanguageCode)
	cc.LangConfig = langconfig.Get(langCode)

	// Threshold config with defaults
	cc.ThresholdCfg = bot.ThresholdConfig
	if cc.ThresholdCfg == nil {
		cc.ThresholdCfg = models.DefaultThresholdConfig()
	}

	// Bot display name
	cc.BotName = bot.Name
	if bot.BotDisplayName != nil && *bot.BotDisplayName != "" {
		cc.BotName = *bot.BotDisplayName
	}

	return cc
}

// =============================================================================
// PIPELINE STEP 2: Conversation Management
// =============================================================================

func (s *ChatService) getOrCreateConversation(ctx context.Context, cc *chatContext) error {
	conv, err := db.GetOrCreateConversationBySessionID(ctx, s.DB, cc.Bot.ID, cc.Request.SessionID)
	if err != nil {
		return err
	}
	if conv == nil {
		return fmt.Errorf("failed to get or create conversation")
	}

	cc.Conversation = conv
	cc.IsNewConv = conv.MessageCount == 0
	return nil
}

// =============================================================================
// PIPELINE STEP 3: Save User Message
// =============================================================================

func (s *ChatService) saveUserMessage(ctx context.Context, cc *chatContext) error {
	msg := &models.Message{
		ConversationID: cc.Conversation.ID,
		Role:           "user",
		Content:        cc.Request.Message,
		TokensUsed:     0,
	}

	if _, err := db.CreateMessage(ctx, s.DB, msg); err != nil {
		return err
	}

	_ = db.IncrementConversationMessageCount(ctx, s.DB, cc.Conversation.ID)
	return nil
}

// =============================================================================
// PIPELINE STEP 4: RAG Search
// =============================================================================

func (s *ChatService) performRAGSearch(ctx context.Context, cc *chatContext) {
	embedder := s.getEmbedder()
	qc := s.getQdrantClient()

	if embedder == nil || qc == nil {
		cc.SearchResult = &rag.TieredSearchResult{Tier: rag.TierLow}
		return
	}

	// Create embedding
	embedding, err := embedder.CreateEmbedding(ctx, cc.Request.Message)
	if err != nil {
		cc.SearchResult = &rag.TieredSearchResult{Tier: rag.TierLow}
		return
	}

	// Perform tiered search
	result, err := rag.SearchContextTiered(
		embedding,
		cc.Bot.ID,
		cc.RAGConfig.TopK,
		cc.RAGConfig.MaxContextTokens,
		cc.ThresholdCfg,
	)
	if err != nil && s.Log != nil {
		s.Log.Warn("tiered_search_error", map[string]any{"error": err.Error(), "chatbot_id": cc.Bot.ID})
	}

	if result == nil {
		cc.SearchResult = &rag.TieredSearchResult{Tier: rag.TierLow}
		return
	}

	cc.SearchResult = result
	cc.ChunkMetas = result.Chunks

	// Build sources list
	for _, m := range result.Chunks {
		cc.Sources = append(cc.Sources, models.SourceUsed{
			ChunkIndex: m.ChunkIndex,
			SourceType: m.SourceType,
		})
	}
}

// =============================================================================
// PIPELINE STEP 5: Build Messages
// =============================================================================

func (s *ChatService) buildMessages(ctx context.Context, cc *chatContext) {
	// Collect tools
	cc.Tools, cc.Actions = s.collectTools(ctx, cc.Bot)

	// Build system prompt
	capabilities := s.getCapabilitySummaries(ctx, cc.Bot.ID)
	systemPrompt := BuildSystemPrompt(
		cc.BotName,
		strings.TrimSpace(cc.Bot.CustomInstruction),
		capabilities,
		cc.LangConfig.Name,
	)

	// Start with system message
	cc.Messages = []rag.ChatMessage{
		{Role: "system", Content: &systemPrompt},
	}

	// Add conversation history
	s.appendConversationHistory(ctx, cc)

	// Add current user message with RAG context
	s.appendUserMessageWithContext(cc)
}

func (s *ChatService) appendConversationHistory(ctx context.Context, cc *chatContext) {
	historyLimit := calculateHistoryLimit(cc.RAGConfig.MaxContextTokens)
	historyMsgs, _ := db.ListRecentMessages(ctx, s.DB, cc.Conversation.ID, historyLimit)

	for _, m := range historyMsgs {
		// Skip the current user message (will be added with context)
		if m.Content == cc.Request.Message && m.Role == "user" {
			continue
		}
		content := m.Content
		cc.Messages = append(cc.Messages, rag.ChatMessage{Role: m.Role, Content: &content})
	}
}

func (s *ChatService) appendUserMessageWithContext(cc *chatContext) {
	contextText := cc.SearchResult.ContextText

	// Add uncertainty note for medium tier
	if cc.SearchResult.Tier == rag.TierMedium && cc.ThresholdCfg.ShowConfidenceWarning && strings.TrimSpace(contextText) != "" {
		contextText = "[Note: The following sources have moderate relevance. Consider expressing appropriate uncertainty in your response.]\n\n" + contextText
	}

	var content string
	if strings.TrimSpace(contextText) != "" {
		content = RAGContextIntroEN + contextText + "\n\nQuestion:\n" + cc.Request.Message
	} else {
		content = cc.Request.Message
	}

	cc.Messages = append(cc.Messages, rag.ChatMessage{Role: "user", Content: &content})
}

// =============================================================================
// PIPELINE STEP 6: Agentic Loop (LLM + Tools)
// =============================================================================

func (s *ChatService) executeAgenticLoop(ctx context.Context, cc *chatContext) error {
	// Check for static fallback shortcut (skip LLM entirely)
	if cc.SearchResult.Tier == rag.TierLow && cc.ThresholdCfg.FallbackMode == "static" {
		cc.Response = s.getStaticFallbackMessage(cc)
		cc.TotalTokens = 0
		return nil
	}

	// Get LLM client
	toolsClient, modelName, err := s.getToolsClient(cc.Bot.Model)
	if err != nil {
		return err
	}

	// Execute agentic loop
	executor := &rag.ToolExecutor{DB: s.DB, Log: s.Log}
	s.runAgenticLoop(ctx, cc, toolsClient, modelName, executor)

	return nil
}

func (s *ChatService) runAgenticLoop(
	ctx context.Context,
	cc *chatContext,
	client rag.ToolsLLMClient,
	modelName string,
	executor *rag.ToolExecutor,
) {
	const maxIterations = 5

	for i := 0; i < maxIterations; i++ {
		response, err := client.CreateCompletionWithTools(
			ctx, cc.Messages, cc.Tools, modelName, cc.Bot.Temperature, cc.Bot.MaxTokens,
		)
		if err != nil {
			if s.Log != nil {
				s.Log.Error("completion_with_tools_error", map[string]any{"error": err.Error()})
			}
			cc.Response = s.getErrorMessage(cc)
			return
		}

		cc.TotalTokens += response.Usage.TotalTokens
		choice := response.Choices[0]

		// Add assistant message to history
		cc.Messages = append(cc.Messages, choice.Message)

		// No tool calls = final response
		if len(choice.Message.ToolCalls) == 0 {
			if choice.Message.Content != nil {
				cc.Response = *choice.Message.Content
			}
			return
		}

		// Execute tool calls
		if s.executeToolCalls(ctx, cc, choice.Message.ToolCalls, executor) {
			return // Handoff occurred, exit loop
		}
	}
}

func (s *ChatService) executeToolCalls(
	ctx context.Context,
	cc *chatContext,
	toolCalls []rag.ToolCall,
	executor *rag.ToolExecutor,
) bool {
	for _, tc := range toolCalls {
		action := findActionByName(cc.Actions, tc.Function.Name)
		result, err := executor.Execute(ctx, tc, action, cc.Bot.ID, cc.Conversation.ID)

		content := ""
		if err != nil {
			content = fmt.Sprintf(`{"error": "%s"}`, err.Error())
		} else {
			content = result.Result
		}

		// Check for handoff
		if tc.Function.Name == "request_human_handoff" && err == nil {
			cc.HandoffReqID = parseHandoffRequestID(result.Result)
			cc.Response = s.getHandoffMessage(cc)
			cc.IsHandoff = true
			return true // Signal to exit loop
		}

		// Add tool result to messages
		cc.Messages = append(cc.Messages, rag.ChatMessage{
			Role:       "tool",
			ToolCallID: tc.ID,
			Content:    &content,
		})
	}

	return false
}

// =============================================================================
// PIPELINE STEP 7: Apply Fallback
// =============================================================================

func (s *ChatService) applyFallback(ctx context.Context, cc *chatContext) {
	if cc.Response != "" {
		return // Already have a response
	}

	switch cc.SearchResult.Tier {
	case rag.TierLow:
		s.applyLowTierFallback(ctx, cc)
	default:
		// For high/medium tiers with empty response, use error message
		cc.Response = s.getErrorMessage(cc)
	}
}

func (s *ChatService) applyLowTierFallback(ctx context.Context, cc *chatContext) {
	switch cc.ThresholdCfg.FallbackMode {
	case "smart":
		resp, tokens, err := s.smartFallback(ctx, cc.Bot, cc.Request.Message, cc.LangConfig.Name)
		if err == nil {
			cc.Response = resp
			cc.TotalTokens += tokens
		} else {
			cc.Response = s.getStaticFallbackMessage(cc)
		}
	case "escalate":
		cc.Response = cc.LangConfig.UserMessages.HandoffSuggestion
	default: // "static"
		cc.Response = s.getStaticFallbackMessage(cc)
	}
}

// =============================================================================
// PIPELINE STEP 8: Save Assistant Message
// =============================================================================

func (s *ChatService) saveAssistantMessage(ctx context.Context, cc *chatContext) string {
	msgType := "normal"
	if cc.IsHandoff {
		msgType = "handoff"
	}

	msg := &models.Message{
		ConversationID: cc.Conversation.ID,
		Role:           "assistant",
		Content:        cc.Response,
		TokensUsed:     cc.TotalTokens,
		Type:           msgType,
	}

	messageID, err := db.CreateMessage(ctx, s.DB, msg)
	if err != nil {
		return ""
	}

	_ = db.IncrementConversationMessageCount(ctx, s.DB, cc.Conversation.ID)

	// Save source usage
	if len(cc.ChunkMetas) > 0 {
		if err := db.SaveMessageSources(ctx, s.DB, messageID, cc.ChunkMetas); err != nil && s.Log != nil {
			s.Log.Warn("save_message_sources_error", map[string]any{"message_id": messageID, "error": err.Error()})
		}
	}

	return messageID
}

// =============================================================================
// PIPELINE STEP 9: Analytics (Async)
// =============================================================================

func (s *ChatService) trackAnalyticsAsync(cc *chatContext, messageID string) {
	isUnanswered := cc.SearchResult.Tier == rag.TierLow
	startTime := cc.StartTime
	botID := cc.Bot.ID
	isNewConv := cc.IsNewConv
	totalTokens := cc.TotalTokens
	isHandoff := cc.IsHandoff
	userMessage := cc.Request.Message

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		responseTime := int(time.Since(startTime).Milliseconds())
		if err := db.IncrementAnalytics(bgCtx, s.DB, botID, time.Now(), isNewConv, totalTokens, isHandoff, responseTime); err != nil && s.Log != nil {
			s.Log.Warn("analytics_error", map[string]any{"chatbot_id": botID, "error": err.Error()})
		}

		if isUnanswered && !isHandoff {
			_ = db.TrackUnansweredQuery(bgCtx, s.DB, botID, userMessage)
		}
	}()
}

// =============================================================================
// PIPELINE STEP 10: Build Result
// =============================================================================

func (s *ChatService) buildChatResult(cc *chatContext, messageID string) *models.ChatResult {
	return &models.ChatResult{
		Response:         cc.Response,
		TokensUsed:       cc.TotalTokens,
		Sources:          cc.Sources,
		ConversationID:   cc.Conversation.ID,
		MessageID:        messageID,
		IsNewConv:        cc.IsNewConv,
		ConfidenceTier:   string(cc.SearchResult.Tier),
		HandoffRequestID: cc.HandoffReqID,
	}
}

// =============================================================================
// HELPER FUNCTIONS - Message Templates
// =============================================================================

func (s *ChatService) getStaticFallbackMessage(cc *chatContext) string {
	if cc.Bot.FallbackMessages != nil && cc.Bot.FallbackMessages.NoInfoFound != "" {
		return cc.Bot.FallbackMessages.NoInfoFound
	}
	return cc.LangConfig.UserMessages.NoInfoFound
}

func (s *ChatService) getErrorMessage(cc *chatContext) string {
	if cc.Bot.FallbackMessages != nil && cc.Bot.FallbackMessages.ErrorMessage != "" {
		return cc.Bot.FallbackMessages.ErrorMessage
	}
	return cc.LangConfig.UserMessages.ErrorMessage
}

func (s *ChatService) getHandoffMessage(cc *chatContext) string {
	if cc.Bot.FallbackMessages != nil && cc.Bot.FallbackMessages.HandoffMessage != "" {
		return cc.Bot.FallbackMessages.HandoffMessage
	}
	return cc.LangConfig.UserMessages.HandoffSuggestion
}

// =============================================================================
// HELPER FUNCTIONS - Client Initialization
// =============================================================================

func (s *ChatService) getEmbedder() rag.EmbeddingClient {
	if s.Embedder != nil {
		return s.Embedder
	}
	client, err := rag.NewOpenAIClientFromEnv()
	if err != nil {
		return nil
	}
	return client
}

func (s *ChatService) getQdrantClient() *rag.QdrantClient {
	if s.QC != nil {
		return s.QC
	}
	client, _ := rag.NewQdrantClientFromEnv()
	return client
}

// parseHandoffRequestID extracts request_id from handoff tool result
func parseHandoffRequestID(result string) string {
	if !strings.Contains(result, "request_id") {
		return ""
	}
	start := strings.Index(result, `"request_id": "`)
	if start == -1 {
		return ""
	}
	start += len(`"request_id": "`)
	end := strings.Index(result[start:], `"`)
	if end == -1 {
		return ""
	}
	return result[start : start+end]
}

// =============================================================================
// EXISTING HELPER FUNCTIONS (unchanged)
// =============================================================================

// collectTools gathers all available tools for the chat based on bot configuration and plan
func (s *ChatService) collectTools(ctx context.Context, bot *models.Chatbot) ([]rag.Tool, []*models.ChatbotAction) {
	// Get external actions from DB
	actions, err := db.GetEnabledActions(ctx, s.DB, bot.ID)
	if err != nil && s.Log != nil {
		s.Log.Warn("get_actions_error", map[string]any{"error": err.Error(), "chatbot_id": bot.ID})
	}

	// Convert actions to tools
	tools := rag.ConvertActionsToTools(actions)

	// Determine if handoff tool should be included
	includeHandoff := bot.HandoffEnabled
	if includeHandoff {
		plan, planErr := db.GetPlanByUserID(ctx, s.DB, bot.UserID)
		if planErr == nil && plan != nil && !plan.Config.Guardrails.CanUseEscalateFallback {
			includeHandoff = false
		}
	}

	// Add built-in tools
	builtinOptions := rag.BuiltinToolOptions{
		IncludeListSources: true,
		IncludeHandoff:     includeHandoff,
	}
	tools = append(tools, rag.GetBuiltinToolsWithOptions(builtinOptions)...)

	return tools, actions
}

// getToolsClient returns a tools-capable LLM client and the model name to use
func (s *ChatService) getToolsClient(botModel string) (rag.ToolsLLMClient, string, error) {
	useOpenAIOnly := !s.Factory.IsProviderConfigured("openrouter")
	modelString := botModel
	if useOpenAIOnly && !strings.HasPrefix(modelString, "openai:") {
		modelString = "openai:" + modelString
	}

	client, modelName, err := s.Factory.GetClientForModel(modelString)
	if err != nil {
		// Fallback chain: OpenRouter → OpenAI
		if orClient, orErr := s.Factory.GetClient("openrouter"); orErr == nil && orClient != nil {
			client = orClient
			modelName = config.ModelOpenRouterGPT4oMini
		} else if oaiClient, oaiErr := rag.NewOpenAIClientFromEnv(); oaiErr == nil && oaiClient != nil {
			client = oaiClient
			modelName = config.ModelGPT4oMini
		}
		if client == nil {
			return nil, "", fmt.Errorf("no LLM client available: %w", err)
		}
	}

	toolsClient, ok := client.(rag.ToolsLLMClient)
	if !ok {
		// Try OpenRouter for tool support
		if orClient, orErr := s.Factory.GetClient("openrouter"); orErr == nil {
			if tc, tcOk := orClient.(rag.ToolsLLMClient); tcOk {
				toolsClient = tc
				if !strings.Contains(modelName, "/") {
					modelName = config.ModelOpenRouterGPT4oMini
				}
			}
		}
		// Final fallback to OpenAI
		if toolsClient == nil {
			c, err := rag.NewOpenAIClientFromEnv()
			if err != nil {
				return nil, "", fmt.Errorf("tool support requires OpenAI or OpenRouter client: %w", err)
			}
			toolsClient = c
			modelName = config.ModelGPT4oMini
		}
	}

	if toolsClient == nil {
		return nil, "", fmt.Errorf("tools client unavailable")
	}

	return toolsClient, modelName, nil
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

// calculateHistoryLimit returns optimal message history count based on context budget
func calculateHistoryLimit(maxContextTokens int) int {
	historyBudget := int(float64(maxContextTokens) * 0.4)
	limit := historyBudget / 150
	if limit < 4 {
		limit = 4
	}
	if limit > 20 {
		limit = 20
	}
	return limit
}

// smartFallback generates a helpful response when no context is available
func (s *ChatService) smartFallback(ctx context.Context, bot *models.Chatbot, userMessage string, langName string) (string, int, error) {
	capabilities := s.getCapabilitySummaries(ctx, bot.ID)

	botName := bot.Name
	if bot.BotDisplayName != nil && *bot.BotDisplayName != "" {
		botName = *bot.BotDisplayName
	}

	systemPrompt := BuildSmartFallbackPrompt(botName, capabilities, langName)

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

	params := models.CompletionParams{
		SystemPrompt: systemPrompt,
		Context:      "",
		UserMessage:  userMessage,
		Model:        modelName,
		Temperature:  0.3,
		MaxTokens:    200,
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

	if len(summaries) > 20 {
		summaries = summaries[:20]
	}

	return strings.Join(summaries, "\n")
}
