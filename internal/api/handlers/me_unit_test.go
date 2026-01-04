package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/middleware"
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

	userRepo := repository.NewPostgresUserRepo(db)
	orgSvc := services.NewOrganizationService(db, nil)
	h := NewMeHandlers(userRepo, orgSvc)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, uid)
	h.Me(rr, req.WithContext(ctx))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestMe_IsPlatformAdmin_DefaultFalse(t *testing.T) {
	db := testdb.OpenTestDB(t)
	var uid string
	var proPlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("meadmin+%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`, email, "x", proPlanID).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	userRepo := repository.NewPostgresUserRepo(db)
	orgSvc := services.NewOrganizationService(db, nil)
	h := NewMeHandlers(userRepo, orgSvc)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, uid)
	h.Me(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rr.Code)
	}

	var resp MeResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// By default, users should not be platform admin
	if resp.IsPlatformAdmin {
		t.Error("expected IsPlatformAdmin to be false by default")
	}
}

func TestMe_IsPlatformAdmin_WhenTrue(t *testing.T) {
	db := testdb.OpenTestDB(t)
	var uid string
	var proPlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("meadmintrue+%d@example.com", time.Now().UnixNano())
	if err := db.QueryRow(`INSERT INTO users (email, password_hash, plan_id, is_platform_admin) VALUES ($1,$2,$3,$4) RETURNING id`, email, "x", proPlanID, true).Scan(&uid); err != nil {
		t.Fatalf("user: %v", err)
	}

	userRepo := repository.NewPostgresUserRepo(db)
	orgSvc := services.NewOrganizationService(db, nil)
	h := NewMeHandlers(userRepo, orgSvc)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, uid)
	h.Me(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rr.Code)
	}

	var resp MeResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Should return true for admin users
	if !resp.IsPlatformAdmin {
		t.Error("expected IsPlatformAdmin to be true for admin user")
	}
}

func TestMe_Unauthorized(t *testing.T) {
	db := testdb.OpenTestDB(t)
	userRepo := repository.NewPostgresUserRepo(db)
	orgSvc := services.NewOrganizationService(db, nil)
	h := NewMeHandlers(userRepo, orgSvc)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	// No user ID in context
	h.Me(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got: %d", rr.Code)
	}
}

func TestMe_UserNotFound(t *testing.T) {
	db := testdb.OpenTestDB(t)
	userRepo := repository.NewPostgresUserRepo(db)
	orgSvc := services.NewOrganizationService(db, nil)
	h := NewMeHandlers(userRepo, orgSvc)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	// Use a non-existent user ID
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, "00000000-0000-0000-0000-000000000000")
	h.Me(rr, req.WithContext(ctx))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got: %d", rr.Code)
	}
}

// MockUserRepo for testing without database
type MockUserRepo struct {
	users map[string]*models.User
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	return m.users[id], nil
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}
