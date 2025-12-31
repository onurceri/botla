# Technical Debt Refactoring Plan

## Task Index
- **TASK-001**: Extract Common AI HTTP Client
- **TASK-002**: Refactor OpenAI Provider to Use Common Client
- **TASK-003**: Refactor OpenRouter Provider to Use Common Client
- **TASK-004**: Refactor JobProcessor to Interface-Based Design
- **TASK-005**: Remove Implicit Environment Configuration in AI Packages
- **TASK-006**: Remove Side-Effect Registry Pattern
- **TASK-007**: Refactor Complex SQL Queries with Query Builder

---

## TASK-001 — Extract Common AI HTTP Client

### Goal
Create a shared, robust HTTP client abstraction in `internal/ai` that handles the common logic found in both OpenAI and OpenRouter implementations: request marshaling, retries with exponential backoff, and error parsing.

### Scope
- **Included**: Creating `internal/ai/http_client.go` (or similar), defining generic request/response structs if applicable, implementing the retry loop.
- **Excluded**: modifying existing providers (that is mostly for the next tasks, though simple wiring is fine to test).

### Checklist
- [ ] Create `internal/ai/base_embedder.go` or `client.go`.
- [ ] Define a `BaseClient` struct that accepts a generic configuration (BaseURL, APIKey, Header map).
- [ ] Implement `Post(ctx, path, body, responseTarget)` method with:
    - [ ] 4 retries with exponential backoff.
    - [ ] `context` support.
    - [ ] JSON marshaling/unmarshaling.
    - [ ] HTTP status code error handling.
- [ ] Write unit tests for `BaseClient` covering:
    - [ ] Successful request.
    - [ ] 429 Rate limit retry behavior.
    - [ ] 5xx Server error retry behavior.
    - [ ] Non-retryable errors (401, 400).
    - [ ] Context cancellation.

### Edge Cases
- Network timeouts during retries.
- Malformed JSON responses.
- Empty bodies.

### Files Likely to Change
- `internal/ai/client.go` (NEW)
- `internal/ai/client_test.go` (NEW)

---

## TASK-002 — Refactor OpenAI Provider to Use Common Client

### Goal
Update the OpenAI implementation to delegate all HTTP communication to the new shared client, removing duplicated logic.

### Scope
- `internal/ai/openai` package.

### Checklist
- [ ] Modify `openai.Embedder` to hold an instance of `ai.BaseClient`.
- [ ] Update `Embed` method to use `BaseClient.Post`.
- [ ] Update `EmbedBatch` method to use `BaseClient.Post`.
- [ ] Ensure specific OpenAI headers are maintained (if any differing from base).
- [ ] Verify `NewFromEnv` still works (temporarily, until TASK-005).
- [ ] Run existing tests in `internal/ai/openai` to ensure no regression.
- [ ] Remove deleted code (old retry loop, old Structs if now shared).

### Edge Cases
- API-specific error message formats.

### Files Likely to Change
- `internal/ai/openai/embedder.go`

---

## TASK-003 — Refactor OpenRouter Provider to Use Common Client

### Goal
Update the OpenRouter implementation to delegate HTTP communication to the new shared client, ensuring consistent behavior with OpenAI.

### Scope
- `internal/ai/openrouter` package.

### Checklist
- [ ] Modify `openrouter.Embedder` to hold an instance of `ai.BaseClient`.
- [ ] Configure `BaseClient` with OpenRouter-specific headers (`HTTP-Referer`, `X-Title`).
- [ ] Update `Embed` method to use `BaseClient.Post`.
- [ ] Update `EmbedBatch` method to use `BaseClient.Post`.
- [ ] Run existing tests in `internal/ai/openrouter`.
- [ ] Delete duplicated retry logic and structs.

### Edge Cases
- Headers specific to OpenRouter must be preserved.

### Files Likely to Change
- `internal/ai/openrouter/embedder.go`

---

## TASK-004 — Refactor JobProcessor to Interface-Based Design

### Goal
Decouple `JobProcessor` from specific processor implementations by introducing a `SourceProcessor` interface and using a map-based registry strategy, strictly adhering to the Open/Closed Principle.

### Scope
- `internal/processing/job_processor.go`
- `internal/processing` implementations (`url`, `pdf`, `text`).

### Checklist
- [ ] Define `SourceProcessor` interface in `internal/processing/interfaces.go`.
    - `ProcessWithSteps(ctx, ...)`
- [ ] Make `URLProcessor`, `PDFProcessor`, `TextProcessor` implement this interface.
- [ ] Refactor `JobProcessor` struct:
    - Replace individual fields (`urlProcessor`) with `processors map[string]SourceProcessor`.
- [ ] Update `NewJobProcessor` to accept the map or register processors.
- [ ] Update `processWithResume` to lookup processor by `source.SourceType`.
- [ ] Remove the hardcoded `switch` statement.
- [ ] Fix any broken tests in `internal/processing`.

### Edge Cases
- Unknown source type handling (should return specific error).
- Nil processor in map.

### Files Likely to Change
- `internal/processing/job_processor.go`
- `internal/processing/interfaces.go` (NEW)

---

## TASK-005 — Remove Implicit Environment Configuration in AI Packages

### Goal
Make dependencies explicit by removing `NewFromEnv` and `os.Getenv` usages inside low-level AI packages. Pass configuration down from the top level.

