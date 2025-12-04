package integration

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

func NewTestMux(cfg *config.Config, pool *sql.DB) http.Handler {
	mux := http.NewServeMux()
	rl := middleware.NewRateLimiterFromEnv()
	hh := &handlers.HealthHandlers{DB: pool, Cfg: cfg}
	mux.Handle("/health", middleware.RateLimitMiddleware(rl)(http.HandlerFunc(hh.Health)))
	ah := &handlers.AuthHandlers{DB: pool, Secret: cfg.JWT_SECRET}
	mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
	mux.HandleFunc("/api/v1/auth/refresh", ah.RefreshHandler)
	mux.HandleFunc("/api/v1/auth/logout", ah.LogoutHandler)
	mux.Handle("/api/v1/protected", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(handlers.ProtectedHandler)))
	mh := &handlers.MeHandlers{DB: pool}
	mux.Handle("/api/v1/me", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(mh.Me)))

	anh := &handlers.AnalyticsHandlers{DB: pool}
	mux.Handle("/api/v1/analytics", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(anh.GetAnalytics)))

	ch := &handlers.ChatbotHandlers{DB: pool}
	mux.Handle("/api/v1/chatbots", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(ch.ListOrCreate)))
	memStore := storage.NewMemoryStorage()
	q, _ := processing.StartSourceQueue(pool, memStore)
	sh := &handlers.SourcesHandlers{DB: pool, Queue: q, Storage: memStore}
	chh := &handlers.ChatHandlers{DB: pool}
	mux.Handle("/api/v1/chatbots/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/sources") {
			sh.ChatbotSources(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			middleware.RateLimitMiddleware(rl)(http.HandlerFunc(chh.Chat)).ServeHTTP(w, r)
			return
		}
		ch.ByID(w, r)
	})))
	mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(sh.GetSourceStatusOrDelete)))
	mux.Handle("/api/v1/messages/", middleware.AuthMiddleware(cfg.JWT_SECRET)(http.HandlerFunc(chh.FeedbackHandler)))
	mux.Handle("/api/v1/public/chatbots/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/public/chatbots/"
		if strings.HasPrefix(r.URL.Path, p) && strings.HasSuffix(r.URL.Path, "/chat") {
			handlers.PublicChat(pool)(w, r)
			return
		}
		handlers.PublicChatbotConfig(pool)(w, r)
	}))
	origins := strings.Split(cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	return cors(mux)
}
