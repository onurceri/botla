# Task 004: Refactor ActionHandlers

## Background

Refactor `ActionHandlers` to use `ActionRepository` interface instead of direct `*sql.DB`. This is the primary handler identified in the issue evidence.

**Depends on:** Task 002, Task 003

**File:** `internal/api/handlers/action.go`

## Current State

```go
type ActionHandlers struct {
    DB *sql.DB  // <--- Direct DB dependency
    ToolNameGenerator *rag.ToolNameGenerator
    WorkspaceService  *services.WorkspaceService
    OrgService        *services.OrganizationService
}
```

## Target State

```go
type ActionHandlers struct {
    ActionRepo        repository.ActionRepository
    ChatbotRepo       repository.ChatbotRepository
    ToolNameGenerator *rag.ToolNameGenerator
    WorkspaceService  *services.WorkspaceService
    OrgService        *services.OrganizationService
}
```

## Implementation Plan

1. **Update struct definition** in `action.go`
2. **Update handler methods** to use repository methods
   - `h.DB` → `h.ActionRepo` / `h.ChatbotRepo`
   - `db.GetActions(ctx, h.DB, ...)` → `h.ActionRepo.List(ctx, ...)`
3. **Update `authorize` helper** to use `ChatbotRepo`
4. **Update wire-up in `cmd/server/main.go`**

## Checklist

- [x] Update `ActionHandlers` struct to use repository interfaces
- [x] Update `List` method to use `ActionRepo.List`
- [x] Update `Create` method to use `ActionRepo.Create`
- [x] Update `Get` method to use `ActionRepo.GetByID`
- [x] Update `Update` method to use `ActionRepo.Update`
- [x] Update `Delete` method to use `ActionRepo.Delete`
- [x] Update `GetLogs` method to use `ActionRepo.GetLogs`
- [x] Update wire-up in `cmd/server/main.go` (done in `router.go` and `testserver.go`)
- [x] Run existing tests to verify no regressions
- [x] Add unit test using mock repository
