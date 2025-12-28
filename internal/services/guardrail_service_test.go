package services

import (
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/stretchr/testify/assert"
)

func TestGuardrailService_EnforceFallbackMode(t *testing.T) {
	svc := NewGuardrailService(nil)

	tests := []struct {
		name       string
		cfg        *models.ThresholdConfig
		guardrails *models.GuardrailsConfig
		expected   string
	}{
		{
			name:       "nil config returns nil",
			cfg:        nil,
			guardrails: &models.GuardrailsConfig{},
			expected:   "",
		},
		{
			name:       "nil guardrails returns unchanged",
			cfg:        &models.ThresholdConfig{FallbackMode: "smart"},
			guardrails: nil,
			expected:   "smart",
		},
		{
			name:       "smart allowed when CanUseSmartFallback=true",
			cfg:        &models.ThresholdConfig{FallbackMode: "smart"},
			guardrails: &models.GuardrailsConfig{CanUseSmartFallback: true},
			expected:   "smart",
		},
		{
			name:       "smart downgraded when CanUseSmartFallback=false",
			cfg:        &models.ThresholdConfig{FallbackMode: "smart"},
			guardrails: &models.GuardrailsConfig{CanUseSmartFallback: false},
			expected:   "static",
		},
		{
			name:       "escalate allowed when CanUseEscalateFallback=true",
			cfg:        &models.ThresholdConfig{FallbackMode: "escalate"},
			guardrails: &models.GuardrailsConfig{CanUseEscalateFallback: true},
			expected:   "escalate",
		},
		{
			name:       "escalate downgraded to smart when smart available",
			cfg:        &models.ThresholdConfig{FallbackMode: "escalate"},
			guardrails: &models.GuardrailsConfig{CanUseEscalateFallback: false, CanUseSmartFallback: true},
			expected:   "smart",
		},
		{
			name:       "escalate downgraded to static when smart unavailable",
			cfg:        &models.ThresholdConfig{FallbackMode: "escalate"},
			guardrails: &models.GuardrailsConfig{CanUseEscalateFallback: false, CanUseSmartFallback: false},
			expected:   "static",
		},
		{
			name:       "static unchanged",
			cfg:        &models.ThresholdConfig{FallbackMode: "static"},
			guardrails: &models.GuardrailsConfig{},
			expected:   "static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.EnforceFallbackMode(tt.cfg, tt.guardrails)
			if tt.cfg == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result.FallbackMode)
			}
		})
	}
}

func TestGuardrailService_EnforceFallbackMode_DoesNotMutateOriginal(t *testing.T) {
	svc := NewGuardrailService(nil)

	original := &models.ThresholdConfig{FallbackMode: "smart"}
	guardrails := &models.GuardrailsConfig{CanUseSmartFallback: false}

	result := svc.EnforceFallbackMode(original, guardrails)

	assert.Equal(t, "static", result.FallbackMode)
	assert.Equal(t, "smart", original.FallbackMode, "original should not be mutated")
}

func TestGuardrailService_GetFallbackMessage(t *testing.T) {
	svc := NewGuardrailService(nil)
	langCfg := langconfig.Get("en")

	tests := []struct {
		name     string
		tier     rag.ContextTier
		messages *models.FallbackMessages
		contains string
	}{
		{
			name:     "low tier returns static fallback",
			tier:     rag.TierLow,
			messages: nil,
			contains: langCfg.UserMessages.NoInfoFound,
		},
		{
			name:     "high tier returns error message",
			tier:     rag.TierHigh,
			messages: nil,
			contains: langCfg.UserMessages.ErrorMessage,
		},
		{
			name:     "medium tier returns error message",
			tier:     rag.TierMedium,
			messages: nil,
			contains: langCfg.UserMessages.ErrorMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.GetFallbackMessage(tt.tier, tt.messages, langCfg)
			assert.Equal(t, tt.contains, result)
		})
	}
}

func TestGuardrailService_GetStaticFallbackMessage(t *testing.T) {
	svc := NewGuardrailService(nil)
	langCfg := langconfig.Get("en")

	t.Run("uses custom message when available", func(t *testing.T) {
		messages := &models.FallbackMessages{NoInfoFound: "Custom no info"}
		result := svc.GetStaticFallbackMessage(messages, langCfg)
		assert.Equal(t, "Custom no info", result)
	})

	t.Run("uses lang config when no custom message", func(t *testing.T) {
		result := svc.GetStaticFallbackMessage(nil, langCfg)
		assert.Equal(t, langCfg.UserMessages.NoInfoFound, result)
	})

	t.Run("uses lang config when custom message is empty", func(t *testing.T) {
		messages := &models.FallbackMessages{NoInfoFound: ""}
		result := svc.GetStaticFallbackMessage(messages, langCfg)
		assert.Equal(t, langCfg.UserMessages.NoInfoFound, result)
	})
}

