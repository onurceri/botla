# Policy Package Migration Guide

Quick reference for migrating code to use `pkg/policy` constants.

## Import the Package

```go
import "github.com/onurceri/botla-co/pkg/policy"
```

## Plan Constants

### Available Constants
```go
policy.PlanFree   // "free"
policy.PlanPro    // "pro"  
policy.PlanUltra  // "ultra"
```

### Common Migrations

#### SQL Queries
```go
// ❌ Before
db.QueryRow(`SELECT id FROM plans WHERE code='free'`).Scan(&id)

// ✅ After (use parameterized query!)
db.QueryRow(`SELECT id FROM plans WHERE code=$1`, policy.PlanFree.String()).Scan(&id)
```

#### Switch Statements
```go
// ❌ Before
switch code {
case "free":
    return 100_000
case "pro":
    return 1_000_000
}

// ✅ After
switch code {
case policy.PlanFree.String():
    return 100_000
case policy.PlanPro.String():
    return 1_000_000
}
```

#### String Comparisons
```go
// ❌ Before
if plan.Code == "free" { ... }

// ✅ After
if plan.Code == policy.PlanFree.String() { ... }
```

#### Test Fixtures
```go
// ❌ Before
testdb.UserFixture{PlanCode: "free"}

// ✅ After
testdb.UserFixture{PlanCode: policy.PlanFree.String()}
```

## Model Constants

### Available Constants
```go
policy.ModelGPT4o         // "gpt-4o"
policy.ModelGPT4oMini     // "gpt-4o-mini"
policy.ModelGPT5          // "gpt-5"
policy.ModelEmbeddingSmall // "text-embedding-3-small"
```

### Common Migrations

#### Default Values
```go
// ❌ Before
Model: "gpt-4o-mini"

// ✅ After
Model: policy.DefaultChatModel().String()
// or explicitly:
Model: policy.ModelGPT4oMini.String()
```

#### SQL Queries
```go
// ❌ Before
db.Exec(`UPDATE chatbots SET model='gpt-4o' WHERE id=$1`, id)

// ✅ After
db.Exec(`UPDATE chatbots SET model=$1 WHERE id=$2`, 
    policy.ModelGPT4o.String(), id)
```

## Helper Functions

```go
// Check if a plan is valid
plan := policy.Plan("free")
if plan.IsValid() {
    // do something
}

// Get all valid plans
for _, p := range policy.AllPlans() {
    fmt.Println(p.String())
}

// Get default models
chatModel := policy.DefaultChatModel()      // ModelGPT4oMini
embedModel := policy.DefaultEmbeddingModel() // ModelEmbeddingSmall
```

## Reference Constants

For validation/testing only (NOT source of truth):

```go
// Token limits (actual values in database)
policy.TokenLimitFree  // 100_000
policy.TokenLimitPro   // 1_000_000
policy.TokenLimitUltra // 10_000_000

// Chatbot limits (actual values in database)
policy.MaxChatbotsFree  // 1
policy.MaxChatbotsPro   // 5  
policy.MaxChatbotsUltra // 100
```

⚠️ **Warning**: These are reference values only! Always fetch actual limits from the database `plans.config` field.

## Finding Code to Migrate

```bash
# Find hardcoded plan codes
grep -rn "='free'" internal/
grep -rn '="free"' internal/
grep -rn 'code = "free"' internal/

# Find hardcoded model names
grep -rn '"gpt-4o"' internal/
grep -rn '"gpt-4o-mini"' internal/

# Find switch on plan codes
grep -rn 'case "free"' internal/
grep -rn 'case "pro"' internal/
```

## Common Pitfalls

### ❌ Don't Hardcode Limits
```go
// ❌ BAD
if userChatbots >= 1 { // hardcoded free limit
    return errors.New("limit exceeded")
}

// ✅ GOOD
planLimits, _ := db.GetPlanByUserID(ctx, userID)
if userChatbots >= planLimits.Config.MaxChatbots {
    return errors.New("limit exceeded")
}
```

### ❌ Don't Skip Parameterization
```go
// ❌ BAD (SQL injection still possible!)
query := fmt.Sprintf("WHERE code='%s'", policy.PlanFree.String())

// ✅ GOOD
query := "WHERE code=$1"
args := []interface{}{policy.PlanFree.String()}
```

### ✅ Do Convert to String
```go
// ❌ Won't compile
var code string = policy.PlanFree

// ✅ Correct
var code string = policy.PlanFree.String()
```

## Testing Your Changes

After migration, always run:

```bash
# Run tests for modified packages
go test ./internal/testdb/... -v
go test ./pkg/policy/... -v

# Build everything to catch compile errors
go build ./...

# Run specific integration tests if modified
go test ./internal/integration/... -run TestYourTest -v
```

## Questions?

See the complete documentation in `pkg/policy/README.md`
