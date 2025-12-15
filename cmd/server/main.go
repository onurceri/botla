package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/ratelimit"
	"github.com/onurceri/botla-co/pkg/storage"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.New("INFO")
	pool, err := db.New(cfg)
	if err != nil {
		log.Error("db_init_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	// Initialize Qdrant
	qdrantClient, err := rag.NewQdrantClientFromEnv()
	if err != nil {
		log.Error("qdrant_init_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	// Ensure embeddings collection exists
	if err := ensureQdrantCollection(qdrantClient, log); err != nil {
		os.Exit(1)
	}
	log.Info("qdrant_collection_ready", nil)

	// Initialize storage service
	var storageService storage.StorageService
	if cfg.R2_ACCOUNT_ID != "" {
		var err error
		storageService, err = storage.NewR2Storage(cfg.R2_ACCOUNT_ID, cfg.R2_ACCESS_KEY_ID, cfg.R2_SECRET_ACCESS_KEY, cfg.R2_BUCKET_NAME)
		if err != nil {
			log.Error("storage_init_failed", map[string]any{"error": err.Error()})
		}
	}

	// Start source processing queue
	q, _ := processing.StartSourceQueue(pool, storageService)

	// Start refresh scheduler
	refreshScheduler := services.NewRefreshScheduler(pool, q, log)
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	refreshScheduler.Start(schedulerCtx)

	// Initialize Redis client for rate limiting
	var redisClient *redis.Client
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Warn("redis_parse_url_failed", map[string]any{"error": err.Error()})
		} else {
			redisClient = redis.NewClient(opts)
			// Test connection
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := redisClient.Ping(ctx).Err(); err != nil {
				log.Warn("redis_ping_failed", map[string]any{"error": err.Error()})
				redisClient.Close()
				redisClient = nil
			} else {
				log.Info("redis_connected", nil)
			}
			cancel()
		}
	}

	// Initialize rate limiter (Redis or in-memory fallback)
	rlConfig := ratelimit.NewConfigFromEnv()
	var globalLimiter, userLimiter ratelimit.Limiter
	if redisClient != nil {
		globalLimiter = ratelimit.NewRedisLimiter(redisClient, rlConfig.Global)
		userLimiter = ratelimit.NewRedisLimiter(redisClient, rlConfig.User)
		log.Info("rate_limiter_initialized", map[string]any{"backend": "redis"})
	} else {
		globalLimiter = ratelimit.NewMemoryLimiter(rlConfig.Global)
		userLimiter = ratelimit.NewMemoryLimiter(rlConfig.User)
		log.Warn("rate_limiter_using_memory", map[string]any{"message": "Redis unavailable, using in-memory rate limiter (not suitable for production)"})
	}
	rl := middleware.NewRateLimiter(globalLimiter, userLimiter, rlConfig)

	mux := buildMux(cfg, pool, log, q, storageService)
	origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	handler := middleware.RecoveryMiddleware(log)(middleware.RequestLogger(log)(middleware.RateLimitMiddleware(rl)(mux)))
	srv := newHTTPServer(cfg.PORT, cors(handler))
	startServerAsync(srv, log, cfg.PORT)
	waitForShutdownSignal()

	// Graceful shutdown
	schedulerCancel()
	refreshScheduler.Stop()
	// Close rate limiters
	globalLimiter.Close()
	userLimiter.Close()
	if redisClient != nil {
		redisClient.Close()
	}
	shutdownServer(srv, log, pool)
}

func buildMux(cfg *config.Config, pool *sql.DB, log *logger.Logger, q *processing.SourceQueue, storageService storage.StorageService) *http.ServeMux {
	mux := http.NewServeMux()
	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	mux.HandleFunc("/health", hh.Health)
	// Organization service (needed by auth for auto-workspace creation)
	orgSvc := services.NewOrganizationService(pool, log)
	workspaceSvc := services.NewWorkspaceService(pool, log)
	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET, OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
	mux.HandleFunc("/api/v1/auth/refresh", ah.RefreshHandler)
	mux.HandleFunc("/api/v1/auth/logout", ah.LogoutHandler)
	mux.Handle("/api/v1/protected", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(handlers.ProtectedHandler)))
	mh := &handlers.MeHandlers{DB: pool}
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))
	ch := &handlers.ChatbotHandlers{DB: pool, Cfg: cfg}
	mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.ExtractTenantContext()(http.HandlerFunc(ch.ListOrCreate))))
	// Sources handler
	sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: storageService, Log: log}
	// Chat service
	factory := rag.NewClientFactory(cfg)
	oaiClient, _ := rag.NewOpenAIClientFromEnv()
	qdClient, _ := rag.NewQdrantClientFromEnv()
	chatSvc := services.NewChatService(pool, factory, oaiClient, qdClient, log)
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc}
	// Composite handler under /api/v1/chatbots/
	// Note: Sources rate limiting is now handled by tiered rate limiter in middleware
	// Pending URLs handler
	puh := &handlers.PendingURLsHandlers{DB: pool, Queue: q, Log: log}
	// Action handler
	acth := &handlers.ActionHandlers{DB: pool}
	// Handoff handler
	hoh := &handlers.HandoffHandlers{DB: pool, Log: log}
	// Analytics handler for chatbot-specific routes
	anhSpec := &handlers.AnalyticsHandlers{DB: pool, AnalyticsService: services.NewAnalyticsService(pool, log), OrgService: orgSvc, WorkspaceService: workspaceSvc}
	// Suggestions handler
	sugh := &handlers.SuggestionsHandlers{DB: pool, Log: log}
	mux.Handle("/api/v1/chatbots/", chatbotsDispatchHandler(cfg.JWT_SECRET, ch, sh, chh, puh, acth, hoh, anhSpec, sugh))

	mux.Handle("/api/v1/public/chatbots/", middleware.OptionalAuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/public/chatbots/"
		// Public email submission for handoff: /api/v1/public/chatbots/:botId/handoff/:requestId/contact
		if strings.HasPrefix(r.URL.Path, p) && strings.Contains(r.URL.Path, "/handoff/") && strings.HasSuffix(r.URL.Path, "/contact") {
			hoh.PublicSubmitEmail(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/handoff") {
			// Public handoff request
			hoh.PublicRequestHandoff(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			// Public handlers
			ph := &handlers.PublicHandlers{DB: pool, ChatService: chatSvc}
			ph.PublicChat(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/feedback") {
			ph := &handlers.PublicHandlers{DB: pool, ChatService: chatSvc}
			ph.SubmitFeedback(w, r)
			return
		}
		handlers.PublicChatbotConfig(pool)(w, r)
	})))

	// Feedback (protected)
	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))

	// Source status, delete, and refresh
	mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/refresh") {
			sh.RefreshSource(w, r)
			return
		}
		sh.GetSourceStatusOrDelete(w, r)
	})))

	anh := &handlers.AnalyticsHandlers{DB: pool, AnalyticsService: services.NewAnalyticsService(pool, log), OrgService: orgSvc, WorkspaceService: workspaceSvc}
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.ExtractTenantContext()(http.HandlerFunc(anh.GetAnalytics))))

	// Organization routes
	oh := &handlers.OrganizationHandlers{OrgService: orgSvc, DB: pool}
	wh := &handlers.WorkspaceHandlers{WorkspaceService: workspaceSvc}
	auth := middleware.AuthMiddleware(cfg.JWT_SECRET)

	// Helper middlewares
	requireMember := middleware.RequireOrganizationAccess(orgSvc, "member")
	requireAdmin := middleware.RequireOrganizationAccess(orgSvc, "admin")
	requireOwner := middleware.RequireOrganizationAccess(orgSvc, "owner")

	// Org List/Create
	mux.Handle("GET /api/v1/organizations", auth(http.HandlerFunc(oh.ListOrCreate)))
	mux.Handle("POST /api/v1/organizations", auth(http.HandlerFunc(oh.ListOrCreate)))

	// Org Management
	mux.Handle("PATCH /api/v1/organizations/{id}", auth(requireOwner(http.HandlerFunc(oh.UpdateOrganization))))
	mux.Handle("DELETE /api/v1/organizations/{id}", auth(requireOwner(http.HandlerFunc(oh.DeleteOrganization))))

	// Workspaces
	mux.Handle("GET /api/v1/organizations/{id}/workspaces", auth(requireMember(http.HandlerFunc(wh.Workspaces))))
	mux.Handle("POST /api/v1/organizations/{id}/workspaces", auth(requireAdmin(http.HandlerFunc(wh.Workspaces))))
	mux.Handle("PATCH /api/v1/organizations/{id}/workspaces/{wsID}", auth(requireAdmin(http.HandlerFunc(wh.UpdateWorkspace))))
	mux.Handle("DELETE /api/v1/organizations/{id}/workspaces/{wsID}", auth(requireAdmin(http.HandlerFunc(wh.DeleteWorkspace))))

	// Members
	mux.Handle("GET /api/v1/organizations/{id}/members", auth(requireMember(http.HandlerFunc(oh.GetMembers))))
	mux.Handle("POST /api/v1/organizations/{id}/members", auth(requireAdmin(http.HandlerFunc(oh.AddMember))))
	mux.Handle("DELETE /api/v1/organizations/{id}/members/{userID}", auth(requireAdmin(http.HandlerFunc(oh.RemoveMember))))
	mux.Handle("PATCH /api/v1/organizations/{id}/members/{userID}", auth(requireAdmin(http.HandlerFunc(oh.UpdateMemberRole))))

	return mux
}

