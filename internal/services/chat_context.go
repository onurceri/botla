package services

import (
	"context"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/langconfig"
)

// =============================================================================
// CHAT CONTEXT - Shared state across pipeline steps
// =============================================================================

// chatContext holds all state during chat processing pipeline.
// This struct is passed through each step of the chat processing pipeline,
// accumulating data as each step completes.
type chatContext struct {
	// Input - provided at initialization
	Request   models.ChatRequest
	Bot       *models.Chatbot
	RAGConfig models.RAGConfig

	// Derived config - computed during initialization
	LangConfig     langconfig.LanguageConfig
	ThresholdCfg   *models.ThresholdConfig
	GuardrailsCfg  *models.GuardrailsConfig // Plan-based guardrails permissions
	BotName        string
	Capabilities   string // Cached capability summaries for fallback

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

// initChatContext creates and initializes a new chat context with all derived values.
// Note: Capabilities are NOT fetched here to keep this function pure and testable.
// They are fetched later in ProcessChat when needed.
func (s *ChatService) initChatContext(
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

	// Threshold config with defaults
	cc.ThresholdCfg = bot.ThresholdConfig
	if cc.ThresholdCfg == nil {
		cc.ThresholdCfg = models.DefaultThresholdConfig()
	}

	// Enforce plan-based fallback mode restrictions
	cc.ThresholdCfg = s.enforcePlanFallbackMode(cc.ThresholdCfg, guardrailsCfg)

	// Bot display name
	cc.BotName = bot.Name
	if bot.BotDisplayName != nil && *bot.BotDisplayName != "" {
		cc.BotName = *bot.BotDisplayName
	}

	return cc
}
