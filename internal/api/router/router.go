package router

import (
	"database/sql"
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

// New creates and configures the main HTTP handler for the application.
func New(cfg *config.Config, pool *sql.DB, log *logger.Logger, q *processing.SourceQueue, storageService storage.StorageService) *http.ServeMux {
	mux := http.NewServeMux()

	// Services
	orgSvc := services.NewOrganizationService(pool, log)
	workspaceSvc := services.NewWorkspaceService(pool, log)
	analyticsSvc := services.NewAnalyticsService(pool, log)

	factory := rag.NewClientFactory(cfg)
	oaiClient, _ := rag.NewOpenAIClientFromEnv()
	qdClient, _ := rag.NewQdrantClientFromEnv()
	chatSvc := services.NewChatService(pool, factory, oaiClient, qdClient, log)

	// Handlers
	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mh := &handlers.MeHandlers{DB: pool}
	plh := &handlers.PlanHandlers{DB: pool}
	uh := &handlers.UsageHandlers{DB: pool}
	ch := &handlers.ChatbotHandlers{DB: pool, Cfg: cfg}
	sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: storageService, Log: log}
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc}
	puh := &handlers.PendingURLsHandlers{DB: pool, Queue: q, Log: log}
	acth := &handlers.ActionHandlers{DB: pool}
	hoh := &handlers.HandoffHandlers{DB: pool, Log: log}
	anh := &handlers.AnalyticsHandlers{DB: pool, AnalyticsService: analyticsSvc, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	sugh := &handlers.SuggestionsHandlers{DB: pool, Log: log}
	ph := &handlers.PublicHandlers{DB: pool, ChatService: chatSvc}
	oh := &handlers.OrganizationHandlers{OrgService: orgSvc, DB: pool}
	wh := &handlers.WorkspaceHandlers{WorkspaceService: workspaceSvc}

	// Health
	mux.HandleFunc("/health", hh.Health)

	// Auth
	registerAuthRoutes(mux, ah, cfg.JWT_SECRET)

	// User
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))
	mux.Handle("/api/v1/me/plan", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(plh.GetPlan)))
	mux.Handle("/api/v1/me/usage", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(uh.GetUsage)))

	// Chatbots (List/Create)
	mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.ExtractTenantContext()(http.HandlerFunc(ch.ListOrCreate))))

	// Chatbots Dispatch (Sub-routes)
	mux.Handle("/api/v1/chatbots/", chatbotsDispatchHandler(cfg.JWT_SECRET, ch, sh, chh, puh, acth, hoh, anh, sugh))

	// Public Routes
	registerPublicRoutes(mux, cfg.JWT_SECRET, hoh, ph, pool)

	// Feedback
	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))

	// Sources
	registerSourceRoutes(mux, cfg.JWT_SECRET, sh)

	// Analytics
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.ExtractTenantContext()(http.HandlerFunc(anh.GetAnalytics))))

	// Organizations & Workspaces
	registerOrgRoutes(mux, cfg.JWT_SECRET, orgSvc, oh, wh)

	return mux
}
