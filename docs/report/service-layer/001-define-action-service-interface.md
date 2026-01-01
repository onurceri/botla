# Task 001: Define ActionService Interface

> **Agent Prompt**: You are introducing a service layer to decouple business logic from HTTP handlers. Your first task is to define the `ActionService` interface in `internal/services/`. This interface will encapsulate all business operations for chatbot actions, including the logic currently embedded in handlers (e.g., ToolName generation when name/description changes). Apply strict TDD: write interface tests first to validate the contract, then define the interface. Do not modify handlers yet—that comes in a later task.

## Background

Currently, `internal/api/handlers/action.go` contains business logic that should be in a service layer:

```go
// Current handler (internal/api/handlers/action.go)
func (h *ActionHandlers) Update(w http.ResponseWriter, r *http.Request) {
    // ... HTTP parsing ...
    
    // Business logic that should NOT be here:
    if nameChanged || descChanged || toolNameMissing {
        toolName, err = h.ToolNameGenerator.Generate(r.Context(), newName, newDesc)
        action.ToolName = &toolName
    }
    
    // ... more business logic mixed with HTTP ...
}
```

## Goal

Define an `ActionService` interface that will contain all action-related business logic, making it:
- Unit-testable without HTTP mocking
- Reusable from CLI, background jobs, or other entry points
- Clear about its responsibilities and dependencies

## Interface Design

Create `internal/services/action_service.go`:

```go
package services

import (
    "context"
    "github.com/onurceri/botla-co/internal/models"
)

// ActionService handles business logic for chatbot actions.
// It orchestrates repository calls and side effects (like LLM-based ToolName generation).
type ActionService interface {
    // CreateAction creates a new action with auto-generated ToolName.
    // It validates the input, generates a ToolName using LLM, and persists the action.
    CreateAction(ctx context.Context, chatbotID string, input CreateActionInput) (*models.ChatbotAction, error)

    // UpdateAction updates an existing action.
    // If name or description changed (or ToolName is missing), it regenerates the ToolName.
    // Returns ErrVersionConflict if optimistic locking fails.
    UpdateAction(ctx context.Context, actionID string, input UpdateActionInput) (*models.ChatbotAction, error)

    // GetAction retrieves an action by ID. Returns nil, nil if not found.
    GetAction(ctx context.Context, actionID string) (*models.ChatbotAction, error)

    // ListActions retrieves all actions for a chatbot.
    ListActions(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error)

    // DeleteAction removes an action by ID.
    DeleteAction(ctx context.Context, actionID string) error

    // GetActionLogs retrieves execution logs for a chatbot's actions.
    GetActionLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error)
}

// CreateActionInput contains the data needed to create a new action.
type CreateActionInput struct {
    Name        string
    Description string
    ActionType  string
    Config      []byte // JSON
    Parameters  []byte // JSON
    Enabled     bool
}

// UpdateActionInput contains the data for updating an action.
// Zero values mean "don't update this field".
type UpdateActionInput struct {
    Name        *string
    Description *string
    ActionType  *string
    Config      []byte
    Parameters  []byte
    Enabled     *bool
}
```

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/services/action_service.go` | **[NEW]** Define interface and input types |

## Checklist

- [ ] Create `internal/services/action_service.go`
- [ ] Define `ActionService` interface with method signatures
- [ ] Define `CreateActionInput` struct
- [ ] Define `UpdateActionInput` struct  
- [ ] Add comprehensive doc comments explaining each method's behavior
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make lint` to check for issues

## Edge Cases to Document

In the interface doc comments, document these behaviors:

1. **CreateAction**: 
   - Returns error if `Name` or `ActionType` is empty
   - Always calls ToolNameGenerator for new actions

2. **UpdateAction**:
   - Returns `ErrVersionConflict` if action was modified concurrently
   - Returns error if action not found
   - Only regenerates ToolName if name/description changed OR ToolName is nil/empty. This decision should mention support for migrating from DB with legacy data.

3. **GetAction**:
   - Returns `nil, nil` for not-found (not an error)

4. **ListActions**:
   - Returns empty slice (not nil) if no actions exist

## Verification

```bash
# Verify compilation
go build ./...

# Run linter
make lint

# Verify interface is importable
grep -r "ActionService" internal/services/
```

## Next Task

After this task, proceed to [002-implement-action-service.md](./002-implement-action-service.md) to implement the interface.
