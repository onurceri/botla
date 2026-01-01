# Task 004: Add Comprehensive ActionService Tests

> **Agent Prompt**: You are adding comprehensive unit and integration tests for the `ActionService` implementation. Focus on testing the business logic that was previously untestable because it was embedded in HTTP handlers. Verify ToolName regeneration logic, validation, error handling, and edge cases. Apply strict TDD principles and aim for high test coverage. Ensure tests are independent and can run in parallel.

## Background

The `ActionService` now contains all business logic that was previously scattered in handlers. This task focuses on:

1. Ensuring comprehensive unit test coverage
2. Testing edge cases that were hard to test in handlers
3. Verifying the ToolName regeneration logic works correctly
4. Creating integration tests that verify end-to-end behavior

## Goal

Achieve comprehensive test coverage for `ActionService` with:
- Unit tests using mock repositories and ToolNameGenerator
- Integration tests using real database
- Edge case coverage
- Parallel-safe tests

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/services/action_service_impl_test.go` | Ensure comprehensive unit tests |
| `internal/integration/action_service_test.go` | **[NEW]** Integration tests |
| `internal/services/mocks/action_mocks.go` | **[NEW]** Mock implementations |

## Unit Test Cases

### CreateAction Tests

```go
func TestActionService_CreateAction(t *testing.T) {
    tests := []struct {
        name           string
        chatbotID      string
        input          CreateActionInput
        mockToolName   string
        mockToolErr    error
        mockCreateErr  error
        wantErr        error
    }{
        {
            name:      "success - creates action with generated tool name",
            chatbotID: "bot-123",
            input: CreateActionInput{
                Name:        "Send Email",
                Description: "Sends an email to user",
                ActionType:  "webhook",
                Enabled:     true,
            },
            mockToolName: "send_email",
            wantErr:      nil,
        },
        {
            name:      "error - empty name",
            chatbotID: "bot-123",
            input: CreateActionInput{
                Name:       "",
                ActionType: "webhook",
            },
            wantErr: ErrActionNameRequired,
        },
        {
            name:      "error - empty action type",
            chatbotID: "bot-123",
            input: CreateActionInput{
                Name:       "Send Email",
                ActionType: "",
            },
            wantErr: ErrActionTypeRequired,
        },
        {
            name:      "error - tool name generation fails",
            chatbotID: "bot-123",
            input: CreateActionInput{
                Name:       "Send Email",
                ActionType: "webhook",
            },
            mockToolErr: errors.New("LLM unavailable"),
            wantErr:     errors.New("failed to generate tool name"),
        },
    }
    // ... test execution ...
}
```

### UpdateAction Tests - ToolName Regeneration Logic

```go
func TestActionService_UpdateAction_ToolNameRegeneration(t *testing.T) {
    tests := []struct {
        name                 string
        existingAction       *models.ChatbotAction
        input                UpdateActionInput
        expectToolNameRegen  bool
        mockNewToolName      string
    }{
        {
            name: "no regeneration when only enabled changes",
            existingAction: &models.ChatbotAction{
                ID:       "action-1",
                Name:     "Send Email",
                ToolName: ptr("send_email"),
            },
            input: UpdateActionInput{
                Enabled: ptr(false),
            },
            expectToolNameRegen: false,
        },
        {
            name: "regenerates when name changes",
            existingAction: &models.ChatbotAction{
                ID:       "action-1",
                Name:     "Send Email",
                ToolName: ptr("send_email"),
            },
            input: UpdateActionInput{
                Name: ptr("Send Notification"),
            },
            expectToolNameRegen: true,
            mockNewToolName:     "send_notification",
        },
        {
            name: "regenerates when description changes",
            existingAction: &models.ChatbotAction{
                ID:          "action-1",
                Name:        "Send Email",
                Description: ptr("Old description"),
                ToolName:    ptr("send_email"),
            },
            input: UpdateActionInput{
                Description: ptr("New description"),
            },
            expectToolNameRegen: true,
            mockNewToolName:     "send_email_v2",
        },
        {
            name: "regenerates when tool name is nil (migration case)",
            existingAction: &models.ChatbotAction{
                ID:       "action-1",
                Name:     "Send Email",
                ToolName: nil,
            },
            input: UpdateActionInput{
                Enabled: ptr(true),
            },
            expectToolNameRegen: true,
            mockNewToolName:     "send_email",
        },
        {
            name: "regenerates when tool name is empty string (migration case)",
            existingAction: &models.ChatbotAction{
                ID:       "action-1",
                Name:     "Send Email",
                ToolName: ptr(""),
            },
            input: UpdateActionInput{
                Enabled: ptr(true),
            },
            expectToolNameRegen: true,
            mockNewToolName:     "send_email",
        },
    }
    // ... test execution with mock verification ...
}
```

### UpdateAction Tests - Error Handling

```go
func TestActionService_UpdateAction_Errors(t *testing.T) {
    tests := []struct {
        name          string
        actionID      string
        mockGetResult *models.ChatbotAction
        mockGetErr    error
        mockUpdateErr error
        wantErr       error
    }{
        {
            name:          "error - action not found (nil)",
            actionID:      "nonexistent",
            mockGetResult: nil,
            wantErr:       ErrActionNotFound,
        },
        {
            name:     "error - version conflict",
            actionID: "action-1",
            mockGetResult: &models.ChatbotAction{
                ID:       "action-1",
                Name:     "Test",
                ToolName: ptr("test"),
            },
            mockUpdateErr: repository.ErrVersionConflict,
            wantErr:       repository.ErrVersionConflict,
        },
    }
    // ... test execution ...
}
```

## Integration Test Cases

Create `internal/integration/action_service_test.go`:

```go
func TestActionService_Integration(t *testing.T) {
    env := fixtures.SetupTestEnv(t)
    defer env.Cleanup()

    // Create test user and chatbot
    user := env.CreateUser("test@example.com")
    chatbot := env.CreateChatbot(user, "TestBot")

    // Initialize real service with real dependencies
    actionService := services.NewActionService(
        env.ActionRepo,
        env.ToolNameGenerator, // Real or mock LLM
    )

    t.Run("CreateAction persists to database", func(t *testing.T) {
        action, err := actionService.CreateAction(env.Ctx, chatbot.ID, services.CreateActionInput{
            Name:       "Create Ticket",
            ActionType: "webhook",
            Enabled:    true,
        })
        
        require.NoError(t, err)
        assert.NotEmpty(t, action.ID)
        assert.NotEmpty(t, action.ToolName)
        
        // Verify persisted
        fetched, err := actionService.GetAction(env.Ctx, action.ID)
        require.NoError(t, err)
        assert.Equal(t, action.Name, fetched.Name)
    })

    t.Run("UpdateAction with version conflict", func(t *testing.T) {
        // Create action
        action, _ := actionService.CreateAction(env.Ctx, chatbot.ID, services.CreateActionInput{
            Name:       "Concurrent Test",
            ActionType: "webhook",
        })

        // Simulate concurrent update by modifying version in DB
        // ... simulate conflict ...

        _, err := actionService.UpdateAction(env.Ctx, action.ID, services.UpdateActionInput{
            Name: ptr("Updated Name"),
        })
        
        assert.ErrorIs(t, err, repository.ErrVersionConflict)
    })
}
```

## Mock Implementations

Create `internal/services/mocks/action_mocks.go`:

```go
package mocks

