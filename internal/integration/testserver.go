package integration

import (
    "database/sql"
    "net/http"
    "strings"

    "github.com/onurceri/botla-co/internal/api/handlers"
    "github.com/onurceri/botla-co/internal/processing"
    "github.com/onurceri/botla-co/pkg/config"
    "github.com/onurceri/botla-co/pkg/middleware"
)

func NewTestMux(cfg *config.Config, pool *sql.DB) http.Handler {
    mux := http.NewServeMux()
    hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
    mux.HandleFunc("/health", hh.Health)
    ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET}
    mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
    mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
    mux.Handle("/api/v1/protected", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(handlers.ProtectedHandler)))
	
	anh := &handlers.AnalyticsHandlers{DB: pool}
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(anh.GetAnalytics)))

	ch := &handlers.ChatbotHandlers{DB: pool}
    rl := middleware.NewRateLimiterFromEnv()
    mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.RateLimitMiddleware(rl)(http.HandlerFunc(ch.ListOrCreate))))
    q, _ := processing.StartSourceQueue(pool, nil)
    sh := &handlers.SourcesHandlers{DB: pool, Queue: q}
    chh := &handlers.ChatHandlers{DB: pool}
    mux.Handle("/api/v1/chatbots/", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.RateLimitMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
    }))))
    mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(cfg.JWT_SECRET)(middleware.RateLimitMiddleware(rl)(http.HandlerFunc(sh.GetSourceStatusOrDelete))))
    origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
    cors := middleware.CORSMiddlewareAllowOrigins(origins)
    return cors(mux)
}
