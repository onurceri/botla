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

	var proPlanID string
	// Try to find 'pro' plan from seeded data
	err := db.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID)
	if err != nil {
		t.Fatalf("pro plan not found in database (migrations might not have run): %v", err)
	}

	// Models should also be seeded by migrations
	var modelCount int
	err = db.QueryRow(`SELECT count(*) FROM ai_models`).Scan(&modelCount)
	if err != nil {
		t.Fatalf("failed to count models: %v", err)
	}
	if modelCount == 0 {
		t.Fatal("no models found in ai_models table (migrations might not have run)")
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
	if _, ok := res["available_models"]; !ok {
		t.Error("missing available_models")
	}
}
