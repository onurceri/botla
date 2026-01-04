# Tech Debt Task 005: Create Plan Repository

> **Agent Prompt**: You are creating a new `PlanRepository` interface and implementation to encapsulate plan-related database operations. Currently, services and handlers call `db.GetPlanByUserID` and related functions directly. Your task is to define the interface in `internal/repository/interfaces.go`, implement `PostgresPlanRepo` using **Squirrel SQL builder** (`github.com/Masterminds/squirrel v1.5.4`), and prepare for consumer migration. Apply strict TDD.

## Background

Plan-related functions are called directly from services:
- `internal/services/chat_service.go`: `db.GetPlanByUserID`
- `internal/services/chat_helpers.go`: `db.GetPlanByUserID`
- `internal/services/refresh_scheduler.go`: `db.GetPlanByUserID`
- `internal/api/handlers/source_create.go`: `db.GetPlanByUserID`
- `internal/api/handlers/source_refresh.go`: `db.GetPlanByUserID`

The `internal/db/plan.go` file contains the implementation.

## Goal

1. Define `PlanRepository` interface in `internal/repository/interfaces.go`
2. Create `internal/repository/plan_repo.go` with `PostgresPlanRepo`
3. Create mock implementation for testing
4. Do NOT migrate consumers yet (that's a separate task)

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/repository/interfaces.go` | Add `PlanRepository` interface |
| `internal/repository/plan_repo.go` | **[NEW]** Implement `PostgresPlanRepo` |
| `internal/repository/plan_repo_test.go` | **[NEW]** Integration tests |
| `internal/repository/mock_plan_repo.go` | **[NEW]** Mock implementation |
| `internal/db/plan.go` | Reference for SQL logic |

## Integration Plan

### Step 1: Analyze Existing db Functions

Review `internal/db/plan.go` to understand:
- What functions exist
- What models are used
- What queries are performed

### Step 2: Define Interface

```go
// internal/repository/interfaces.go

// PlanRepository defines the interface for plan data access operations.
// Plans define feature limits and pricing tiers for users.
type PlanRepository interface {
    // GetByUserID retrieves the active plan for a user.
    // Returns the default plan if user has no explicit plan.
    GetByUserID(ctx context.Context, userID string) (*models.Plan, error)
    
    // GetByCode retrieves a plan by its code (e.g., "free", "pro", "enterprise").
    GetByCode(ctx context.Context, code string) (*models.Plan, error)
    
    // GetAll retrieves all active plans.
    GetAll(ctx context.Context) ([]models.Plan, error)
}
```

### Step 3: Implement with Squirrel

```go
// internal/repository/plan_repo.go
package repository

import (
    "context"
    "database/sql"
    
    sq "github.com/Masterminds/squirrel"
    "github.com/onurceri/botla-app/internal/models"
)

type PostgresPlanRepo struct {
    pool *sql.DB
}

var _ PlanRepository = (*PostgresPlanRepo)(nil)

func NewPostgresPlanRepo(pool *sql.DB) *PostgresPlanRepo {
    return &PostgresPlanRepo{pool: pool}
}

func (r *PostgresPlanRepo) GetByUserID(ctx context.Context, userID string) (*models.Plan, error) {
    // Query user's plan, fallback to default
    query, args, err := psql.
        Select("p.id", "p.code", "p.name", "p.config", "p.is_active").
        From("plans p").
        Join("users u ON u.plan_id = p.id").
        Where(sq.Eq{"u.id": userID}).
        Where(sq.Eq{"p.is_active": true}).
        ToSql()
    // ...
}
```

### Step 4: Create Mock

```go
// internal/repository/mock_plan_repo.go
type MockPlanRepo struct {
    GetByUserIDFunc func(ctx context.Context, userID string) (*models.Plan, error)
    GetByCodeFunc   func(ctx context.Context, code string) (*models.Plan, error)
    GetAllFunc      func(ctx context.Context) ([]models.Plan, error)
}
```

## Edge Cases

- **User with no plan**: Should return default/free plan
- **Inactive plans**: Should not be returned
- **Plan config parsing**: Plans have JSONB config that needs unmarshaling

## Checklist

- [ ] Add `PlanRepository` interface to `interfaces.go`
- [ ] Create `plan_repo.go` with struct skeleton
- [ ] Write test for `GetByUserID`
- [ ] Implement `GetByUserID` using Squirrel
- [ ] Write test for `GetByCode`
- [ ] Implement `GetByCode` using Squirrel
- [ ] Write test for `GetAll`
- [ ] Implement `GetAll` using Squirrel
- [ ] Create `mock_plan_repo.go`
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify all tests pass

## Verification

```bash
go test ./internal/repository/plan_repo_test.go -v
make test-all
```
