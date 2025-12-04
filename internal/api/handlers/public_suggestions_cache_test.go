package handlers

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/onurceri/botla-co/internal/db"
    "github.com/onurceri/botla-co/internal/models"
)

func TestPublicChatbotConfig_SuggestionsCacheKeyedByUpdatedAt(t *testing.T) {
    pool, cleanup := mustInitDB(t)
    defer cleanup()
    var uid string
    email := "cache_" + time.Now().Format("150405.000000") + "@test"
    if err := pool.QueryRow(`INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id`, email, "x").Scan(&uid); err != nil { t.Fatalf("user: %v", err) }
    bot := &models.Chatbot{
        UserID:               uid,
        Name:                 "Bot",
        SystemPrompt:         "p",
        Language:             "en",
        Model:                "gpt-3.5-turbo",
        Temperature:          0.1,
        MaxTokens:            64,
        ThemeColor:           "#000000",
        WelcomeMessage:       "hi",
        Position:             "bottom-right",
        BotMessageColor:      "#000000",
        UserMessageColor:     "#ffffff",
        BotMessageTextColor:  "#ffffff",
        UserMessageTextColor: "#000000",
        ChatFontFamily:       "Inter",
        ChatHeaderColor:      "#000000",
        ChatHeaderTextColor:  "#ffffff",
        ChatBackgroundColor:  "#ffffff",
        SuggestedQuestions:   []string{"A", "B"},
        SuggestionsEnabled:   true,
    }
    bid, err := db.CreateChatbot(nil, pool, bot)
    if err != nil { t.Fatalf("create: %v", err) }

    req1 := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+bid, nil)
    w1 := httptest.NewRecorder()
    PublicChatbotConfig(pool)(w1, req1)
    if w1.Code != http.StatusOK { t.Fatalf("status1: %d", w1.Code) }
    var m1 map[string]any
    _ = json.Unmarshal(w1.Body.Bytes(), &m1)
    if len(m1["suggested_questions"].([]any)) != 2 { t.Fatalf("len1") }

    time.Sleep(10 * time.Millisecond)
    if _, err := pool.Exec(`UPDATE chatbots SET suggested_questions=$1, updated_at=NOW() WHERE id=$2`, jsonArr([]string{"C"}), bid); err != nil { t.Fatalf("upd: %v", err) }

    req2 := httptest.NewRequest(http.MethodGet, "/api/v1/public/chatbots/"+bid, nil)
    w2 := httptest.NewRecorder()
    PublicChatbotConfig(pool)(w2, req2)
    var m2 map[string]any
    _ = json.Unmarshal(w2.Body.Bytes(), &m2)
    if len(m2["suggested_questions"].([]any)) != 1 { t.Fatalf("len2") }
}

func jsonArr(in []string) []byte { b, _ := json.Marshal(in); return b }

func mustInitDB(t *testing.T) (*sql.DB, func()) {
    t.Helper()
    dsn := "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable"
    db, err := sql.Open("pgx", dsn)
    if err != nil { t.Fatalf("open: %v", err) }
    return db, func() { db.Close() }
}
