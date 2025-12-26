# Task 04: Configuration Centralization - Eliminate Primitive Obsession

## Priority
**Medium** - Improves code readability and reduces configuration errors

## Problem Statement

System-wide configurations (Plan limits, Model names, System Prompts) are handled as raw strings or integers. This leads to "magic values" scattered across the codebase, making changes error-prone.

## Evidence

```go
// Scattered magic values
model := "gpt-4o-mini"
if plan.Name == "free" { ... }
if tokenCount > 100000 { ... }
```

## Implementation Plan

### Phase 1: Create Policy Package (`pkg/policy/`)

**File**: `pkg/policy/plans.go`
```go
type Plan string
const (
    PlanFree  Plan = "free"
    PlanPro   Plan = "pro"
    PlanTeam  Plan = "team"
)
```

**File**: `pkg/policy/models.go`
```go
type Model string
const (
    ModelGPT4o         Model = "gpt-4o"
    ModelGPT4oMini     Model = "gpt-4o-mini"
)

var modelRegistry = map[Model]ModelInfo{...}
func DefaultChatModel() Model { return ModelGPT4oMini }
```

**File**: `pkg/policy/limits.go`
```go
type PlanLimits struct {
    MaxMonthlyTokens   int64
    MaxChatbots        int
    AllowedModels      []Model
}

var DefaultLimits = map[Plan]PlanLimits{
    PlanFree: {MaxMonthlyTokens: 100_000, MaxChatbots: 1, ...},
    PlanPro:  {MaxMonthlyTokens: 1_000_000, MaxChatbots: 5, ...},
}
```

### Phase 2: Update Consumers

Replace raw strings with typed constants throughout the codebase.

## Affected Files

| File | Action | Description |
|------|--------|-------------|
| `pkg/policy/` | NEW | Configuration package |
| `internal/db/queries.go` | MODIFY | Use `policy.Plan` type |
| `internal/rag/*.go` | MODIFY | Use `policy.Model` type |

## Acceptance Criteria

- [ ] All plan types as typed constants
- [ ] Model identifiers in central registry
- [ ] No raw "free"/"pro" strings in business logic
- [ ] All existing tests pass

## Estimated Effort
**Size**: Medium (2-3 days)
