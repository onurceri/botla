package fixtures

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/internal/api/router"
	"github.com/onurceri/botla-app/internal/processing"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/scraper"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/workers"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/ratelimit"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/onurceri/botla-app/pkg/tokenizer"
	"github.com/onurceri/botla-app/pkg/urlutil"
)

func NewTestMux(cfg *config.Config, pool *sql.DB, vs handlers.VectorStore, llm rag.LLMClient, vc rag.VectorClient) (http.Handler, *processing.SourceQueue, *middleware.RateLimiter, *workers.WorkerPool, *handlers.SourcesHandlers, *scraper.MockScraper) {
	mux := http.NewServeMux()
	log := logger.New("INFO")

	// Initialize rate limiter (in-memory for tests)
	rlSettings := ratelimit.Settings{
		GlobalRequestsPerMinute: cfg.RateLimitGlobalRequestsPerMinute,
		GlobalWindowSeconds:     cfg.RateLimitGlobalWindowSeconds,
		UserRequestsPerMinute:   cfg.RateLimitUserRequestsPerMinute,
		UserWindowSeconds:       cfg.RateLimitUserWindowSeconds,
		EndpointChat:            cfg.RateLimitEndpointChat,
		EndpointSources:         cfg.RateLimitEndpointSources,
		AuthLogin:               cfg.RateLimitAuthLogin,
		AuthRegister:            cfg.RateLimitAuthRegister,
		AuthRefresh:             cfg.RateLimitAuthRefresh,
	}

	rlConfig := ratelimit.NewConfig(rlSettings)
	globalLimiter := ratelimit.NewMemoryLimiter(rlConfig.Global)
	rl := middleware.NewRateLimiter(globalLimiter, nil, rlConfig) // nil Redis client for tests

	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	mux.Handle("/health", middleware.RateLimitMiddleware(rl)(http.HandlerFunc(hh.Health)))

	// Repositories
	actionRepo := repository.NewPostgresActionRepo(pool)
	chatbotRepo := repository.NewPostgresChatbotRepo(pool)
	adminChatbotRepo := repository.NewPostgresAdminChatbotRepo(pool)
	adminRepo := repository.NewPostgresAdminRepo(pool)
	planRepo := repository.NewPostgresPlanRepo(pool, nil) // nil Redis for tests
	conversationRepo := repository.NewPostgresConversationRepo(pool)
	analyticsRepo := repository.NewPostgresAnalyticsRepo(pool)
	privacyRepo := repository.NewPostgresPrivacyRepo(pool)
	handoffRepo := repository.NewPostgresHandoffRepo(pool)
	userRepo := repository.NewPostgresUserRepo(pool)
	queueRepo := repository.NewPostgresQueueRepo(pool)
	organizationRepo := repository.NewPostgresOrganizationRepo(pool)
	sourceRepo := repository.NewPostgresSourceRepo(pool)
	trainingJobRepo := repository.NewPostgresTrainingJobRepo(pool)
	usageRepo := repository.NewPostgresUsageRepo(pool)

	// Services
	orgSvc := services.NewOrganizationService(pool, log)
	workspaceSvc := services.NewWorkspaceService(pool, log)
	analyticsSvc := services.NewAnalyticsService(analyticsRepo, log)

	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
	mux.HandleFunc("/api/v1/auth/refresh", ah.RefreshHandler)
	mux.HandleFunc("/api/v1/auth/logout", ah.LogoutHandler)
	mux.Handle("/api/v1/protected", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(handlers.ProtectedHandler)))

	mh := handlers.NewMeHandlers(userRepo, orgSvc)
	ph := handlers.NewPlanHandlers(userRepo, planRepo, pool)
	onbh := &handlers.OnboardingHandlers{DB: pool}
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))
	mux.Handle("/api/v1/me/plan", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(ph.GetPlan)))

	// PlanService with caching for public plan endpoints
	planSvc := services.NewPlanService(planRepo, nil) // nil Redis for tests
	plansH := handlers.NewPlansHandlers(planSvc)
	mux.HandleFunc("GET /api/v1/plans", plansH.GetAllPlans)
	mux.HandleFunc("GET /api/v1/plans/{code}", plansH.GetPlanByCode)

	// Onboarding
	mux.Handle("GET /api/v1/me/onboarding", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.GetOnboardingState)))
	mux.Handle("PUT /api/v1/me/onboarding", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.UpdateOnboardingState)))
	mux.Handle("POST /api/v1/me/onboarding/skip", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.SkipOnboarding)))
	mux.Handle("POST /api/v1/me/onboarding/complete", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.CompleteOnboarding)))

	anh := &handlers.AnalyticsHandlers{AnalyticsService: analyticsSvc, OrgService: orgSvc, WorkspaceService: workspaceSvc, AnalyticsRepo: analyticsRepo, ChatbotRepo: chatbotRepo}
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(anh.GetAnalytics)))

	// Organization routes
	oh := &handlers.OrganizationHandlers{OrgService: orgSvc, UserRepo: userRepo}
	wh := &handlers.WorkspaceHandlers{WorkspaceService: workspaceSvc}

	auth := middleware.AuthMiddleware(cfg.JWT_SECRET)
	planLoader := middleware.PlanLoaderMiddleware(planRepo, log)
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
		ChatbotService:   services.NewChatbotService(chatbotRepo, planRepo, log),
		OrgService:       orgSvc,
		WorkspaceService: workspaceSvc,
		Logger:           log,
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

	// Create mock scraper for tests
	ms := scraper.NewMockScraper()

	// Initialize tokenizer loader for tests (mock storage)
	tokLoader := tokenizer.NewLoader(memStore)

	q, err := processing.StartSourceQueue(trainingJobRepo, memStore, actualLLM, actualVC, ms, tokLoader, 2)
	if err != nil {
		logger.New("WARN").Warn("failed to start source queue in testmux", map[string]any{"error": err})
	}
	sh := &handlers.SourcesHandlers{
		DB:               pool,
		Queue:            q,
		Storage:          memStore,
		QdrantClient:     actualVC,
		WorkspaceService: workspaceSvc,
		OrgService:       orgSvc,
		SSRFValidator:    urlutil.NewSSRFValidator(true),
	}
	tjh := &handlers.TrainingJobHandlers{Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc, Queue: q, TrainingJobRepo: trainingJobRepo, SourceRepo: sourceRepo, ChatbotRepo: chatbotRepo}
	factory := rag.NewClientFactory(cfg)
	// Register the actual LLM client (either passed in or created from config)
	if actualLLM != nil {
		factory.RegisterClient("openai", actualLLM)
		factory.RegisterClient("openrouter", actualLLM)
	}
	chatSvc := services.NewChatService(planRepo, conversationRepo, analyticsRepo, actionRepo, sourceRepo, handoffRepo, factory, nil, actualVC, log)
	chatSvc.SyncAnalytics = true
	if llmEmbed, ok := actualLLM.(rag.EmbeddingClient); ok {
		chatSvc.Embedder = llmEmbed
	}
	workerPool := workers.NewWorkerPool(log, 5)
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc, WorkspaceService: workspaceSvc, OrgService: orgSvc, WorkerPool: workerPool, Logger: log}

	// Create mock tool name generator for tests
	mockClient := &MockToolsClient{}
	tng := rag.NewToolNameGenerator(mockClient)
	actionService := services.NewActionService(actionRepo, tng)
	acth := &handlers.ActionHandlers{ActionService: actionService, ChatbotRepo: chatbotRepo, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	handoffSvc := services.NewHandoffService(handoffRepo, conversationRepo, analyticsRepo, log)
	hoh := &handlers.HandoffHandlers{DB: pool, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc, HandoffService: handoffSvc}
	puh := &handlers.PendingURLsHandlers{Log: log, Queue: q, WorkspaceService: workspaceSvc, OrgService: orgSvc, PendingURLRepo: repository.NewPostgresPendingURLRepo(pool), SourceRepo: sourceRepo, ChatbotRepo: chatbotRepo}
	sugh := &handlers.SuggestionsHandlers{Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc, SuggestionJobRepo: repository.NewPostgresSuggestionJobRepo(pool), ChatbotRepo: chatbotRepo}

	// Actions routes
	// Note: Actions routes are now handled by ChatbotsDispatchHandler

	// Chatbots Dispatch (Sub-routes)
	mux.Handle("/api/v1/chatbots/", protected(middleware.ExtractTenantContext()(router.ChatbotsRawHandler(ch, sh, chh, puh, acth, hoh, anh, sugh))))

	// Explicitly handle /api/v1/chatbots/{id} (no trailing slash)
	mux.Handle("/api/v1/chatbots/{id}", protected(http.HandlerFunc(ch.ByID)))

	// Sources
	router.RegisterSourceRoutes(mux, cfg.JWT_SECRET, sh, tjh)

	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))
	mux.Handle("/api/v1/public/chatbots/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/public/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/handoff") {
			hoh.PublicRequestHandoff(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			ph := &handlers.PublicHandlers{ChatService: chatSvc, ChatbotRepo: chatbotRepo, PlanRepo: planRepo, UsageRepo: usageRepo, AnalyticsRepo: analyticsRepo}
			ph.PublicChat(w, r)
			return
		}
		handlers.PublicChatbotConfig(chatbotRepo)(w, r)
	}))
	// Admin
	adminSvc := services.NewAdminService(adminRepo, log)
	privacySvc := services.NewPrivacyService(privacyRepo, log, memStore)

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

	adh := handlers.NewAdminHandlers(adminSvc, userRepo, organizationRepo)
	adhh := handlers.NewAdminHealthHandlers(pool, nil, cfg) // nil Redis client for tests
	aqh := handlers.NewAdminQueueHandlers(adminSvc, queueRepo, sourceRepo)
	aeh := handlers.NewAdminErrorHandlers(adminRepo)
	aah := handlers.NewAdminAuditHandlers(adminRepo)
	aph := &handlers.PrivacyHandlers{DB: pool, PrivacyService: privacySvc, AdminService: adminSvc}

	// RAG service and queue wrapper for admin chatbot/source handlers
	var embedder rag.EmbeddingClient
	if actualLLM != nil {
		if e, ok := actualLLM.(rag.EmbeddingClient); ok {
			embedder = e
		}
	}
	ragSubsystem := rag.NewRAGSubsystem(embedder, actualVC, actualLLM)
	ragSvc := services.NewRAGService(pool, ragSubsystem, log)
	queueWrapper := &services.Queue{SourceQueue: q}
	ach := handlers.NewAdminChatbotHandlers(adminChatbotRepo, adminSvc, ragSvc, queueWrapper)
	ash := handlers.NewAdminSourceHandlers(adminRepo, adminSvc, ragSvc, queueWrapper)
	router.RegisterAdminRoutes(mux, adh, adhh, aqh, aeh, aah, aph, ach, ash, cfg.JWT_SECRET)

	origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	return cors(middleware.RequestID(mux)), q, rl, workerPool, sh, ms
}
