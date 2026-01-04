package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
)

// TestMockPlanRepo_InterfaceCompliance ensures MockPlanRepo implements PlanRepository
func TestMockPlanRepo_InterfaceCompliance(t *testing.T) {
	var _ PlanRepository = (*MockPlanRepo)(nil)
}

// TestPostgresPlanRepo_InterfaceCompliance ensures PostgresPlanRepo implements PlanRepository
func TestPostgresPlanRepo_InterfaceCompliance(t *testing.T) {
	var _ PlanRepository = (*PostgresPlanRepo)(nil)
}

// TestNewMockPlanRepo verifies that NewMockPlanRepo creates a valid mock
func TestNewMockPlanRepo(t *testing.T) {
	mock := NewMockPlanRepo()
	if mock == nil {
		t.Fatal("NewMockPlanRepo returned nil")
	}
}

// TestNewPostgresPlanRepo verifies that NewPostgresPlanRepo creates a valid repo
func TestNewPostgresPlanRepo(t *testing.T) {
	repo := NewPostgresPlanRepo(nil, nil)
	if repo == nil {
		t.Fatal("NewPostgresPlanRepo returned nil")
	}
}

// TestMockPlanRepo_GetByUserID tests the GetByUserID mock functionality
func TestMockPlanRepo_GetByUserID(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPlanRepo()
		result, err := mock.GetByUserID(context.Background(), "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByUserID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByUserID))
		}
		if mock.Calls.GetByUserID[0].UserID != "user-1" {
			t.Errorf("expected call with UserID 'user-1', got: %s", mock.Calls.GetByUserID[0].UserID)
		}
	})

	t.Run("custom function returns plan", func(t *testing.T) {
		mock := NewMockPlanRepo()
		expectedPlan := &models.Plan{
			ID:    "plan-1",
			Code:  "pro",
			Price: 29.99,
		}
		mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
			return expectedPlan, nil
		}

		result, err := mock.GetByUserID(context.Background(), "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != expectedPlan {
			t.Errorf("expected plan, got: %v", result)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockPlanRepo()
		expectedErr := errors.New("database connection failed")
		mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
			return nil, expectedErr
		}

		_, err := mock.GetByUserID(context.Background(), "any-user")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockPlanRepo_GetByCode tests the GetByCode mock functionality
func TestMockPlanRepo_GetByCode(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPlanRepo()
		result, err := mock.GetByCode(context.Background(), "pro")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByCode) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByCode))
		}
	})

	t.Run("custom function returns plan", func(t *testing.T) {
		mock := NewMockPlanRepo()
		mock.GetByCodeFunc = func(ctx context.Context, code string) (*models.Plan, error) {
			if code == "pro" {
				return &models.Plan{ID: "plan-pro", Code: "pro", Price: 29.99}, nil
			}
			return nil, nil
		}

		result, err := mock.GetByCode(context.Background(), "pro")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil || result.Code != "pro" {
			t.Errorf("expected pro plan, got: %v", result)
		}
	})
}

// TestMockPlanRepo_GetAll tests the GetAll mock functionality
func TestMockPlanRepo_GetAll(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPlanRepo()
		result, err := mock.GetAll(context.Background())
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetAll) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetAll))
		}
	})

	t.Run("custom function returns plans", func(t *testing.T) {
		mock := NewMockPlanRepo()
		mock.GetAllFunc = func(ctx context.Context) ([]models.Plan, error) {
			return []models.Plan{
				{ID: "plan-free", Code: "free", Price: 0},
				{ID: "plan-pro", Code: "pro", Price: 29.99},
			}, nil
		}

		result, err := mock.GetAll(context.Background())
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 plans, got: %d", len(result))
		}
	})
}

