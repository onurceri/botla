package repository

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

// MockPlanRepo is a mock implementation of PlanRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockPlanRepo struct {
	// GetByUserIDFunc is called when GetByUserID is invoked.
	GetByUserIDFunc func(ctx context.Context, userID string) (*models.Plan, error)

	// GetByCodeFunc is called when GetByCode is invoked.
	GetByCodeFunc func(ctx context.Context, code string) (*models.Plan, error)

	// GetAllFunc is called when GetAll is invoked.
	GetAllFunc func(ctx context.Context) ([]models.Plan, error)

	// GetByIDFunc is called when GetByID is invoked.
	GetByIDFunc func(ctx context.Context, id string) (*models.Plan, error)

	// GetPlanWithLimitsFunc is called when GetPlanWithLimits is invoked.
	GetPlanWithLimitsFunc func(ctx context.Context, userID string) (*models.Plan, error)

	// GetAllPlansWithLimitsFunc is called when GetAllPlansWithLimits is invoked.
	GetAllPlansWithLimitsFunc func(ctx context.Context) ([]models.Plan, error)

	// InvalidateCacheFunc is called when InvalidateCache is invoked.
	InvalidateCacheFunc func(ctx context.Context, userID string) error

	// Invocation tracking for test assertions
	Calls struct {
		GetByUserID           []PlanGetByUserIDCall
		GetByCode             []PlanGetByCodeCall
		GetAll                []PlanGetAllCall
		GetByID               []PlanGetByIDCall
		GetPlanWithLimits     []PlanGetByUserIDCall
		GetAllPlansWithLimits []PlanGetAllCall
		InvalidateCache       []PlanInvalidateCacheCall
	}
}

// Call recording types for test verification
type PlanGetByUserIDCall struct {
	UserID string
}

type PlanGetByCodeCall struct {
	Code string
}

type PlanGetAllCall struct{}

type PlanGetByIDCall struct {
	ID string
}

type PlanInvalidateCacheCall struct {
	UserID string
}

// Compile-time check that MockPlanRepo implements PlanRepository.
var _ PlanRepository = (*MockPlanRepo)(nil)

// NewMockPlanRepo creates a new MockPlanRepo with default no-op behavior.
func NewMockPlanRepo() *MockPlanRepo {
	return &MockPlanRepo{}
}

// GetByUserID retrieves the active plan for a user.
func (m *MockPlanRepo) GetByUserID(ctx context.Context, userID string) (*models.Plan, error) {
	m.Calls.GetByUserID = append(m.Calls.GetByUserID, PlanGetByUserIDCall{UserID: userID})
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

// GetByCode retrieves a plan by its code.
func (m *MockPlanRepo) GetByCode(ctx context.Context, code string) (*models.Plan, error) {
	m.Calls.GetByCode = append(m.Calls.GetByCode, PlanGetByCodeCall{Code: code})
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	return nil, nil
}

// GetAll retrieves all active plans.
func (m *MockPlanRepo) GetAll(ctx context.Context) ([]models.Plan, error) {
	m.Calls.GetAll = append(m.Calls.GetAll, PlanGetAllCall{})
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

// GetPlanWithLimits retrieves a plan by user ID with all limits populated.
func (m *MockPlanRepo) GetPlanWithLimits(ctx context.Context, userID string) (*models.Plan, error) {
	m.Calls.GetPlanWithLimits = append(m.Calls.GetPlanWithLimits, PlanGetByUserIDCall{UserID: userID})
	if m.GetPlanWithLimitsFunc != nil {
		return m.GetPlanWithLimitsFunc(ctx, userID)
	}
	return nil, nil
}

// GetAllPlansWithLimits retrieves all active plans with their limits.
func (m *MockPlanRepo) GetAllPlansWithLimits(ctx context.Context) ([]models.Plan, error) {
	m.Calls.GetAllPlansWithLimits = append(m.Calls.GetAllPlansWithLimits, PlanGetAllCall{})
	if m.GetAllPlansWithLimitsFunc != nil {
		return m.GetAllPlansWithLimitsFunc(ctx)
	}
	return nil, nil
}

// GetByID retrieves a plan by its unique identifier.
func (m *MockPlanRepo) GetByID(ctx context.Context, id string) (*models.Plan, error) {
	m.Calls.GetByID = append(m.Calls.GetByID, PlanGetByIDCall{ID: id})
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// InvalidateCache removes the cached plan for a user.
func (m *MockPlanRepo) InvalidateCache(ctx context.Context, userID string) error {
	m.Calls.InvalidateCache = append(m.Calls.InvalidateCache, PlanInvalidateCacheCall{UserID: userID})
	if m.InvalidateCacheFunc != nil {
		return m.InvalidateCacheFunc(ctx, userID)
	}
	return nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockPlanRepo) Reset() {
	m.Calls.GetByUserID = nil
	m.Calls.GetByCode = nil
	m.Calls.GetAll = nil
	m.Calls.GetByID = nil
	m.Calls.GetPlanWithLimits = nil
	m.Calls.GetAllPlansWithLimits = nil
	m.Calls.InvalidateCache = nil
}
