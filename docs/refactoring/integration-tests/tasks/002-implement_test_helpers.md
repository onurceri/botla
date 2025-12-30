## TASK-002 — Implement Test Helpers (TDD)

Goal:
Add high-level helper methods to `fixtures.TestEnv` to simplify object creation (User, Chatbot, Source).

Scope:
- Implement `CreateUser`, `CreateChatbot`, `CreateSource` in `fixtures` package.
- Use TDD: Write a new test in `internal/integration/fixtures/helpers_test.go` that *uses* these methods before they exist (failing compilation/test), then implement them.

Checklist:
[x] Create `internal/integration/fixtures/helpers_test.go`
[x] Write failing test using `env.CreateUser("test@example.com")`
[x] Implement `CreateUser` in `fixtures`
[x] Write failing test using `env.CreateChatbot(user, "MyBot")`
[x] Implement `CreateChatbot` in `fixtures`
[x] Write failing test using `env.CreateSource(bot, "http://example.com")`
[x] Implement `CreateSource` in `fixtures` (handling multipart/DB interactions inside)
[x] Run tests to verify helpers

Edge Cases:
- `CreateChatbot` needs to handle Plan limits? (Should force Pro plan by default or allow specifying?)
- Helper should probably auto-assign a scalable plan (e.g., Pro) to avoid "plan limit reached" errors in tests unless testing limits specifically.

Files Likely to Change:
- `internal/integration/fixtures/helpers.go`
- `internal/integration/fixtures/helpers_test.go`
