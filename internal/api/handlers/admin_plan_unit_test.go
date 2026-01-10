package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
)

func TestAdminPlanHandlers_ListPlans(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	planRepo := repository.NewPostgresPlanRepo(db, nil)
	planSvc := services.NewPlanService(planRepo, nil)
	h := NewAdminPlanHandlers(planSvc, planRepo)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plans", nil)
	h.ListPlans(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListPlans: expected 200, got %d", rr.Code)
	}

	var response AdminPlanListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("ListPlans: failed to unmarshal response: %v", err)
	}

	if response.Total == 0 {
		t.Log("ListPlans: no plans found (might be expected in test environment)")
	}
}

func TestAdminPlanHandlers_GetPlan_InvalidID(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	planRepo := repository.NewPostgresPlanRepo(db, nil)
	planSvc := services.NewPlanService(planRepo, nil)
	h := NewAdminPlanHandlers(planSvc, planRepo)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plans/", nil)
	h.GetPlan(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("GetPlan: expected 400, got %d", rr.Code)
	}
}

func TestAdminPlanHandlers_UpdateLimits_InvalidID(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	planRepo := repository.NewPostgresPlanRepo(db, nil)
	planSvc := services.NewPlanService(planRepo, nil)
	h := NewAdminPlanHandlers(planSvc, planRepo)

	body := map[string]interface{}{"max_chatbots": 10}
	jb, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/plans//limits", bytes.NewReader(jb))
	h.UpdateLimits(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("UpdateLimits: expected 400, got %d", rr.Code)
	}
}

func TestAdminPlanHandlers_InvalidateCache_InvalidID(t *testing.T) {
	t.Parallel()
	db := testdb.OpenTestDB(t)
	planRepo := repository.NewPostgresPlanRepo(db, nil)
	planSvc := services.NewPlanService(planRepo, nil)
	h := NewAdminPlanHandlers(planSvc, planRepo)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plans//cache-invalidate", nil)
	h.InvalidateCache(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("InvalidateCache: expected 400, got %d", rr.Code)
	}
}
