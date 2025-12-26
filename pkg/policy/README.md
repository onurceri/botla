# Policy Package

This package provides typed constants for plan codes and model identifiers to eliminate "magic strings" and "primitive obsession" throughout the codebase.

## Purpose

Before this package, plan codes and model names were scattered as raw strings:

```go
// ❌ BAD: Magic strings everywhere
if plan.Code == "free" { ... }
model := "gpt-4o-mini"
if tokenCount > 100000 { ... }
```

Now, we use typed constants:

```go
// ✅ GOOD: Type-safe constants
if plan.Code == policy.PlanFree.String() { ... }
model := policy.DefaultChatModel()
if tokenCount > policy.TokenLimitFree { ... }
```

## Design Principles

1. **Database is the Source of Truth**: This package does NOT duplicate plan configurations from the database. Plan limits, allowed models, and other configurations are stored in the `plans` table.

2. **Type Safety**: Use typed constants (`policy.Plan`, `policy.Model`) instead of raw strings.

3. **Reference Values Only**: The limit constants (`TokenLimitFree`, etc.) are for reference in tests and validation, NOT as the source of truth.

## Usage Examples

### Working with Plans

```go
import "github.com/onurceri/botla-co/pkg/policy"

// Define plan type
var userPlan policy.Plan = policy.PlanFree

// Validate plan
if !userPlan.IsValid() {
    return errors.New("invalid plan")
}

// Convert to string for database queries
planCode := userPlan.String() // "free"

// Iterate all plans
for _, plan := range policy.AllPlans() {
    fmt.Println(plan)
}
```

### Working with Models

```go
import "github.com/onurceri/botla-co/pkg/policy"

// Use default model
defaultModel := policy.DefaultChatModel() // ModelGPT4oMini

// Validate model
userModel := policy.Model("gpt-4o")
if !userModel.IsValid() {
    return errors.New("unknown model")
}

// Convert to string for API calls
modelName := userModel.String() // "gpt-4o"
```

### Reference Limits (Testing/Validation)

```go
import "github.com/onurceri/botla-co/pkg/policy"

// Use reference constants for validation
// NOTE: Always fetch actual limits from the database plan.Config
if estimatedTokens > policy.TokenLimitFree {
    // This is a quick check, but always verify against database limits
}
```

## Migration Guide

When replacing magic strings with policy constants:

1. **For Plan Codes**:
   ```go
   // Before
   if code == "free" { ... }
   
   // After
   if code == policy.PlanFree.String() { ... }
   ```

2. **For Model Names**:
   ```go
   // Before
   model := "gpt-4o-mini"
   
   // After
   model := policy.ModelGPT4oMini.String()
   // Or for defaults:
   model := policy.DefaultChatModel().String()
   ```

3. **For Limits** - ALWAYS fetch from database:
   ```go
   // ❌ DON'T hardcode limits
   if userChatbots > 1 { ... }
   
   // ✅ DO fetch from database
   planLimits, err := db.GetPlanByUserID(ctx, userID)
   if userChatbots >= planLimits.Config.MaxChatbots { ... }
   ```

## What's NOT in This Package

- **Plan Configurations**: Stored in the `plans` table
- **Model Metadata**: Stored in the `ai_models` table
- **Business Logic**: This is just constants, not business rules

## Files

- `plans.go` - Plan type and constants (Free, Pro, Ultra)
- `models.go` - Model type and constants (GPT-4o, GPT-4o-mini, etc.)
- `limits.go` - Reference limit constants (for testing/validation only)
- `*_test.go` - Comprehensive unit tests

## Testing

```bash
go test ./pkg/policy/...
```

All constants have full test coverage ensuring validity and consistency.
