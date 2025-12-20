package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestMe_Success(t *testing.T) {
	db := testdb.OpenTestDB(t)
	var uid string
	var proPlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("meuniq+%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", proPlanID).Scan(&uid); err != nil {
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