import (
    "context"
    "github.com/onurceri/botla-co/internal/models"
)

// MockToolNameGenerator for testing
type MockToolNameGenerator struct {
    GenerateFunc func(ctx context.Context, name, desc string) (string, error)
    CallCount    int
    LastName     string
    LastDesc     string
}

func (m *MockToolNameGenerator) Generate(ctx context.Context, name, desc string) (string, error) {
    m.CallCount++
    m.LastName = name
    m.LastDesc = desc
    if m.GenerateFunc != nil {
        return m.GenerateFunc(ctx, name, desc)
    }
    return "mock_tool_name", nil
}
```

## Checklist

- [ ] Create mock implementations in `internal/services/mocks/`
- [ ] Add unit tests for `CreateAction` (success + all error cases)
- [ ] Add unit tests for `UpdateAction` - ToolName regeneration logic
- [ ] Add unit tests for `UpdateAction` - error handling
- [ ] Add unit tests for `GetAction`
- [ ] Add unit tests for `ListActions`
- [ ] Add unit tests for `DeleteAction`
- [ ] Add unit tests for `GetActionLogs`
- [ ] Create integration test file `internal/integration/action_service_test.go`
- [ ] Add integration tests for persistence
- [ ] Add integration tests for version conflict handling
- [ ] Verify ToolNameGenerator mock tracks call count (to assert regeneration)
- [ ] Run `go test ./internal/services/... -cover` and verify coverage
- [ ] Run `make test-all` to verify all tests pass
- [ ] Run `make lint` to check for issues

## Coverage Targets

| Package | Target Coverage |
|---------|-----------------|
| `internal/services` (ActionService) | ≥ 90% |

## Verification

```bash
# Run service tests with coverage
go test ./internal/services/... -v -cover -run Action

# Run integration tests
go test ./internal/integration/... -v -run ActionService

# Check for race conditions
go test ./internal/services/... -race -run Action

# Full test suite
make test-all

# Lint
make lint
```

## Benefits Achieved

After completing this task, you will have:

1. **Testable Business Logic**: ToolName regeneration logic is now unit-testable
2. **No HTTP Mocking**: Service tests don't require mocking HTTP requests/responses
3. **Reusable Logic**: Service can be called from CLI tools, background jobs, etc.
4. **Clear Separation**: Handlers only handle HTTP, services handle business logic
5. **High Confidence**: Comprehensive tests give confidence for future refactoring
