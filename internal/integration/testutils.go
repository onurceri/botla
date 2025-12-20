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

	// Migrate Up
	//nolint:gosec // this is a test helper using dynamic path
	cmdUp := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "up")
	if output, err := cmdUp.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("migration up failed: %s, %v", output, err)
	}

	// Clean up only data, don't drop tables
	// This makes parallel tests slightly better (though still
	// not perfect since they share the same schema).
	// We use TRUNCATE for speed.
	// NOTE: We don't truncate 'plans' and 'languages' as they are seeded.

	cfg := config.LoadConfig()
	db, err := dbpkg.New(cfg)
	if err != nil {
		return nil, err
	}

	// Set search path first
	_ = os.Setenv("DB_SCHEMA", "test")
	_, _ = db.Exec("SET search_path TO test")

	_, _ = db.Exec(`TRUNCATE TABLE chatbots, users, organizations, workspaces, data_sources, analytics, handoff_requests, messages, conversations CASCADE`)

	db.SetMaxOpenConns(1)

	// Restore plans to a clean state from migration 000035
	restorePlans(db)

	// Relax rate limits and limits for free plan in test environment
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits,requests_per_minute}', '1000'::jsonb) WHERE code = 'free'`)
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_set(config, '{max_chatbots}', '100'::jsonb) WHERE code = 'free'`)

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

func restorePlans(db *sql.DB) {
	// Re-apply migration 000035 configs to ensure clean state
	// Free
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_build_object(
    'scraping', jsonb_build_object('dynamic_enabled', false, 'max_urls_per_bot', 1, 'max_pages_per_crawl', 5),
    'files', jsonb_build_object('ocr_enabled', false, 'max_size_mb', 5, 'max_files_per_bot', 1, 'max_files_total', 5, 'total_storage_mb', 10, 'max_text_length', 400000),
    'chat', jsonb_build_object('default_model', 'openai/gpt-4o-mini', 'allowed_models', '["openai/gpt-4o-mini"]'::jsonb, 'max_monthly_tokens', 100000, 'rag', jsonb_build_object('top_k', 3, 'max_context_tokens', 2000), 'max_suggested_questions', 3),
    'refresh', jsonb_build_object('enabled', false, 'max_monthly', 0),
    'security', jsonb_build_object('secure_embed_enabled', false),
    'guardrails', jsonb_build_object('can_customize_thresholds', false, 'can_use_smart_fallback', false, 'can_use_escalate_fallback', false, 'can_manage_topics', false, 'can_customize_messages', false),
    'branding', jsonb_build_object('can_hide_branding', false, 'can_custom_branding', false),
    'rate_limits', jsonb_build_object('requests_per_minute', 100, 'window_seconds', 60, 'endpoints', jsonb_build_object('chat', jsonb_build_object('requests_per_minute', 30, 'window_seconds', 60), 'sources', jsonb_build_object('requests_per_minute', 10, 'window_seconds', 60))),
    'max_chatbots', 1, 'max_monthly_ingestions', 50, 'max_monthly_embedding_tokens', 250000, 'min_readd_cooldown_minutes', 60
) WHERE code = 'free'`)

	// Pro
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_build_object(
    'scraping', jsonb_build_object('dynamic_enabled', true, 'max_urls_per_bot', 10, 'max_pages_per_crawl', 50),
    'files', jsonb_build_object('ocr_enabled', true, 'max_size_mb', 20, 'max_files_per_bot', 20, 'max_files_total', 100, 'total_storage_mb', 500, 'max_text_length', 400000),
    'chat', jsonb_build_object('default_model', 'openai/gpt-4o', 'allowed_models', '["openai/gpt-4o-mini", "openai/gpt-4o"]'::jsonb, 'max_monthly_tokens', 1000000, 'rag', jsonb_build_object('top_k', 5, 'max_context_tokens', 4000), 'max_suggested_questions', 6),
    'refresh', jsonb_build_object('enabled', true, 'max_monthly', 5),
    'security', jsonb_build_object('secure_embed_enabled', true),
    'guardrails', jsonb_build_object('can_customize_thresholds', true, 'can_use_smart_fallback', true, 'can_use_escalate_fallback', false, 'can_manage_topics', true, 'can_customize_messages', true),
    'branding', jsonb_build_object('can_hide_branding', true, 'can_custom_branding', false),
    'rate_limits', jsonb_build_object('requests_per_minute', 500, 'window_seconds', 60, 'endpoints', jsonb_build_object('chat', jsonb_build_object('requests_per_minute', 100, 'window_seconds', 60), 'sources', jsonb_build_object('requests_per_minute', 30, 'window_seconds', 60))),
    'max_chatbots', 10, 'max_monthly_ingestions', 500, 'max_monthly_embedding_tokens', 2500000, 'min_readd_cooldown_minutes', 30
) WHERE code = 'pro'`)

	// Ultra
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_build_object(
    'scraping', jsonb_build_object('dynamic_enabled', true, 'max_urls_per_bot', 50, 'max_pages_per_crawl', 200),
    'files', jsonb_build_object('ocr_enabled', true, 'max_size_mb', 50, 'max_files_per_bot', 100, 'max_files_total', 1000, 'total_storage_mb', 2000, 'max_text_length', 400000),
    'chat', jsonb_build_object('default_model', 'openai/gpt-4o', 'allowed_models', '["openai/gpt-4o-mini", "openai/gpt-4o", "openai/gpt-5"]'::jsonb, 'max_monthly_tokens', 5000000, 'rag', jsonb_build_object('top_k', 10, 'max_context_tokens', 8000), 'max_suggested_questions', 10),
    'refresh', jsonb_build_object('enabled', true, 'max_monthly', 100),
    'security', jsonb_build_object('secure_embed_enabled', true),
    'guardrails', jsonb_build_object('can_customize_thresholds', true, 'can_use_smart_fallback', true, 'can_use_escalate_fallback', true, 'can_manage_topics', true, 'can_customize_messages', true),
    'branding', jsonb_build_object('can_hide_branding', true, 'can_custom_branding', true),
    'rate_limits', jsonb_build_object('requests_per_minute', 2000, 'window_seconds', 60, 'endpoints', jsonb_build_object('chat', jsonb_build_object('requests_per_minute', 500, 'window_seconds', 60), 'sources', jsonb_build_object('requests_per_minute', 100, 'window_seconds', 60))),
    'max_chatbots', 100, 'max_monthly_ingestions', 10000, 'max_monthly_embedding_tokens', 100000000, 'min_readd_cooldown_minutes', 0
) WHERE code = 'ultra'`)
}
