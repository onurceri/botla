package fixtures

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
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/internal/db"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/processing"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/scraper"
	"github.com/onurceri/botla-app/internal/workers"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/policy"
	"github.com/stretchr/testify/mock"
)

// TestPassword is a strong password that meets complexity requirements for tests.
// Requires: uppercase, lowercase, digit, and special character.
const TestPassword = "Test@123"

type TestEnv struct {
	Cfg             *config.Config
	DB              *sql.DB
	Schema          string
	Server          *httptest.Server
	VectorStore     *MockVectorStore
	MockVC          *rag.MockVectorClient
	MockLLM         *rag.MockFullClient
	MockScraper     *scraper.MockScraper // For tests that need to configure scraper responses
	RealVC          *rag.QdrantClient    // Added for cleanup
	CollectionName  string               // Added for cleanup
	Queue           *processing.SourceQueue
	Limiter         *middleware.RateLimiter
	WorkerPool      *workers.WorkerPool
	SourcesHandlers *handlers.SourcesHandlers
}

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
		// Terminate connections first? NO. This kills parallel tests in other packages.
		// We rely on tests closing connections.

		if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: failed to drop schema %s: %v\n", schema, err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "cleaned up integration test schema: %s\n", schema)
		}
	}

	return nil
}

func SetupTestEnv() (*TestEnv, error) {
	return SetupTestEnvWithConfigAndMocks(nil, true)
}

func SetupTestEnvWithMocks() (*TestEnv, error) {
	return SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_KEY = ""
		cfg.QDRANT_URL = ""
	}, true)
}

// ConfigOverride is a function that modifies the test config.
// Use this to customize config values without t.Setenv() calls.
type ConfigOverride func(*config.Config)

// SetupTestEnvWithConfig creates a test environment with optional config overrides.
// This enables parallel test execution by avoiding t.Setenv() calls.
//
// Usage:
//
//	func TestExample(t *testing.T) {
//		te, err := fixtures.SetupTestEnvWithConfig(func(cfg *config.Config) {
//			cfg.OPENAI_API_BASE = oai.URL
//			cfg.QDRANT_URL = qd.URL
//		})
//		if err != nil {
//			t.Fatalf("setup failed: %v", err)
//		}
//		defer fixtures.TeardownTestEnv(te)
//		// ... test code
//	}
func SetupTestEnvWithConfig(override ConfigOverride) (*TestEnv, error) {
	return setupTestEnvCommon(true, override)
}

// SetupTestEnvWithConfigAndMocks creates a test environment with mocks and optional config overrides.
func SetupTestEnvWithConfigAndMocks(override ConfigOverride, useMocks bool) (*TestEnv, error) {
	return setupTestEnvCommon(useMocks, override)
}

// defaultTestConfig returns a Config struct with test values.
// This avoids t.Setenv() calls, enabling parallel test execution.
func defaultTestConfig() *config.Config {
	return &config.Config{
		DB_HOST:                "localhost",
		DB_PORT:                "5432",
		DB_NAME:                "botla_test",
		DB_USER:                "botla",
		DB_PASSWORD:            "botla",
		DB_SCHEMA:              "public",
		DB_SSLMODE:             "disable",
		REDIS_URL:              "redis://localhost:6379",
		QDRANT_URL:             "http://localhost:6333",
		QDRANT_API_KEY:         "",
		OPENAI_API_KEY:         "test-key",
		OPENAI_API_BASE:        "https://api.openai.com",
		OPENAI_TIMEOUT_MS:      30000,
		OPENROUTER_API_KEY:     "test-key",
		OPENROUTER_API_BASE:    "https://openrouter.ai/api/v1",
		OPENROUTER_TIMEOUT_MS:  30000,
		IYZICO_API_KEY:         "",
		IYZICO_SECRET_KEY:      "",
		JWT_SECRET:             "test-secret-for-testing-only",
		PORT:                   "8080",
		CORS_ALLOWED_ORIGINS:   "http://localhost:5173",
		WORKER_COUNT:           4,
		ANALYTICS_WORKER_COUNT: 10,
		R2_ACCOUNT_ID:          "",
		R2_ACCESS_KEY_ID:       "",
		R2_SECRET_ACCESS_KEY:   "",
		R2_BUCKET_NAME:         "",
		DEFAULT_CHATBOT_MODEL:  "gpt-4o-mini",
		RAG_TOPK:               5,
		RAG_MAX_CONTEXT_TOKENS: 2000,
		CHAT_TIMEOUT_MS:        60000,
		GO_ENV:                 "test",
		CookieSecure:           false,
	}
}

