package services

import (
	"context"
	"fmt"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/config"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
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
		// Smart fallback uses restricted LLM prompt to handle greetings naturally
		// while refusing factual questions. The RestrictedFallbackPrompt has
		// strong guardrails that prevent the bot from answering with general knowledge.
		// Even with empty capabilities, we call the LLM to handle greetings/small talk.
		resp, tokens, err := s.restrictedSmartFallback(ctx, cc)
		if err == nil {
			cc.Response = resp
			cc.TotalTokens += tokens
		} else {
			// If LLM fails, use the softer empty state message instead of hard refusal
			cc.Response = s.getEmptyStateMessage(cc)
		}
	case "escalate":
		cc.Response = cc.LangConfig.UserMessages.HandoffSuggestion
	default: // "static"
		cc.Response = s.getStaticFallbackMessage(cc)
	}
}

// =============================================================================
// MESSAGE TEMPLATES - Delegated to GuardrailService
// =============================================================================

// getStaticFallbackMessage returns the "no info found" message.
func (s *ChatService) getStaticFallbackMessage(cc *chatContext) string {
	return s.Guardrails.GetStaticFallbackMessage(cc.Bot.FallbackMessages, cc.LangConfig)
}

// getErrorMessage returns the error message.
func (s *ChatService) getErrorMessage(cc *chatContext) string {
	return s.Guardrails.GetErrorMessage(cc.Bot.FallbackMessages, cc.LangConfig)
}

// getHandoffMessage returns the handoff suggestion message.
func (s *ChatService) getHandoffMessage(cc *chatContext) string {
	return s.Guardrails.GetHandoffMessage(cc.Bot.FallbackMessages, cc.LangConfig)
}

// getEmptyStateMessage returns a softer message for when the bot has no knowledge sources.
func (s *ChatService) getEmptyStateMessage(cc *chatContext) string {
	return s.Guardrails.GetEmptyStateMessage(cc.Bot.FallbackMessages, cc.LangConfig)
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
		c, e := s.Factory.GetClient("openai")
		if e != nil || c == nil {
			return "", 0, pkgerrors.Wrapf(e, "openai client not configured")
		}
		client = c
		modelName = config.DefaultModelName
	}

	if client == nil {
		return "", 0, fmt.Errorf("model client unavailable")
	}

	params := models.CompletionParams{
		SystemPrompt: systemPrompt,
		Context:      "",
		UserMessage:  cc.Request.Message,
		Model:        modelName,
		Temperature:  cc.Bot.Temperature,
		MaxTokens:    cc.Bot.MaxTokens,
	}

	res, err := client.CreateCompletion(ctx, params)
	if err != nil {
		if s.Log != nil {
			s.Log.Error("restricted_smart_fallback_error", map[string]any{"error": err.Error(), "model": cc.Bot.Model})
		}
		return "", 0, pkgerrors.Wrapf(err, "create low-tier completion")
	}

	return res.Content, res.UsageTokens, nil
}