// TestMockPlanRepo_GetByID tests the GetByID mock functionality
func TestMockPlanRepo_GetByID(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPlanRepo()
		result, err := mock.GetByID(context.Background(), "plan-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetByID) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetByID))
		}
	})

	t.Run("custom function returns plan", func(t *testing.T) {
		mock := NewMockPlanRepo()
		mock.GetByIDFunc = func(ctx context.Context, id string) (*models.Plan, error) {
			if id == "plan-1" {
				return &models.Plan{ID: "plan-1", Code: "pro"}, nil
			}
			return nil, nil
		}

		result, err := mock.GetByID(context.Background(), "plan-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil || result.ID != "plan-1" {
			t.Errorf("expected plan-1, got: %v", result)
		}
	})
}

// TestMockPlanRepo_InvalidateCache tests the InvalidateCache mock functionality
func TestMockPlanRepo_InvalidateCache(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPlanRepo()
		err := mock.InvalidateCache(context.Background(), "user-1")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.InvalidateCache) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.InvalidateCache))
		}
		if mock.Calls.InvalidateCache[0].UserID != "user-1" {
			t.Errorf("expected user-1, got: %s", mock.Calls.InvalidateCache[0].UserID)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockPlanRepo()
		expectedErr := errors.New("redis connection failed")
		mock.InvalidateCacheFunc = func(ctx context.Context, userID string) error {
			return expectedErr
		}

		err := mock.InvalidateCache(context.Background(), "user-1")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

// TestMockPlanRepo_Reset tests that Reset clears all recorded calls
func TestMockPlanRepo_Reset(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	// Make various calls
	_, _ = mock.GetByUserID(ctx, "user-1")
	_, _ = mock.GetByCode(ctx, "pro")
	_, _ = mock.GetAll(ctx)
	_, _ = mock.GetByID(ctx, "plan-1")
	_ = mock.InvalidateCache(ctx, "user-1")

	// Verify calls were recorded
	if len(mock.Calls.GetByUserID) != 1 {
		t.Error("expected GetByUserID call to be recorded")
	}

	// Reset
	mock.Reset()

	// Verify all calls are cleared
	if len(mock.Calls.GetByUserID) != 0 || len(mock.Calls.GetByCode) != 0 ||
		len(mock.Calls.GetAll) != 0 || len(mock.Calls.GetByID) != 0 ||
		len(mock.Calls.InvalidateCache) != 0 {
		t.Error("expected all calls to be cleared after Reset")
	}
}

// TestMockPlanRepo_MultipleCalls tests recording multiple calls
func TestMockPlanRepo_MultipleCalls(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	// Make multiple GetByUserID calls
	_, _ = mock.GetByUserID(ctx, "user-1")
	_, _ = mock.GetByUserID(ctx, "user-2")
	_, _ = mock.GetByUserID(ctx, "user-3")

	if len(mock.Calls.GetByUserID) != 3 {
		t.Errorf("expected 3 calls, got: %d", len(mock.Calls.GetByUserID))
	}

	// Verify user IDs are correctly recorded
	expectedIDs := []string{"user-1", "user-2", "user-3"}
	for i, call := range mock.Calls.GetByUserID {
		if call.UserID != expectedIDs[i] {
			t.Errorf("call %d: expected UserID %s, got %s", i, expectedIDs[i], call.UserID)
		}
	}
}

// TestMockPlanRepo_ContextPropagation verifies context is passed to custom functions
func TestMockPlanRepo_ContextPropagation(t *testing.T) {
	mock := NewMockPlanRepo()
	type ctxKey string
	key := ctxKey("testKey")
	ctx := context.WithValue(context.Background(), key, "testValue")

	var receivedCtx context.Context
	mock.GetByUserIDFunc = func(c context.Context, userID string) (*models.Plan, error) {
		receivedCtx = c
		return nil, nil
	}

	_, _ = mock.GetByUserID(ctx, "test-user")

	if receivedCtx.Value(key) != "testValue" {
		t.Error("context was not properly propagated")
	}
}

// TestMockPlanRepo_ComplexScenario demonstrates a realistic usage pattern
func TestMockPlanRepo_ComplexScenario(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	// Simulate a store for plans
	plansByUser := make(map[string]*models.Plan)

	mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
		return plansByUser[userID], nil
	}

	mock.GetByCodeFunc = func(ctx context.Context, code string) (*models.Plan, error) {
		for _, p := range plansByUser {
			if p.Code == code {
				return p, nil
			}
		}
		return nil, nil
	}

	// Set up some plans
	plansByUser["user-free"] = &models.Plan{ID: "plan-free", Code: "free", Price: 0}
	plansByUser["user-pro"] = &models.Plan{ID: "plan-pro", Code: "pro", Price: 29.99}

	// Test: Get user's plan
	plan, err := mock.GetByUserID(ctx, "user-pro")
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}
	if plan == nil || plan.Code != "pro" {
		t.Errorf("expected pro plan, got: %v", plan)
	}

	// Verify the sequence of operations
	if len(mock.Calls.GetByUserID) != 1 {
		t.Errorf("expected 1 GetByUserID call, got: %d", len(mock.Calls.GetByUserID))
	}
}

