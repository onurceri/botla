# Refactoring Plan: Integration Test Fixtures

## Problem
Integration tests suffer from massive code duplication in setup logic (`SetupTestEnv`, user creation, plan assignment, chatbot creation). This makes refactoring difficult and reduces readability.

## Goal
Introduce a `fixtures` package to encapsulate common test selection, environment setup, and entity creation logic.

## Task Index
TASK-001: Extract Core Test Infrastructure
TASK-002: Implement Test Helpers (TDD)
TASK-003: Refactor Evidence Tests

---

## TASK-001 — Extract Core Test Infrastructure

Goal:
Move the underlying test infrastructure (`TestEnv`, `SetupTestEnv`, `NewTestMux`, Mocks) to a new `internal/integration/fixtures` package and update all tests to use it directly.

Scope:
- Create `internal/integration/fixtures/` package.
- Move core logic from `internal/integration/testutils.go` and `internal/integration/testserver.go` to the new package.
- Update all existing integration tests (100+ files) to import `fixtures` and use the new package directly.
- Ensure all tests pass.

Checklist:
[ ] Create directory `internal/integration/fixtures`
[ ] Move `TestEnv` struct and `SetupTestEnv` logic to `fixtures` package
[ ] Move `MockLLM`, `MockVectorStore` to `fixtures` package
[ ] Move `NewTestMux` (and server setup) to `fixtures` package
[ ] Use `sed` or similar to bulk update all `*_test.go` files to use `fixtures.SetupTestEnv` and `fixtures.TestEnv`
[ ] Verify `make test-all` passes

Edge Cases:
- Circular dependencies if `fixtures` imports `integration`. ensure `fixtures` is standalone.
- Environment variables consistency between packages.

Files Likely to Change:
- `internal/integration/fixtures/env.go` (NEW)
- `internal/integration/fixtures/server.go` (NEW)
- `internal/integration/fixtures/mocks.go` (NEW)
- `internal/integration/testutils.go`
- `internal/integration/testserver.go` (may be deleted or empty)

---

## TASK-002 — Implement Test Helpers (TDD)

Goal:
Add high-level helper methods to `fixtures.TestEnv` to simplify object creation (User, Chatbot, Source).

Scope:
- Implement `CreateUser`, `CreateChatbot`, `CreateSource` in `fixtures` package.
- Use TDD: Write a new test in `internal/integration/fixtures/helpers_test.go` that *uses* these methods before they exist (failing compilation/test), then implement them.

Checklist:
[ ] Create `internal/integration/fixtures/helpers_test.go`
[ ] Write failing test using `env.CreateUser("test@example.com")`
[ ] Implement `CreateUser` in `fixtures`
[ ] Write failing test using `env.CreateChatbot(user, "MyBot")`
[ ] Implement `CreateChatbot` in `fixtures`
[ ] Write failing test using `env.CreateSource(bot, "http://example.com")`
[ ] Implement `CreateSource` in `fixtures` (handling multipart/DB interactions inside)
[ ] Run tests to verify helpers

Edge Cases:
- `CreateChatbot` needs to handle Plan limits? (Should force Pro plan by default or allow specifying?)
- Helper should probably auto-assign a scalable plan (e.g., Pro) to avoid "plan limit reached" errors in tests unless testing limits specifically.

Files Likely to Change:
- `internal/integration/fixtures/helpers.go` (NEW)
- `internal/integration/fixtures/helpers_test.go` (NEW)

---

## TASK-003 — Refactor Evidence Tests

Goal:
Refactor the specific evidence files to use the new `fixtures` package and helpers, eliminating boilerplate.

Scope:
- `internal/integration/analytics_full_coverage_test.go`
- `internal/integration/url_discovery_test.go`
- Replace manual setup code with `fixtures` calls.

Checklist:
[ ] Refactor `internal/integration/analytics_full_coverage_test.go`
[ ] Refactor `internal/integration/url_discovery_test.go`
[ ] Verify both tests pass
[ ] Lint check

Edge Cases:
- `analytics_full_coverage_test.go` does specific analytics setups (manual DB inserts). Ensure helpers don't conflict or obscure necessary details.
- `url_discovery_test.go` sets specific discovery modes (`auto`, `pending`). Ensure `CreateChatbot` helper accepts options or config.

Files Likely to Change:
- `internal/integration/analytics_full_coverage_test.go`
- `internal/integration/url_discovery_test.go`
