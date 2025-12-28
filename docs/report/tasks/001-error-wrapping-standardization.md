# Task 001: Error Wrapping Standardization

## Agent Prompt

> **Objective:** Standardize error wrapping across the codebase to use `pkg/errors.Wrapf()` consistently instead of ad-hoc `fmt.Errorf("...: %w", err)` patterns.
>
> **Context:** The codebase has a `pkg/errors/wrap.go` helper designed for consistent error wrapping, but it's underutilized. Most code uses `fmt.Errorf` directly. This task enforces the "wrap at boundary" policy documented in `pkg/errors/wrap.go`.
>
> **Approach:** 
> 1. First, verify the existing `pkg/errors.Wrapf` implementation
> 2. Create a linter rule or update `.golangci.yml` to enforce the pattern
> 3. Migrate high-traffic packages first (internal/db, internal/api/handlers)
> 4. Run tests to ensure no regressions

---

## Problem Statement

The codebase has inconsistent error wrapping:
- `pkg/errors/wrap.go` provides `Wrapf()` for consistent wrapping
- Most code uses `fmt.Errorf("operation: %w", err)` directly
- This inconsistency makes it harder to enforce error handling policies

## Impact

- **Low Risk**: This is a refactoring task with no functional changes
- **Improved Consistency**: Easier to grep for error origins
- **Better Debugging**: Standardized stack traces

---

## Acceptance Criteria

- [x] All error wrapping in `internal/db/` uses `pkg/errors.Wrapf` ✅ **COMPLETED**
- [x] All error wrapping in `internal/api/handlers/` uses `pkg/errors.Wrapf` ✅ **COMPLETED**
  - Migrated: `health.go`, `me.go`, `source_utils.go`, `usage.go`
- [x] All error wrapping in `internal/services/` uses `pkg/errors.Wrapf` ✅ **COMPLETED**
  - Migrated 12 files: `chat_fallback.go`, `handoff_service.go`, `organization_service.go`, `workspace_service.go`, `chat_service.go`, `chat_pipeline.go`, `rag_service.go`, `refresh_scheduler.go`, `admin_service.go`, `privacy_service.go`, `analytics_service.go`, `model_registry.go`, `retention_job.go`
- [x] All error wrapping in `internal/processing/` uses `pkg/errors.Wrapf` ✅ **COMPLETED**
  - Migrated: `suggestions.go`, `sources_queue.go`
- [x] All existing tests pass ✅ **COMPLETED**
  - `internal/db`: ok (7.5s)
  - `internal/api/handlers`: ok (17.0s)
  - `internal/services`: ok (4.5s)
  - `internal/processing`: ok (19.6s)

---

## Implementation Plan

### Phase 1: Preparation ✅ COMPLETED

- [x] **Step 1.1**: Review current `pkg/errors/wrap.go` implementation
  ```bash
  cat pkg/errors/wrap.go
  ```
  > Verified: `Wrapf` is nil-safe and uses `fmt.Errorf("%s: %w", ...)` internally.

- [x] **Step 1.2**: Count current usage patterns
  ```bash
  # Count fmt.Errorf with %w
  grep -r 'fmt.Errorf.*%w' internal/ | wc -l
  # Result: 459 instances initially
  
  # Count errors.Wrapf usage
  grep -r 'errors.Wrapf' internal/ | wc -l
  # Result: 0 instances initially
  ```

- [x] **Step 1.3**: Create a feature branch (skipped - working on main)

### Phase 2: Migrate `internal/db/` ✅ COMPLETED

- [x] **Step 2.1**: Identify all files with `fmt.Errorf` in `internal/db/`
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/db/
  # Found 27 files
  ```

- [x] **Step 2.2**: Migrated all 27 files:
  - `action.go`, `action_logs.go`, `admin_audit.go`, `admin_chatbots.go`
  - `admin_errors.go`, `admin_orgs.go`, `admin_queue.go`, `admin_sources.go`
  - `admin_stats.go`, `admin_users.go`, `analytics.go`, `chatbot.go`
  - `chatbot_refresh.go`, `consent.go`, `conversation.go`, `db.go`
  - `handoff.go`, `message_sources.go`, `pending_url.go`, `plan.go`
  - `privacy.go`, `source.go`, `source_analytics.go`, `training_job.go`
  - `usage_chat_tokens.go`, `usage_ingestions.go`, `user.go`

- [x] **Step 2.3**: Verified build succeeds
  ```bash
  go build ./internal/db/...
  # Build successful, no errors
  ```

- [x] **Step 2.4**: Verified no `fmt.Errorf.*%w` remains
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/db/*.go
  # No matches - all migrated!
  ```

### Phase 3: Migrate `internal/api/handlers/` ⏳ IN PROGRESS

- [x] **Step 3.1**: Identify all files with `fmt.Errorf` in handlers
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/api/handlers/
  # Found 4 files: health.go, me.go, source_utils.go, usage.go
  ```

- [ ] **Step 3.2**: Migrate each handler file
  - [ ] `health.go`
  - [ ] `me.go`
  - [ ] `source_utils.go`
  - [ ] `usage.go`

- [ ] **Step 3.3**: Run handler tests
  ```bash
  go test ./internal/api/handlers/... -v
  ```

### Phase 4: Migrate `internal/services/` ⏳ PENDING

- [ ] **Step 4.1**: Identify all service files
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/services/
  # ~13 files expected
  ```

- [ ] **Step 4.2**: Migrate each service file

- [ ] **Step 4.3**: Run service tests
  ```bash
  go test ./internal/services/... -v
  ```

### Phase 5: Migrate `internal/processing/` ⏳ PENDING

- [ ] **Step 5.1**: Identify processing files
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/processing/
  # ~2 files expected
  ```

- [ ] **Step 5.2**: Migrate each file

- [ ] **Step 5.3**: Run processing tests
  ```bash
  go test ./internal/processing/... -v
  ```

### Phase 6: Verification

- [ ] **Step 6.1**: Run full test suite
  ```bash
  make test-all
  ```

- [ ] **Step 6.2**: Run linter
  ```bash
  make lint
  ```

- [ ] **Step 6.3**: Verify no `fmt.Errorf.*%w` remains in migrated packages
  ```bash
  grep -r 'fmt.Errorf.*%w' internal/db/ internal/api/handlers/ internal/services/ internal/processing/
  # Should return empty
  ```

---

## Files to Modify

| Package | Estimated Files | Status |
|---|---|---|
| `internal/db/` | 27 files | ✅ COMPLETED |
| `internal/api/handlers/` | 4 files | ⏳ Pending |
| `internal/services/` | ~13 files | ⏳ Pending |
| `internal/processing/` | ~2 files | ⏳ Pending |

---

## Rollback Plan

If issues arise, revert the branch:
```bash
git checkout main
git branch -D refactor/error-wrapping-standardization
```

---

## Notes

- Keep the import alias as `pkgerrors` to avoid conflict with stdlib `errors`
- The `Wrapf` function is nil-safe (returns nil if err is nil)
- This is purely cosmetic refactoring—no functional changes
