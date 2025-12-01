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
    mux := buildMux(cfg, pool, log)
    origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
    cors := middleware.CORSMiddlewareAllowOrigins(origins)
    rl := middleware.NewRateLimiterFromEnv()
    handler := middleware.RequestLogger(log)(middleware.RateLimitMiddleware(rl)(mux))
    srv := newHTTPServer(cfg.PORT, cors(handler))
	startServerAsync(srv, log, cfg.PORT)
	waitForShutdownSignal()
	shutdownServer(srv, log, pool)
}

func buildMux(cfg *config.Config, pool *sql.DB, log *logger.Logger) *http.ServeMux {
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
	ch := &handlers.ChatbotHandlers{DB: pool}
	mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(ch.ListOrCreate)))
	// Storage
	var storageService storage.StorageService
	if cfg.R2_ACCOUNT_ID != "" {
		var err error
		storageService, err = storage.NewR2Storage(cfg.R2_ACCOUNT_ID, cfg.R2_ACCESS_KEY_ID, cfg.R2_SECRET_ACCESS_KEY, cfg.R2_BUCKET_NAME)
		if err != nil {
			log.Error("storage_init_failed", map[string]any{"error": err.Error()})
		}
	}

	// Sources queue
	q, _ := processing.StartSourceQueue(pool, storageService)
    sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: storageService, Log: log}
	chh := &handlers.ChatHandlers{DB: pool}
	// Composite handler under /api/v1/chatbots/
	mux.Handle("/api/v1/chatbots/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sources") {
			sh.ChatbotSources(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			chh.Chat(w, r)
			return
		}
		ch.ByID(w, r)
	})))
	
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

	// Source status and delete
	mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(sh.GetSourceStatusOrDelete)))
	
	anh := &handlers.AnalyticsHandlers{DB: pool}
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(anh.GetAnalytics)))
	
	return mux
}

func newHTTPServer(port string, h http.Handler) *http.Server {
	return &http.Server{Addr: ":" + port, Handler: h}
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
	srv.Shutdown(ctx)
	pool.Close()
}
