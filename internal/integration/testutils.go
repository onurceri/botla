package integration

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/stretchr/testify/mock"
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
	Schema      string
	Server      *httptest.Server
	VectorStore *MockVectorStore
	MockVC      *rag.MockVectorClient
	MockLLM     *rag.MockFullClient
	Queue       *processing.SourceQueue
}

// cleanupOnce ensures we only run cleanup once per test suite
var cleanupOnce sync.Once

// CleanupAllIntegrationSchemas removes all botla_it_* schemas from the test database.
// This should be called at the start of a test run to ensure a clean state.
func CleanupAllIntegrationSchemas() error {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnvOrDefault("DB_USER", "botla"),
		getEnvOrDefault("DB_PASSWORD", "botla"),
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"),
		"botla_test",
	)

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(`
		SELECT nspname 
		FROM pg_namespace 
		WHERE nspname LIKE 'botla_it_%'
	`)
	if err != nil {
		return fmt.Errorf("list schemas: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err == nil {
			schemas = append(schemas, schema)
		}
	}

	for _, schema := range schemas {
		// Terminate connections first
		_, _ = db.Exec(`
			SELECT pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = current_database()
			  AND pid <> pg_backend_pid()
		`)

		if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: failed to drop schema %s: %v\n", schema, err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "cleaned up integration test schema: %s\n", schema)
		}
	}

	return nil
}

