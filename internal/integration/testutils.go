package integration

import (
	"database/sql"
	"net/http/httptest"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/config"
)

type TestEnv struct {
	Cfg    *config.Config
	DB     *sql.DB
	Server *httptest.Server
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
        ('free','active','lifetime',0,'TRY',0,'{}'::jsonb),
        ('pro','active','monthly',199,'TRY',7,'{}'::jsonb)
        ON CONFLICT (code) DO NOTHING`)
	// ensure users has plan_id column for tests
	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS plan_id UUID`)
	// ensure chatbots has language_id column for tests
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS language_id UUID REFERENCES languages(id)`)
	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_language_id UUID`)
	// ensure columns exist for tests
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS allowed_domains TEXT`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS embed_secret VARCHAR(255)`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS secure_embed_enabled BOOLEAN DEFAULT false`)
	mux := NewTestMux(cfg, db)
	srv := httptest.NewServer(mux)
	return &TestEnv{Cfg: cfg, DB: db, Server: srv}, nil
}

func TeardownTestEnv(te *TestEnv) {
	if te == nil {
		return
	}
	if te.DB != nil {
		_, _ = te.DB.Exec("TRUNCATE TABLE refresh_tokens, messages, conversations, analytics, payments, data_sources, chatbots, users RESTART IDENTITY CASCADE")
	}
	if te.Server != nil {
		te.Server.Close()
	}
	if te.DB != nil {
		_ = te.DB.Close()
	}
}
