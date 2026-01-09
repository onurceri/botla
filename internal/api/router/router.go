package router

import (
	"database/sql"
	"net/http"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/internal/processing"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/workers"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/redis/go-redis/v9"
)

// New creates and configures the main HTTP handler for the application.
func New(cfg *config.Config, pool *sql.DB, log *logger.Logger, q *processing.SourceQueue, storageService storage.StorageService, qdClient *rag.QdrantClient, redisClient *redis.Client, workerPool *workers.WorkerPool) *http.ServeMux {
	mux := http.NewServeMux()

	// Repositories
	actionRepo := repository.NewPostgresActionRepo(pool)
	chatbotRepo := repository.NewPostgresChatbotRepo(pool)
	adminChatbotRepo := repository.NewPostgresAdminChatbotRepo(pool)
	planRepo := repository.NewPostgresPlanRepo(pool, redisClient)
	conversationRepo := repository.NewPostgresConversationRepo(pool)
	analyticsRepo := repository.NewPostgresAnalyticsRepo(pool)
	privacyRepo := repository.NewPostgresPrivacyRepo(pool)
	handoffRepo := repository.NewPostgresHandoffRepo(pool)
	sourceRepo := repository.NewPostgresSourceRepo(pool)
	userRepo := repository.NewPostgresUserRepo(pool)
	usageRepo := repository.NewPostgresUsageRepo(pool)
	queueRepo := repository.NewPostgresQueueRepo(pool)
	trainingJobRepo := repository.NewPostgresTrainingJobRepo(pool)
	suggestionJobRepo := repository.NewPostgresSuggestionJobRepo(pool)
	organizationRepo := repository.NewPostgresOrganizationRepo(pool)
	adminRepo := repository.NewPostgresAdminRepo(pool)

	// Services
	orgSvc := services.NewOrganizationService(pool, log)
	workspaceSvc := services.NewWorkspaceService(pool, log)
	analyticsSvc := services.NewAnalyticsService(analyticsRepo, log)
	chatbotSvc := services.NewChatbotService(chatbotRepo, planRepo, log)
	adminSvc := services.NewAdminService(adminRepo, log)
	privacySvc := services.NewPrivacyService(privacyRepo, log, storageService)

	factory := rag.NewClientFactory(cfg)
	// Initialize circuit breakers for LLM fault tolerance
	_ = factory.InitCircuitBreakers()

	oaiClient, _ := rag.NewOpenAIClient(cfg)
	// qdClient is passed in
	chatSvc := services.NewChatService(planRepo, conversationRepo, analyticsRepo, actionRepo, sourceRepo, handoffRepo, factory, oaiClient, qdClient, log)
	toolNameGenerator := rag.NewToolNameGenerator(oaiClient)

	// Handlers
	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg, Queue: q, LLMFactory: factory}
	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET, CookieSecure: cfg.CookieSecure, CookieDomain: cfg.CookieDomain, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mh := handlers.NewMeHandlers(userRepo, orgSvc)
	plh := handlers.NewPlanHandlers(userRepo, planRepo, pool)
	uh := &handlers.UsageHandlers{
		UserRepo:    userRepo,
		ChatbotRepo: chatbotRepo,
		UsageRepo:   usageRepo,
		Log:         log,
	}
	onbh := &handlers.OnboardingHandlers{DB: pool, UserRepo: userRepo}
	ch := &handlers.ChatbotHandlers{
		DB:               pool,
		Cfg:              cfg,
		ChatbotService:   chatbotSvc,
		ChatbotRepo:      chatbotRepo,
		PlanRepo:         planRepo,
		OrgService:       orgSvc,
		WorkspaceService: workspaceSvc,
		Logger:           log,
	}
	sh := &handlers.SourcesHandlers{
		DB:               pool,
		Queue:            q,
		Storage:          storageService,
		QdrantClient:     qdClient,
		Log:              log,
		WorkspaceService: workspaceSvc,
		OrgService:       orgSvc,
		PlanRepo:         planRepo,
		SourceRepo:       repository.NewPostgresSourceRepo(pool),
		UsageRepo:        usageRepo,
		ChatbotRepo:      chatbotRepo,
	}
	tjh := &handlers.TrainingJobHandlers{
		Log:              log,
		WorkspaceService: workspaceSvc,
		OrgService:       orgSvc,
		Queue:            q,
		TrainingJobRepo:  trainingJobRepo,
		SourceRepo:       repository.NewPostgresSourceRepo(pool),
		ChatbotRepo:      chatbotRepo,
	}
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc, WorkspaceService: workspaceSvc, OrgService: orgSvc, WorkerPool: workerPool, Logger: log, AnalyticsRepo: analyticsRepo, ChatbotRepo: chatbotRepo}
	puh := &handlers.PendingURLsHandlers{Queue: q, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc, PendingURLRepo: repository.NewPostgresPendingURLRepo(pool), SourceRepo: sourceRepo, ChatbotRepo: chatbotRepo}
	actionService := services.NewActionService(actionRepo, toolNameGenerator)
	acth := &handlers.ActionHandlers{ActionService: actionService, ChatbotRepo: chatbotRepo, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	hoh := &handlers.HandoffHandlers{DB: pool, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc, HandoffService: services.NewHandoffService(handoffRepo, conversationRepo, analyticsRepo, log), ChatbotRepo: chatbotRepo, ConversationRepo: conversationRepo, HandoffRepo: handoffRepo}
	anh := &handlers.AnalyticsHandlers{
		AnalyticsService: analyticsSvc,
		OrgService:       orgSvc,
		WorkspaceService: workspaceSvc,
		AnalyticsRepo:    analyticsRepo,
		ChatbotRepo:      chatbotRepo,
	}
	sugh := &handlers.SuggestionsHandlers{
		Log:               log,
		WorkspaceService:  workspaceSvc,
		OrgService:        orgSvc,
		SuggestionJobRepo: suggestionJobRepo,
		ChatbotRepo:       chatbotRepo,
	}
	ph := &handlers.PublicHandlers{ChatService: chatSvc, Log: log, ChatbotRepo: chatbotRepo, PlanRepo: planRepo, UsageRepo: usageRepo, AnalyticsRepo: analyticsRepo}
	oh := &handlers.OrganizationHandlers{OrgService: orgSvc, UserRepo: userRepo}
	wh := &handlers.WorkspaceHandlers{WorkspaceService: workspaceSvc}
	adh := handlers.NewAdminHandlers(adminSvc, userRepo, organizationRepo)
	adhh := handlers.NewAdminHealthHandlers(pool, redisClient, cfg)
	aqh := handlers.NewAdminQueueHandlers(adminSvc, queueRepo, repository.NewPostgresSourceRepo(pool))
	aeh := handlers.NewAdminErrorHandlers(adminRepo)
	aah := handlers.NewAdminAuditHandlers(adminRepo)
	aph := &handlers.PrivacyHandlers{DB: pool, PrivacyService: privacySvc, AdminService: adminSvc, PrivacyRepo: privacyRepo}
	uph := &handlers.UserPrivacyHandlers{DB: pool, PrivacyService: privacySvc, UserRepo: userRepo, PrivacyRepo: privacyRepo}

	// PlanService with Redis caching for all plan operations
	planSvc := services.NewPlanService(planRepo, redisClient)
	plansH := handlers.NewPlansHandlers(planSvc)

	// RAG service for admin operations (vector deletion)
	ragSubsystem := rag.NewRAGSubsystem(oaiClient, qdClient, oaiClient)
	ragSvc := services.NewRAGService(pool, ragSubsystem, log)
	// Queue wrapper for source processing
	queueWrapper := &services.Queue{SourceQueue: q}
	ach := handlers.NewAdminChatbotHandlers(adminChatbotRepo, adminSvc, ragSvc, queueWrapper)
	ash := handlers.NewAdminSourceHandlers(adminRepo, adminSvc, ragSvc, queueWrapper)

	// Health
	mux.HandleFunc("/health", hh.Health)

	// Auth
	registerAuthRoutes(mux, ah, cfg.JWT_SECRET)

	// User
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))
	mux.Handle("/api/v1/me/plan", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(plh.GetPlan)))
	mux.Handle("/api/v1/me/usage", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.ExtractTenantContext()(http.HandlerFunc(uh.GetUsage))))

	// User Privacy
	mux.Handle("GET /api/v1/me/privacy/consents", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.GetMyConsents)))
	mux.Handle("PATCH /api/v1/me/privacy/consents", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.UpdateMyConsents)))
	mux.Handle("POST /api/v1/me/privacy/export", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.RequestMyDataExport)))
	mux.Handle("POST /api/v1/me/privacy/correction", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.RequestDataCorrection)))
	mux.Handle("POST /api/v1/me/privacy/delete", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.RequestAccountDeletion)))
	mux.Handle("GET /api/v1/me/privacy/requests/{id}", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.GetMyPrivacyRequest)))
	mux.Handle("GET /api/v1/me/privacy/requests/{id}/download", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.DownloadMyPrivacyExport)))
	mux.Handle("GET /api/v1/me/privacy/exports/{id}/download", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uph.DownloadMyDataExport)))

	// Onboarding
	mux.Handle("GET /api/v1/me/onboarding", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.GetOnboardingState)))
	mux.Handle("PUT /api/v1/me/onboarding", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.UpdateOnboardingState)))
	mux.Handle("POST /api/v1/me/onboarding/skip", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.SkipOnboarding)))
	mux.Handle("POST /api/v1/me/onboarding/complete", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(onbh.CompleteOnboarding)))

	// Chatbots (List/Create)
	mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.ExtractTenantContext()(http.HandlerFunc(ch.ListOrCreate))))

	// Chatbots Dispatch (Sub-routes)
	mux.Handle("/api/v1/chatbots/", ChatbotsDispatchHandler(cfg.JWT_SECRET, ch, sh, chh, puh, acth, hoh, anh, sugh))

	// Public Routes
	registerPublicRoutes(mux, cfg.JWT_SECRET, hoh, ph, chatbotRepo)

	// Public Plan Routes (no auth required)
	mux.HandleFunc("GET /api/v1/plans", plansH.GetAllPlans)
	mux.HandleFunc("GET /api/v1/plans/{code}", plansH.GetPlanByCode)

	// Feedback
	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))

	// Sources
	RegisterSourceRoutes(mux, cfg.JWT_SECRET, sh, tjh)

	// Analytics
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.ExtractTenantContext()(http.HandlerFunc(anh.GetAnalytics))))

	// Organizations & Workspaces
	registerOrgRoutes(mux, cfg.JWT_SECRET, orgSvc, oh, wh)

	// Admin
	RegisterAdminRoutes(mux, adh, adhh, aqh, aeh, aah, aph, ach, ash, cfg.JWT_SECRET)

	// OpenAPI Spec
	mux.HandleFunc("GET /api/openapi.yaml", handlers.ServeOpenAPI)

	return mux
}