func SetupTestEnv() (*TestEnv, error) {
	// Run cleanup once at the start of the test suite to remove stale schemas
	cleanupOnce.Do(func() {
		_ = CleanupAllIntegrationSchemas()
	})

	if os.Getenv("DB_HOST") == "" {
		_ = os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		_ = os.Setenv("DB_PORT", "5432")
	}

	// Always use test database for integration tests
	_ = os.Setenv("DB_NAME", "botla_test")

	if os.Getenv("DB_USER") == "" {
		_ = os.Setenv("DB_USER", "botla")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		_ = os.Setenv("DB_PASSWORD", "botla")
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

	randBytes := make([]byte, 4)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, fmt.Errorf("failed to generate schema suffix: %w", err)
	}
	schema := "botla_it_" + hex.EncodeToString(randBytes)

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
		os.Getenv("DB_NAME"), schema)

	baseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	baseDB, err := sql.Open("pgx", baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open base db: %w", err)
	}
	if pingErr := baseDB.Ping(); pingErr != nil {
		_ = baseDB.Close()
		return nil, fmt.Errorf("failed to ping base db: %w", pingErr)
	}
	if _, schemaCreateErr := baseDB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)); schemaCreateErr != nil {
		_ = baseDB.Close()
		return nil, fmt.Errorf("failed to create schema %s: %w", schema, schemaCreateErr)
	}
	_ = baseDB.Close()

	// Migrate Up
	//nolint:gosec // this is a test helper using dynamic path
	cmdUp := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "up")
	output, migrateErr := cmdUp.CombinedOutput()
	if migrateErr != nil {
		return nil, fmt.Errorf("migration up failed: %s, %v", output, migrateErr)
	}

	// Clean up only data, don't drop tables
	// This makes parallel tests slightly better (though still
	// not perfect since they share the same schema).
	// We use TRUNCATE for speed.
	// NOTE: We don't truncate 'plans' and 'languages' as they are seeded.

	cfg := config.LoadConfig()
	cfg.DB_SCHEMA = schema
	db, err := dbpkg.New(cfg)
	if err != nil {
		return nil, err
	}

	// Set search path first
	_, _ = db.Exec("SET search_path TO " + schema)

	_, _ = db.Exec(`TRUNCATE TABLE chatbots, users, organizations, workspaces, data_sources, analytics, handoff_requests, messages, conversations CASCADE`)

	db.SetMaxOpenConns(1)

	// Restore plans to a clean state from migration 000035
	restorePlans(db)

	// Relax rate limits and limits for free plan in test environment
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_set(config, '{rate_limits,requests_per_minute}', '1000'::jsonb) WHERE code = 'free'`)
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_set(config, '{max_chatbots}', '100'::jsonb) WHERE code = 'free'`)

	// Insert dummy data for stub sources to prevent foreign key violations
	dummyUUID := "00000000-0000-0000-0000-000000000001"

	// Dummy User
	if _, insertUserErr := db.Exec(`INSERT INTO users (id, email, password_hash, plan_id, created_at, is_platform_admin, onboarding_completed, onboarding_step, onboarding_skipped) 
		VALUES ($1, 'dummy@example.com', 'hash', (SELECT id FROM plans WHERE code='free'), NOW(), false, true, 0, false)
		ON CONFLICT (id) DO NOTHING`, dummyUUID); insertUserErr != nil {
		return nil, fmt.Errorf("failed to insert dummy user: %w", insertUserErr)
	}

	// Dummy Chatbot
	if _, insertChatbotErr := db.Exec(`INSERT INTO chatbots (
		id, user_id, name, model, temperature, max_tokens, 
		theme_color, welcome_message, created_at, updated_at, 
		position, bot_message_color, user_message_color, bot_message_text_color, user_message_text_color, 
		chat_font_family, chat_header_color, chat_header_text_color, chat_background_color, bubble_radius, 
		input_background_color, input_text_color, send_button_color, 
		discovery_mode, refresh_policy, hide_branding, confidence_threshold, handoff_enabled, handoff_type, 
		language_id, secure_embed_enabled, suggestions_enabled
	) VALUES (
		$1, $1, 'Dummy Bot', 'gpt-4o-mini', 0.7, 4096, 
		'#000000', 'Hello', NOW(), NOW(), 
		'bottom-right', '#ffffff', '#000000', '#000000', '#ffffff', 
		'Inter', '#ffffff', '#000000', '#ffffff', '12px', 
		'#ffffff', '#000000', '#000000', 
		'auto', 'manual', false, 0.5, false, 'email', 
		(SELECT id FROM languages WHERE code='en-US'), false, false
	) ON CONFLICT (id) DO NOTHING`, dummyUUID); insertChatbotErr != nil {
		return nil, fmt.Errorf("failed to insert dummy chatbot: %w", insertChatbotErr)
	}

	// Dummy Source
	if _, insertSourceErr := db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, status, chunk_count, size_bytes, is_discovered, created_at)
		VALUES ($1, $1, 'text', 'completed', 1, 100, false, NOW())
		ON CONFLICT (id) DO NOTHING`, dummyUUID); insertSourceErr != nil {
		return nil, fmt.Errorf("failed to insert dummy source: %w", insertSourceErr)
	}

	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)
	mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockVC.On("DeleteBySourceID", mock.Anything, mock.Anything).Return(nil)
	mockVC.On("SearchSimilar", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]rag.SearchResult{}, nil)

	mockLLM := &rag.MockFullClient{}
	mockLLM.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
		Content: "Mock response",
	}, nil)
	mockLLM.On("CreateEmbedding", mock.Anything, mock.Anything).Return([]float32{0.1}, nil)
	mockLLM.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1}}, nil)
	mockLLM.On("CreateCompletionWithTools", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&rag.ChatResponseWithTools{
		Choices: []struct {
			Message      rag.ChatMessage `json:"message"`
			FinishReason string          `json:"finish_reason"`
		}{{
			Message: rag.ChatMessage{Content: strPtr("Mock tool response")},
		}},
		Usage: struct {
			TotalTokens int `json:"total_tokens"`
		}{TotalTokens: 10},
	}, nil)

	vs := &MockVectorStore{}
	// For now, continue using real clients by default to keep existing tests passing
	// New tests can manually create a mux with mocks.
	mux, q := NewTestMux(cfg, db, vs, nil, nil)
	srv := httptest.NewServer(mux)
	return &TestEnv{Cfg: cfg, DB: db, Schema: schema, Server: srv, VectorStore: vs, MockVC: mockVC, MockLLM: mockLLM, Queue: q}, nil
}

func strPtr(s string) *string {
	return &s
}

func TeardownTestEnv(te *TestEnv) {
	if te == nil {
		return
	}
	if te.Server != nil {
		te.Server.Close()
	}
	if te.Queue != nil {
		te.Queue.Stop()
	}

	schema := te.Schema
	if te.DB != nil {
		_ = te.DB.Close()
	}

	// Use a fresh connection for cleanup to avoid issues with the closed connection
	if schema != "" {
		dropIntegrationSchema(schema)
	}
}

// dropIntegrationSchema drops an integration test schema using a fresh connection.
// This is more reliable than using the test's DB connection which may have issues.
func dropIntegrationSchema(schema string) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnvOrDefault("DB_USER", "botla"),
		getEnvOrDefault("DB_PASSWORD", "botla"),
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"),
		"botla_test",
	)

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to open db for schema cleanup: %v\n", err)
		return
	}
	defer func() { _ = db.Close() }()

	// Terminate all connections to this schema to allow drop
	_, _ = db.Exec(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = current_database()
		  AND pid <> pg_backend_pid()
		  AND state = 'idle'
	`)

	// Retry drop up to 3 times
	for i := 0; i < 3; i++ {
		if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); err == nil {
			return
		}
		// Small delay between retries
		if i < 2 {
			select {
			case <-context.Background().Done():
				return
			default:
			}
		}
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func restorePlans(db *sql.DB) {
	// Re-apply plan configs with bare model names (matching migration 000040)
	// Free
	_, _ = db.Exec(`UPDATE plans SET config = jsonb_build_object(
    'scraping', jsonb_build_object('dynamic_enabled', false, 'max_urls_per_bot', 1, 'max_pages_per_crawl', 5),
    'files', jsonb_build_object('ocr_enabled', false, 'max_size_mb', 5, 'max_files_per_bot', 1, 'max_files_total', 5, 'total_storage_mb', 10, 'max_text_length', 400000),
    'chat', jsonb_build_object('default_model', 'gpt-4o-mini', 'allowed_models', '["gpt-4o-mini"]'::jsonb, 'max_monthly_tokens', 100000, 'rag', jsonb_build_object('top_k', 3, 'max_context_tokens', 2000), 'max_suggested_questions', 3),
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
    'chat', jsonb_build_object('default_model', 'gpt-4o', 'allowed_models', '["gpt-4o-mini", "gpt-4o"]'::jsonb, 'max_monthly_tokens', 1000000, 'rag', jsonb_build_object('top_k', 5, 'max_context_tokens', 4000), 'max_suggested_questions', 6),
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
    'chat', jsonb_build_object('default_model', 'gpt-4o', 'allowed_models', '["gpt-4o-mini", "gpt-4o", "gpt-5"]'::jsonb, 'max_monthly_tokens', 5000000, 'rag', jsonb_build_object('top_k', 10, 'max_context_tokens', 8000), 'max_suggested_questions', 10),
    'refresh', jsonb_build_object('enabled', true, 'max_monthly', 100),
    'security', jsonb_build_object('secure_embed_enabled', true),
    'guardrails', jsonb_build_object('can_customize_thresholds', true, 'can_use_smart_fallback', true, 'can_use_escalate_fallback', true, 'can_manage_topics', true, 'can_customize_messages', true),
    'branding', jsonb_build_object('can_hide_branding', true, 'can_custom_branding', true),
    'rate_limits', jsonb_build_object('requests_per_minute', 2000, 'window_seconds', 60, 'endpoints', jsonb_build_object('chat', jsonb_build_object('requests_per_minute', 500, 'window_seconds', 60), 'sources', jsonb_build_object('requests_per_minute', 100, 'window_seconds', 60))),
    'max_chatbots', 100, 'max_monthly_ingestions', 10000, 'max_monthly_embedding_tokens', 100000000, 'min_readd_cooldown_minutes', 0
) WHERE code = 'ultra'`)
}
