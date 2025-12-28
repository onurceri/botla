package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

// =============================================================================
// CHAT SERVICE - Core chat processing with RAG and tool support
//
// This service is split across multiple files for better organization:
//   - chat_service.go  : Service struct and main entry point (this file)
//   - chat_context.go  : chatContext struct and initialization
//   - chat_pipeline.go : RAG search, message building, agentic loop
//   - chat_fallback.go : Fallback logic and plan enforcement
//   - chat_helpers.go  : Client initialization, utilities, analytics
//   - chat_prompts.go  : LLM prompt templates
// =============================================================================

// ChatService handles core chat logic, shared between authenticated and public endpoints.
type ChatService struct {
	DB            *sql.DB
	Factory       *rag.ClientFactory
	Embedder      rag.EmbeddingClient
	QC            rag.VectorClient
	Log           *logger.Logger
	Guardrails    *GuardrailService
	SyncAnalytics bool // When true, analytics run synchronously (useful for testing)
}

// NewChatService creates a new ChatService instance.
func NewChatService(db *sql.DB, factory *rag.ClientFactory, embedder rag.EmbeddingClient, qc rag.VectorClient, log *logger.Logger) *ChatService {
	if embedder == nil {
		if client, err := factory.GetClient("openai"); err == nil {
			if e, ok := client.(rag.EmbeddingClient); ok {
				embedder = e
			}
		}
	}
	return &ChatService{
		DB:         db,
		Factory:    factory,
		Embedder:   embedder,
		QC:         qc,
		Log:        log,
		Guardrails: NewGuardrailService(log),
	}
}

// ChatRequestWithUser encapsulates resources needed for a chat request
type ChatRequestWithUser struct {
	UserID      string
	Chatbot     *models.Chatbot
	ChatRequest models.ChatRequest
}

// ProcessChatWithValidation handles the complete chat flow including plan validation, quota enforcement, and model adjustment.
func (s *ChatService) ProcessChatWithValidation(ctx context.Context, req ChatRequestWithUser) (*models.ChatResult, error) {
	// 1. Get and validate plan
	plan, err := db.GetPlanByUserID(ctx, s.DB, req.UserID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get plan")
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}

	// 2. Validate and adjust model
	if len(plan.Config.Chat.AllowedModels) > 0 {
		allowed := false
		for _, m := range plan.Config.Chat.AllowedModels {
			if m == req.Chatbot.Model {
				allowed = true
				break
			}
		}
		// If requested model is not allowed, default to the first allowed model
		if !allowed {
			req.Chatbot.Model = plan.Config.Chat.AllowedModels[0]
		}
	}

	// 3. Token Check & Reservation (Atomic)
	maxMonthlyTokens := plan.Config.Chat.MaxMonthlyTokens
	var estimatedTokens int

	if maxMonthlyTokens > 0 {
		estimatedTokens = req.Chatbot.MaxTokens
		if estimatedTokens <= 0 {
			estimatedTokens = 512 // Default estimate if not set
		}

		// Reserve tokens optimistically
		err := db.ReserveChatTokens(ctx, s.DB, req.UserID, estimatedTokens, maxMonthlyTokens)
		if err != nil {
			if errors.Is(err, db.ErrTokenQuotaExceeded) {
				return nil, ErrTokenQuotaExceeded
			}
			return nil, pkgerrors.Wrapf(err, "reserve tokens")
		}
	}

	// 4. Process Chat
	result, err := s.ProcessChat(ctx, req.ChatRequest, req.Chatbot, plan.Config.Chat.RAG)

	// 5. Adjust Token Usage
	if maxMonthlyTokens > 0 {
		if err != nil {
			// On error, refund the reserved tokens
			_ = db.AdjustChatTokens(context.Background(), s.DB, req.UserID, -estimatedTokens)
		} else {
			// Adjust based on actual usage
			delta := result.TokensUsed - estimatedTokens
			if delta != 0 {
				_ = db.AdjustChatTokens(context.Background(), s.DB, req.UserID, delta)
			}
		}
	}

	return result, err
}

// =============================================================================
// MAIN ENTRY POINT - ProcessChat orchestrates the pipeline
// =============================================================================

// ProcessChat handles the complete chat flow with unified tool support.
// It orchestrates a pipeline of steps:
//  1. Initialize context (config, language, capabilities)
//  2. Get or create conversation
//  3. Save user message
//  4. Perform RAG search
//  5. Build messages for LLM
//  6. Execute agentic loop (LLM + tools)
//  7. Apply fallback if needed
//  8. Save assistant message
//  9. Track analytics (async)
//  10. Build and return result
func (s *ChatService) ProcessChat(ctx context.Context, req models.ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig) (*models.ChatResult, error) {
	// Step 0: Fetch plan for guardrails enforcement
	var guardrailsCfg *models.GuardrailsConfig
	if plan, err := db.GetPlanByUserID(ctx, s.DB, bot.UserID); err == nil && plan != nil {
		guardrailsCfg = &plan.Config.Guardrails
	}

	// Step 1: Initialize chat context
	cc := s.initChatContext(ctx, req, bot, ragConfig, guardrailsCfg)

	// Step 2: Get or create conversation
	if err := s.getOrCreateConversation(ctx, cc); err != nil {
		return nil, err
	}

	// Step 3: Save user message
	if err := s.saveUserMessage(ctx, cc); err != nil {
		return nil, err
	}

	// Step 4: Perform RAG search
	s.performRAGSearch(ctx, cc)

	// Step 4b: Fetch capabilities for potential fallback use
	if cc.SearchResult.Tier == rag.TierLow {
		cc.Capabilities = s.getCapabilitySummaries(ctx, bot.ID)
	}

	// Step 5: Build messages for LLM (collects tools, builds prompt)
	s.buildMessages(ctx, cc)

	// Step 6: Execute agentic loop (LLM + tools)
	if err := s.executeAgenticLoop(ctx, cc); err != nil {
		// Log error but proceed to fallback
		if s.Log != nil {
			s.Log.Error("chat_agentic_loop_failed", map[string]any{"error": err.Error(), "chatbot_id": bot.ID})
		}
	}

	// Step 7: Apply fallback if needed
	s.applyFallback(ctx, cc)

	// Step 8: Save assistant message
	messageID := s.saveAssistantMessage(ctx, cc)

	// Step 9: Track analytics (async)
	s.trackAnalyticsAsync(cc, messageID)

	// Step 10: Build and return result
	return s.buildChatResult(cc, messageID), nil
}
