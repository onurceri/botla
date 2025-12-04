package integration

import (
    "database/sql"
    "net/http/httptest"
    "os"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/onurceri/botla-co/pkg/config"
    dbpkg "github.com/onurceri/botla-co/internal/db"
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
	// ensure columns exist for tests
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS allowed_domains TEXT`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS embed_secret VARCHAR(255)`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS secure_embed_enabled BOOLEAN DEFAULT false`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS language VARCHAR(10) DEFAULT 'tr'`)
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
