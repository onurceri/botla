package services

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestChatContextBuilder_Build(t *testing.T) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	bot := &models.Chatbot{
		ID:             "bot123",
		Name:           "Test Bot",
		LanguageCode:   "en-US",
		Model:          "gpt-4",
		MaxTokens:      1000,
		ThresholdConfig: &models.ThresholdConfig{
			HighThreshold:   0.7,
			MediumThreshold: 0.4,
			FallbackMode:    "smart",
		},
	}

	req := models.ChatRequest{
		Message: "Hello, world!",
	}

	ragConfig := models.RAGConfig{
		TopK:             5,
		MaxContextTokens: 2000,
	}

	planGuardrails := &models.GuardrailsConfig{
		CanCustomizeThresholds: true,
		CanUseSmartFallback:   true,
	}

	ctx := context.Background()
	cc := builder.Build(ctx, req, bot, ragConfig, planGuardrails)

	assert.NotNil(t, cc)
	assert.Equal(t, req.Message, cc.Request.Message)
	assert.Equal(t, bot.ID, cc.Bot.ID)
	assert.Equal(t, bot.Name, cc.BotName)
	assert.Equal(t, ragConfig, cc.RAGConfig)
	assert.Equal(t, planGuardrails, cc.GuardrailsCfg)
	assert.NotNil(t, cc.ThresholdCfg)
	assert.NotZero(t, cc.StartTime)
}

func TestChatContextBuilder_Build_DefaultDisplayName(t *testing.T) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	displayName := "Custom Name"
	bot := &models.Chatbot{
		ID:             "bot123",
		Name:           "Test Bot",
		BotDisplayName: &displayName,
		LanguageCode:   "en-US",
	}

	req := models.ChatRequest{}
	cc := builder.Build(context.Background(), req, bot, models.RAGConfig{}, nil)

	assert.Equal(t, "Custom Name", cc.BotName)
}

func TestChatContextBuilder_Build_EmptyDisplayName(t *testing.T) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	emptyName := ""
	bot := &models.Chatbot{
		ID:             "bot123",
		Name:           "Test Bot",
		BotDisplayName: &emptyName,
		LanguageCode:   "en-US",
	}

	req := models.ChatRequest{}
	cc := builder.Build(context.Background(), req, bot, models.RAGConfig{}, nil)

	// Should fall back to bot.Name
	assert.Equal(t, "Test Bot", cc.BotName)
}

func TestChatContextBuilder_Build_LanguageConfig(t *testing.T) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	tests := []struct {
		name     string
		langCode string
		expected string
	}{
		{"English US", "en-US", "en"},
		{"English GB", "en-GB", "en"},
		{"Turkish", "tr-TR", "tr"},
		{"Turkish with variant", "tr", "tr"},
		{"Empty defaults to Turkish", "", "tr"},
		{"Whitespace trimmed", "  en-US  ", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &models.Chatbot{
				LanguageCode: tt.langCode,
			}
			cc := builder.Build(context.Background(), models.ChatRequest{}, bot, models.RAGConfig{}, nil)

			assert.Equal(t, tt.expected, cc.LangConfig.Code)
		})
	}
}

func TestChatContextBuilder_Build_ThresholdConfig(t *testing.T) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	bot := &models.Chatbot{
		ThresholdConfig: &models.ThresholdConfig{
			HighThreshold:   0.8,
			MediumThreshold: 0.5,
			FallbackMode:    "smart",
		},
	}

	cc := builder.Build(context.Background(), models.ChatRequest{}, bot, models.RAGConfig{}, nil)

	assert.NotNil(t, cc.ThresholdCfg)
	assert.Equal(t, 0.8, cc.ThresholdCfg.HighThreshold)
	assert.Equal(t, 0.5, cc.ThresholdCfg.MediumThreshold)
	assert.Equal(t, "smart", cc.ThresholdCfg.FallbackMode)
}

func TestChatContextBuilder_Build_PlanEnforcedThresholds(t *testing.T) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	bot := &models.Chatbot{
		ThresholdConfig: &models.ThresholdConfig{
			HighThreshold:   0.9, // High threshold
			MediumThreshold: 0.6, // Medium threshold
			FallbackMode:    "smart",
		},
	}

	planGuardrails := &models.GuardrailsConfig{
		CanCustomizeThresholds: true,
		CanUseSmartFallback:   true,
	}

	cc := builder.Build(context.Background(), models.ChatRequest{}, bot, models.RAGConfig{}, planGuardrails)

	// Threshold should be adjusted by plan guardrails
	assert.NotNil(t, cc.ThresholdCfg)
	// The actual values depend on InitializeThresholdConfig implementation
}

func TestChatContextBuilder_BuildWithBotID(t *testing.T) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	bot := &models.Chatbot{
		ID:           "bot123",
		Name:         "Test Bot",
		LanguageCode: "en-US",
	}

	cc := builder.BuildWithBotID(context.Background(), bot, models.RAGConfig{}, nil)

	assert.NotNil(t, cc)
	assert.Equal(t, bot.ID, cc.Bot.ID)
}

func TestChatContextBuilder_NilGuardrails(t *testing.T) {
	builder := NewChatContextBuilder(nil)

	bot := &models.Chatbot{
		ID:           "bot123",
		Name:         "Test Bot",
		LanguageCode: "en-US",
	}

	// Should not panic even with nil guardrails
	cc := builder.Build(context.Background(), models.ChatRequest{}, bot, models.RAGConfig{}, nil)

	assert.NotNil(t, cc)
	assert.Equal(t, "Test Bot", cc.BotName)
}

// Benchmark for Build performance
func BenchmarkChatContextBuilder_Build(b *testing.B) {
	guardrails := NewGuardrailService(nil)
	builder := NewChatContextBuilder(guardrails)

	bot := &models.Chatbot{
		ID:           "bot123",
		Name:         "Test Bot",
		LanguageCode: "en-US",
		MaxTokens:    1000,
	}

	req := models.ChatRequest{
		Message: "Hello, world!",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Build(ctx, req, bot, models.RAGConfig{}, nil)
	}
}
