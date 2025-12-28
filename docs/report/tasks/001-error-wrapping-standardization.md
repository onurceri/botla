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

- [ ] All error wrapping in `internal/db/` uses `pkg/errors.Wrapf`
- [ ] All error wrapping in `internal/api/handlers/` uses `pkg/errors.Wrapf`
- [ ] All error wrapping in `internal/services/` uses `pkg/errors.Wrapf`
- [ ] All error wrapping in `internal/processing/` uses `pkg/errors.Wrapf`
- [ ] All existing tests pass
- [ ] Documentation updated in `pkg/errors/wrap.go` if needed

---

## Implementation Plan

### Phase 1: Preparation

- [ ] **Step 1.1**: Review current `pkg/errors/wrap.go` implementation
  ```bash
  cat pkg/errors/wrap.go
  ```

- [ ] **Step 1.2**: Count current usage patterns
  ```bash
  # Count fmt.Errorf with %w
  grep -r 'fmt.Errorf.*%w' internal/ | wc -l
  
  # Count errors.Wrapf usage
  grep -r 'errors.Wrapf' internal/ | wc -l
  ```

- [ ] **Step 1.3**: Create a feature branch
  ```bash
  git checkout -b refactor/error-wrapping-standardization
  ```

### Phase 2: Migrate `internal/db/`

- [ ] **Step 2.1**: Identify all files with `fmt.Errorf` in `internal/db/`
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/db/
  ```

- [ ] **Step 2.2**: For each file, replace:
  ```go
  // Before
  return nil, fmt.Errorf("query chatbots by user id: %w", err)
  
  // After
  import pkgerrors "github.com/onurceri/botla-co/pkg/errors"
  return nil, pkgerrors.Wrapf(err, "query chatbots by user id")
  ```

- [ ] **Step 2.3**: Run tests for db package
  ```bash
  go test ./internal/db/... -v
  ```

### Phase 3: Migrate `internal/api/handlers/`

- [ ] **Step 3.1**: Identify all files with `fmt.Errorf` in handlers
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/api/handlers/
  ```

- [ ] **Step 3.2**: Migrate each handler file (prioritize by size)

- [ ] **Step 3.3**: Run handler tests
  ```bash
  go test ./internal/api/handlers/... -v
  ```

### Phase 4: Migrate `internal/services/`

- [ ] **Step 4.1**: Identify all service files
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/services/
  ```

- [ ] **Step 4.2**: Migrate each service file

- [ ] **Step 4.3**: Run service tests
  ```bash
  go test ./internal/services/... -v
  ```

### Phase 5: Migrate `internal/processing/`

- [ ] **Step 5.1**: Identify processing files
  ```bash
  grep -l 'fmt.Errorf.*%w' internal/processing/
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

| Package | Estimated Files | Priority |
|---|---|---|
| `internal/db/` | ~15 files | High |
| `internal/api/handlers/` | ~30 files | High |
| `internal/services/` | ~12 files | Medium |
| `internal/processing/` | ~6 files | Medium |

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
