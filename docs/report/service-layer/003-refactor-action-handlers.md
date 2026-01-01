# Task 003: Refactor Action Handlers to Use ActionService

> **Agent Prompt**: You are refactoring `internal/api/handlers/action.go` to use the newly created `ActionService`. The handler should become a thin HTTP layer that only handles request parsing, authentication, calling the service, and formatting responses. All business logic (ToolName generation, validation, version conflict handling) has already been moved to `ActionService` in Task 002. Apply strict TDD: update handler tests first to expect the new behavior, then modify the handlers. Ensure all existing API behavior remains unchanged.

## Background

After completing Task 002, we have an `ActionService` that contains all business logic. Now we need to:

1. Update `ActionHandlers` to depend on `ActionService` instead of `ActionRepository` + `ToolNameGenerator`
2. Simplify handler methods to only handle HTTP concerns
3. Keep all API behavior identical (same status codes, response formats)

## Goal

Transform handlers from "fat controllers" to thin HTTP adapters:

**Before (current):**
```go
type ActionHandlers struct {
    ActionRepo        repository.ActionRepository
    ChatbotRepo       repository.ChatbotRepository
    ToolNameGenerator *rag.ToolNameGenerator
    // ...
}

func (h *ActionHandlers) Update(w http.ResponseWriter, r *http.Request) {
    // ... lots of business logic ...
    if nameChanged || descChanged || toolNameMissing {
        toolName, err = h.ToolNameGenerator.Generate(...)
    }
    // ... more business logic ...
}
```

**After (target):**
```go
type ActionHandlers struct {
    ActionService    services.ActionService
    ChatbotRepo      repository.ChatbotRepository
    // ...
}

func (h *ActionHandlers) Update(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    var req updateActionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 2. Call service
    action, err := h.ActionService.UpdateAction(r.Context(), actionID, req.toInput())
    
    // 3. Handle errors -> HTTP status codes
    if errors.Is(err, services.ErrActionNotFound) {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    if errors.Is(err, repository.ErrVersionConflict) {
        http.Error(w, "Action was modified...", http.StatusConflict)
        return
    }
    
    // 4. Format response
    api.WriteJSON(w, http.StatusOK, action)
}
```

## Files to Modify

| File | Action |
|------|--------|
| `internal/api/handlers/action.go` | Refactor to use `ActionService` |
| `internal/api/handlers/action_unit_test.go` | Update tests for new structure |
| `internal/api/router.go` | Update handler initialization |
| `internal/integration/testserver.go` | Update test server initialization |

## Implementation Steps

### Step 1: Update ActionHandlers Struct

```go
type ActionHandlers struct {
    ActionService    services.ActionService
    ChatbotRepo      repository.ChatbotRepository
    WorkspaceService *services.WorkspaceService
    OrgService       *services.OrganizationService
}
```

### Step 2: Refactor Each Handler Method

#### Create Handler
```go
func (h *ActionHandlers) Create(w http.ResponseWriter, r *http.Request) {
    botID, _, ok := h.authorize(w, r)
    if !ok {
        return
    }

    var req createActionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    input := services.CreateActionInput{
        Name:        req.Name,
        Description: req.Description,
        ActionType:  req.ActionType,
        Config:      req.Config,
        Parameters:  req.Parameters,
        Enabled:     req.Enabled,
    }

    action, err := h.ActionService.CreateAction(r.Context(), botID, input)
    if err != nil {
        if errors.Is(err, services.ErrActionNameRequired) || 
           errors.Is(err, services.ErrActionTypeRequired) {
            api.WriteErrorCode(w, http.StatusBadRequest, api.ErrNameAndActionTypeRequired)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    api.WriteJSON(w, http.StatusCreated, action)
}
```

#### Update Handler
```go
func (h *ActionHandlers) Update(w http.ResponseWriter, r *http.Request) {
    botID, _, ok := h.authorize(w, r)
    if !ok {
        return
    }

    actionID := r.PathValue("actionId")
    
    // Verify action belongs to this chatbot
    action, err := h.ActionService.GetAction(r.Context(), actionID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if action == nil || action.ChatbotID != botID {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    var req createActionRequest
    if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    input := services.UpdateActionInput{}
    if req.Name != "" {
        input.Name = &req.Name
    }
    if req.Description != "" {
        input.Description = &req.Description
    }
    if req.ActionType != "" {
        input.ActionType = &req.ActionType
    }
    input.Config = req.Config
    input.Parameters = req.Parameters
    input.Enabled = &req.Enabled

    updatedAction, err := h.ActionService.UpdateAction(r.Context(), actionID, input)
    if err != nil {
        if errors.Is(err, services.ErrActionNotFound) {
            w.WriteHeader(http.StatusNotFound)
            return
        }
        if errors.Is(err, repository.ErrVersionConflict) {
            http.Error(w, "Action was modified by another request, please refresh and try again", http.StatusConflict)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    api.WriteJSON(w, http.StatusOK, updatedAction)
}
```

### Step 3: Update Router

```go
// In router.go
actionService := services.NewActionService(actionRepo, toolNameGenerator)

actionHandlers := &handlers.ActionHandlers{
    ActionService:    actionService,
    ChatbotRepo:      chatbotRepo,
    WorkspaceService: workspaceService,
    OrgService:       orgService,
}
```

## Checklist

- [x] Update `ActionHandlers` struct to use `ActionService`
- [x] Refactor `List` handler
- [x] Refactor `Create` handler
- [x] Refactor `Get` handler
- [x] Refactor `Update` handler
- [x] Refactor `Delete` handler
- [x] Refactor `GetLogs` handler
- [x] Update router initialization in `router.go`
- [x] Update test server initialization in `testserver.go`
- [x] Update unit tests in `action_unit_test.go`
- [x] Run `go build ./...` to verify compilation
- [x] Run `make test-all` to verify all tests pass
- [x] Run integration tests to ensure API behavior unchanged
- [x] Run `make lint` to check for issues

## API Behavior Verification

Ensure these behaviors remain unchanged:

| Endpoint | Scenario | Expected Response |
|----------|----------|-------------------|
| POST /actions | Valid request | 201 Created + action JSON |
| POST /actions | Missing name | 400 Bad Request |
| PUT /actions/:id | Valid update | 200 OK + updated action |
| PUT /actions/:id | Version conflict | 409 Conflict |
| PUT /actions/:id | Action not found | 404 Not Found |
| GET /actions/:id | Exists | 200 OK + action |
| GET /actions/:id | Not found | 404 Not Found |
| DELETE /actions/:id | Exists | 204 No Content |

## Verification

```bash
# Verify compilation
go build ./...

# Run all tests
make test-all

# Run handler unit tests specifically
go test ./internal/api/handlers/... -v -run Action

# Run integration tests
go test ./internal/integration/... -v -run Action

# Lint
make lint
```

## Next Task

After this task, proceed to [004-add-action-service-tests.md](./004-add-action-service-tests.md) for comprehensive service testing.
