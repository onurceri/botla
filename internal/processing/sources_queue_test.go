package processing

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAggregateAndPersistChatbotSuggestions(t *testing.T) {
	db, cleanup := mustInitDB(t)
	defer cleanup()
	_, _ = db.Exec(`ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS suggested_questions JSONB`)
	_, _ = db.Exec(`ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS suggested_questions JSONB`)
	// create user and bot
	var uid, bid string
	email := "test-" + time.Now().Format("20060102150405.000") + "@example.com"
	if err := db.QueryRow(`INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id`, email, "x").Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO chatbots (user_id, name) VALUES ($1,$2) RETURNING id`, uid, "X").Scan(&bid); err != nil {
		t.Fatalf("bot: %v", err)
	}
	// insert sources
	arr1, _ := json.Marshal([]string{"S1", "S2"})
	arr2, _ := json.Marshal([]string{"S2", "S3"})
	if _, err := db.Exec(`INSERT INTO data_sources (chatbot_id, source_type, suggested_questions) VALUES ($1,'text',$2)`, bid, arr1); err != nil {
		t.Fatalf("src1: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO data_sources (chatbot_id, source_type, suggested_questions) VALUES ($1,'url',$2)`, bid, arr2); err != nil {
		t.Fatalf("src2: %v", err)
	}
	aggregateAndPersistChatbotSuggestions(context.Background(), db, bid)
	var js []byte
	if err := db.QueryRow(`SELECT suggested_questions FROM chatbots WHERE id=$1`, bid).Scan(&js); err != nil {
		t.Fatalf("chatbot: %v", err)
	}
	var out []string
	_ = json.Unmarshal(js, &out)
	// We expect S1, S2 (from existing) + S3 (from new source) = 3 total
	if len(out) != 3 {
		t.Fatalf("expected 3 merged suggestions, got %d: %v", len(out), out)
	}
	// Check content
	hasS3 := false
	for _, s := range out {
		if s == "S3" {
			hasS3 = true
		}
	}
	if !hasS3 {
		t.Fatalf("expected S3 in suggestions, got %v", out)
	}
}

// minimal db init using env config from integration tests
func mustInitDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	te, err := integrationSetupEnv()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return te.DB, func() { te.DB.Close() }
}

type testEnv struct{ DB *sql.DB }

func integrationSetupEnv() (*testEnv, error) {
	dsn := "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &testEnv{DB: db}, nil
}
