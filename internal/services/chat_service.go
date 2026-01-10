package services

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/logger"
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
// ChatService is composed of focused sub-services for better maintainability:
//   - Quota: Token quota enforcement
//   - Context: Chat context initialization
//   - Guardrails: Content filtering and fallback messages
type ChatService struct {
	PlanRepo         repository.PlanRepository
	ConversationRepo repository.ConversationRepository
	AnalyticsRepo    repository.AnalyticsRepository
	ActionRepo       repository.ActionRepository
	SourceRepo       repository.SourceRepository
	HandoffRepo      repository.HandoffRepository
	Factory          *rag.ClientFactory
	Embedder         rag.EmbeddingClient
	QC               rag.VectorClient
	Log              *logger.Logger
	Guardrails       *GuardrailService
	Quota            *QuotaEnforcer
	Context          *ChatContextBuilder
	SyncAnalytics    bool
}

// NewChatService creates a new ChatService instance with all sub-services composed.
func NewChatService(
	planRepo repository.PlanRepository,
	conversationRepo repository.ConversationRepository,
	analyticsRepo repository.AnalyticsRepository,
	actionRepo repository.ActionRepository,
	sourceRepo repository.SourceRepository,
	handoffRepo repository.HandoffRepository,
	factory *rag.ClientFactory,
	embedder rag.EmbeddingClient,
	qc rag.VectorClient,
	usageRepo repository.UsageRepository,
	log *logger.Logger,
) *ChatService {
	if embedder == nil {
		if client, err := factory.GetClient("openai"); err == nil {
			if e, ok := client.(rag.EmbeddingClient); ok {
				embedder = e
			}
		}
	}
	guardrails := NewGuardrailService(log)
	return &ChatService{
		PlanRepo:         planRepo,
		ConversationRepo: conversationRepo,
		AnalyticsRepo:    analyticsRepo,
		ActionRepo:       actionRepo,
		SourceRepo:       sourceRepo,
		HandoffRepo:      handoffRepo,
		Factory:          factory,
		Embedder:         embedder,
		QC:               qc,
		Log:              log,
		Guardrails:       guardrails,
		Quota:            NewQuotaEnforcer(usageRepo),
		Context:          NewChatContextBuilder(guardrails),
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
	plan, err := s.PlanRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get plan")
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}

	// 2. Validate and adjust model
	if len(plan.Limits.ChatAllowedModels) > 0 {
		allowed := false
		for _, m := range plan.Limits.ChatAllowedModels {
			if m == req.Chatbot.Model {
				allowed = true
				break
			}
		}
		// If requested model is not allowed, default to the first allowed model
		if !allowed {
			req.Chatbot.Model = plan.Limits.ChatAllowedModels[0]
		}
	}

	// 3. Token Check & Reservation (Atomic) - delegated to QuotaEnforcer
	maxMonthlyTokens := plan.Limits.ChatMaxMonthlyTokens
	var estimatedTokens int

	if maxMonthlyTokens > 0 {
		estimatedTokens = req.Chatbot.MaxTokens
		if estimatedTokens <= 0 {
			estimatedTokens = GetDefaultTokenEstimate()
		}

		// Reserve tokens optimistically
		if err = s.Quota.ReserveTokens(ctx, req.UserID, estimatedTokens, maxMonthlyTokens); err != nil {
			return nil, err
		}
	}

	// 4. Process Chat
	result, err := s.ProcessChat(ctx, req.ChatRequest, req.Chatbot, models.RAGConfig{
		TopK:             plan.Limits.ChatRAGTopK,
		MaxContextTokens: plan.Limits.ChatRAGMaxContextTokens,
	})

	// 5. Adjust Token Usage - delegated to QuotaEnforcer
	if maxMonthlyTokens > 0 {
		if err != nil {
			// On error, refund the reserved tokens
			s.Quota.RefundTokens(context.Background(), req.UserID, estimatedTokens)
		} else {
			// Adjust based on actual usage
			s.Quota.AdjustTokens(context.Background(), req.UserID, estimatedTokens, result.TokensUsed)
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
	if plan, err := s.PlanRepo.GetByUserID(ctx, bot.UserID); err == nil && plan != nil {
		guardrailsCfg = &models.GuardrailsConfig{
			CanCustomizeThresholds: plan.Limits.GuardrailsCanCustomizeThresholds,
			CanUseSmartFallback:    plan.Limits.GuardrailsCanUseSmartFallback,
			CanUseEscalateFallback: plan.Limits.GuardrailsCanUseEscalateFallback,
			CanManageTopics:        plan.Limits.GuardrailsCanManageTopics,
			CanCustomizeMessages:   plan.Limits.GuardrailsCanCustomizeMessages,
		}
	}

	// Step 1: Initialize chat context - delegated to ChatContextBuilder
	cc := s.Context.Build(ctx, req, bot, ragConfig, guardrailsCfg)

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
