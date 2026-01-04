package main

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
)

func TestNewHTTPServer_Config(t *testing.T) {
	mux := http.NewServeMux()
	srv := newHTTPServer("8081", mux)
	if srv.Addr != ":8081" {
		t.Fatalf("addr got %s want %s", srv.Addr, ":8081")
	}
	if srv.Handler == nil {
		t.Fatalf("handler should be set")
	}
}

func TestStartAndShutdownServer(t *testing.T) {
	mux := http.NewServeMux()
	srv := newHTTPServer("0", mux)
	log := logger.New("ERROR")
	startServerAsync(srv, log, "0")
	db := testdb.OpenTestDB(t)
	defer db.Close()
	shutdownServer(srv, log, db)
}

func TestInitRedisClient_MissingURL(t *testing.T) {
	// Save and clear REDIS_URL
	originalRedisURL := os.Getenv("REDIS_URL")
	os.Unsetenv("REDIS_URL")
	defer func() {
		if originalRedisURL != "" {
			os.Setenv("REDIS_URL", originalRedisURL)
		}
	}()

	log := logger.New("ERROR")
	client, err := initRedisClient(log)

	if client != nil {
		client.Close()
		t.Fatal("expected nil client when REDIS_URL is missing")
	}
	if err == nil {
		t.Fatal("expected error when REDIS_URL is missing")
	}
	if !errors.Is(err, ErrRedisURLMissing) {
		t.Fatalf("expected ErrRedisURLMissing, got: %v", err)
	}
}

func TestInitRedisClient_InvalidURL(t *testing.T) {
	// Save original REDIS_URL
	originalRedisURL := os.Getenv("REDIS_URL")
	defer func() {
		if originalRedisURL != "" {
			os.Setenv("REDIS_URL", originalRedisURL)
		} else {
			os.Unsetenv("REDIS_URL")
		}
	}()

	// Set invalid URL
	os.Setenv("REDIS_URL", "not-a-valid-redis-url")

	log := logger.New("ERROR")
	client, err := initRedisClient(log)

	if client != nil {
		client.Close()
		t.Fatal("expected nil client with invalid REDIS_URL")
	}
	if err == nil {
		t.Fatal("expected error with invalid REDIS_URL")
	}
	// Error should indicate invalid URL
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestInitRedisClient_ConnectionFailure(t *testing.T) {
	// Save original REDIS_URL
	originalRedisURL := os.Getenv("REDIS_URL")
	defer func() {
		if originalRedisURL != "" {
			os.Setenv("REDIS_URL", originalRedisURL)
		} else {
			os.Unsetenv("REDIS_URL")
		}
	}()

	// Set URL pointing to non-existent Redis server
	os.Setenv("REDIS_URL", "redis://localhost:59999")

	log := logger.New("ERROR")
	client, err := initRedisClient(log)

	if client != nil {
		client.Close()
		t.Fatal("expected nil client when connection fails")
	}
	if err == nil {
		t.Fatal("expected error when connection fails")
	}
	if !errors.Is(err, ErrRedisConnectionFailed) {
		t.Fatalf("expected ErrRedisConnectionFailed, got: %v", err)
	}
}

func TestInitRedisClient_Success(t *testing.T) {
	// This test requires a running Redis instance
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis connection test")
	}

	log := logger.New("ERROR")
	client, err := initRedisClient(log)

	if err != nil {
		t.Fatalf("unexpected error with valid REDIS_URL: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client with valid REDIS_URL")
	}
	defer client.Close()
}

func TestNewApplication_DBInitFailure(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping DB init failure test")
	}

	originalDBURL := os.Getenv("DATABASE_URL")
	os.Setenv("DATABASE_URL", "postgres://invalid:invalid@localhost:5432/invalid")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	log := logger.New("ERROR")
	cfg := config.LoadConfig()

	_, err := newApplication(cfg, log)

	if err == nil {
		t.Fatal("expected error when DB init fails")
	}
}

func TestNewApplication_QdrantInitFailure(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Qdrant init failure test")
	}

	originalQdrantURL := os.Getenv("QDRANT_URL")
	os.Setenv("QDRANT_URL", "http://localhost:99999")
	defer func() {
		if originalQdrantURL != "" {
			os.Setenv("QDRANT_URL", originalQdrantURL)
		} else {
			os.Unsetenv("QDRANT_URL")
		}
	}()

	log := logger.New("ERROR")
	cfg := config.LoadConfig()

	_, err := newApplication(cfg, log)

	if err == nil {
		t.Fatal("expected error when Qdrant init fails")
	}
}

func TestNewApplication_PlanValidationFailure(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping plan validation test")
	}

	log := logger.New("ERROR")
	cfg := config.LoadConfig()

	app, err := newApplication(cfg, log)

	if err != nil {
		t.Fatalf("plan validation should pass with valid test DB: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil application on success")
	}
}

func TestNewApplication_RedisInitFailure(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping Redis init failure test")
	}

	originalRedisURL := os.Getenv("REDIS_URL")
	os.Setenv("REDIS_URL", "redis://localhost:59999")
	defer func() {
		if originalRedisURL != "" {
			os.Setenv("REDIS_URL", originalRedisURL)
		} else {
			os.Unsetenv("REDIS_URL")
		}
	}()

	log := logger.New("ERROR")
	cfg := config.LoadConfig()

	_, err := newApplication(cfg, log)

	if err == nil {
		t.Fatal("expected error when Redis init fails")
	}
}

func TestNewApplication_OpenAIInitFailure(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping OpenAI init failure test")
	}

	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("OPENAI_API_KEY", originalAPIKey)
		}
	}()

	log := logger.New("ERROR")
	cfg := config.LoadConfig()

	_, err := newApplication(cfg, log)

	if err == nil {
		t.Fatal("expected error when OpenAI init fails")
	}
}

func TestNewApplication_SourceQueueInitFailure(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping source queue test")
	}

	log := logger.New("ERROR")
	cfg := config.LoadConfig()

	app, err := newApplication(cfg, log)

	if err != nil {
		t.Fatalf("application init should succeed with valid config: %v", err)
	}
	if app.queue == nil {
		t.Fatal("expected non-nil queue")
	}
	if app.workerPool == nil {
		t.Fatal("expected non-nil workerPool")
	}
}

func TestNewApplication_ApplicationFieldsPopulated(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set, skipping field population test")
	}

	log := logger.New("ERROR")
	cfg := config.LoadConfig()

	app, err := newApplication(cfg, log)

	if err != nil {
		t.Fatalf("application init should succeed: %v", err)
	}

	if app.cfg == nil {
		t.Error("cfg should not be nil")
	}
	if app.log == nil {
		t.Error("log should not be nil")
	}
	if app.db == nil {
		t.Error("db should not be nil")
	}
	if app.redisClient == nil {
		t.Error("redisClient should not be nil")
	}
	if app.qdrantClient == nil {
		t.Error("qdrantClient should not be nil")
	}
	if app.rateLimiter == nil {
		t.Error("rateLimiter should not be nil")
	}
	if app.globalLimiter == nil {
		t.Error("globalLimiter should not be nil")
	}
	if app.refreshScheduler == nil {
		t.Error("refreshScheduler should not be nil")
	}
	if app.retentionJob == nil {
		t.Error("retentionJob should not be nil")
	}
	if app.workerPool == nil {
		t.Error("workerPool should not be nil")
	}
}
