package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/config"
)

// =============================================================================
// FALLBACK LOGIC - Handles responses when RAG context is low/unavailable
// =============================================================================

// applyFallback sets a response when the main LLM loop didn't produce one.
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

// applyLowTierFallback handles responses when no relevant RAG context was found.
func (s *ChatService) applyLowTierFallback(ctx context.Context, cc *chatContext) {
	switch cc.ThresholdCfg.FallbackMode {
	case "smart":
		// Smart fallback requires capabilities to redirect the user.
		// If no capabilities, fall back to static message to avoid
		// the bot answering with general LLM knowledge.
		if strings.TrimSpace(cc.Capabilities) == "" {
			cc.Response = s.getStaticFallbackMessage(cc)
			return
		}
		resp, tokens, err := s.restrictedSmartFallback(ctx, cc)
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
// MESSAGE TEMPLATES - Localized fallback messages
// =============================================================================

// getStaticFallbackMessage returns the "no info found" message.
func (s *ChatService) getStaticFallbackMessage(cc *chatContext) string {
	if cc.Bot.FallbackMessages != nil && cc.Bot.FallbackMessages.NoInfoFound != "" {
		return cc.Bot.FallbackMessages.NoInfoFound
	}
	return cc.LangConfig.UserMessages.NoInfoFound
}

// getErrorMessage returns the error message.
func (s *ChatService) getErrorMessage(cc *chatContext) string {
	if cc.Bot.FallbackMessages != nil && cc.Bot.FallbackMessages.ErrorMessage != "" {
		return cc.Bot.FallbackMessages.ErrorMessage
	}
	return cc.LangConfig.UserMessages.ErrorMessage
}

// getHandoffMessage returns the handoff suggestion message.
func (s *ChatService) getHandoffMessage(cc *chatContext) string {
	if cc.Bot.FallbackMessages != nil && cc.Bot.FallbackMessages.HandoffMessage != "" {
		return cc.Bot.FallbackMessages.HandoffMessage
	}
	return cc.LangConfig.UserMessages.HandoffSuggestion
}

// =============================================================================
// SMART FALLBACK - AI-generated redirection when no context found
// =============================================================================

// restrictedSmartFallback generates a controlled response when no RAG context is available.
// This version uses a stricter prompt and lower token limit to prevent
// the bot from answering factual questions with general LLM knowledge.
func (s *ChatService) restrictedSmartFallback(ctx context.Context, cc *chatContext) (string, int, error) {
	systemPrompt := BuildRestrictedFallbackPrompt(cc.BotName, cc.Capabilities, cc.LangConfig.Name)

	client, modelName, err := s.Factory.GetClientForModel(cc.Bot.Model)
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
		UserMessage:  cc.Request.Message,
		Model:        modelName,
		Temperature:  0.3,
		MaxTokens:    80, // Low limit to prevent detailed factual responses
	}

	res, err := client.CreateCompletion(ctx, params)
	if err != nil {
		if s.Log != nil {
			s.Log.Error("restricted_smart_fallback_error", map[string]any{"error": err.Error(), "model": cc.Bot.Model})
		}
		return "", 0, err
	}

	return res.Content, res.UsageTokens, nil
}

// =============================================================================
// PLAN ENFORCEMENT - Ensures fallback mode respects plan restrictions
// =============================================================================

// enforcePlanFallbackMode ensures the fallback mode respects plan restrictions.
// Free plan cannot use smart/escalate, non-Ultra cannot use escalate.
func (s *ChatService) enforcePlanFallbackMode(cfg *models.ThresholdConfig, guardrails *models.GuardrailsConfig) *models.ThresholdConfig {
	if cfg == nil || guardrails == nil {
		return cfg
	}

	// Create a copy to avoid modifying the original
	result := *cfg

	switch result.FallbackMode {
	case "smart":
		if !guardrails.CanUseSmartFallback {
			result.FallbackMode = "static"
		}
	case "escalate":
		if !guardrails.CanUseEscalateFallback {
			// Downgrade to smart if allowed, otherwise static
			if guardrails.CanUseSmartFallback {
				result.FallbackMode = "smart"
			} else {
				result.FallbackMode = "static"
			}
		}
	}

	return &result
}
