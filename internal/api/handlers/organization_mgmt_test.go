package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db connection failed: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping test: db not available: %v", err)
	}
	return db
}

func createTestUser(t *testing.T, db *sql.DB) string {
	var uid string
	email := fmt.Sprintf("testuser_%d@example.com", time.Now().UnixNano())
	// Ensure plan exists
	var planID string
	err := db.QueryRow(`SELECT id FROM plans LIMIT 1`).Scan(&planID)
	if err != nil {
		// Create a dummy plan if none exists
		planID = "basic"
		_, _ = db.Exec(`INSERT INTO plans (id, name, monthly_price, yearly_price, currency, features) VALUES ($1, 'Basic', 0, 0, 'USD', '{}') ON CONFLICT DO NOTHING`, planID)
	}

	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "hash", planID).Scan(&uid); err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	return uid
}

func createTestOrg(t *testing.T, db *sql.DB, ownerID string) (string, *services.OrganizationService) {
	log := logger.New("test")
	svc := services.NewOrganizationService(db, log)

	name := fmt.Sprintf("Test Org %d", time.Now().UnixNano())
	slug := fmt.Sprintf("test-org-%d", time.Now().UnixNano())

	org, err := svc.CreateOrganization(context.Background(), name, slug, ownerID)
	if err != nil {
		t.Fatalf("create org failed: %v", err)
	}
	return org.ID, svc
}

func TestOrganizationManagement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ownerID := createTestUser(t, db)
	orgID, svc := createTestOrg(t, db, ownerID)
	h := &OrganizationHandlers{OrgService: svc, DB: db}

	t.Run("UpdateOrganization", func(t *testing.T) {
		newSlug := fmt.Sprintf("updated-slug-%d", time.Now().UnixNano())
		body := map[string]string{
			"name": "Updated Name",
			"slug": newSlug,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/organizations/"+orgID, bytes.NewBuffer(jsonBody))

		// Mock Context
		ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, ownerID)
		ctx = context.WithValue(ctx, middleware.ContextKeyOrgID, orgID)

		rr := httptest.NewRecorder()
		h.UpdateOrganization(rr, req.WithContext(ctx))

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}

		// Verify update
		var name string
		err := db.QueryRow("SELECT name FROM organizations WHERE id = $1", orgID).Scan(&name)
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if name != "Updated Name" {
			t.Errorf("expected name 'Updated Name', got '%s'", name)
		}
	})

	t.Run("AddMember", func(t *testing.T) {
		// Create another user to add
		memberID := createTestUser(t, db)
		var memberEmail string
		db.QueryRow("SELECT email FROM users WHERE id = $1", memberID).Scan(&memberEmail)

		body := map[string]string{
			"email": memberEmail,
			"role":  "member",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/organizations/%s/members", orgID), bytes.NewBuffer(jsonBody))

		// Mock Context (Owner is adding)
		ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, ownerID)
		ctx = context.WithValue(ctx, middleware.ContextKeyOrgID, orgID)

		rr := httptest.NewRecorder()
		h.AddMember(rr, req.WithContext(ctx))

		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201 Created, got %d", rr.Code)
		}

		// Verify membership
		exists := false
		db.QueryRow("SELECT EXISTS(SELECT 1 FROM memberships WHERE organization_id=$1 AND user_id=$2)", orgID, memberID).Scan(&exists)
		if !exists {
			t.Error("member not added to database")
		}
	})

	t.Run("DeleteWorkspace", func(t *testing.T) {
		// Create workspace 1
		ws1, _ := svc.CreateWorkspace(context.Background(), orgID, "WS1", "ws1", nil)
		// Create workspace 2
		ws2, _ := svc.CreateWorkspace(context.Background(), orgID, "WS2", "ws2", nil)

		// Delete WS1 (should succeed as count=2)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+ws1.ID, nil)
		req.SetPathValue("wsID", ws1.ID)
		ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, ownerID)
		ctx = context.WithValue(ctx, middleware.ContextKeyOrgID, orgID)

		rr := httptest.NewRecorder()
		h.DeleteWorkspace(rr, req.WithContext(ctx))

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content for DeleteWorkspace, got %d", rr.Code)
		}

		// Delete WS2 (should fail as it is the last one)
		req2 := httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/"+ws2.ID, nil)
		req2.SetPathValue("wsID", ws2.ID)
		ctx2 := context.WithValue(req2.Context(), middleware.ContextKeyUserID, ownerID)
		ctx2 = context.WithValue(ctx2, middleware.ContextKeyOrgID, orgID)

		rr2 := httptest.NewRecorder()
		h.DeleteWorkspace(rr2, req2.WithContext(ctx2))

		if rr2.Code != http.StatusBadRequest {
			t.Errorf("expected 400 BadRequest for last workspace deletion, got %d", rr2.Code)
		}
	})

	t.Run("DeleteOrganization", func(t *testing.T) {
		// Create a second organization so we can delete the first one
		secondOrgSlug := fmt.Sprintf("second-org-%d", time.Now().UnixNano())
		secondOrg, err := svc.CreateOrganization(context.Background(), "Second Org", secondOrgSlug, ownerID)
		if err != nil {
			t.Fatalf("failed to create second org: %v", err)
		}
		if secondOrg == nil {
			t.Fatalf("second org is nil")
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+orgID, nil)

		// Mock Context
		ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, ownerID)
		ctx = context.WithValue(ctx, middleware.ContextKeyOrgID, orgID)

		rr := httptest.NewRecorder()
		h.DeleteOrganization(rr, req.WithContext(ctx))

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content, got %d", rr.Code)
		}

		// Verify deletion
		exists := false
		db.QueryRow("SELECT EXISTS(SELECT 1 FROM organizations WHERE id=$1)", orgID).Scan(&exists)
		if exists {
			t.Error("organization not deleted from database")
		}
	})

	t.Run("DeleteLastOrganization_Fail", func(t *testing.T) {
		// Find the remaining org (the second one created above)
		var remainingOrgID string
		db.QueryRow("SELECT id FROM organizations WHERE owner_id=$1 LIMIT 1", ownerID).Scan(&remainingOrgID)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+remainingOrgID, nil)
		ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, ownerID)
		ctx = context.WithValue(ctx, middleware.ContextKeyOrgID, remainingOrgID)

		rr := httptest.NewRecorder()
		h.DeleteOrganization(rr, req.WithContext(ctx))

		// Should be 400 BadRequest (because service returns specific error)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400 BadRequest for last org deletion, got %d", rr.Code)
		}
	})
}
