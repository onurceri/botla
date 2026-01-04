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

	"github.com/onurceri/botla-app/internal/api/router"
	"github.com/onurceri/botla-app/internal/db"
	"github.com/onurceri/botla-app/internal/processing"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/scraper"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/workers"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/ratelimit"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/onurceri/botla-app/pkg/tokenizer"
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
		return fmt.Errorf("new application: %w", err)
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
	workerPool       *workers.WorkerPool
}

type infraDeps struct {
	db      *sql.DB
	qdrant  *rag.QdrantClient
	storage storage.StorageService
	loader  *tokenizer.Loader
}

type rateLimitDeps struct {
	redisClient   *redis.Client
	globalLimiter ratelimit.Limiter
	rateLimiter   *middleware.RateLimiter
}

type processingDeps struct {
	queue      *processing.SourceQueue
	workerPool *workers.WorkerPool
}

type schedulerDeps struct {
	refreshScheduler *services.RefreshScheduler
	retentionJob     *services.RetentionJob
}

func initInfrastructure(cfg *config.Config, log *logger.Logger) (*infraDeps, error) {
	pool, err := db.New(cfg)
	if err != nil {
		log.Error("db_init_failed", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("init db: %w", err)
	}

	planRepo := repository.NewPostgresPlanRepo(pool, nil)
	planSvc := services.NewPlanService(planRepo, nil)
	if vErr := planSvc.ValidateAllPlans(context.Background()); vErr != nil {
		log.Error("plan_validation_failed", map[string]any{
			"error":   vErr.Error(),
			"message": "Application failed to start because one or more plans have invalid configurations in the database",
		})
		return nil, fmt.Errorf("validate plans: %w", vErr)
	}
	log.Info("plan_validation_success", nil)

	qdrantClient, err := rag.NewQdrantClient(&rag.QdrantConfig{
		URL:     cfg.QDRANT_URL,
		APIKey:  cfg.QDRANT_API_KEY,
		Timeout: 15 * time.Second,
	})
	if err != nil {
		log.Error("qdrant_init_failed", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("init qdrant: %w", err)
	}

	err = ensureQdrantCollection(qdrantClient, log)
	if err != nil {
		return nil, fmt.Errorf("ensure qdrant collection: %w", err)
	}
	log.Info("qdrant_collection_ready", nil)

	var storageService storage.StorageService
	if cfg.R2_ACCOUNT_ID != "" {
		storageService, err = storage.NewR2Storage(cfg.R2_ACCOUNT_ID, cfg.R2_ACCESS_KEY_ID, cfg.R2_SECRET_ACCESS_KEY, cfg.R2_BUCKET_NAME)
		if err != nil {
			log.Error("storage_init_failed", map[string]any{"error": err.Error()})
		}
	}

	var tokLoader *tokenizer.Loader
	if storageService != nil {
		tokLoader = tokenizer.NewLoader(storageService)
		if tokErr := tokLoader.Preload(context.Background()); tokErr != nil {
			log.Warn("tokenizer_init_fallback", map[string]any{"error": tokErr.Error()})
		} else {
			log.Info("tokenizer_loaded", nil)
		}
	}

	return &infraDeps{
		db:      pool,
		qdrant:  qdrantClient,
		storage: storageService,
		loader:  tokLoader,
	}, nil
}

func initRateLimiting(cfg *config.Config, log *logger.Logger, pool *sql.DB) (*rateLimitDeps, error) {
	redisClient, err := initRedisClient(log)
	if err != nil {
		log.Error("redis_required", map[string]any{
			"error":   err.Error(),
			"message": "Redis is required for rate limiting to ensure consistent behavior across distributed instances",
		})
		return nil, fmt.Errorf("init redis: %w", err)
	}

	rlSettings := ratelimit.Settings{
		GlobalRequestsPerMinute: cfg.RateLimitGlobalRequestsPerMinute,
		GlobalWindowSeconds:     cfg.RateLimitGlobalWindowSeconds,
		UserRequestsPerMinute:   cfg.RateLimitUserRequestsPerMinute,
		UserWindowSeconds:       cfg.RateLimitUserWindowSeconds,
		EndpointChat:            cfg.RateLimitEndpointChat,
		EndpointSources:         cfg.RateLimitEndpointSources,
		AuthLogin:               cfg.RateLimitAuthLogin,
		AuthRegister:            cfg.RateLimitAuthRegister,
		AuthRefresh:             cfg.RateLimitAuthRefresh,
	}

	rlConfig := ratelimit.NewConfig(rlSettings)
	globalLimiter := ratelimit.NewRedisLimiter(redisClient, rlConfig.Global)
	log.Info("rate_limiter_initialized", map[string]any{"backend": "redis", "mode": "plan-based"})
	rl := middleware.NewRateLimiter(globalLimiter, redisClient, rlConfig)

	return &rateLimitDeps{
		redisClient:   redisClient,
		globalLimiter: globalLimiter,
		rateLimiter:   rl,
	}, nil
}

func initProcessing(cfg *config.Config, log *logger.Logger, pool *sql.DB, storageService storage.StorageService, qdrantClient *rag.QdrantClient, tokLoader *tokenizer.Loader) (*processingDeps, error) {
	oaiClient, err := rag.NewOpenAIClient(cfg)
	if err != nil {
		log.Error("openai_init_failed", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("init openai: %w", err)
	}

	scConfig := scraper.CollectorConfig{
		AllowedDomains:  strings.Split(cfg.SCRAPER_ALLOWED_DOMAINS, ","),
		Timeout:         30 * time.Second,
		RateLimitPerSec: 2,
	}
	sdConfig := scraper.DynamicConfig{
		PoolSize:   cfg.SCRAPER_BROWSER_POOL_SIZE,
		IdleTTL:    time.Duration(cfg.SCRAPER_DYNAMIC_IDLE_SECS) * time.Second,
		NavTimeout: time.Duration(cfg.SCRAPER_NAV_TIMEOUT_MS) * time.Millisecond,
		Allowed:    strings.Split(cfg.SCRAPER_ALLOWED_DOMAINS, ","),
	}
	bScraper, err := scraper.NewBrowserScraper(sdConfig)
	if err != nil {
		log.Error("browser_scraper_init_failed", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("init browser scraper: %w", err)
	}

	scrp := scraper.NewDefaultScraper(scConfig, bScraper)

	// Create repositories needed for the source queue
	trainingJobRepo := repository.NewPostgresTrainingJobRepo(pool)

	q, err := processing.StartSourceQueue(trainingJobRepo, storageService, oaiClient, qdrantClient, scrp, tokLoader, cfg.WORKER_COUNT)
	if err != nil {
		log.Error("source_queue_init_failed", map[string]any{"error": err.Error()})
		return nil, fmt.Errorf("init source queue: %w", err)
	}

	workerPool := workers.NewWorkerPool(log, cfg.ANALYTICS_WORKER_COUNT)

	return &processingDeps{
		queue:      q,
		workerPool: workerPool,
	}, nil
}

func initSchedulers(pool *sql.DB, redisClient *redis.Client, log *logger.Logger, queue *processing.SourceQueue, storageService storage.StorageService) *schedulerDeps {
	chatbotRepo := repository.NewPostgresChatbotRepo(pool)
	sourceRepo := repository.NewPostgresSourceRepo(pool)
	planRepo := repository.NewPostgresPlanRepo(pool, redisClient)
	analyticsRepo := repository.NewPostgresAnalyticsRepo(pool)

	return &schedulerDeps{
		refreshScheduler: services.NewRefreshScheduler(chatbotRepo, sourceRepo, planRepo, analyticsRepo, queue, log),
		retentionJob:     services.NewRetentionJob(pool, log, storageService),
	}
}

func newApplication(cfg *config.Config, log *logger.Logger) (*application, error) {
	infra, err := initInfrastructure(cfg, log)
	if err != nil {
		return nil, err
	}

	rl, err := initRateLimiting(cfg, log, infra.db)
	if err != nil {
		return nil, err
	}

	proc, err := initProcessing(cfg, log, infra.db, infra.storage, infra.qdrant, infra.loader)
	if err != nil {
		return nil, err
	}

	sched := initSchedulers(infra.db, rl.redisClient, log, proc.queue, infra.storage)

	return &application{
		cfg:              cfg,
		log:              log,
		db:               infra.db,
		qdrantClient:     infra.qdrant,
		storageService:   infra.storage,
		redisClient:      rl.redisClient,
		globalLimiter:    rl.globalLimiter,
		rateLimiter:      rl.rateLimiter,
		queue:            proc.queue,
		workerPool:       proc.workerPool,
		refreshScheduler: sched.refreshScheduler,
		retentionJob:     sched.retentionJob,
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

	mux := router.New(app.cfg, app.db, app.log, app.queue, app.storageService, app.qdrantClient, app.redisClient, app.workerPool)
	origins := strings.Split(app.cfg.CORS_ALLOWED_ORIGINS, ",")
	cors := middleware.CORSMiddlewareAllowOrigins(origins)
	// Middleware chain: RequestID -> Security -> Recovery -> Logger -> MaxBytes -> PlanLoader -> RateLimit -> Mux
	planRepo := repository.NewPostgresPlanRepo(app.db, app.redisClient)
	planLoader := middleware.PlanLoaderMiddleware(planRepo, app.log)
	handler := middleware.RequestID(
		middleware.SecurityHeadersMiddleware()(
			middleware.RecoveryMiddleware(app.log, app.cfg.GO_ENV)(
				middleware.RequestLogger(app.log)(
					middleware.MaxBytesMiddleware(1 * 1024 * 1024)( // 1MB limit
						planLoader(
							middleware.RateLimitMiddleware(app.rateLimiter)(mux)))))))

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
	if app.workerPool != nil {
		app.workerPool.Shutdown(10 * time.Second)
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
		return fmt.Errorf("ensure qdrant embeddings collection: %w", err)
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