func TestGuardrailService_GetErrorMessage(t *testing.T) {
	svc := NewGuardrailService(nil)
	langCfg := langconfig.Get("en")

	t.Run("uses custom message when available", func(t *testing.T) {
		messages := &models.FallbackMessages{ErrorMessage: "Custom error"}
		result := svc.GetErrorMessage(messages, langCfg)
		assert.Equal(t, "Custom error", result)
	})

	t.Run("uses lang config when no custom message", func(t *testing.T) {
		result := svc.GetErrorMessage(nil, langCfg)
		assert.Equal(t, langCfg.UserMessages.ErrorMessage, result)
	})
}

func TestGuardrailService_GetHandoffMessage(t *testing.T) {
	svc := NewGuardrailService(nil)
	langCfg := langconfig.Get("en")

	t.Run("uses custom message when available", func(t *testing.T) {
		messages := &models.FallbackMessages{HandoffMessage: "Custom handoff"}
		result := svc.GetHandoffMessage(messages, langCfg)
		assert.Equal(t, "Custom handoff", result)
	})

	t.Run("uses lang config when no custom message", func(t *testing.T) {
		result := svc.GetHandoffMessage(nil, langCfg)
		assert.Equal(t, langCfg.UserMessages.HandoffSuggestion, result)
	})
}

func TestGuardrailService_GetEmptyStateMessage(t *testing.T) {
	svc := NewGuardrailService(nil)
	langCfg := langconfig.Get("en")

	t.Run("uses empty state message when available", func(t *testing.T) {
		result := svc.GetEmptyStateMessage(nil, langCfg)
		if langCfg.UserMessages.EmptyStateMessage != "" {
			assert.Equal(t, langCfg.UserMessages.EmptyStateMessage, result)
		} else {
			assert.Equal(t, langCfg.UserMessages.NoInfoFound, result)
		}
	})
}

func TestGuardrailService_InitializeThresholdConfig(t *testing.T) {
	svc := NewGuardrailService(nil)

	t.Run("creates defaults when config is nil", func(t *testing.T) {
		result := svc.InitializeThresholdConfig(nil, nil)
		assert.NotNil(t, result)
		assert.Equal(t, 0.50, result.HighThreshold)
		assert.Equal(t, 0.30, result.MediumThreshold)
	})

	t.Run("applies enforcement when config and guardrails provided", func(t *testing.T) {
		cfg := &models.ThresholdConfig{FallbackMode: "smart", HighThreshold: 0.6}
		guardrails := &models.GuardrailsConfig{CanUseSmartFallback: false}

		result := svc.InitializeThresholdConfig(cfg, guardrails)
		assert.Equal(t, "static", result.FallbackMode)
		assert.Equal(t, 0.6, result.HighThreshold)
	})
}

func TestGuardrailService_PermissionChecks(t *testing.T) {
	svc := NewGuardrailService(nil)

	t.Run("CanCustomizeThresholds", func(t *testing.T) {
		assert.False(t, svc.CanCustomizeThresholds(nil))
		assert.False(t, svc.CanCustomizeThresholds(&models.GuardrailsConfig{}))
		assert.True(t, svc.CanCustomizeThresholds(&models.GuardrailsConfig{CanCustomizeThresholds: true}))
	})

	t.Run("CanManageTopics", func(t *testing.T) {
		assert.False(t, svc.CanManageTopics(nil))
		assert.False(t, svc.CanManageTopics(&models.GuardrailsConfig{}))
		assert.True(t, svc.CanManageTopics(&models.GuardrailsConfig{CanManageTopics: true}))
	})

	t.Run("CanCustomizeMessages", func(t *testing.T) {
		assert.False(t, svc.CanCustomizeMessages(nil))
		assert.False(t, svc.CanCustomizeMessages(&models.GuardrailsConfig{}))
		assert.True(t, svc.CanCustomizeMessages(&models.GuardrailsConfig{CanCustomizeMessages: true}))
	})
}
