package services

import (
	"context"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
)

// =============================================================================
// PIPELINE STEPS - RAG Search, Message Building, Agentic Loop
// =============================================================================

// getOrCreateConversation retrieves or creates a conversation for this session.
func (s *ChatService) getOrCreateConversation(ctx context.Context, cc *chatContext) error {
	conv, err := db.GetOrCreateConversationBySessionID(ctx, s.DB, cc.Bot.ID, cc.Request.SessionID)
	if err != nil {
		return err
	}
	if conv == nil {
		return errConversationCreateFailed
	}

	cc.Conversation = conv
	cc.IsNewConv = conv.MessageCount == 0
	return nil
}

// saveUserMessage saves the user's message to the database.
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

// performRAGSearch searches for relevant context using vector similarity.
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
		ctx,
		qc,
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

// buildMessages constructs the message array for the LLM call.
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
		cc.Bot.TopicRestrictions,
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

// appendConversationHistory adds recent messages to provide context.
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

// appendUserMessageWithContext adds the user message with RAG context if available.
func (s *ChatService) appendUserMessageWithContext(cc *chatContext) {
	contextText := cc.SearchResult.ContextText

	// Add uncertainty note for medium tier
	if cc.SearchResult.Tier == rag.TierMedium && cc.ThresholdCfg.ShowConfidenceWarning && strings.TrimSpace(contextText) != "" {
		contextText = "[Note: The following sources have moderate relevance. Consider expressing appropriate uncertainty in your response.]\n\n" + contextText
	}

	var content string
	if strings.TrimSpace(contextText) != "" {
		content = RAGContextIntroEN + contextText + "\n\nQuestion:\n" + cc.Request.Message
	} else if cc.SearchResult.Tier == rag.TierLow && len(cc.Actions) > 0 {
		// Tool-only mode: No RAG context but custom actions available.
		// Prevent LLM from using general knowledge - only allow tool usage.
		content = "[IMPORTANT: You have NO knowledge sources available for this query. " +
			"You may ONLY use the provided tools to help the user. " +
			"If no tool can help with their request, politely say you don't have information on this topic. " +
			"Do NOT make up facts or use general knowledge.]\n\n" + cc.Request.Message
	} else {
		content = cc.Request.Message
	}

	cc.Messages = append(cc.Messages, rag.ChatMessage{Role: "user", Content: &content})
}

// executeAgenticLoop runs the LLM with tools for High/Medium tier responses.
// For Low tier (no relevant context), this is skipped and fallback handles the response.
// Exception: If custom actions are defined, we still run the loop to allow tool usage.
func (s *ChatService) executeAgenticLoop(ctx context.Context, cc *chatContext) error {
	// Skip agentic loop for Low tier when no custom actions are defined.
	// Built-in tools (list_sources) alone don't justify running the LLM loop.
	// This ensures the fallback logic handles the "no sources" case properly.
	if cc.SearchResult.Tier == rag.TierLow && len(cc.Actions) == 0 {
		return nil
	}

	// Only proceed for High/Medium tier where we have RAG context
	toolsClient, modelName, err := s.getToolsClient(cc.Bot.Model)
	if err != nil {
		return err
	}

	// Execute agentic loop
	executor := &rag.ToolExecutor{DB: s.DB, Log: s.Log}
	s.runAgenticLoop(ctx, cc, toolsClient, modelName, executor)

	return nil
}

// runAgenticLoop executes the LLM call with tool support, iterating as needed.
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

		// CR-001: Guard against empty choices array to prevent panic
		if len(response.Choices) == 0 {
			if s.Log != nil {
				s.Log.Error("llm_empty_choices", map[string]any{"chatbot_id": cc.Bot.ID})
			}
			cc.Response = s.getErrorMessage(cc)
			return
		}
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

// executeToolCalls processes tool calls from the LLM response.
// Returns true if a handoff occurred (signals exit from agentic loop).
func (s *ChatService) executeToolCalls(
	ctx context.Context,
	cc *chatContext,
	toolCalls []rag.ToolCall,
	executor *rag.ToolExecutor,
) bool {
	for _, tc := range toolCalls {
		action := executor.FindActionByToolName(cc.Actions, tc.Function.Name)
		result, err := executor.Execute(ctx, tc, action, cc.Bot.ID, cc.Conversation.ID)

		content := ""
		if err != nil {
			content = `{"error": "` + err.Error() + `"}`
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

// saveAssistantMessage persists the assistant's response to the database.
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

// buildChatResult constructs the final response object.
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
