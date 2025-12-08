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
	"github.com/onurceri/botla-co/pkg/storage"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := qdrantClient.EnsureEmbeddingsCollection(ctx); err != nil {
		log.Error("qdrant_ensure_collection_failed", map[string]any{"error": err.Error()})
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

	mux := buildMux(cfg, pool, log, q, storageService)
	origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	rl := middleware.NewRateLimiterFromEnv()
	handler := middleware.RequestLogger(log)(middleware.RateLimitMiddleware(rl)(mux))
	srv := newHTTPServer(cfg.PORT, cors(handler))
	startServerAsync(srv, log, cfg.PORT)
	waitForShutdownSignal()

	// Graceful shutdown
	schedulerCancel()
	refreshScheduler.Stop()
	shutdownServer(srv, log, pool)
}

func buildMux(cfg *config.Config, pool *sql.DB, log *logger.Logger, q *processing.SourceQueue, storageService storage.StorageService) *http.ServeMux {
	mux := http.NewServeMux()
	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	mux.HandleFunc("/health", hh.Health)
	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET}
	mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
	mux.HandleFunc("/api/v1/auth/refresh", ah.RefreshHandler)
	mux.HandleFunc("/api/v1/auth/logout", ah.LogoutHandler)
	mux.Handle("/api/v1/protected", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(handlers.ProtectedHandler)))
	mh := &handlers.MeHandlers{DB: pool}
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))
	ch := &handlers.ChatbotHandlers{DB: pool, Cfg: cfg}
	mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(ch.ListOrCreate)))
	// Sources handler
	sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: storageService, Log: log}
	// Chat service
	factory := rag.NewClientFactory()
	oaiClient, _ := rag.NewOpenAIClientFromEnv()
	qdClient, _ := rag.NewQdrantClientFromEnv()
	chatSvc := services.NewChatService(pool, factory, oaiClient, qdClient, log)
	chh := &handlers.ChatHandlers{DB: pool, ChatService: chatSvc}
	// Composite handler under /api/v1/chatbots/
	// Per-route limiter for sources endpoints
	rlSources := middleware.NewRateLimiterFromEnvWithPrefix("SOURCES")
	// Pending URLs handler
	puh := &handlers.PendingURLsHandlers{DB: pool, Log: log}
	mux.Handle("/api/v1/chatbots/", chatbotsDispatchHandlerWithSourcesRL(cfg.JWT_SECRET, ch, sh, chh, puh, rlSources))

	mux.Handle("/api/v1/public/chatbots/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/public/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			handlers.PublicChat(pool)(w, r)
			return
		}
		handlers.PublicChatbotConfig(pool)(w, r)
	}))

	// Feedback (protected)
	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))

	// Source status, delete, and refresh
	mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.RateLimitMiddleware(rlSources)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/refresh") {
			sh.RefreshSource(w, r)
			return
		}
		sh.GetSourceStatusOrDelete(w, r)
	}))))

	anh := &handlers.AnalyticsHandlers{DB: pool}
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(anh.GetAnalytics)))

	return mux
}

func chatbotsDispatchHandlerWithSourcesRL(secret string, ch *handlers.ChatbotHandlers, sh *handlers.SourcesHandlers, chh *handlers.ChatHandlers, puh *handlers.PendingURLsHandlers, rlSources *middleware.RateLimiter) http.Handler {
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
		// Sitemap discovery endpoint
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sitemap/discover") {
			middleware.RateLimitMiddleware(rlSources)(http.HandlerFunc(sh.DiscoverSitemap)).ServeHTTP(w, r)
			return
		}
		// Bulk sources creation endpoint
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sources/bulk") {
			middleware.RateLimitMiddleware(rlSources)(http.HandlerFunc(sh.BulkCreateSources)).ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sources") {
			middleware.RateLimitMiddleware(rlSources)(http.HandlerFunc(sh.ChatbotSources)).ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			chh.Chat(w, r)
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

// Backward-compatible dispatcher used by tests
func chatbotsDispatchHandler(secret string, ch *handlers.ChatbotHandlers, sh *handlers.SourcesHandlers, chh *handlers.ChatHandlers, puh *handlers.PendingURLsHandlers) http.Handler {
	rlSources := middleware.NewRateLimiterFromEnvWithPrefix("SOURCES")
	return chatbotsDispatchHandlerWithSourcesRL(secret, ch, sh, chh, puh, rlSources)
}
