package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestMe_Success(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	var uid string
	email := fmt.Sprintf("meuniq+%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, subscription_plan) VALUES ($1,$2,$3) RETURNING id`, email, "x", "pro").Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}
	h := &MeHandlers{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, uid)
	h.Me(rr, req.WithContext(ctx))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
}