// TestMockPlanRepo_PlanWithLimits tests handling of plans with limits
func TestMockPlanRepo_PlanWithLimits(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	planWithLimits := &models.Plan{
		ID:    "plan-pro",
		Code:  "pro",
		
		Price: 29.99,
		Limits: &models.PlanLimits{
			MaxChatbots:          10,
			MaxMonthlyIngestions: 500,
			ChatDefaultModel:     "openai/gpt-4o",
		},
	}

	mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
		return planWithLimits, nil
	}

	result, err := mock.GetByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected plan, got nil")
	}
	if result.Limits == nil {
		t.Error("expected limits to be set")
	}
	if result.Limits.MaxChatbots != 10 {
		t.Errorf("expected MaxChatbots 10, got: %d", result.Limits.MaxChatbots)
	}
}

// TestMockPlanRepo_ErrorScenarios tests various error handling scenarios
func TestMockPlanRepo_ErrorScenarios(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(*MockPlanRepo)
		execute   func(context.Context, *MockPlanRepo) error
		wantErr   bool
	}{
		{
			name: "GetByUserID database error",
			setupMock: func(m *MockPlanRepo) {
				m.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
					return nil, errors.New("connection refused")
				}
			},
			execute: func(ctx context.Context, m *MockPlanRepo) error {
				_, err := m.GetByUserID(ctx, "user-1")
				return err
			},
			wantErr: true,
		},
		{
			name: "GetByCode not found",
			setupMock: func(m *MockPlanRepo) {
				m.GetByCodeFunc = func(ctx context.Context, code string) (*models.Plan, error) {
					return nil, nil
				}
			},
			execute: func(ctx context.Context, m *MockPlanRepo) error {
				_, err := m.GetByCode(ctx, "nonexistent")
				return err
			},
			wantErr: false,
		},
		{
			name: "GetAll database error",
			setupMock: func(m *MockPlanRepo) {
				m.GetAllFunc = func(ctx context.Context) ([]models.Plan, error) {
					return nil, errors.New("query timeout")
				}
			},
			execute: func(ctx context.Context, m *MockPlanRepo) error {
				_, err := m.GetAll(ctx)
				return err
			},
			wantErr: true,
		},
		{
			name: "InvalidateCache redis error",
			setupMock: func(m *MockPlanRepo) {
				m.InvalidateCacheFunc = func(ctx context.Context, userID string) error {
					return errors.New("redis connection failed")
				}
			},
			execute: func(ctx context.Context, m *MockPlanRepo) error {
				return m.InvalidateCache(ctx, "user-1")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockPlanRepo()
			tc.setupMock(mock)
			err := tc.execute(context.Background(), mock)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// TestMockPlanRepo_Stateful demonstrates using the mock for stateful CRUD testing
func TestMockPlanRepo_Stateful(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	// Create an in-memory store
	plans := make(map[string]*models.Plan)

	mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
		return plans[userID], nil
	}

	mock.GetByCodeFunc = func(ctx context.Context, code string) (*models.Plan, error) {
		for _, p := range plans {
			if p.Code == code {
				return p, nil
			}
		}
		return nil, nil
	}

	// Test: Set a plan for a user
	plans["user-1"] = &models.Plan{ID: "plan-pro", Code: "pro", Price: 29.99}

	// Retrieve and verify
	got, err := mock.GetByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}
	if got == nil || got.Code != "pro" {
		t.Errorf("expected pro plan, got: %v", got)
	}

	// Verify by code
	byCode, err := mock.GetByCode(ctx, "pro")
	if err != nil {
		t.Fatalf("GetByCode failed: %v", err)
	}
	if byCode == nil || byCode.ID != "plan-pro" {
		t.Errorf("expected plan-pro, got: %v", byCode)
	}
}

