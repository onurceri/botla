package integration

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/ratelimit"
	"github.com/onurceri/botla-co/pkg/storage"
)

func NewTestMux(cfg *config.Config, pool *sql.DB, vs handlers.VectorStore) http.Handler {
	mux := http.NewServeMux()
	log := logger.New("INFO")

	// Initialize rate limiter (in-memory for tests)
	rlConfig := ratelimit.NewConfigFromEnv()
	globalLimiter := ratelimit.NewMemoryLimiter(rlConfig.Global)
	rl := middleware.NewRateLimiter(globalLimiter, nil, rlConfig) // nil Redis client for tests

	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	mux.Handle("/health", middleware.RateLimitMiddleware(rl)(http.HandlerFunc(hh.Health)))

	// Org Service
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
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))
	mux.Handle("/api/v1/me/plan", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(ph.GetPlan)))

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
		DB:             pool,
		VectorStore:    vs,
		ChatbotService: services.NewChatbotService(pool, log),
	}
	// Add ExtractTenantContext to support X-Workspace-ID header
	mux.Handle("/api/v1/chatbots", protected(middleware.ExtractTenantContext()(http.HandlerFunc(ch.ListOrCreate))))
	memStore := storage.NewMemoryStorage()
	q, _ := processing.StartSourceQueue(pool, memStore)
	sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: memStore}
	factory := rag.NewClientFactory(cfg)
	chatSvc := services.NewChatService(pool, factory, nil, nil, log) // nil embedder/qc -> lazy init
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc}
	acth := &handlers.ActionHandlers{DB: pool}
	handh := &handlers.HandoffHandlers{DB: pool, Log: log}
	puh := &handlers.PendingURLsHandlers{DB: pool, Log: log, Queue: q}

	mux.Handle("/api/v1/chatbots/", protected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.Contains(r.URL.Path, "/pending-urls") {
			if strings.HasSuffix(r.URL.Path, "/pending-urls/approve") {
				puh.ApprovePendingURLs(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/pending-urls/reject") {
				puh.RejectPendingURLs(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/pending-urls/clear") {
				puh.ClearPendingURLs(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/pending-urls") {
				puh.ListPendingURLs(w, r)
				return
			}
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			chh.Chat(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.Contains(r.URL.Path, "/actions") {
			acth.Dispatch(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.Contains(r.URL.Path, "/handoff-requests") {
			if r.Method == http.MethodGet {
				handh.ListHandoffRequests(w, r)
				return
			}
			if r.Method == http.MethodPatch {
				handh.UpdateHandoffRequest(w, r)
				return
			}
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/analytics/trends") {
			anh.GetChatbotAnalyticsTrends(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/analytics/overview") {
			anh.GetChatbotAnalyticsOverview(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/analytics/sources") {
			anh.GetSourceUsage(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/basic-info") && r.Method == http.MethodPut {
			ch.UpdateBasicInfo(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/appearance") && r.Method == http.MethodPut {
			ch.UpdateAppearance(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/model") && r.Method == http.MethodPut {
			ch.UpdateModelSettings(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/security") && r.Method == http.MethodPut {
			ch.UpdateSecuritySettings(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/guardrails") && r.Method == http.MethodPut {
			ch.UpdateGuardrails(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/handoff") && r.Method == http.MethodPut {
			ch.UpdateHandoff(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/refresh") && r.Method == http.MethodPut {
			ch.UpdateRefresh(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/scraping") && r.Method == http.MethodPut {
			ch.UpdateScrapingConfig(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sources") && !strings.Contains(r.URL.Path, "/analytics/") {
			sh.ChatbotSources(w, r)
			return
		}
		ch.ByID(w, r)
	})))
	mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/refresh") {
			sh.RefreshSource(w, r)
			return
		}
		sh.GetSourceStatusOrDelete(w, r)
	})))
	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))
	mux.Handle("/api/v1/public/chatbots/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/public/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/handoff") {
			handh.PublicRequestHandoff(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			ph := &handlers.PublicHandlers{DB: pool, ChatService: chatSvc}
			ph.PublicChat(w, r)
			return
		}
		handlers.PublicChatbotConfig(pool)(w, r)
	}))
	// mux.Handle("/api/public/", http.HandlerFunc(handh.PublicRequestHandoff))
	origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	return cors(mux)
}