func chatbotsDispatchHandler(secret string, ch *handlers.ChatbotHandlers, sh *handlers.SourcesHandlers, chh *handlers.ChatHandlers, puh *handlers.PendingURLsHandlers, acth *handlers.ActionHandlers, hoh *handlers.HandoffHandlers, anh *handlers.AnalyticsHandlers, sugh *handlers.SuggestionsHandlers) http.Handler {
	return middleware.AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/chatbots/"
		// Pending URLs endpoints
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
		// Actions endpoint
		if strings.HasPrefix(r.URL.Path, p) && strings.Contains(r.URL.Path, "/actions") {
			acth.Dispatch(w, r)
			return
		}
		// Handoff requests endpoint
		if strings.HasPrefix(r.URL.Path, p) && strings.Contains(r.URL.Path, "/handoff-requests") {
			// Has request ID if path contains more segments after handoff-requests
			parts := strings.Split(r.URL.Path, "/")
			// /api/v1/chatbots/:id/handoff-requests/:requestId -> parts[6] is requestId
			if len(parts) >= 7 && parts[6] != "" {
				if r.Method == http.MethodGet {
					// GET specific request: GET /api/v1/chatbots/:id/handoff-requests/:requestId
					hoh.GetHandoffRequestDetail(w, r)
				} else {
					// Update specific request: PATCH /api/v1/chatbots/:id/handoff-requests/:requestId
					hoh.UpdateHandoffRequest(w, r)
				}
			} else {
				// List requests: GET /api/v1/chatbots/:id/handoff-requests
				hoh.ListHandoffRequests(w, r)
			}
			return
		}
		// Analytics endpoints
		if strings.HasPrefix(r.URL.Path, p) && strings.Contains(r.URL.Path, "/analytics") {
			if strings.HasSuffix(r.URL.Path, "/analytics/overview") {
				anh.GetChatbotAnalyticsOverview(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/analytics/trends") {
				anh.GetChatbotAnalyticsTrends(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/analytics/sources") {
				anh.GetSourceUsage(w, r)
				return
			}
		}
		// Suggestions regenerate endpoint
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/suggestions/regenerate") {
			sugh.RegenerateSuggestions(w, r)
			return
		}
		// Sitemap discovery endpoint
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sitemap/discover") {
			sh.DiscoverSitemap(w, r)
			return
		}
		// Bulk sources creation endpoint
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sources/bulk") {
			sh.BulkCreateSources(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sources") && !strings.Contains(r.URL.Path, "/analytics/") {
			sh.ChatbotSources(w, r)
			return
		}
		ch.ByID(w, r)
	}))
}

func newHTTPServer(port string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + port,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func startServerAsync(srv *http.Server, log *logger.Logger, port string) {
	go func() {
		log.Info("server_start", map[string]any{"port": port})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server_error", map[string]any{"error": err.Error()})
		}
	}()
}

func waitForShutdownSignal() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
}

func shutdownServer(srv *http.Server, log *logger.Logger, pool *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info("server_shutdown", map[string]any{})
	_ = srv.Shutdown(ctx)
	_ = pool.Close()
}

func ensureQdrantCollection(qdrantClient *rag.QdrantClient, log *logger.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := qdrantClient.EnsureEmbeddingsCollection(ctx); err != nil {
		log.Error("qdrant_ensure_collection_failed", map[string]any{"error": err.Error()})
		return err
	}
	return nil
}
