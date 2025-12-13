# Test Conventions

This document outlines the testing conventions and best practices for the Botla backend codebase.

## Database Setup

### Database Naming

- **Tests MUST use `botla_test` database** - Never use `botla_dev` in tests
- The test database should be separate from development to prevent data corruption
- Integration tests use the `test` schema within `botla_test` database

### Centralized Test Database Utilities

All tests should use the centralized test database utilities from `internal/testdb`:

```go
import "github.com/onurceri/botla-co/internal/testdb"

func TestMyFeature(t *testing.T) {
    db := testdb.OpenTestDB(t)
    defer db.Close()
    // ... test code
}
```

**Benefits:**

- Single source of truth for database connection
- Consistent configuration across all tests
- Environment variable support for CI/CD
- Automatic schema setup

### Available Functions

- `testdb.OpenTestDB(t)` - Opens connection to test database with default schema
- `testdb.OpenTestDBWithSchema(t, schema)` - Opens connection with custom schema
- `testdb.WithTx(t, db, fn)` - Runs test code within a transaction (auto-rollback)

## Test Structure

### Unit Tests

- Located alongside source files: `*_test.go`
- Use `testdb.OpenTestDB(t)` for database access
- Keep tests focused on single functions/units

### Integration Tests

- Located in `internal/integration/`
- Use `integration.SetupTestEnv()` for full environment setup
- Include HTTP server, database, and mock services

## Best Practices

### Error Handling

```go
// ✅ Good: Declare err in if statement
if err := db.QueryRow(...).Scan(&id); err != nil {
    t.Fatalf("query failed: %v", err)
}

// ❌ Bad: Using undefined err variable
err = db.QueryRow(...).Scan(&id)
if err != nil {
    t.Fatalf("query failed: %v", err)
}
```

### Test Helpers

- Mark helper functions with `t.Helper()` for better error reporting
- Keep helpers simple and focused

### Database Cleanup

- Use `defer db.Close()` for connection cleanup
- Integration tests use `TeardownTestEnv()` for full cleanup
- Unit tests should clean up test data in `t.Cleanup()` when needed

### Skipping Tests

- Use `t.Skipf()` when database is unavailable (handled automatically by `testdb.OpenTestDB`)
- Don't fail tests due to missing external dependencies

## Environment Variables

Test utilities respect these environment variables (with sensible defaults):

- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `5432`)
- `DB_NAME` (default: `botla_test`)
- `DB_USER` (default: `botla`)
- `DB_PASSWORD` (default: `botla`)
- `DB_SCHEMA` (default: `test`)

## Running Tests

```bash
# Run all tests (without PDF support)
make test-no-pdf

# Run tests with coverage
make test-all

# Run specific test
go test -v ./internal/api/handlers -run TestFunctionName
```

## Common Patterns

### Creating Test Users

```go
func createTestUser(t *testing.T, db *sql.DB) string {
    t.Helper()
    var userID string
    email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
    if err := db.QueryRow(
        `INSERT INTO users (email, password_hash, plan_id) 
         VALUES ($1, $2, $3) RETURNING id`,
        email, "hash", planID,
    ).Scan(&userID); err != nil {
        t.Fatalf("create user: %v", err)
    }
    return userID
}
```

### Using Transactions

```go
testdb.WithTx(t, db, func(ctx context.Context, tx *sql.Tx) {
    // All changes in this block will be rolled back
    _, err := tx.Exec("INSERT INTO ...")
    // ...
})
```

## Migration

When updating existing tests:

1. Replace hardcoded connection strings with `testdb.OpenTestDB(t)`
2. Remove duplicate database setup code
3. Ensure all tests use `botla_test` database
4. Update imports to include `internal/testdb`

