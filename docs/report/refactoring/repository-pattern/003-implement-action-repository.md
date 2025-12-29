# Task 003: Implement ActionRepository

## Background

Implement the `ActionRepository` interface by wrapping existing `db.*` functions from `internal/db/action.go`.

**Depends on:** Task 001

## Implementation Plan

1. **Create PostgresActionRepo**
   - File: `internal/repository/action_repo.go`
   - Wraps existing db functions:
     - `db.GetActions` → `List`
     - `db.GetActionByID` → `GetByID`
     - `db.CreateAction` → `Create`
     - `db.UpdateAction` → `Update`
     - `db.DeleteAction` → `Delete`
     - `db.GetActionLogs` → `GetLogs`

2. **Create mock implementation**
   - File: `internal/repository/mock_action_repo.go`

## Files to Create

| File | Purpose |
|------|---------|
| `internal/repository/action_repo.go` | PostgreSQL implementation |
| `internal/repository/mock_action_repo.go` | Mock for unit tests |

## Checklist

- [x] Create `action_repo.go` with `PostgresActionRepo` struct
- [x] Implement all `ActionRepository` interface methods
- [x] Create `mock_action_repo.go` with mock implementation
- [x] Run `go build ./...` to verify compilation
- [x] Write comprehensive tests for the implemented repository

