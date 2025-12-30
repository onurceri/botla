## Task Index
TASK-001: Define Sentinel Errors
TASK-002: Refactor OpenAI Client Error Handling
TASK-003: Refactor Scraper Error Handling
TASK-004: Refactor Job Processor and Verification

## TASK-001 — Define Sentinel Errors

Goal:
Define a set of reusable, exported sentinel errors in the `pkg/errors` package to replace magic strings throughout the application.

Scope:
- Create `pkg/errors/sentinel.go`
- Define errors for: Rate Limit, Timeout, Network/Temporary, Not Found, Context Cancelled.
- Ensure they are standard Go errors (using `errors.New`).

Checklist:
[ ] Create `pkg/errors/sentinel.go`.
[ ] Define `ErrRateLimit`.
[ ] Define `ErrTimeout`.
[ ] Define `ErrNetwork` (or `ErrTemporary`).
[ ] Define `ErrNotFound`.
[ ] Define `ErrContextCancelled`.
[ ] Create `pkg/errors/sentinel_test.go` to ensure no error variable collisions (optional but good for TDD verify).
[ ] Run `go test ./pkg/errors/...`.

Edge Cases:
- Variable naming conflicts.
- Ensuring error messages are descriptive.

Files Likely to Change:
- `pkg/errors/sentinel.go`
- `pkg/errors/sentinel_test.go`

## TASK-002 — Refactor OpenAI Client Error Handling

Goal:
Update the OpenAI client (`internal/rag`) to catch specific HTTP errors (like 429, 5xx) and wrap them with the new sentinel errors using `fmt.Errorf("... %w", ErrX)`.

Scope:
- `internal/rag/openai.go`
- `internal/rag` tests.

Checklist:
[ ] Create/Update `internal/rag/error_test.go` (or similar) to simulate HTTP 429 and 500 responses.
[ ] Assert that current implementation FAILS `errors.Is(err, pkgErrors.ErrRateLimit)` (TDD).
[ ] Modify `internal/rag/openai.go` to check `res.StatusCode` and return wrapped sentinel errors.
[ ] Handle Context Cancellation by wrapping `pkgErrors.ErrContextCancelled`.
[ ] Run tests and verify `errors.Is` checks pass.

Edge Cases:
- API returns 200 checks but body contains error (handled in existing logic, ensure wrapping there too if applicable).
- Timeout errors from `http.Client` vs Context cancellation.

Files Likely to Change:
- `internal/rag/openai.go`
- `internal/rag/openai_test.go`

## TASK-003 — Refactor Scraper Error Handling

Goal:
Update the Scraper logic (`internal/scraper`) to use/wrap sentinel errors, specifically for network issues, timeouts, and known blocking responses (403/429).

Scope:
- `internal/scraper/errors.go`
- `internal/scraper` parsing/fetching logic.

Checklist:
[ ] Update `internal/scraper/errors_test.go` to assert that specific error conditions return wrapped sentinel errors.
[ ] Modify `internal/scraper/errors.go` or the fetcher (e.g. `colly.go` or `default_scraper.go`) to map status codes to sentinel errors.
[ ] Run tests.

Edge Cases:
- `colly` might mask some errors, check how it exposes status codes.
- Partial failures.

Files Likely to Change:
- `internal/scraper/errors.go`
- `internal/scraper/colly.go` (or where requests are made)

## TASK-004 — Refactor Job Processor and Verification

Goal:
Update the `JobProcessor` to use `errors.Is` for retry logic instead of string matching. Verify that the entire flow works with the new error types.

Scope:
- `internal/processing/job_processor.go`
- `internal/processing/retry_test.go`

Checklist:
[ ] Modify `internal/processing/retry_test.go`:
    - Replace string-based test cases with sentinel error cases.
    - Add test cases where the error is wrapped (e.g., `fmt.Errorf("context: %w", ErrRateLimit)`).
    - Assert `isRetryableError` returns true for these wrapped errors.
[ ] Run test (Expect FAIL for encapsulated errors if logic not updated, or PASS if only testing the function against new inputs).
[ ] Modify `internal/processing/job_processor.go`:
    - Remove/Deprecate the string matching list.
    - Use `errors.Is(err, pkgErrors.ErrRateLimit)`, etc.
[ ] Run all tests in `internal/processing`.
[ ] Run all tests in `internal/integration` (to ensure no regressions in full flows).

Edge Cases:
- Mixed errors (string + wrapped). The refactor should prefer `errors.Is` but might need to keep string matching as a fallback if not all errors are converted yet (safe transition). *Decision: For this task, we aim for full replacement, but a fallback is acceptable if deemed safer.*

Files Likely to Change:
- `internal/processing/job_processor.go`
- `internal/processing/retry_test.go`