func setupTestEnvCommon(useMocks bool, override ConfigOverride) (*TestEnv, error) {
	// Build config without relying on environment variables
	cfg := defaultTestConfig()

	// Apply any overrides provided by the caller
	if override != nil {
		override(cfg)
	}

	// Only set defaults from env if not overridden
	// This allows tests to fully control config values

	randBytes := make([]byte, 4)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, fmt.Errorf("failed to generate schema suffix: %w", err)
	}
	schemaSuffix := hex.EncodeToString(randBytes)
	schema := "botla_it_" + schemaSuffix
	collectionName := "embeddings_" + schemaSuffix

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
		cfg.DB_USER, cfg.DB_PASSWORD,
		cfg.DB_HOST, cfg.DB_PORT,
		cfg.DB_NAME, schema)

	baseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB_USER, cfg.DB_PASSWORD,
		cfg.DB_HOST, cfg.DB_PORT,
		cfg.DB_NAME,
	)
	baseDB, err := sql.Open("pgx", baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open base db: %w", err)
	}
	baseDB.SetMaxOpenConns(1)
	if pingErr := baseDB.Ping(); pingErr != nil {
		_ = baseDB.Close()
		return nil, fmt.Errorf("failed to ping base db: %w", pingErr)
	}
	if _, schemaCreateErr := baseDB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)); schemaCreateErr != nil {
		_ = baseDB.Close()
		return nil, fmt.Errorf("failed to create schema %s: %w", schema, schemaCreateErr)
	}
	if closeErr := baseDB.Close(); closeErr != nil {
		return nil, fmt.Errorf("failed to close base db: %w", closeErr)
	}
	time.Sleep(50 * time.Millisecond)

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

	// Use the config we built earlier (with overrides applied)
	cfg.DB_SCHEMA = schema
	db, err := db.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("new db: %w", err)
	}

	// Set search path first
	_, _ = db.Exec("SET search_path TO " + schema)

	_, _ = db.Exec(`TRUNCATE TABLE chatbots, users, organizations, workspaces, data_sources, analytics, handoff_requests, messages, conversations CASCADE`)

	db.SetMaxOpenConns(10)

	// Restore plans to a clean state from migration 000035
	RestorePlans(db)

	// Relax rate limits and limits for free plan in test environment
	_ = updatePlanLimitField(context.Background(), db, policy.PlanFree.String(), "rate_limits_requests_per_minute", 1000)
	_ = updatePlanLimitField(context.Background(), db, policy.PlanFree.String(), "max_chatbots", 100)

	// Insert dummy data for stub sources to prevent foreign key violations
	dummyUUID := "00000000-0000-0000-0000-000000000001"

	// Dummy User
	if _, insertUserErr := db.Exec(`INSERT INTO users (id, email, password_hash, plan_id, created_at, is_platform_admin, onboarding_completed, onboarding_step, onboarding_skipped) 
		VALUES ($1, 'dummy@example.com', 'hash', (SELECT id FROM plans WHERE code=$2), NOW(), false, true, 0, false)
		ON CONFLICT (id) DO NOTHING`, dummyUUID, policy.PlanFree.String()); insertUserErr != nil {
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
		$1, $1, 'Dummy Bot', $2, 0.7, 4096, 
		'#000000', 'Hello', NOW(), NOW(), 
		'bottom-right', '#ffffff', '#000000', '#000000', '#ffffff', 
		'Inter', '#ffffff', '#000000', '#ffffff', '12px', 
		'#ffffff', '#000000', '#000000', 
		'auto', 'manual', false, 0.5, false, 'email', 
		(SELECT id FROM languages WHERE code='en-US'), false, false
	) ON CONFLICT (id) DO NOTHING`, dummyUUID, policy.ModelGPT4oMini.String()); insertChatbotErr != nil {
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
			Message: rag.ChatMessage{Content: StrPtr("Mock tool response")},
		}},
		Usage: struct {
			TotalTokens int `json:"total_tokens"`
		}{TotalTokens: 10},
	}, nil)

	vs := &MockVectorStore{}

	var llm rag.LLMClient
	var vc rag.VectorClient
	var realVC *rag.QdrantClient

	if useMocks {
		// Always use mocks when useMocks=true, regardless of config values
		llm = mockLLM
		vc = mockVC
	} else {
		// useMocks=false: Try to create real clients, fall back to mocks if not configured
		// This path is only for explicit real-service testing scenarios

		// Vector client
		if cfg.QDRANT_URL != "" {
			var err error
			realVC, err = rag.NewQdrantClient(&rag.QdrantConfig{
				URL:            cfg.QDRANT_URL,
				APIKey:         cfg.QDRANT_API_KEY,
				Timeout:        15 * time.Second,
				CollectionName: collectionName,
			})
			if err == nil {
				vc = realVC
			} else {
				fmt.Printf("WARNING: failed to create qdrant client, using mock: %v\n", err)
				vc = mockVC
			}
		} else {
			// No QDRANT_URL configured, use mock
			vc = mockVC
		}

		// LLM client - leave nil to let NewTestMux create real client
		// when useMocks=false. Tests can point cfg.OPENAI_API_BASE to a
		// mock HTTP server to intercept LLM calls.
		// llm remains nil here, NewTestMux will create the real client
	}

	mux, q, rl, wp, sh, ms := NewTestMux(cfg, db, vs, llm, vc)
	srv := httptest.NewServer(mux)
	return &TestEnv{
		Cfg:             cfg,
		DB:              db,
		Schema:          schema,
		Server:          srv,
		VectorStore:     vs,
		MockVC:          mockVC,
		MockLLM:         mockLLM,
		MockScraper:     ms,
		RealVC:          realVC,
		CollectionName:  collectionName,
		Queue:           q,
		Limiter:         rl,
		WorkerPool:      wp,
		SourcesHandlers: sh,
	}, nil
}

func StrPtr(s string) *string {
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
	if te.Limiter != nil {
		_ = te.Limiter.Close()
	}
	if te.WorkerPool != nil {
		te.WorkerPool.Shutdown(1 * time.Second)
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
	// WARNING: This kills all connections to the DB, which breaks parallel tests.
	// We rely on TeardownTestEnv closing the test's DB connection first.
	// If DROP SCHEMA fails due to locks, we will retry.

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
				time.Sleep(100 * time.Millisecond) // Added explicit sleep since helper context was background
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

// RestorePlans resets plan_limits to their default values for testing.
func RestorePlans(db *sql.DB) {
	ctx := context.Background()

	// Free plan limits
	freeLimits := models.DefaultPlanLimits()
	updatePlanLimits(ctx, db, "free", freeLimits)

	// Pro plan limits
	proLimits := models.ProPlanLimits()
	updatePlanLimits(ctx, db, "pro", proLimits)

	// Ultra plan limits
	ultraLimits := models.UltraPlanLimits()
	updatePlanLimits(ctx, db, "ultra", ultraLimits)
}

// updatePlanLimits updates all fields in plan_limits for a given plan code.
func updatePlanLimits(ctx context.Context, db *sql.DB, code string, limits models.PlanLimits) {
	// Use the db package helper for all fields
	_ = updatePlanLimitField(ctx, db, code, "max_chatbots", limits.MaxChatbots)
	_ = updatePlanLimitField(ctx, db, code, "max_monthly_ingestions", limits.MaxMonthlyIngestions)
	_ = updatePlanLimitField(ctx, db, code, "max_monthly_embedding_tokens", limits.MaxMonthlyEmbeddingTokens)
	_ = updatePlanLimitField(ctx, db, code, "min_readd_cooldown_minutes", limits.MinReAddCooldownMinutes)
	_ = updatePlanLimitField(ctx, db, code, "scraping_dynamic_enabled", limits.ScrapingDynamicEnabled)
	_ = updatePlanLimitField(ctx, db, code, "scraping_max_urls_per_bot", limits.ScrapingMaxURLsPerBot)
	_ = updatePlanLimitField(ctx, db, code, "scraping_max_pages_per_crawl", limits.ScrapingMaxPagesPerCrawl)
	_ = updatePlanLimitField(ctx, db, code, "files_max_size_mb", limits.FilesMaxSizeMB)
	_ = updatePlanLimitField(ctx, db, code, "files_max_files_per_bot", limits.FilesMaxFilesPerBot)
	_ = updatePlanLimitField(ctx, db, code, "files_max_files_total", limits.FilesMaxFilesTotal)
	_ = updatePlanLimitField(ctx, db, code, "files_total_storage_mb", limits.FilesTotalStorageMB)
	_ = updatePlanLimitField(ctx, db, code, "files_max_text_length", limits.FilesMaxTextLength)
	_ = updatePlanLimitField(ctx, db, code, "chat_default_model", limits.ChatDefaultModel)
	_ = updatePlanLimitField(ctx, db, code, "chat_max_monthly_tokens", limits.ChatMaxMonthlyTokens)
	_ = updatePlanLimitField(ctx, db, code, "chat_rag_top_k", limits.ChatRAGTopK)
	_ = updatePlanLimitField(ctx, db, code, "chat_rag_max_context_tokens", limits.ChatRAGMaxContextTokens)
	_ = updatePlanLimitField(ctx, db, code, "chat_max_suggested_questions", limits.ChatMaxSuggestedQuestions)
	_ = updatePlanLimitField(ctx, db, code, "chat_max_manual_questions", limits.ChatMaxManualQuestions)
	_ = updatePlanLimitField(ctx, db, code, "chat_min_response_token_limit", limits.ChatMinResponseTokenLimit)
	_ = updatePlanLimitField(ctx, db, code, "chat_max_response_token_limit", limits.ChatMaxResponseTokenLimit)
	_ = updatePlanLimitField(ctx, db, code, "refresh_enabled", limits.RefreshEnabled)
	_ = updatePlanLimitField(ctx, db, code, "refresh_max_monthly", limits.RefreshMaxMonthly)
	_ = updatePlanLimitField(ctx, db, code, "security_secure_embed_enabled", limits.SecuritySecureEmbedEnabled)
	_ = updatePlanLimitField(ctx, db, code, "guardrails_can_customize_thresholds", limits.GuardrailsCanCustomizeThresholds)
	_ = updatePlanLimitField(ctx, db, code, "guardrails_can_use_smart_fallback", limits.GuardrailsCanUseSmartFallback)
	_ = updatePlanLimitField(ctx, db, code, "guardrails_can_use_escalate_fallback", limits.GuardrailsCanUseEscalateFallback)
	_ = updatePlanLimitField(ctx, db, code, "guardrails_can_manage_topics", limits.GuardrailsCanManageTopics)
	_ = updatePlanLimitField(ctx, db, code, "guardrails_can_customize_messages", limits.GuardrailsCanCustomizeMessages)
	_ = updatePlanLimitField(ctx, db, code, "branding_can_hide_branding", limits.BrandingCanHideBranding)
	_ = updatePlanLimitField(ctx, db, code, "branding_can_custom_branding", limits.BrandingCanCustomBranding)
	_ = updatePlanLimitField(ctx, db, code, "rate_limits_requests_per_minute", limits.RateLimitsRequestsPerMinute)
	_ = updatePlanLimitField(ctx, db, code, "rate_limits_window_seconds", limits.RateLimitsWindowSeconds)
	_ = updatePlanLimitField(ctx, db, code, "rate_limits_chat_rpm", limits.RateLimitsChatRPM)
	_ = updatePlanLimitField(ctx, db, code, "rate_limits_chat_window", limits.RateLimitsChatWindow)
	_ = updatePlanLimitField(ctx, db, code, "rate_limits_sources_rpm", limits.RateLimitsSourcesRPM)
	_ = updatePlanLimitField(ctx, db, code, "rate_limits_sources_window", limits.RateLimitsSourcesWindow)
}

// UpdatePlanLimit updates a single limit field for a plan. This is a convenience
// wrapper around updatePlanLimitField for use in tests.
func (te *TestEnv) UpdatePlanLimit(planCode, field string, value any) error {
	return updatePlanLimitField(context.Background(), te.DB, planCode, field, value)
}

// updatePlanLimitField updates a single field in the plan_limits table for a given plan code.
// This is a local implementation that replaces the deprecated db.UpdatePlanLimitField.
func updatePlanLimitField(ctx context.Context, db *sql.DB, planCode, field string, value any) error {
	query := fmt.Sprintf(`
		UPDATE plan_limits
		SET value = $1, updated_at = NOW()
		WHERE plan_id = (SELECT id FROM plans WHERE code = $2)
		  AND field = $3
	`)
	_, err := db.ExecContext(ctx, query, value, planCode, field)
	return err
}
