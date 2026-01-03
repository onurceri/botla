package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/redis/go-redis/v9"
)

// RealServiceTestEnv provides real service connections for integration testing.
// This is separate from the mocked TestEnv used in most integration tests.
type RealServiceTestEnv struct {
	t      *testing.T
	Cfg    *config.Config
	DB     *sql.DB
	PGPool *pgxpool.Pool
	Redis  *redis.Client
	Qdrant *rag.QdrantClient
	Schema string
}

// SetupRealServices creates connections to real services for integration testing.
// Requires services to be running (use docker-compose.integration.yml).
func SetupRealServices(t *testing.T) *RealServiceTestEnv {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping real service integration test in short mode")
	}

	cfg := buildTestConfig()

	pgPool, err := createPostgreSQLPool(t, cfg)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL pool: %v", err)
	}

	testDB := testdb.OpenParallelTestDB(t)
	db := testDB

	var schema string
	err = db.QueryRow("SELECT current_schema()").Scan(&schema)
	if err != nil {
		t.Fatalf("Failed to get schema name: %v", err)
	}

	redisClient, err := createRedisClient(t, cfg)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}

	qdrantClient, err := createQdrantClient(t, cfg)
	if err != nil {
		t.Fatalf("Failed to create Qdrant client: %v", err)
	}

	env := &RealServiceTestEnv{
		t:      t,
		Cfg:    cfg,
		DB:     db,
		PGPool: pgPool,
		Redis:  redisClient,
		Qdrant: qdrantClient,
		Schema: schema,
	}

	t.Cleanup(env.Cleanup)

	return env
}

// buildTestConfig creates configuration for real service integration tests.
func buildTestConfig() *config.Config {
	return &config.Config{
		DB_HOST:               getEnvOrDefault("DB_HOST", "localhost"),
		DB_PORT:               getEnvOrDefault("DB_PORT", "5433"),
		DB_NAME:               getEnvOrDefault("DB_NAME", "botla_integration"),
		DB_USER:               getEnvOrDefault("DB_USER", "botla"),
		DB_PASSWORD:           getEnvOrDefault("DB_PASSWORD", "botla"),
		DB_SSLMODE:            "disable",
		DB_SCHEMA:             "",
		REDIS_URL:             getEnvOrDefault("REDIS_URL", "redis://localhost:6380"),
		QDRANT_URL:            getEnvOrDefault("QDRANT_URL", "http://localhost:6334"),
		QDRANT_API_KEY:        "",
		OPENAI_API_KEY:        getEnvOrDefault("OPENAI_API_KEY", ""),
		OPENAI_API_BASE:       getEnvOrDefault("OPENAI_API_BASE", "https://api.openai.com"),
		OPENAI_TIMEOUT_MS:     30000,
		JWT_SECRET:            "test-secret-for-integration-tests-only",
		PORT:                  "8080",
		CORS_ALLOWED_ORIGINS:  "http://localhost:5173",
		WORKER_COUNT:          2,
		DEFAULT_CHATBOT_MODEL: "gpt-4o-mini",
		GO_ENV:                "test",
	}
}

// createPostgreSQLPool creates a PostgreSQL connection pool for tests.
func createPostgreSQLPool(t *testing.T, cfg *config.Config) (*pgxpool.Pool, error) {
	t.Helper()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB_USER,
		cfg.DB_PASSWORD,
		cfg.DB_HOST,
		cfg.DB_PORT,
		cfg.DB_NAME)

	pgConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DB config: %w", err)
	}

	pgConfig.MaxConns = 10
	pgConfig.MinConns = 2
	pgConfig.MaxConnLifetime = 5 * time.Minute
	pgConfig.MaxConnIdleTime = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// createRedisClient creates a Redis client for tests.
func createRedisClient(t *testing.T, cfg *config.Config) (*redis.Client, error) {
	t.Helper()

	opts, err := redis.ParseURL(cfg.REDIS_URL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}

	opts.PoolSize = 10
	opts.MinIdleConns = 2
	opts.MaxRetries = 3
	opts.PoolTimeout = 5 * time.Second

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return client, nil
}

// createQdrantClient creates a Qdrant client for tests.
func createQdrantClient(t *testing.T, cfg *config.Config) (*rag.QdrantClient, error) {
	t.Helper()

	client, err := rag.NewQdrantClient(&rag.QdrantConfig{
		URL:            cfg.QDRANT_URL,
		APIKey:         cfg.QDRANT_API_KEY,
		Timeout:        15 * time.Second,
		CollectionName: "test_embeddings_integration",
	})
	if err != nil {
		return nil, fmt.Errorf("create Qdrant client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.EnsureEmbeddingsCollection(ctx); err != nil {
		t.Logf("Qdrant health check failed (will skip Qdrant tests): %v", err)
		return client, nil
	}

	return client, nil
}

// Cleanup closes all service connections.
func (e *RealServiceTestEnv) Cleanup() {
	if e.Redis != nil {
		if err := e.Redis.Close(); err != nil && e.t != nil {
			e.t.Logf("warning: failed to close Redis: %v", err)
		}
	}
	if e.PGPool != nil {
		e.PGPool.Close()
	}
	if e.DB != nil {
		e.DB.Close()
	}
}

// getEnvOrDefault returns environment variable or default value.
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
