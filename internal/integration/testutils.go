package integration

import (
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/pkg/config"
)

type TestEnv struct {
	Cfg    *config.Config
	DB     *sql.DB
	Server *httptest.Server
}

func SetupTestEnv() (*TestEnv, error) {
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "botla_dev")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "botla")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		os.Setenv("DB_PASSWORD", "botla")
	}
	if os.Getenv("QDRANT_URL") == "" {
		os.Setenv("QDRANT_URL", "http://localhost:6333")
	}
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret")
	}
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8080")
	}
	if os.Getenv("OPENAI_API_KEY") == "" {
		os.Setenv("OPENAI_API_KEY", "test-key")
	}

	cfg := config.LoadConfig()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DB_USER, cfg.DB_PASSWORD, cfg.DB_HOST, cfg.DB_PORT, cfg.DB_NAME)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
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
	if te.Server != nil {
		te.Server.Close()
	}
	if te.DB != nil {
		te.DB.Close()
	}
}
