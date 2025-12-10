package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"
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
		_ = os.Setenv("DB_NAME", "botla_dev")
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
	if os.Getenv("QDRANT_URL") == "" {
		_ = os.Setenv("QDRANT_URL", "http://localhost:6333")
	}
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "test-secret")
	}
	if os.Getenv("PORT") == "" {
		_ = os.Setenv("PORT", "8080")
	}
	if os.Getenv("OPENAI_API_KEY") == "" {
		_ = os.Setenv("OPENAI_API_KEY", "test-key")
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
	// ensure extensions and canonical tables for plans/languages exist in test schema
	_, _ = db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS languages (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        code TEXT UNIQUE NOT NULL,
        name TEXT NOT NULL,
        rtl BOOLEAN DEFAULT false,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ,
        deleted_at TIMESTAMPTZ
    )`)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS plans (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        code TEXT UNIQUE NOT NULL,
        status TEXT NOT NULL DEFAULT 'active',
        billing_cycle TEXT NOT NULL DEFAULT 'monthly',
        price NUMERIC(10,2) NOT NULL DEFAULT 0,
        currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
        trial_days INTEGER NOT NULL DEFAULT 0,
        config JSONB,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ,
        deleted_at TIMESTAMPTZ
    )`)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS plan_translations (
        plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
        language_id UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
        name TEXT NOT NULL,
        description TEXT,
        UNIQUE (plan_id, language_id)
    )`)
	_, _ = db.Exec(`INSERT INTO languages (code, name, rtl) VALUES
        ('tr-TR','Turkish (Türkiye)',false),
        ('en-US','English (United States)',false)
        ON CONFLICT (code) DO NOTHING`)
	_, _ = db.Exec(`INSERT INTO plans (code, status, billing_cycle, price, currency, trial_days, config) VALUES
        ('free','active','lifetime',0,'TRY',0,'{"scraping": {"max_pages_per_crawl": 10, "max_urls_per_bot": 100}}'::jsonb),
        ('pro','active','monthly',199,'TRY',7,'{"scraping": {"max_pages_per_crawl": 100, "max_urls_per_bot": 1000}}'::jsonb)
        ON CONFLICT (code) DO UPDATE SET config = EXCLUDED.config`)
	// ensure users has plan_id column for tests
	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS plan_id UUID`)
	// ensure chatbots has language_id column for tests
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS language_id UUID REFERENCES languages(id)`)
	// ensure analytics has total_tokens_used
	_, _ = db.Exec(`ALTER TABLE analytics ADD COLUMN IF NOT EXISTS total_tokens_used INTEGER DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_language_id UUID`)
	// ensure columns exist for tests
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS allowed_domains TEXT`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS embed_secret VARCHAR(255)`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS secure_embed_enabled BOOLEAN DEFAULT false`)
	// Add missing columns for tests
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS include_paths JSONB`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS exclude_paths JSONB`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS selector_whitelist JSONB`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS discovery_mode TEXT DEFAULT 'auto'`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS refresh_policy TEXT DEFAULT 'manual'`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS refresh_frequency TEXT`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS next_refresh_at TIMESTAMPTZ`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS last_refresh_at TIMESTAMPTZ`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS bot_icon TEXT`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS bot_display_name TEXT`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS suggested_questions JSONB`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS suggestions_enabled BOOLEAN DEFAULT false`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS hide_branding BOOLEAN DEFAULT false`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS custom_branding JSONB`)

	// ensure new source columns exist for tests
	_, _ = db.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS hash VARCHAR(128)`)
	_, _ = db.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`)
	_, _ = db.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS size_bytes BIGINT DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS is_discovered BOOLEAN DEFAULT false`)
	// create usage_ingestions table for tests
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS usage_ingestions (
		user_id VARCHAR(64) NOT NULL,
		period_month DATE NOT NULL,
		sources_count INT NOT NULL DEFAULT 0,
		embedding_tokens INT NOT NULL DEFAULT 0,
		refresh_count INT NOT NULL DEFAULT 0,
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (user_id, period_month)
	)`)
	// Create pending_discovered_urls table for tests
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS pending_discovered_urls (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
		source_id UUID REFERENCES data_sources(id) ON DELETE SET NULL,
		url TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		discovered_at TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE (chatbot_id, url)
	)`)
	// Add refresh tracking columns
	_, _ = db.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_refreshed_at TIMESTAMPTZ`)
	_, _ = db.Exec(`ALTER TABLE usage_ingestions ADD COLUMN IF NOT EXISTS refresh_count INT DEFAULT 0`)
	// Create chatbot_actions table
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS chatbot_actions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
        name TEXT NOT NULL,
        description TEXT,
        action_type TEXT NOT NULL,
        config JSONB,
        parameters JSONB,
        enabled BOOLEAN DEFAULT false,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ,
        deleted_at TIMESTAMPTZ
    )`)
	// ensure cascade deletes are set up for tests
	_, _ = db.Exec(`ALTER TABLE chatbots DROP CONSTRAINT IF EXISTS chatbots_workspace_id_fkey`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD CONSTRAINT chatbots_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE`)
	_, _ = db.Exec(`ALTER TABLE chatbots DROP CONSTRAINT IF EXISTS chatbots_organization_id_fkey`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD CONSTRAINT chatbots_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE`)

	vs := &MockVectorStore{}
	mux := NewTestMux(cfg, db, vs)
	srv := httptest.NewServer(mux)
	return &TestEnv{Cfg: cfg, DB: db, Server: srv, VectorStore: vs}, nil
}

func TeardownTestEnv(te *TestEnv) {
	if te == nil {
		return
	}
	if te.DB != nil {
		_, _ = te.DB.Exec("TRUNCATE TABLE refresh_tokens, messages, conversations, analytics, payments, data_sources, chatbots, users, organizations, memberships, workspaces RESTART IDENTITY CASCADE")
	}
	if te.Server != nil {
		te.Server.Close()
	}
	if te.DB != nil {
		_ = te.DB.Close()
	}
}
