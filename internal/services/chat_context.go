package services

import (
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
//
// Use ChatContextBuilder to create instances of this struct.
type chatContext struct {
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
