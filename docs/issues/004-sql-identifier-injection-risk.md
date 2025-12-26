# Issue 004: Risk of SQL Identifier Injection in Admin Updates

## Priority: Medium
## Confidence: Medium

## Summary

The `AdminUpdateUser` function constructs SQL UPDATE statements by interpolating column names via `fmt.Sprintf`. While protected by a whitelist, this architectural pattern is error-prone and could lead to SQL injection if the whitelist is bypassed or expanded incorrectly.

## Evidence

**File:** [admin_users.go](file:///Users/onur/Documents/workspace/botla-co/internal/db/admin_users.go#L96-L110)

```go
setParts := []string{}
for k, v := range updates {
    // Basic validation of keys to prevent SQL injection (though we should use a proper builder)
    allowedKeys := map[string]bool{
        "full_name":         true,
        "plan_id":           true,
        "is_platform_admin": true,
    }
    if !allowedKeys[k] {
        continue
    }

    setParts = append(setParts, fmt.Sprintf("%s = $%d", k, argIdx))
    args = append(args, v)
    argIdx++
}
```

## Risk Analysis

### Current State
- **Mitigated by**: Static whitelist of allowed column names
- **Current Risk Level**: Low (whitelist is effective)

### Future Risk Scenarios

1. **Whitelist Expansion Error**
   - Developer adds a new key that's derived from user input
   - Example: `allowedKeys[untrustedKey] = true`
   - Result: Direct SQL injection path

2. **Whitelist Bypass**
   - Keys are checked but not sanitized for SQL special characters
   - If a key like `plan_id; DROP TABLE users--` passed whitelist, it would be injected

3. **Copy-Paste Propagation**
   - This pattern might be copied to other functions without the whitelist
   - Leads to inconsistent security posture

4. **Configuration-Driven Keys**
   - Whitelist loaded from config file or environment variable
   - Attackers who compromise config can inject SQL

## Recommended Fix

### Option A: Use a Query Builder Library (Recommended)

Use `squirrel` for safe query construction:

```go
import sq "github.com/Masterminds/squirrel"

func AdminUpdateUser(ctx context.Context, pool *sql.DB, userID string, updates map[string]any) error {
    if len(updates) == 0 {
        return nil
    }

    // Define allowed columns as constants
    allowedColumns := map[string]bool{
        "full_name":         true,
        "plan_id":           true,
        "is_platform_admin": true,
    }

    builder := sq.Update("users").
        Where(sq.Eq{"id": userID}).
        Where("deleted_at IS NULL").
        PlaceholderFormat(sq.Dollar)

    hasUpdates := false
    for k, v := range updates {
        if !allowedColumns[k] {
            continue
        }
        builder = builder.Set(k, v)
        hasUpdates = true
    }

    if !hasUpdates {
        return nil
    }

    query, args, err := builder.ToSql()
    if err != nil {
        return fmt.Errorf("build query: %w", err)
    }

    _, err = pool.ExecContext(ctx, query, args...)
    if err != nil {
        return fmt.Errorf("update user: %w", err)
    }

    return nil
}
```

### Option B: Switch Statement with Predefined Columns

Eliminate string interpolation entirely:

```go
func AdminUpdateUser(ctx context.Context, pool *sql.DB, userID string, updates map[string]any) error {
    type columnUpdate struct {
        column string
        value  any
    }
    
    var cols []columnUpdate
    
    for k, v := range updates {
        switch k {
        case "full_name":
            cols = append(cols, columnUpdate{"full_name", v})
        case "plan_id":
            cols = append(cols, columnUpdate{"plan_id", v})
        case "is_platform_admin":
            cols = append(cols, columnUpdate{"is_platform_admin", v})
        // Unknown keys are silently ignored
        }
    }
    
    if len(cols) == 0 {
        return nil
    }
    
    // Build query with constant strings only
    // ...
}
```

### Option C: Type-Safe Update Struct

Define a struct that enforces column names at compile-time:

```go
type UserUpdate struct {
    FullName        *string `db:"full_name"`
    PlanID          *string `db:"plan_id"`
    IsPlatformAdmin *bool   `db:"is_platform_admin"`
}

func AdminUpdateUser(ctx context.Context, pool *sql.DB, userID string, update UserUpdate) error {
    // Use reflection on struct tags (which are compile-time constants)
    // Or manually check each field
}
```

## Verification

1. Code review: Ensure no `fmt.Sprintf` with user-controlled column names
2. Static analysis: Use `gosec` to detect SQL construction patterns
3. Unit test: Verify that unexpected keys are ignored
4. Fuzz test: Inject malicious column names and verify they're rejected

## Related Files

- `internal/db/admin_users.go` - Contains vulnerable pattern
- Other admin DB files that may follow similar patterns
