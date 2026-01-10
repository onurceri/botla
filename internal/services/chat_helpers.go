package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/pkg/config"
)

// =============================================================================
// ERRORS
// =============================================================================

var errConversationCreateFailed = errors.New("failed to get or create conversation")

// =============================================================================
// CLIENT INITIALIZATION HELPERS
// =============================================================================

// getEmbedder returns the embedding client, creating one if needed.
func (s *ChatService) getEmbedder() rag.EmbeddingClient {
	if s.Embedder != nil {
		return s.Embedder
	}
	// Try getting from factory
	client, err := s.Factory.GetClient("openai")
	if err == nil && client != nil {
		if e, ok := client.(rag.EmbeddingClient); ok {
			return e
		}
	}
	return nil
}

// getQdrantClient returns the Qdrant client.
func (s *ChatService) getQdrantClient() rag.VectorClient {
	return s.QC
}

// getToolsClient returns a tools-capable LLM client and the model name to use.
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
			modelName = config.DefaultModelName // Bare model name, resolved to API format at call time
		} else if oaiClient, oaiErr := s.Factory.GetClient("openai"); oaiErr == nil && oaiClient != nil {
			client = oaiClient
			modelName = config.DefaultModelName
		}
		if client == nil {
			return nil, "", errors.New("no LLM client available")
		}
	}

	toolsClient, ok := client.(rag.ToolsLLMClient)
	if !ok {
		// Try OpenRouter for tool support
		if orClient, orErr := s.Factory.GetClient("openrouter"); orErr == nil {
			if tc, tcOk := orClient.(rag.ToolsLLMClient); tcOk {
				toolsClient = tc
				// Use bare model name, API format resolved at call time
				modelName = config.DefaultModelName
			}
		}
		// Final fallback to OpenAI
		if toolsClient == nil {
			c, err := s.Factory.GetClient("openai")
			if err != nil {
				return nil, "", errors.New("tool support requires OpenAI or OpenRouter client")
			}
			if tc, ok := c.(rag.ToolsLLMClient); ok {
				toolsClient = tc
				modelName = config.DefaultModelName
			} else {
				return nil, "", errors.New("openai client does not support tools")
			}
		}
	}

	if toolsClient == nil {
		return nil, "", errors.New("tools client unavailable")
	}

	return toolsClient, modelName, nil
}

// =============================================================================
// TOOL HELPERS
// =============================================================================

// collectTools gathers all available tools for the chat based on bot configuration and plan.
func (s *ChatService) collectTools(ctx context.Context, bot *models.Chatbot) ([]rag.Tool, []*models.ChatbotAction) {
	// Get external actions from repository
	var actions []*models.ChatbotAction
	if s.ActionRepo != nil {
		fetchedActions, listErr := s.ActionRepo.List(ctx, bot.ID)
		if listErr != nil {
			if s.Log != nil {
				s.Log.Warn("get_actions_error", map[string]any{"error": listErr.Error(), "chatbot_id": bot.ID})
			}
		} else {
			actions = fetchedActions
		}
	}

	// Convert actions to tools
	tools := rag.ConvertActionsToTools(actions)

	// Determine if handoff tool should be included
	includeHandoff := bot.HandoffEnabled
	if includeHandoff {
		plan, planErr := s.PlanRepo.GetByUserID(ctx, bot.UserID)
		if planErr == nil && plan != nil && !plan.Limits.GuardrailsCanUseEscalateFallback {
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

// parseHandoffRequestID extracts request_id from handoff tool result.
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
// DATA HELPERS
// =============================================================================

func (s *ChatService) getCapabilitySummaries(ctx context.Context, chatbotID string) string {
	if s.SourceRepo == nil {
		return ""
	}

	summaries, err := s.SourceRepo.GetCapabilitySummaries(ctx, chatbotID)
	if err != nil {
		if s.Log != nil {
			s.Log.Warn("get_capability_summaries_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID})
		}
		return ""
	}

	if len(summaries) == 0 {
		return ""
	}

	return "- " + strings.Join(summaries, "\n- ")
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// normalizeLangCode extracts the language code prefix (e.g., "tr" from "tr-TR").
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

// calculateHistoryLimit returns optimal message history count based on context budget.
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

// =============================================================================
// ANALYTICS
// =============================================================================

// trackAnalyticsAsync tracks chat analytics asynchronously.
func (s *ChatService) trackAnalyticsAsync(cc *chatContext, messageID string) {
	isUnanswered := cc.SearchResult.Tier == rag.TierLow
	startTime := cc.StartTime
	botID := cc.Bot.ID
	isNewConv := cc.IsNewConv
	totalTokens := cc.TotalTokens
	isHandoff := cc.IsHandoff
	userMessage := cc.Request.Message

	fn := func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		responseTime := int(time.Since(startTime).Milliseconds())
		if err := s.AnalyticsRepo.IncrementAnalytics(bgCtx, botID, isNewConv, totalTokens, isHandoff, responseTime); err != nil && s.Log != nil {
			s.Log.Warn("analytics_error", map[string]any{"chatbot_id": botID, "error": err.Error()})
		}

		if isUnanswered && !isHandoff {
			_ = s.AnalyticsRepo.TrackUnansweredQuery(bgCtx, botID, userMessage)
		}
	}

	if s.SyncAnalytics {
		fn()
	} else {
		go fn()
	}
}