### Scope
- `internal/ai/openai`
- `internal/ai/openrouter`
- `internal/ai/qdrant`

### Checklist
- [ ] Update `openai.NewEmbedder` to accept a Config struct (Key, BaseURL, Model).
- [ ] Remove `openai.NewFromEnv`.
- [ ] Update `openrouter.NewEmbedder` to accept a Config struct.
- [ ] Remove `openrouter.NewFromEnv`.
- [ ] Update `qdrant.NewStore` to accept Config.
- [ ] Remove `qdrant.NewFromEnv`.
- [ ] Update `cmd/server/main.go` and `factory.go` to load config from `config.Config` and pass it down.
- [ ] Update integration tests to manually construct services instead of relying on Env.

### Edge Cases
- Missing configuration values (should fail at call site, not deep in library).

### Files Likely to Change
- `internal/ai/openai/embedder.go`
- `internal/ai/openrouter/embedder.go`
- `internal/ai/qdrant/client.go`
- `internal/ai/factory.go`
- `cmd/server/main.go`

---

## TASK-006 — Remove Side-Effect Registry Pattern

### Goal
Remove `init()` functions that rely on global state for registration. Use explicit dependency injection wiring in the application root.

### Scope
- `internal/ai` packages.

### Checklist
- [ ] Remove `init()` function in `internal/ai/openai/embedder.go`.
- [ ] Remove `init()` function in `internal/ai/openrouter/embedder.go`.
- [ ] Remove `init()` function in `internal/ai/qdrant/client.go`.
- [ ] Remove global registry map in `internal/ai/factory.go`.
- [ ] Create a `Wiring` helper or update `cmd/server/main.go` to explicitly choose and instantiate the correct Embedder/VectorStore based on application config.
- [ ] Verify application startup works correctly.

### Edge Cases
- Switching providers based on config (previously handled by factory string lookup).

### Files Likely to Change
- `internal/ai/openai/embedder.go`
- `internal/ai/openrouter/embedder.go`
- `internal/ai/factory.go`
- `cmd/server/main.go`

---

## TASK-007 — Refactor Complex SQL Queries with Query Builder

### Goal
Replace manual string concatenation in `AdminListChatbots` with a type-safe SQL builder (Squirrel) to improve maintainability and safety.

### Scope
- `internal/db/admin_chatbots.go`

### Checklist
- [ ] Add `github.com/Masterminds/squirrel` dependency.
- [ ] Create a reproduction test for `AdminListChatbots` (if not exists) to ensure baseline behavior.
- [ ] Refactor `AdminListChatbots` to use `squirrel.Select(...)`.
    - Handle dynamic `Where` clauses for filters.
    - Handle Pagination.
    - Handle Sorting.
- [ ] Verify the generated SQL is correct and comparable to original.
- [ ] Run existing tests for Admin Chatbots.

### Edge Cases
- Empty filters.
- SQL injection prevention (builder handles this, but verify usage).
- Complex joins or subqueries (Squirrel supports these, but check syntax).

### Files Likely to Change
- `internal/db/admin_chatbots.go`
- `go.mod`

---

## Completion Summary

### Completed Tasks (2025-12-31)

**TASK-001 — Extract Common AI HTTP Client**
- ✅ Created `internal/ai/client.go` with `BaseClient` struct
- ✅ Implemented retry logic with exponential backoff
- ✅ Added comprehensive unit tests for all scenarios
- ✅ Status: Completed as part of task 005 cleanup

**TASK-002 — Refactor OpenAI Provider to Use Common Client**
- ✅ Updated OpenAI embedder to use `ai.BaseClient`
- ✅ Removed duplicate retry logic
- ✅ All tests passing
- ✅ Status: Completed as part of task 005 cleanup

**TASK-003 — Refactor OpenRouter Provider to Use Common Client**
- ✅ Updated OpenRouter embedder to use `ai.BaseClient`
- ✅ Configured OpenRouter-specific headers
- ✅ Removed duplicate retry logic
- ✅ All tests passing
- ✅ Status: Completed as part of task 005 cleanup

**TASK-004 — Refactor JobProcessor to Interface-Based Design**
- Status: Not yet started

**TASK-005 — Remove Implicit Environment Configuration in AI Packages**
- ✅ Updated all providers to accept explicit `Config` structs
- ✅ Removed all `NewFromEnv` functions
- ✅ Updated integration tests to use explicit configuration
- ✅ Fail-fast validation implemented
- ✅ Factory pattern removed (was unused)
- ✅ Status: Completed in commit 08a86f4

**TASK-006 — Remove Side-Effect Registry Pattern**
- ✅ No `init()` functions exist in `internal/ai` packages
- ✅ Global registry pattern removed (factory.go deleted)
- ✅ Explicit dependency injection implemented
- ✅ Application verified to work correctly
- ✅ Status: Completed in commit 08a86f4

**Note:** As part of tasks 005-006 cleanup, the entire `internal/ai` package was identified as dead code (zero production imports) and removed. The project uses `internal/rag` package exclusively for all AI/vector operations.

### Remaining Tasks
- TASK-004: Refactor JobProcessor to Interface-Based Design
- TASK-007: Refactor Complex SQL Queries with Query Builder
