package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func TestGetOnboardingState_Success(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Create test user
	var userID string
	email := fmt.Sprintf("onboarding-get+%d@example.com", time.Now().UnixNano())
	var planID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id, onboarding_step, onboarding_completed) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		email, "hash", planID, 2, false).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := &OnboardingHandlers{DB: pool, UserRepo: repository.NewPostgresUserRepo(pool)}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/onboarding", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.GetOnboardingState(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp["completed"] != false {
		t.Errorf("expected completed=false, got %v", resp["completed"])
	}
	if resp["step"] != float64(2) {
		t.Errorf("expected step=2, got %v", resp["step"])
	}
}

func TestUpdateOnboardingState_Success(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Create test user
	var userID string
	email := fmt.Sprintf("onboarding-update+%d@example.com", time.Now().UnixNano())
	var planID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "hash", planID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := &OnboardingHandlers{DB: pool, UserRepo: repository.NewPostgresUserRepo(pool)}

	// Update to step 2
	reqBody := map[string]interface{}{
		"step": 2,
		"data": map[string]interface{}{
			"bot_name":    "Test Bot",
			"source_type": "text",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/me/onboarding", bytes.NewReader(bodyBytes))
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.UpdateOnboardingState(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify in DB
	user, err := repository.NewPostgresUserRepo(pool).GetByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.OnboardingStep != 2 {
		t.Errorf("expected step=2, got %d", user.OnboardingStep)
	}
	if user.OnboardingData == nil || user.OnboardingData.BotName != "Test Bot" {
		t.Errorf("expected bot_name='Test Bot', got %v", user.OnboardingData)
	}
}

func TestSkipOnboarding_Success(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Create test user
	var userID string
	email := fmt.Sprintf("onboarding-skip+%d@example.com", time.Now().UnixNano())
	var planID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "hash", planID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := &OnboardingHandlers{DB: pool, UserRepo: repository.NewPostgresUserRepo(pool)}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding/skip", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.SkipOnboarding(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Verify in DB
	user, err := repository.NewPostgresUserRepo(pool).GetByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if !user.OnboardingSkipped {
		t.Error("expected onboarding_skipped=true")
	}
}

func TestCompleteOnboarding_Success(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Create test user
	var userID string
	email := fmt.Sprintf("onboarding-complete+%d@example.com", time.Now().UnixNano())
	var planID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "hash", planID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := &OnboardingHandlers{DB: pool, UserRepo: repository.NewPostgresUserRepo(pool)}

	reqBody := map[string]string{"bot_id": "bot-123"}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/onboarding/complete", bytes.NewReader(bodyBytes))
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.CompleteOnboarding(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify in DB
	user, err := repository.NewPostgresUserRepo(pool).GetByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if !user.OnboardingCompleted {
		t.Error("expected onboarding_completed=true")
	}
	if user.OnboardingStep != 4 {
		t.Errorf("expected step=4, got %d", user.OnboardingStep)
	}
	if user.OnboardingData == nil || user.OnboardingData.CreatedBotID != "bot-123" {
		t.Errorf("expected created_bot_id='bot-123', got %v", user.OnboardingData)
	}
}

func TestGetOnboardingState_Unauthorized(t *testing.T) {
	pool := testdb.OpenTestDB(t)
	h := &OnboardingHandlers{DB: pool, UserRepo: repository.NewPostgresUserRepo(pool)}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/onboarding", nil)
	rr := httptest.NewRecorder()

	h.GetOnboardingState(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestUpdateOnboardingState_InvalidJSON(t *testing.T) {
	pool := testdb.OpenTestDB(t)

	// Create test user
	var userID string
	email := fmt.Sprintf("onboarding-invalid+%d@example.com", time.Now().UnixNano())
	var planID string
	if err := pool.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&planID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := pool.QueryRow(`INSERT INTO users (email, password_hash, plan_id) VALUES ($1,$2,$3) RETURNING id`,
		email, "hash", planID).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := &OnboardingHandlers{DB: pool, UserRepo: repository.NewPostgresUserRepo(pool)}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/me/onboarding", bytes.NewReader([]byte("invalid json")))
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	rr := httptest.NewRecorder()

	h.UpdateOnboardingState(rr, req.WithContext(ctx))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
