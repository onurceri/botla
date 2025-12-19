package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestGetPlan_Success(t *testing.T) {
	db := testdb.OpenTestDB(t)
	defer db.Close()

	var proPlanID string
	// Try to find 'pro' plan, or fallback to any plan if pro is not seeded in testdb
	err := db.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID)
	if err != nil {
		// Fallback: create a dummy plan if not exists
		proPlanID = "00000000-0000-0000-0000-000000000001"
		_, err = db.Exec(`INSERT INTO plans (id, code, price, currency, config) VALUES ($1, 'pro', 100, 'USD', '{}') ON CONFLICT DO NOTHING`, proPlanID)
		if err != nil {
			// If conflict (e.g. ID exists or CODE exists), select whatever is there
			if err := db.QueryRow(`SELECT id FROM plans LIMIT 1`).Scan(&proPlanID); err != nil {
				t.Fatalf("no plans available (insert error: %v): %v", err, err)
			}
		}
	}

	var uid string
	email := fmt.Sprintf("plantest+%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", proPlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	h := &PlanHandlers{DB: db}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/plan", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, uid)
	h.GetPlan(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body: %s", rr.Code, rr.Body.String())
	}

	var res map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Check new structure
	if _, ok := res["limits"]; !ok {
		t.Error("missing limits")
	}
	if _, ok := res["features"]; !ok {
		t.Error("missing features")
	}
	if _, ok := res["config"]; ok {
		t.Error("config field should not be present")
	}
}
