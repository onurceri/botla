package main

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/logger"
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