// TestMockPlanRepo_CacheInvalidation demonstrates cache invalidation workflow
func TestMockPlanRepo_CacheInvalidation(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	var invalidateCalls int

	mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
		return &models.Plan{ID: "plan-" + userID, Code: "free"}, nil
	}

	mock.InvalidateCacheFunc = func(ctx context.Context, userID string) error {
		invalidateCalls++
		return nil
	}

	// First call - should not use cache
	plan1, _ := mock.GetByUserID(ctx, "user-1")
	_ = mock.InvalidateCache(ctx, "user-1")

	// Second call - after invalidation, still returns fresh data
	plan2, _ := mock.GetByUserID(ctx, "user-1")

	// Both calls should succeed
	if plan1 == nil || plan2 == nil {
		t.Error("expected plans to be returned")
	}

	// Verify invalidation was called once
	if invalidateCalls != 1 {
		t.Errorf("expected 1 invalidation call, got: %d", invalidateCalls)
	}
}

// TestMockPlanRepo_PlanJSONSerialization tests that plans can be serialized to JSON
func TestMockPlanRepo_PlanJSONSerialization(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	plan := &models.Plan{
		ID:    "plan-1",
		Code:  "pro",
		
		Price: 29.99,
		Limits: &models.PlanLimits{
			MaxChatbots:          10,
			ChatAllowedModels:    []string{"openai/gpt-4o", "openai/gpt-4o-mini"},
			ChatMaxMonthlyTokens: 1000000,
		},
	}

	mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
		return plan, nil
	}

	result, err := mock.GetByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}

	// Serialize to JSON (simulates Redis caching)
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal plan: %v", err)
	}

	// Deserialize back
	var unmarshaled models.Plan
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal plan: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != plan.ID {
		t.Errorf("expected ID %s, got: %s", plan.ID, unmarshaled.ID)
	}
	if unmarshaled.Code != plan.Code {
		t.Errorf("expected Code %s, got: %s", plan.Code, unmarshaled.Code)
	}
	if unmarshaled.Limits == nil {
		t.Error("expected Limits to be preserved")
	}
	if len(unmarshaled.Limits.ChatAllowedModels) != 2 {
		t.Errorf("expected 2 allowed models, got: %d", len(unmarshaled.Limits.ChatAllowedModels))
	}
}

// TestMockPlanRepo_DefaultPlanLimits tests that default plan limits can be used
func TestMockPlanRepo_DefaultPlanLimits(t *testing.T) {
	mock := NewMockPlanRepo()
	ctx := context.Background()

	plan := &models.Plan{
		ID:    "plan-free",
		Code:  "free",
		
		Price: 0,
		Limits: &models.PlanLimits{
			MaxChatbots:               1,
			MaxMonthlyIngestions:      50,
			MaxMonthlyEmbeddingTokens: 250000,
		},
	}

	mock.GetByUserIDFunc = func(ctx context.Context, userID string) (*models.Plan, error) {
		return plan, nil
	}

	result, err := mock.GetByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetByUserID failed: %v", err)
	}

	if result.Limits.MaxChatbots != 1 {
		t.Errorf("expected MaxChatbots 1, got: %d", result.Limits.MaxChatbots)
	}
	if result.Limits.MaxMonthlyIngestions != 50 {
		t.Errorf("expected MaxMonthlyIngestions 50, got: %d", result.Limits.MaxMonthlyIngestions)
	}
}

// Helper function to create a pointer to a time
func ptrTime(t time.Time) *time.Time {
	return &t
}
