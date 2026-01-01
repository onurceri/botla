# Task 002: Implement ActionService

> **Agent Prompt**: You are implementing the `ActionService` interface defined in Task 001. Create a concrete implementation that encapsulates all business logic currently scattered in `internal/api/handlers/action.go`. The service should orchestrate repository calls, LLM-based ToolName generation, and version conflict handling. Apply strict TDD: write failing tests first for each method, then implement the minimal code to pass. This service must be fully testable without HTTP mocking.

## Background

The `ActionService` interface was defined in Task 001. Now we need to implement it by:

1. Moving business logic out of `internal/api/handlers/action.go`
2. Injecting dependencies (repository, ToolNameGenerator)
3. Making the service fully unit-testable

## Goal

Create `ActionServiceImpl` that implements `ActionService` with:
- All business logic extracted from handlers
- Proper dependency injection
- Clear separation of concerns
- Full unit test coverage

## Implementation

Create `internal/services/action_service_impl.go`:

```go
package services

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"

    "github.com/onurceri/botla-co/internal/models"
    "github.com/onurceri/botla-co/internal/rag"
    "github.com/onurceri/botla-co/internal/repository"
)

// Sentinel errors for ActionService
var (
    ErrActionNameRequired     = errors.New("action name is required")
    ErrActionTypeRequired     = errors.New("action type is required")
    ErrActionNotFound         = errors.New("action not found")
)

// ActionServiceImpl implements ActionService
type ActionServiceImpl struct {
    actionRepo        repository.ActionRepository
    toolNameGenerator *rag.ToolNameGenerator
}

// NewActionService creates a new ActionService implementation
func NewActionService(
    actionRepo repository.ActionRepository,
    toolNameGenerator *rag.ToolNameGenerator,
) *ActionServiceImpl {
    return &ActionServiceImpl{
        actionRepo:        actionRepo,
        toolNameGenerator: toolNameGenerator,
    }
}

func (s *ActionServiceImpl) CreateAction(ctx context.Context, chatbotID string, input CreateActionInput) (*models.ChatbotAction, error) {
    // Validation
    if input.Name == "" {
        return nil, ErrActionNameRequired
    }
    if input.ActionType == "" {
        return nil, ErrActionTypeRequired
    }

    // Business logic: Generate ToolName (this was in the handler)
    toolName, err := s.toolNameGenerator.Generate(ctx, input.Name, input.Description)
    if err != nil {
        return nil, fmt.Errorf("failed to generate tool name: %w", err)
    }

    // Prepare model
    action := &models.ChatbotAction{
        ChatbotID:  chatbotID,
        Name:       input.Name,
        ActionType: models.ActionType(input.ActionType),
        ToolName:   &toolName,
        Enabled:    input.Enabled,
    }

    if input.Description != "" {
        action.Description = &input.Description
    }
    if len(input.Config) > 0 {
        cfg := json.RawMessage(input.Config)
        action.Config = &cfg
    }
    if len(input.Parameters) > 0 {
        params := json.RawMessage(input.Parameters)
        action.Parameters = &params
    }

    // Persist
    if err := s.actionRepo.Create(ctx, action); err != nil {
        return nil, fmt.Errorf("failed to create action: %w", err)
    }

    return action, nil
}

func (s *ActionServiceImpl) UpdateAction(ctx context.Context, actionID string, input UpdateActionInput) (*models.ChatbotAction, error) {
    // Fetch existing action
    action, err := s.actionRepo.GetByID(ctx, actionID)
    if err != nil {
        return nil, fmt.Errorf("failed to get action: %w", err)
    }
    if action == nil {
        return nil, ErrActionNotFound
    }

    // Determine if ToolName needs regeneration (business logic moved from handler)
    nameChanged := input.Name != nil && *input.Name != action.Name
    descChanged := input.Description != nil && (action.Description == nil || *input.Description != *action.Description)
    toolNameMissing := action.ToolName == nil || *action.ToolName == ""

    if nameChanged || descChanged || toolNameMissing {
        newName := action.Name
        if input.Name != nil {
            newName = *input.Name
        }
        
        newDesc := ""
        if input.Description != nil {
            newDesc = *input.Description
        } else if action.Description != nil {
            newDesc = *action.Description
        }

        toolName, err := s.toolNameGenerator.Generate(ctx, newName, newDesc)
        if err != nil {
            return nil, fmt.Errorf("failed to generate tool name: %w", err)
        }
        action.ToolName = &toolName
    }

    // Apply updates
    if input.Name != nil {
        action.Name = *input.Name
    }
    if input.Description != nil {
        action.Description = input.Description
    }
    if input.ActionType != nil {
        action.ActionType = models.ActionType(*input.ActionType)
    }
    if len(input.Config) > 0 {
        cfg := json.RawMessage(input.Config)
        action.Config = &cfg
    }
    if len(input.Parameters) > 0 {
        params := json.RawMessage(input.Parameters)
        action.Parameters = &params
    }
    if input.Enabled != nil {
        action.Enabled = *input.Enabled
    }

    // Persist with version check
    if err := s.actionRepo.Update(ctx, action); err != nil {
        if errors.Is(err, repository.ErrVersionConflict) {
            return nil, repository.ErrVersionConflict
        }
        return nil, fmt.Errorf("failed to update action: %w", err)
    }

    return action, nil
}

func (s *ActionServiceImpl) GetAction(ctx context.Context, actionID string) (*models.ChatbotAction, error) {
    return s.actionRepo.GetByID(ctx, actionID)
}

func (s *ActionServiceImpl) ListActions(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
    actions, err := s.actionRepo.List(ctx, chatbotID)
    if err != nil {
        return nil, err
    }
    if actions == nil {
        return []*models.ChatbotAction{}, nil
    }
    return actions, nil
}

func (s *ActionServiceImpl) DeleteAction(ctx context.Context, actionID string) error {
    return s.actionRepo.Delete(ctx, actionID)
}

func (s *ActionServiceImpl) GetActionLogs(ctx context.Context, chatbotID string, limit, offset int) ([]*models.ActionExecutionLog, error) {
    logs, err := s.actionRepo.GetLogs(ctx, chatbotID, limit, offset)
    if err != nil {
        return nil, err
    }
    if logs == nil {
        return []*models.ActionExecutionLog{}, nil
    }
    return logs, nil
}
```

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/services/action_service_impl.go` | **[NEW]** Implement ActionService |
| `internal/services/action_service_impl_test.go` | **[NEW]** Unit tests for implementation |

## Checklist

- [ ] Create `internal/services/action_service_impl.go`
- [ ] Define sentinel errors (`ErrActionNameRequired`, `ErrActionTypeRequired`, `ErrActionNotFound`)
- [ ] Implement `NewActionService` constructor
- [ ] Implement `CreateAction` with ToolName generation
- [ ] Implement `UpdateAction` with conditional ToolName regeneration
- [ ] Implement `GetAction`
- [ ] Implement `ListActions` (ensure empty slice, not nil)
- [ ] Implement `DeleteAction`
- [ ] Implement `GetActionLogs` (ensure empty slice, not nil)
- [ ] Create mock `ToolNameGenerator` for testing
- [ ] Write unit tests for all methods
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify all tests pass
- [ ] Run `make lint` to check for issues

## Test Cases

Write tests for these scenarios:

### CreateAction
- [x] Success case: creates action with generated ToolName
- [x] Error: empty name returns `ErrActionNameRequired`
- [x] Error: empty action type returns `ErrActionTypeRequired`
- [x] Error: ToolName generation failure propagates

### UpdateAction
- [x] Success: updates action without regenerating ToolName (no name/desc change)
- [x] Success: regenerates ToolName when name changes
- [x] Success: regenerates ToolName when description changes
- [x] Success: regenerates ToolName when ToolName is nil/empty (migration case)
- [x] Error: returns `ErrActionNotFound` for non-existent action
- [x] Error: returns `ErrVersionConflict` on concurrent modification

### Other Methods
- [x] GetAction: returns nil for non-existent action
- [x] ListActions: returns empty slice for chatbot without actions
- [x] GetActionLogs: returns empty slice when no logs exist

## Verification

```bash
# Verify compilation
go build ./...

# Run all tests
make test-all

# Run lint
make lint

# Run only service tests
go test ./internal/services/... -v -run Action
```

## Next Task

After this task, proceed to [003-refactor-action-handlers.md](./003-refactor-action-handlers.md) to update handlers to use the service.
