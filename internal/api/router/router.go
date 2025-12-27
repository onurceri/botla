package router

import (
	"database/sql"
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/internal/workers"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
	"github.com/redis/go-redis/v9"
)

// New creates and configures the main HTTP handler for the application.
func New(cfg *config.Config, pool *sql.DB, log *logger.Logger, q *processing.SourceQueue, storageService storage.StorageService, qdClient *rag.QdrantClient, redisClient *redis.Client, workerPool *workers.WorkerPool) *http.ServeMux {
	mux := http.NewServeMux()

	// Services
	orgSvc := services.NewOrganizationService(pool, log)
	workspaceSvc := services.NewWorkspaceService(pool, log)
	analyticsSvc := services.NewAnalyticsService(pool, log)
	chatbotSvc := services.NewChatbotService(pool, log)
	adminSvc := services.NewAdminService(pool, log)
	privacySvc := services.NewPrivacyService(pool, log, storageService)

	factory := rag.NewClientFactory(cfg)
	oaiClient, _ := rag.NewOpenAIClient(cfg)
	// qdClient is passed in
	chatSvc := services.NewChatService(pool, factory, oaiClient, qdClient, log)
	toolNameGenerator := rag.NewToolNameGenerator(oaiClient)

	// Handlers
	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mh := &handlers.MeHandlers{DB: pool}
	plh := &handlers.PlanHandlers{DB: pool}
	uh := &handlers.UsageHandlers{DB: pool}
	onbh := &handlers.OnboardingHandlers{DB: pool}
	ch := &handlers.ChatbotHandlers{
		DB:               pool,
		Cfg:              cfg,
		ChatbotService:   chatbotSvc,
		OrgService:       orgSvc,
		WorkspaceService: workspaceSvc,
		Logger:           log,
	}
	sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: storageService, QdrantClient: qdClient, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	tjh := &handlers.TrainingJobHandlers{DB: pool, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc, Queue: q}
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc, WorkspaceService: workspaceSvc, OrgService: orgSvc, WorkerPool: workerPool, Logger: log}
	puh := &handlers.PendingURLsHandlers{DB: pool, Queue: q, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	acth := &handlers.ActionHandlers{DB: pool, ToolNameGenerator: toolNameGenerator, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	hoh := &handlers.HandoffHandlers{DB: pool, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	anh := &handlers.AnalyticsHandlers{DB: pool, AnalyticsService: analyticsSvc, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	sugh := &handlers.SuggestionsHandlers{DB: pool, Log: log, WorkspaceService: workspaceSvc, OrgService: orgSvc}
	ph := &handlers.PublicHandlers{DB: pool, ChatService: chatSvc, Log: log}
	oh := &handlers.OrganizationHandlers{OrgService: orgSvc, DB: pool}
	wh := &handlers.WorkspaceHandlers{WorkspaceService: workspaceSvc}
	adh := handlers.NewAdminHandlers(pool, adminSvc)
	adhh := handlers.NewAdminHealthHandlers(pool, redisClient, cfg)
	aqh := handlers.NewAdminQueueHandlers(pool, adminSvc)
	aeh := handlers.NewAdminErrorHandlers(pool)
	aah := handlers.NewAdminAuditHandlers(pool)
	aph := &handlers.PrivacyHandlers{DB: pool, PrivacyService: privacySvc, AdminService: adminSvc}
	uph := &handlers.UserPrivacyHandlers{DB: pool, PrivacyService: privacySvc}

	// RAG service for admin operations (vector deletion)
	ragSvc := services.NewRAGService(pool, qdClient, log)
	// Queue wrapper for source processing
	queueWrapper := &services.Queue{SourceQueue: q}
	ach := handlers.NewAdminChatbotHandlers(pool, adminSvc, ragSvc, queueWrapper)
	ash := handlers.NewAdminSourceHandlers(pool, adminSvc, ragSvc, queueWrapper)

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
	registerPublicRoutes(mux, cfg.JWT_SECRET, hoh, ph, pool)

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

	return mux
}
