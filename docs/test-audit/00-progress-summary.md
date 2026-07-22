# Test Coverage Audit — Progress Summary

> **Date**: 2026-07-20
> **Goal**: Verify that every feature scenario documented in `docs/features/` has equivalent automated test coverage before deploying to PROD.

---

## 1. Scope & Inventory — DONE

- **Source of truth**: `docs/features/` — 16 feature areas, **2,855 documented scenario checkboxes** (not to be confused with `frontend/src/features/`, which is code modules).
- **Test suites inventoried**:
  - Backend: 266 `*_test.go` files (unit next to source; integration in `internal/integration/` + `internal/repository/*_integration_test.go`); Makefile targets `test-fast` / `test-integration` / `test-integration-real`; 90% coverage gate (`make cover-gate`); `make ci` = vet + lint + test-no-pdf.
  - Frontend: ~85 Vitest files + 15 Playwright e2e specs; coverage thresholds 80/80/75/80 in `vite.config.js`.
  - Widget: 7 Vitest files + 6 Playwright e2e specs (5 browser projects); **no coverage thresholds**; vitest v2 vs frontend v4 (version skew).

## 2. Test Setup Health — DONE (live verification)

### Backend (Go)
- `go build ./...` OK. Pure unit packages (auth, services, rag, processing, models, scraper, pkg…) green.
- **P0 PROD BLOCKER — server cannot connect to Postgres**: `internal/db/db.go:19` builds the DSN with `default_exec_mode=simple_protocol`. pgx v5 does **not** know this parameter (correct name: `default_query_exec_mode`), so pgx forwards it to Postgres as a startup parameter → `FATAL: unrecognized configuration parameter "default_exec_mode" (SQLSTATE 42704)`. **Every DB connection fails; the server cannot boot.** Reproduced live against a healthy dev Postgres via the `cmd/server` test path. Introduced by commit `849eb4d` (2026-01-17, "feat: update DSN construction to use simple protocol").
- `internal/db` `TestBuildDSN` is red — expectation was never updated for the new DSN param (would have caught the bug above if run).
- `cmd/server/main_test.go`: 3 tests (`TestNewApplication_*`) are red — they require a live DB (non-hermetic unit tests) and also hit the DSN bug.
- `-short` gating is inconsistent: `internal/integration/main_test.go` (TestMain), many integration files, and all `internal/repository/*_integration_test.go` do **not** check `testing.Short()` → `make test-fast` (documented as "fast unit tests only") still executes DB-dependent tests and fails without a live DB.
- Integration fixtures (`internal/integration/fixtures/env.go`) and `internal/testdb` build their own DSNs without the bad param → integration suite is not blocked by the DSN bug once a DB is up.

### Frontend
- **28 failing unit tests across 8 files** (of 420 tests / 77 files): stale `isValidToken.test.ts` (imports deleted export from `../App`, 13 tests), `OnboardingWizard` (7), `ColorsSection` (3), `PrivateRoute`, `client.test`, `DashboardLayout`, `SourceCard`, `PlanPage.hidden-limits`.
- Stubbed Playwright e2e: **186 passed, 1 skipped** — but the skipped one is `smoke.spec.ts`, the **only full user-journey test (register → create chatbot → add source → chat)**, and it always skips unless `E2E_API_BASE` is set (real backend). Default `npm run e2e` therefore never exercises the true happy path.

### Widget
- Unit: 91/91 green (7 files).
- e2e runs against `vite preview` → requires a fresh `npm run build` first; `test:e2e` script does not build → risk of testing stale bundles.

### Environment incident
- During verification the dev `botla-postgres` container disappeared mid-run (external cause). Restored via `make up`; verified healthy. The integration compose stack (`botla-co-*-integration-1`) is currently up.
