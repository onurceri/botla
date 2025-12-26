package integration

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/api/router"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/ratelimit"
	"github.com/onurceri/botla-co/pkg/storage"
)

func NewTestMux(cfg *config.Config, pool *sql.DB, vs handlers.VectorStore, llm rag.LLMClient, vc rag.VectorClient) (http.Handler, *processing.SourceQueue) {
	mux := http.NewServeMux()
	log := logger.New("INFO")

	// Initialize rate limiter (in-memory for tests)
	rlConfig := ratelimit.NewConfigFromEnv()
	globalLimiter := ratelimit.NewMemoryLimiter(rlConfig.Global)
	rl := middleware.NewRateLimiter(globalLimiter, nil, rlConfig) // nil Redis client for tests

	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	mux.Handle("/health", middleware.RateLimitMiddleware(rl)(http.HandlerFunc(hh.Health)))

	// Services
	orgSvc := services.NewOrganizationService(pool, log)
	workspaceSvc := services.NewWorkspaceService(pool, log)
	analyticsSvc := services.NewAnalyticsService(pool, log)

	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
	mux.HandleFunc("/api/v1/auth/refresh", ah.RefreshHandler)
	mux.HandleFunc("/api/v1/auth/logout", ah.LogoutHandler)
	mux.Handle("/api/v1/protected", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(handlers.ProtectedHandler)))

	mh := &handlers.MeHandlers{DB: pool}
	ph := &handlers.PlanHandlers{DB: pool}
	onbh := &handlers.OnboardingHandlers{DB: pool}
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))
	mux.Handle("/api/v1/me/plan", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(ph.GetPlan)))

	// Onboarding
	mux.Handle("GET /api/v1/me/onboarding", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.GetOnboardingState)))
	mux.Handle("PUT /api/v1/me/onboarding", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.UpdateOnboardingState)))
	mux.Handle("POST /api/v1/me/onboarding/skip", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.SkipOnboarding)))
	mux.Handle("POST /api/v1/me/onboarding/complete", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.CompleteOnboarding)))

	anh := &handlers.AnalyticsHandlers{DB: pool, AnalyticsService: analyticsSvc, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(anh.GetAnalytics)))

	// Organization routes
	oh := &handlers.OrganizationHandlers{OrgService: orgSvc, DB: pool}
	wh := &handlers.WorkspaceHandlers{WorkspaceService: workspaceSvc}

	auth := middleware.AuthMiddleware(cfg.JWT_SECRET)
	planLoader := middleware.PlanLoaderMiddleware(pool, log)
	rateLimit := middleware.RateLimitMiddleware(rl)

	// Protected middleware chain
	protected := func(h http.Handler) http.Handler {
		return auth(planLoader(rateLimit(h)))
	}

	requireMember := middleware.RequireOrganizationAccess(orgSvc, "member")
	requireAdmin := middleware.RequireOrganizationAccess(orgSvc, "admin")
	requireOwner := middleware.RequireOrganizationAccess(orgSvc, "owner")

	mux.Handle("GET /api/v1/organizations", protected(http.HandlerFunc(oh.ListOrCreate)))
	mux.Handle("POST /api/v1/organizations", protected(http.HandlerFunc(oh.ListOrCreate)))
	mux.Handle("GET /api/v1/organizations/{id}", protected(requireMember(http.HandlerFunc(oh.GetOrganization))))
	mux.Handle("PATCH /api/v1/organizations/{id}", protected(requireOwner(http.HandlerFunc(oh.UpdateOrganization))))
	mux.Handle("DELETE /api/v1/organizations/{id}", protected(requireOwner(http.HandlerFunc(oh.DeleteOrganization))))

	mux.Handle("GET /api/v1/organizations/{id}/workspaces", protected(requireMember(http.HandlerFunc(wh.Workspaces))))
	mux.Handle("POST /api/v1/organizations/{id}/workspaces", protected(requireAdmin(http.HandlerFunc(wh.Workspaces))))
	mux.Handle("PATCH /api/v1/organizations/{id}/workspaces/{wsID}", protected(requireAdmin(http.HandlerFunc(wh.UpdateWorkspace))))
	mux.Handle("DELETE /api/v1/organizations/{id}/workspaces/{wsID}", protected(requireAdmin(http.HandlerFunc(wh.DeleteWorkspace))))

	mux.Handle("GET /api/v1/organizations/{id}/members", protected(requireMember(http.HandlerFunc(oh.GetMembers))))
	mux.Handle("POST /api/v1/organizations/{id}/members", protected(requireAdmin(http.HandlerFunc(oh.AddMember))))
	mux.Handle("PATCH /api/v1/organizations/{id}/members/{userID}", protected(requireAdmin(http.HandlerFunc(oh.UpdateMemberRole))))
	mux.Handle("DELETE /api/v1/organizations/{id}/members/{userID}", protected(requireAdmin(http.HandlerFunc(oh.RemoveMember))))

	ch := &handlers.ChatbotHandlers{
		DB:               pool,
		VectorStore:      vs,
		ChatbotService:   services.NewChatbotService(pool, log),
		OrgService:       orgSvc,
		WorkspaceService: workspaceSvc,
	}
	// Add ExtractTenantContext to support X-Workspace-ID header
	mux.Handle("GET /api/v1/chatbots", protected(middleware.ExtractTenantContext()(http.HandlerFunc(ch.ListOrCreate))))
	mux.Handle("POST /api/v1/chatbots", protected(middleware.ExtractTenantContext()(http.HandlerFunc(ch.ListOrCreate))))
	memStore := storage.NewMemoryStorage()

	// Use real clients if mocks are nil
	var actualLLM = llm
	if actualLLM == nil {
		if c, err := rag.NewOpenAIClient(cfg); err == nil {
			actualLLM = c
		}
	}
	var actualVC = vc
	if actualVC == nil {
		if c, err := rag.NewQdrantClientFromEnv(); err == nil {
			actualVC = c
		}
	}

	q, err := processing.StartSourceQueue(pool, memStore, actualLLM, actualVC)
	if err != nil {
		logger.New("WARN").Warn("failed to start source queue in testmux", map[string]any{"error": err})
	}
	sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: memStore, QdrantClient: actualVC, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	factory := rag.NewClientFactory(cfg)
	if llm != nil {
		factory.RegisterClient("openai", llm)
		factory.RegisterClient("openrouter", llm)
	}
	chatSvc := services.NewChatService(pool, factory, nil, actualVC, log)
	if llmEmbed, ok := actualLLM.(rag.EmbeddingClient); ok {
		chatSvc.Embedder = llmEmbed
	}
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc, WorkspaceService: workspaceSvc, OrgService: orgSvc}

	// Create mock tool name generator for tests
	mockClient := &mockToolsClient{}
	tng := rag.NewToolNameGenerator(mockClient)
	acth := &handlers.ActionHandlers{DB: pool, ToolNameGenerator: tng, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	hoh := &handlers.HandoffHandlers{DB: pool, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	puh := &handlers.PendingURLsHandlers{DB: pool, Log: log, Queue: q, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	sugh := &handlers.SuggestionsHandlers{DB: pool, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc}

	// Actions routes
	// Note: Actions routes are now handled by ChatbotsDispatchHandler

	// Chatbots Dispatch (Sub-routes)
	mux.Handle("/api/v1/chatbots/", protected(middleware.ExtractTenantContext()(router.ChatbotsRawHandler(ch, sh, chh, puh, acth, hoh, anh, sugh))))

	// Explicitly handle /api/v1/chatbots/{id} (no trailing slash)
	mux.Handle("/api/v1/chatbots/{id}", protected(http.HandlerFunc(ch.ByID)))

	// Sources
	router.RegisterSourceRoutes(mux, cfg.JWT_SECRET, sh)

	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))
	mux.Handle("/api/v1/public/chatbots/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/public/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/handoff") {
			hoh.PublicRequestHandoff(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			ph := &handlers.PublicHandlers{DB: pool, ChatService: chatSvc}
			ph.PublicChat(w, r)
			return
		}
		handlers.PublicChatbotConfig(pool)(w, r)
	}))
	// Admin
	adminSvc := services.NewAdminService(pool, log)
	privacySvc := services.NewPrivacyService(pool, log, memStore)

	// User Privacy
	uph := &handlers.UserPrivacyHandlers{DB: pool, PrivacyService: privacySvc}
	mux.Handle("GET /api/v1/me/privacy/consents", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.GetMyConsents)))
	mux.Handle("PATCH /api/v1/me/privacy/consents", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.UpdateMyConsents)))
	mux.Handle("POST /api/v1/me/privacy/export", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.RequestMyDataExport)))
	mux.Handle("POST /api/v1/me/privacy/correction", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.RequestDataCorrection)))
	mux.Handle("POST /api/v1/me/privacy/delete", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.RequestAccountDeletion)))
	mux.Handle("GET /api/v1/me/privacy/requests/{id}", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.GetMyPrivacyRequest)))
	mux.Handle("GET /api/v1/me/privacy/requests/{id}/download", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.DownloadMyPrivacyExport)))
	mux.Handle("GET /api/v1/me/privacy/exports/{id}/download", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.DownloadMyDataExport)))

	adh := handlers.NewAdminHandlers(pool, adminSvc)
	adhh := handlers.NewAdminHealthHandlers(pool, nil, cfg) // nil Redis client for tests
	aqh := handlers.NewAdminQueueHandlers(pool, adminSvc)
	aeh := handlers.NewAdminErrorHandlers(pool)
	aah := handlers.NewAdminAuditHandlers(pool)
	aph := &handlers.PrivacyHandlers{DB: pool, PrivacyService: privacySvc, AdminService: adminSvc}

	// RAG service and queue wrapper for admin chatbot/source handlers
	ragSvc := services.NewRAGService(pool, actualVC, log)
	queueWrapper := &services.Queue{SourceQueue: q}
	ach := handlers.NewAdminChatbotHandlers(pool, adminSvc, ragSvc, queueWrapper)
	ash := handlers.NewAdminSourceHandlers(pool, adminSvc, ragSvc, queueWrapper)
	router.RegisterAdminRoutes(mux, adh, adhh, aqh, aeh, aah, aph, ach, ash, cfg.JWT_SECRET)

	origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	return cors(mux), q
}
