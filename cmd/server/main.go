package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/onurceri/botla-co/internal/api/router"
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
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg := config.LoadConfig()
	log := logger.New("INFO")

	app, err := newApplication(cfg, log)
	if err != nil {
		// Specific errors are logged in newApplication
		return err
	}

	app.start()
	waitForShutdownSignal()
	app.shutdown()

	return nil
}

type application struct {
	cfg              *config.Config
	log              *logger.Logger
	db               *sql.DB
	redisClient      *redis.Client
	qdrantClient     *rag.QdrantClient
	storageService   storage.StorageService
	queue            *processing.SourceQueue
	refreshScheduler *services.RefreshScheduler
	retentionJob     *services.RetentionJob
	rateLimiter      *middleware.RateLimiter
	globalLimiter    ratelimit.Limiter
	server           *http.Server
	schedulerCancel  context.CancelFunc
}

func newApplication(cfg *config.Config, log *logger.Logger) (*application, error) {
	pool, err := db.New(cfg)
	if err != nil {
		log.Error("db_init_failed", map[string]any{"error": err.Error()})
		return nil, err
	}

	// Initialize Qdrant
	qdrantClient, err := rag.NewQdrantClientFromEnv()
	if err != nil {
		log.Error("qdrant_init_failed", map[string]any{"error": err.Error()})
		return nil, err
	}

	// Ensure embeddings collection exists
	err = ensureQdrantCollection(qdrantClient, log)
	if err != nil {
		return nil, err
	}
	log.Info("qdrant_collection_ready", nil)

	// Initialize storage service
	var storageService storage.StorageService
	if cfg.R2_ACCOUNT_ID != "" {
		storageService, err = storage.NewR2Storage(cfg.R2_ACCOUNT_ID, cfg.R2_ACCESS_KEY_ID, cfg.R2_SECRET_ACCESS_KEY, cfg.R2_BUCKET_NAME)
		if err != nil {
			log.Error("storage_init_failed", map[string]any{"error": err.Error()})
		}
	}

	// Initialize OpenAI client
	oaiClient, err := rag.NewOpenAIClient(cfg)
	if err != nil {
		log.Error("openai_init_failed", map[string]any{"error": err.Error()})
		return nil, err
	}

	// Start source processing queue
	q, err := processing.StartSourceQueue(pool, storageService, oaiClient, qdrantClient)
	if err != nil {
		log.Error("source_queue_init_failed", map[string]any{"error": err.Error()})
		return nil, err
	}

	// Initialize Redis client for rate limiting (mandatory)
	redisClient, err := initRedisClient(log)
	if err != nil {
		log.Error("redis_required", map[string]any{
			"error":   err.Error(),
			"message": "Redis is required for rate limiting to ensure consistent behavior across distributed instances",
		})
		return nil, err
	}

	// Initialize rate limiter (Redis-based, mandatory)
	rlConfig := ratelimit.NewConfigFromEnv()
	globalLimiter := ratelimit.NewRedisLimiter(redisClient, rlConfig.Global)
	log.Info("rate_limiter_initialized", map[string]any{"backend": "redis", "mode": "plan-based"})
	rl := middleware.NewRateLimiter(globalLimiter, redisClient, rlConfig)

	// Start refresh scheduler
	refreshScheduler := services.NewRefreshScheduler(pool, q, log)

	// Initialize retention job
	retentionJob := services.NewRetentionJob(pool, log, storageService)

	return &application{
		cfg:              cfg,
		log:              log,
		db:               pool,
		redisClient:      redisClient,
		qdrantClient:     qdrantClient,
		storageService:   storageService,
		queue:            q,
		refreshScheduler: refreshScheduler,
		retentionJob:     retentionJob,
		rateLimiter:      rl,
		globalLimiter:    globalLimiter,
	}, nil
}

func (app *application) start() {
	// Start refresh scheduler
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	app.schedulerCancel = schedulerCancel
	app.refreshScheduler.Start(schedulerCtx)

	// Start retention job (daily at 03:00 AM)
	go func() {
		// Run once on startup (in background)
		go func() {
			if err := app.retentionJob.Run(schedulerCtx); err != nil {
				app.log.Error("initial_retention_job_failed", map[string]any{"error": err.Error()})
			}
		}()

		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}

			select {
			case <-schedulerCtx.Done():
				return
			case <-time.After(time.Until(next)):
				if err := app.retentionJob.Run(schedulerCtx); err != nil {
					app.log.Error("retention_job_failed", map[string]any{"error": err.Error()})
				}
			}
		}
	}()

	mux := router.New(app.cfg, app.db, app.log, app.queue, app.storageService, app.qdrantClient, app.redisClient)
	origins := strings.Split(app.cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	// Middleware chain: Recovery -> Logger -> PlanLoader -> RateLimit -> Mux
	planLoader := middleware.PlanLoaderMiddleware(app.db, app.log)
	handler := middleware.RecoveryMiddleware(app.log)(middleware.RequestLogger(app.log)(planLoader(middleware.RateLimitMiddleware(app.rateLimiter)(mux))))

	app.server = newHTTPServer(app.cfg.PORT, cors(handler))
	startServerAsync(app.server, app.log, app.cfg.PORT)
}

func (app *application) shutdown() {
	// Graceful shutdown
	if app.schedulerCancel != nil {
		app.schedulerCancel()
	}
	if app.refreshScheduler != nil {
		app.refreshScheduler.Stop()
	}
	// Close rate limiters
	if app.globalLimiter != nil {
		if err := app.globalLimiter.Close(); err != nil {
			app.log.Error("global_limiter_close_failed", map[string]any{"error": err.Error()})
		}
	}
	// Note: Plan-based limiters are managed internally by RateLimiter
	if app.redisClient != nil {
		if err := app.redisClient.Close(); err != nil {
			app.log.Error("redis_close_failed", map[string]any{"error": err.Error()})
		}
	}
	if app.server != nil {
		shutdownServer(app.server, app.log, app.db)
	}
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

// ErrRedisURLMissing indicates REDIS_URL environment variable is not set
var ErrRedisURLMissing = errors.New("REDIS_URL environment variable is required")

// ErrRedisConnectionFailed indicates Redis connection could not be established
var ErrRedisConnectionFailed = errors.New("failed to connect to Redis")

// initRedisClient initializes and validates a Redis connection.
// Returns an error if Redis is not configured or connection fails.
func initRedisClient(log *logger.Logger) (*redis.Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, ErrRedisURLMissing
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("%w: %v", ErrRedisConnectionFailed, err)
	}

	log.Info("redis_connected", nil)
	return client, nil
}
