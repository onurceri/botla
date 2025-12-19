package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/config"
)

// MockVectorStore for testing
type MockVectorStore struct {
	mu               sync.Mutex
	DeletedSourceIDs []string
}

func (m *MockVectorStore) DeleteBySourceID(ctx context.Context, sourceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeletedSourceIDs = append(m.DeletedSourceIDs, sourceID)
	return nil
}

type TestEnv struct {
	Cfg         *config.Config
	DB          *sql.DB
	Server      *httptest.Server
	VectorStore *MockVectorStore
}

func SetupTestEnv() (*TestEnv, error) {
	if os.Getenv("DB_HOST") == "" {
		_ = os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		_ = os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_NAME") == "" {
		_ = os.Setenv("DB_NAME", "botla_test")
	}
	if os.Getenv("DB_USER") == "" {
		_ = os.Setenv("DB_USER", "botla")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		_ = os.Setenv("DB_PASSWORD", "botla")
	}
	if os.Getenv("DB_SCHEMA") == "" {
		_ = os.Setenv("DB_SCHEMA", "test")
	}
	if os.Getenv("DB_NAME") != "botla_test" {
		return nil, fmt.Errorf("tests must use botla_test database, got %s", os.Getenv("DB_NAME"))
	}
	if os.Getenv("DB_SCHEMA") != "test" {
		return nil, fmt.Errorf("tests must use test schema, got %s", os.Getenv("DB_SCHEMA"))
	}
	if os.Getenv("QDRANT_URL") == "" {
		_ = os.Setenv("QDRANT_URL", "http://localhost:6333")
	}
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "test-secret")
	}
	if os.Getenv("PORT") == "" {
		_ = os.Setenv("PORT", "8080")
	}
	if _, ok := os.LookupEnv("OPENAI_API_KEY"); !ok {
		_ = os.Setenv("OPENAI_API_KEY", "test-key")
	}

	// Run migrations
	// We use the migrate CLI tool to ensure the test database is in a clean state.
	// This replaces manual SQL setup and ensures we test against the latest schema.
	wd, _ := os.Getwd()
	// Find the project root by looking for go.mod
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			// Could not find go.mod, fallback to relative path assumption
			projectRoot = filepath.Join(wd, "../..")
			break
		}
		projectRoot = parent
	}
	migrationsPath := filepath.Join(projectRoot, "db/migrations")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"), os.Getenv("DB_SCHEMA"))

	// Migrate Down (Reset)
	// We use -all to ensure we revert all migrations
	cmdDown := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "down", "-all")
	// If down fails (e.g. inconsistent state), we try to drop -f
	if output, err := cmdDown.CombinedOutput(); err != nil {
		// Try drop -f as fallback
		cmdDrop := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "drop", "-f")
		if outDrop, errDrop := cmdDrop.CombinedOutput(); errDrop != nil {
			return nil, fmt.Errorf("migration reset failed: down output: %s, drop output: %s, err: %v", output, outDrop, errDrop)
		}
	}

	// Migrate Up
	cmdUp := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "up")
	if output, err := cmdUp.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("migration up failed: %s, %v", output, err)
	}

	cfg := config.LoadConfig()
	db, err := dbpkg.New(cfg)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	// Validate schema name to prevent SQL injection
	for _, c := range cfg.DB_SCHEMA {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return nil, fmt.Errorf("invalid schema name: %s", cfg.DB_SCHEMA)
		}
	}
	_, _ = db.Exec("SET search_path TO " + cfg.DB_SCHEMA)
	var currentSchema string
	if err := db.QueryRow("SELECT current_schema()").Scan(&currentSchema); err != nil {
		return nil, err
	}
	if currentSchema != "test" {
		return nil, fmt.Errorf("expected current_schema() to be test, got %s", currentSchema)
	}

	// Relax rate limits for free plan in test environment to prevent 429 errors during tests
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits,requests_per_minute}', '1000'::jsonb) WHERE code = 'free'`)

	vs := &MockVectorStore{}
	mux := NewTestMux(cfg, db, vs)
	srv := httptest.NewServer(mux)
	return &TestEnv{Cfg: cfg, DB: db, Server: srv, VectorStore: vs}, nil
}

func TeardownTestEnv(te *TestEnv) {
	if te == nil {
		return
	}
	if te.Server != nil {
		te.Server.Close()
	}
	if te.DB != nil {
		_ = te.DB.Close()
	}
}
