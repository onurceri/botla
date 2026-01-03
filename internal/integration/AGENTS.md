# AGENTS.md - internal/integration

## OVERVIEW
Integration tests with real dependencies (PostgreSQL, Qdrant, Redis via Docker fixtures).

## WHERE TO LOOK
- **Setup**: `fixtures/env.go` - `SetupTestEnvWithConfig()`, schema isolation per test
- **Main entry**: `main_test.go` - schema cleanup on startup
- **Fixtures**: `fixtures/server.go` - HTTP test server with all services
- **Test utilities**: `../testdb/` - `OpenParallelTestDB()`, high-level fixtures
- **BOT tests**: `lifecycle_test.go:54-107` - BOT-002, BOT-005 expectations

## CONVENTIONS
- Each test gets unique schema: `botla_it_<hex>` for isolation
- Use `fixtures.SetupTestEnv()` then `fixtures.TeardownTestEnv(te)`
- TRUNCATE tables (except plans/languages) between tests for clean state
- Favor `fixtures.SetupTestEnvWithConfig()` over `t.Setenv()` for config overrides
- High-level fixtures in `testdb/fixtures.go`: `CreateUser`, `CreateChatbot`, etc.

## HTTP RESPONSE BODY HANDLING
Always use `drainBody(res)` instead of `res.Body.Close()` to prevent goroutine leaks:

```go
// GOOD - drains body before closing for connection reuse
defer drainBody(res)

// BAD - causes goroutine leaks in parallel tests
defer res.Body.Close()
```

The `drainBody()` utility in `http.go` ensures response bodies are fully consumed before closing, allowing HTTP connection reuse and preventing goroutine leaks in parallel integration tests.

## HTTP CLIENT USAGE
Always use test HTTP client utilities instead of standard `http` package functions to prevent connection leaks:

```go
// GOOD - test HTTP client with DisableKeepAlives
client := testHTTPClient()
resp, err := client.Do(req)
resp, err := testHTTPGet(url)
resp, err := testHTTPPost(url, contentType, body)

// BAD - uses http.DefaultClient internally, causing connection leaks
resp, err := http.DefaultClient.Do(req)
resp, err := http.Get(url)
resp, err := http.Post(url, contentType, body)
```

All test HTTP utilities (`testHTTPClient()`, `testHTTPGet()`, `testHTTPPost()`) return clients configured with `DisableKeepAlives: true` to prevent persistent connections that leak goroutines in parallel integration tests.

## ANTI-PATTERNS
- **DB connection killing**: `dropIntegrationSchema()` attempts DROP without terminating connections first. Relies on `TeardownTestEnv()` closing connections first. Breaks parallel tests.
- **t.Setenv() usage**: 42+ test files blocked from parallelization. Refactor to `SetupTestEnvWithConfig()` pattern.
- **Schema cleanup race**: No explicit connection termination in teardown. Retry logic (3x, 100ms) masks underlying race conditions.
- **BOT-002/BOT-005 failures**: Currently log warnings only, not hard failures. `t.Errorf()` required for CI enforcement.
