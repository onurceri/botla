# Task 05: Error Handling Standardization

## Priority
**High** - Critical for incident response and debugging

## Problem Statement

Error wrapping and logging is inconsistent across the codebase:
- Some layers return raw errors without context
- Others wrap with redundant information
- Mixed usage of `log.Printf` and structured `pkg/logger`

This increases Mean Time to Repair (MTTR) during incidents.

## Evidence

```go
// internal/scraper/sitemap_parser.go - raw error returned
resp, err := http.Get(url)
if err != nil {
    return nil, err  // No context about what operation failed
}

// Mixed logging
log.Printf("error: %v", err)        // Unstructured
logger.Error("operation failed", "error", err, "user_id", userID)  // Structured
```

## Implementation Plan

### Phase 1: Enforce "Wrap at Boundary" Policy

All errors from external packages must be wrapped:

```go
// Before
resp, err := http.Get(url)
if err != nil {
    return nil, err
}

// After
resp, err := http.Get(url)
if err != nil {
    return nil, fmt.Errorf("fetching sitemap %s: %w", url, err)
}
```

### Phase 2: Add wrapcheck Linter

**File**: `.golangci.yml`
```yaml
linters:
  enable:
    - wrapcheck

linters-settings:
  wrapcheck:
    ignoreSigs:
      - .Errorf(
      - errors.New(
    ignorePackageGlobs:
      - github.com/onurceri/botla-co/*
```

### Phase 3: Standardize on Structured Logger

Replace all `log.Printf` with `pkg/logger`:

```go
// Before
log.Printf("failed to process: %v", err)

// After  
logger.Error("failed to process", 
    "error", err,
    "request_id", ctx.Value("request_id"),
    "user_id", userID,
)
```

### Phase 4: Create Error Context Helpers

**File**: `pkg/errors/wrap.go`
```go
package errors

import "fmt"

func Wrapf(err error, format string, args ...interface{}) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
```

## Affected Files

| File | Action | Description |
|------|--------|-------------|
| `.golangci.yml` | MODIFY | Add wrapcheck linter |
| `internal/scraper/*.go` | MODIFY | Wrap external errors |
| `internal/rag/*.go` | MODIFY | Wrap external errors |
| All files with `log.Printf` | MODIFY | Use structured logger |

## Acceptance Criteria

- [ ] wrapcheck linter enabled and passing
- [ ] No raw errors from external packages
- [ ] All logging uses `pkg/logger`
- [ ] Error traces contain full context
- [ ] Existing tests pass

## Estimated Effort
**Size**: Medium (2-3 days)
