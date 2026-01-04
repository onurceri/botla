package services

import (
	"context"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/pkg/langconfig"
)

// ChatContextBuilder initializes chat context for processing.
type ChatContextBuilder struct {
	guardrails *GuardrailService
}

// NewChatContextBuilder creates a new context builder.
func NewChatContextBuilder(guardrails *GuardrailService) *ChatContextBuilder {
	return &ChatContextBuilder{
		guardrails: guardrails,
	}
}

// Build creates a new chatContext from request parameters.
func (b *ChatContextBuilder) Build(
	ctx context.Context,
	req models.ChatRequest,
	bot *models.Chatbot,
	ragConfig models.RAGConfig,
	guardrailsCfg *models.GuardrailsConfig,
) *chatContext {
	cc := &chatContext{
		Request:       req,
		Bot:           bot,
		RAGConfig:     ragConfig,
		GuardrailsCfg: guardrailsCfg,
		StartTime:     time.Now(),
	}

	// Language config
	langCode := normalizeLangCode(bot.LanguageCode)
	cc.LangConfig = langconfig.Get(langCode)

	// Threshold config with defaults and plan enforcement
	cc.ThresholdCfg = b.guardrails.InitializeThresholdConfig(bot.ThresholdConfig, guardrailsCfg)

	// Bot display name
	cc.BotName = bot.Name
	if bot.BotDisplayName != nil && *bot.BotDisplayName != "" {
		cc.BotName = *bot.BotDisplayName
	}

	return cc
}

// BuildWithBotID creates a chat context when only bot ID is available.
// This is used when we need to fetch the bot first.
func (b *ChatContextBuilder) BuildWithBotID(
	ctx context.Context,
	bot *models.Chatbot,
	ragConfig models.RAGConfig,
	guardrailsCfg *models.GuardrailsConfig,
) *chatContext {
	return b.Build(ctx, models.ChatRequest{}, bot, ragConfig, guardrailsCfg)
}

// ChatContext holds all state during chat processing pipeline.
// This struct is passed through each step of the chat processing pipeline,
// accumulating data as each step completes.
type ChatContext struct {
	// Input - provided at initialization
	Request   models.ChatRequest
	Bot       *models.Chatbot
	RAGConfig models.RAGConfig

	// Derived config - computed during initialization
	LangConfig    langconfig.LanguageConfig
	ThresholdCfg  *models.ThresholdConfig
	GuardrailsCfg *models.GuardrailsConfig // Plan-based guardrails permissions
	BotName       string
	Capabilities  string // Cached capability summaries for fallback

	// Conversation state - set after conversation lookup
	Conversation *models.Conversation
	IsNewConv    bool

	// RAG results - populated after search
	SearchResult *rag.TieredSearchResult
	ChunkMetas   []models.ChunkMetadata
	Sources      []models.SourceUsed

	// Messages for LLM - built before agentic loop
	Messages []rag.ChatMessage
	Tools    []rag.Tool
	Actions  []*models.ChatbotAction

	// Response state - set by agentic loop or fallback
	Response     string
	TotalTokens  int
	IsHandoff    bool
	HandoffReqID string

	// Timing
	StartTime time.Time
}
