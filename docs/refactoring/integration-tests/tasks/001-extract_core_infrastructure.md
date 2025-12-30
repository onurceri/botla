## TASK-001 — Extract Core Test Infrastructure

Goal:
Move the underlying test infrastructure (`TestEnv`, `SetupTestEnv`, `NewTestMux`, Mocks) to a new `internal/integration/fixtures` package and update all tests to use it directly.

Scope:
- Create `internal/integration/fixtures/` package.
- Move core logic from `internal/integration/testutils.go` and `internal/integration/testserver.go` to the new package.
- Update all existing integration tests (100+ files) to import `fixtures` and use the new package directly.
- Ensure all tests pass.

Checklist:
[x] Create directory `internal/integration/fixtures`
[x] Move `TestEnv` struct and `SetupTestEnv` logic to `fixtures` package
[x] Move `MockLLM`, `MockVectorStore` to `fixtures` package
[x] Move `NewTestMux` (and server setup) to `fixtures` package
[x] Use `sed` or similar to bulk update all `*_test.go` files to use `fixtures.SetupTestEnv` and `fixtures.TestEnv`
[x] Verify `make test-all` passes

Edge Cases:
- Circular dependencies if `fixtures` imports `integration`. ensure `fixtures` is standalone.
- Environment variables consistency between packages.

Files Likely to Change:
- `internal/integration/fixtures/env.go`
- `internal/integration/fixtures/server.go`
- `internal/integration/fixtures/mocks.go`
- `internal/integration/testutils.go`
- `internal/integration/testserver.go`
