# Task 002: Middleware Clarification ✅ Completed

## Agent Prompt

> **Objective:** Clarify the distinction between `internal/api/middleware/` and `pkg/middleware/` to reduce cognitive overhead for new contributors.
>
> **Context:** The codebase has two middleware locations serving different purposes. While logically sound, this can confuse new developers. This task documents the distinction and optionally renames the API-specific middleware for clarity.
>
> **Approach:**
> 1. Add documentation to both middleware directories
> 2. Optionally rename `internal/api/middleware/` to `internal/api/guards/`
> 3. Update all imports if renaming

---

## Problem Statement

Two middleware locations exist:
- `internal/api/middleware/` (2 files) - API-specific authentication guards
- `pkg/middleware/` (21 files) - Reusable infrastructure middleware

New contributors may not understand which to use for new middleware.

## Impact

- **Low Risk**: Documentation or rename with import updates
- **Improved DX**: Clearer code organization
- **Reduced Onboarding Time**: Self-documenting structure

---

## Acceptance Criteria

### Option A: Documentation Only (Recommended for minimal change)

- [ ] `internal/api/middleware/README.md` created with clear purpose
- [ ] `pkg/middleware/README.md` created with clear purpose
- [ ] Each middleware file has package-level doc comment

### Option B: Rename + Documentation (More invasive but clearer)

- [x] `internal/api/middleware/` renamed to `internal/api/guards/`
- [x] All imports updated across codebase
- [x] Both directories have README.md files
- [x] All tests pass

---

## Implementation Plan

### Option A: Documentation Only

#### Step 1: Create internal/api/middleware/README.md

- [ ] **Step 1.1**: Create README
  ```bash
  cat > internal/api/middleware/README.md << 'EOF'
  # API Middleware (Guards)
  
  This package contains **API-specific middleware** that enforces access control rules
  for specific route groups.
  
  ## Contents
  
  | File | Purpose |
  |---|---|
  | `admin.go` | Validates admin access for `/admin/*` routes |
  
  ## When to Use This Package
  
  Add middleware here when:
  - It's specific to the HTTP API layer
  - It enforces authorization rules (not authentication)
  - It applies to specific route groups, not globally
  
  ## When to Use `pkg/middleware/` Instead
  
  Use `pkg/middleware/` for:
  - Cross-cutting concerns (logging, CORS, rate limiting)
  - Reusable middleware that could apply to any HTTP handler
  - Authentication (JWT validation)
  - Request enrichment (request ID, plan loading)
  EOF
  ```

#### Step 2: Create pkg/middleware/README.md

- [ ] **Step 2.1**: Create README
  ```bash
  cat > pkg/middleware/README.md << 'EOF'
  # Infrastructure Middleware
  
  This package contains **reusable, cross-cutting middleware** that applies
  globally or to broad categories of routes.
  
  ## Contents
  
  | File | Purpose |
  |---|---|
  | `auth.go` | JWT authentication |
  | `cors.go` | CORS handling |
  | `ratelimit.go` | Rate limiting |
  | `security.go` | Security headers (CSP, HSTS, etc.) |
  | `request_id.go` | Request ID propagation |
  | `recovery.go` | Panic recovery |
  | `requestlog.go` | Request logging |
  | `plan_loader.go` | Load user plan into context |
  | `plan_context.go` | Plan context helpers |
  | `maxbytes.go` | Request size limits |
  | `organization.go` | Tenant context extraction |
  
  ## When to Use This Package
  
  Add middleware here when:
  - It's a cross-cutting concern (applies broadly)
  - It's reusable across different applications
  - It handles authentication (not authorization)
  - It enriches requests with context
  
  ## Middleware Chain Order
  
  The recommended order is defined in `cmd/server/main.go`:
  
  ```
  RequestID → Security → Recovery → Logger → MaxBytes → PlanLoader → RateLimit → Router
  ```
  EOF
  ```

#### Step 3: Add Package Doc Comments

- [ ] **Step 3.1**: Update `internal/api/middleware/admin.go`
  ```go
  // Package middleware provides API-specific access control guards.
  // For reusable infrastructure middleware, see pkg/middleware.
  package middleware
  ```

- [ ] **Step 3.2**: Update `pkg/middleware/auth.go` (first file alphabetically)
  ```go
  // Package middleware provides reusable HTTP middleware for cross-cutting concerns
  // including authentication, CORS, rate limiting, logging, and security headers.
  // For API-specific authorization guards, see internal/api/middleware.
  package middleware
  ```

---

### Option B: Rename to `internal/api/guards/` ✅ Completed

#### Step 1: Rename Directory

- [x] **Step 1.1**: Create new directory and move files
  ```bash
  mkdir -p internal/api/guards
  mv internal/api/middleware/*.go internal/api/guards/
  mv internal/api/middleware/*_test.go internal/api/guards/
  rmdir internal/api/middleware
  ```

- [x] **Step 1.2**: Update package name in files
  ```bash
  sed -i '' 's/package middleware/package guards/' internal/api/guards/*.go
  ```

#### Step 2: Update Imports

- [x] **Step 2.1**: Find all imports
  ```bash
  grep -r '"github.com/onurceri/botla-co/internal/api/middleware"' --include="*.go"
  ```

- [x] **Step 2.2**: Update each import
  ```go
  // Before
  import "github.com/onurceri/botla-co/internal/api/middleware"

  // After
  import "github.com/onurceri/botla-co/internal/api/guards"
  ```

- [x] **Step 2.3**: Update usage references
  ```go
  // Before
  middleware.AdminRequired()

  // After
  guards.AdminRequired()
  ```

#### Step 3: Add Documentation

- [x] Follow steps from Option A to add README files

#### Step 4: Verification

- [x] **Step 4.1**: Run tests
  ```bash
  go test ./internal/api/guards/...
  ```

- [x] **Step 4.2**: Run linter
  ```bash
  go build ./...
  ```

- [x] **Step 4.3**: Verify builds
  ```bash
  go build ./...
  ```

---

## Files to Modify

### Option A
| File | Action |
|---|---|
| `internal/api/middleware/README.md` | Create |
| `pkg/middleware/README.md` | Create |
| `internal/api/middleware/admin.go` | Add doc comment |
| `pkg/middleware/auth.go` | Add doc comment |

### Option B (Additional)
| File | Action |
|---|---|
| `internal/api/guards/*` | Create (renamed) |
| `internal/api/router/router.go` | Update imports |
| `internal/api/router/admin.go` | Update imports |

---

## Recommendation

**Start with Option A** (documentation only). It achieves the goal with minimal risk. Option B can be done later if the team feels stronger naming would help.

---

## Rollback Plan

For Option A: Delete the README files
For Option B: Reverse the rename with git
```bash
git checkout main -- internal/api/middleware/
rm -rf internal/api/guards/
```
