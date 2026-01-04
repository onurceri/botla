package services

import (
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/pkg/langconfig"
	"github.com/onurceri/botla-app/pkg/logger"
)

// GuardrailService handles guardrail-related logic including:
// - Fallback mode enforcement based on plan permissions
// - Fallback message selection based on context tier
// - Topic restriction formatting for prompts
type GuardrailService struct {
	log *logger.Logger
}

// NewGuardrailService creates a new GuardrailService.
func NewGuardrailService(log *logger.Logger) *GuardrailService {
	return &GuardrailService{log: log}
}

// EnforceFallbackMode ensures the fallback mode respects plan restrictions.
// Free plan cannot use smart/escalate, non-Ultra cannot use escalate.
func (g *GuardrailService) EnforceFallbackMode(cfg *models.ThresholdConfig, guardrails *models.GuardrailsConfig) *models.ThresholdConfig {
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

// GetFallbackMessage returns the appropriate fallback message based on tier.
func (g *GuardrailService) GetFallbackMessage(tier rag.ContextTier, fallbackMessages *models.FallbackMessages, langConfig langconfig.LanguageConfig) string {
	switch tier {
	case rag.TierLow:
		return g.GetStaticFallbackMessage(fallbackMessages, langConfig)
	default:
		return g.GetErrorMessage(fallbackMessages, langConfig)
	}
}

// GetStaticFallbackMessage returns the "no info found" message.
func (g *GuardrailService) GetStaticFallbackMessage(fallbackMessages *models.FallbackMessages, langConfig langconfig.LanguageConfig) string {
	if fallbackMessages != nil && fallbackMessages.NoInfoFound != "" {
		return fallbackMessages.NoInfoFound
	}
	return langConfig.UserMessages.NoInfoFound
}

// GetErrorMessage returns the error message.
func (g *GuardrailService) GetErrorMessage(fallbackMessages *models.FallbackMessages, langConfig langconfig.LanguageConfig) string {
	if fallbackMessages != nil && fallbackMessages.ErrorMessage != "" {
		return fallbackMessages.ErrorMessage
	}
	return langConfig.UserMessages.ErrorMessage
}

// GetHandoffMessage returns the handoff suggestion message.
func (g *GuardrailService) GetHandoffMessage(fallbackMessages *models.FallbackMessages, langConfig langconfig.LanguageConfig) string {
	if fallbackMessages != nil && fallbackMessages.HandoffMessage != "" {
		return fallbackMessages.HandoffMessage
	}
	return langConfig.UserMessages.HandoffSuggestion
}

// GetEmptyStateMessage returns a softer message for when the bot has no knowledge sources.
func (g *GuardrailService) GetEmptyStateMessage(fallbackMessages *models.FallbackMessages, langConfig langconfig.LanguageConfig) string {
	if langConfig.UserMessages.EmptyStateMessage != "" {
		return langConfig.UserMessages.EmptyStateMessage
	}
	return g.GetStaticFallbackMessage(fallbackMessages, langConfig)
}

// InitializeThresholdConfig returns the threshold config with defaults applied.
func (g *GuardrailService) InitializeThresholdConfig(cfg *models.ThresholdConfig, guardrails *models.GuardrailsConfig) *models.ThresholdConfig {
	if cfg == nil {
		cfg = models.DefaultThresholdConfig()
	}
	return g.EnforceFallbackMode(cfg, guardrails)
}

// CanCustomizeThresholds checks if the plan allows threshold customization.
func (g *GuardrailService) CanCustomizeThresholds(guardrails *models.GuardrailsConfig) bool {
	return guardrails != nil && guardrails.CanCustomizeThresholds
}

// CanManageTopics checks if the plan allows topic management.
func (g *GuardrailService) CanManageTopics(guardrails *models.GuardrailsConfig) bool {
	return guardrails != nil && guardrails.CanManageTopics
}

// CanCustomizeMessages checks if the plan allows message customization.
func (g *GuardrailService) CanCustomizeMessages(guardrails *models.GuardrailsConfig) bool {
	return guardrails != nil && guardrails.CanCustomizeMessages
}
