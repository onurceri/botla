# Phase 1: Backend Foundation

> **Estimated Time:** 3-4 days  
> **Priority:** Critical (Week 1)  
> **Depends On:** Nothing  

This phase establishes the admin authentication system, database schema, and core API endpoints.

---

## Step 1.1: Database Migration

Create the database migration for admin tables.

### Tasks

- [ ] **Create migration file** `db/migrations/000041_admin_platform.up.sql`
  
  Add the following tables and columns:
  
  ```sql
  -- Admin role flag on users table
  ALTER TABLE users ADD COLUMN is_platform_admin BOOLEAN DEFAULT FALSE;
  CREATE INDEX idx_users_is_platform_admin ON users(is_platform_admin) WHERE is_platform_admin = TRUE;
  
  -- Admin audit log
  CREATE TABLE admin_audit_logs (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      admin_user_id UUID REFERENCES users(id),
      action TEXT NOT NULL,
      target_type TEXT NOT NULL,
      target_id UUID,
      details JSONB,
      ip_address INET,
      user_agent TEXT,
      created_at TIMESTAMPTZ DEFAULT NOW()
  );
  
  CREATE INDEX idx_admin_audit_logs_admin ON admin_audit_logs(admin_user_id);
  CREATE INDEX idx_admin_audit_logs_action ON admin_audit_logs(action);
  CREATE INDEX idx_admin_audit_logs_created ON admin_audit_logs(created_at DESC);
  CREATE INDEX idx_admin_audit_logs_target ON admin_audit_logs(target_type, target_id);
  
  -- Error tracking table
  CREATE TABLE error_logs (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      error_type TEXT NOT NULL,
      message TEXT NOT NULL,
      stack_trace TEXT,
      request_path TEXT,
      request_method TEXT,
      user_id UUID,
      chatbot_id UUID,
      organization_id UUID,
      severity TEXT DEFAULT 'error' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
      context JSONB,
      created_at TIMESTAMPTZ DEFAULT NOW()
  );
  
  CREATE INDEX idx_error_logs_type ON error_logs(error_type);
  CREATE INDEX idx_error_logs_created ON error_logs(created_at DESC);
  CREATE INDEX idx_error_logs_severity ON error_logs(severity);
  CREATE INDEX idx_error_logs_chatbot ON error_logs(chatbot_id) WHERE chatbot_id IS NOT NULL;
  
  -- Platform metrics (time-series for trends)
  CREATE TABLE platform_metrics (
      id BIGSERIAL PRIMARY KEY,
      metric_name TEXT NOT NULL,
      metric_value BIGINT NOT NULL,
      recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      dimensions JSONB
  );
  
  CREATE INDEX idx_platform_metrics_name_time ON platform_metrics(metric_name, recorded_at DESC);
  ```

- [ ] **Create down migration** `db/migrations/000041_admin_platform.down.sql`
  
  ```sql
  DROP TABLE IF EXISTS platform_metrics;
  DROP TABLE IF EXISTS error_logs;
  DROP TABLE IF EXISTS admin_audit_logs;
  DROP INDEX IF EXISTS idx_users_is_platform_admin;
  ALTER TABLE users DROP COLUMN IF EXISTS is_platform_admin;
  ```

- [ ] **Run migration**
  ```bash
  make migrate-up
  ```

- [ ] **Verify migration**
  ```bash
  make psql
  # Then run:
  \d users  -- Should show is_platform_admin column
  \d admin_audit_logs
  \d error_logs
  \d platform_metrics
  ```

---

## Step 1.2: Update User Model

Update the User model to include the admin flag.

### Tasks

- [ ] **Update `internal/models/user.go`**
  
  Add field to User struct:
  ```go
  type User struct {
      // ... existing fields
      IsPlatformAdmin bool `json:"is_platform_admin"`
  }
  ```

- [ ] **Update `internal/db/user.go`**
  
  Add `is_platform_admin` to SELECT queries in:
  - `GetUserByID()`
  - `GetUserByEmail()`
  
  Example:
  ```go
  SELECT id, email, full_name, avatar_url, plan_id, preferred_language_id, 
         created_at, onboarding_completed, onboarding_step, onboarding_skipped, 
         onboarding_data, is_platform_admin
  FROM users WHERE id=$1 AND deleted_at IS NULL
  ```

- [ ] **Update auth token generation** in `internal/auth/jwt.go`
  
  Add `is_platform_admin` to JWT claims so it's available on every request.

---

## Step 1.3: Admin Middleware

Create middleware to protect admin routes.

### Tasks

- [ ] **Create `internal/api/middleware/admin.go`**
  
  ```go
  package middleware
  
  import (
      "net/http"
      
      "github.com/onurceri/botla-co/internal/auth"
  )
  
  // RequirePlatformAdmin ensures the request is from a platform admin
  func RequirePlatformAdmin(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          user := auth.UserFromContext(r.Context())
          if user == nil {
              http.Error(w, "Unauthorized", http.StatusUnauthorized)
              return
          }
          if !user.IsPlatformAdmin {
              http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
              return
          }
          next.ServeHTTP(w, r)
      })
  }
  ```

- [ ] **Write unit tests** `internal/api/middleware/admin_test.go`
  
  Test cases:
  - Request without user context returns 401
  - Request with non-admin user returns 403
  - Request with admin user passes through

---

## Step 1.4: Admin Audit Service

Create service for logging admin actions.

### Tasks

- [ ] **Create `internal/db/admin_audit.go`**
  
  ```go
  package db
  
  import (
      "context"
      "database/sql"
  )
  
  type AuditLogEntry struct {
      ID          string
      AdminUserID string
      Action      string
      TargetType  string
      TargetID    *string
      Details     map[string]any
      IPAddress   string
      UserAgent   string
      CreatedAt   time.Time
  }
  
  func InsertAuditLog(ctx context.Context, pool *sql.DB, entry AuditLogEntry) error {
      // Implementation
  }
  
  func ListAuditLogs(ctx context.Context, pool *sql.DB, filter AuditFilter, limit, offset int) ([]AuditLogEntry, int, error) {
      // Implementation with pagination
  }
  ```

- [ ] **Create `internal/services/admin_service.go`**
  
  ```go
  package services
  
  type AdminService struct {
      DB  *sql.DB
      Log *logger.Logger
  }
  
  func (s *AdminService) LogAction(ctx context.Context, adminID, action, targetType string, targetID *string, details map[string]any, r *http.Request) error {
      entry := db.AuditLogEntry{
          AdminUserID: adminID,
          Action:      action,
          TargetType:  targetType,
          TargetID:    targetID,
          Details:     details,
          IPAddress:   r.RemoteAddr,
          UserAgent:   r.Header.Get("User-Agent"),
      }
      return db.InsertAuditLog(ctx, s.DB, entry)
  }
  ```

---

## Step 1.5: Admin API Routes Setup

Set up the admin route group.

### Tasks

- [ ] **Create `internal/api/routes/admin.go`**
  
  ```go
  package routes
  
  import (
      "github.com/go-chi/chi/v5"
      "github.com/onurceri/botla-co/internal/api/handlers"
      "github.com/onurceri/botla-co/pkg/middleware"
  )
  
  func RegisterAdminRoutes(r chi.Router, h *handlers.AdminHandlers) {
      r.Route("/admin", func(r chi.Router) {
          // All admin routes require platform admin
          r.Use(middleware.RequirePlatformAdmin)
          
          // Stats & Overview
          r.Get("/stats/overview", h.GetOverviewStats)
          
          // Users
          r.Get("/users", h.ListUsers)
          r.Get("/users/{id}", h.GetUser)
          r.Patch("/users/{id}", h.UpdateUser)
          
          // Organizations
          r.Get("/organizations", h.ListOrganizations)
          r.Get("/organizations/{id}", h.GetOrganization)
          
          // System Health
          r.Get("/health/detailed", h.GetDetailedHealth)
          
          // Audit Logs
          r.Get("/audit-logs", h.ListAuditLogs)
      })
  }
  ```

- [ ] **Register routes in main router** (`cmd/server/main.go` or router setup file)
  
  Add:
  ```go
  routes.RegisterAdminRoutes(r, adminHandlers)
  ```

---

## Step 1.6: Admin Overview Stats Handler

Create the main dashboard stats endpoint.

### Tasks

- [ ] **Create `internal/api/handlers/admin_stats.go`**
  
  ```go
  package handlers
  
  type AdminStatsHandlers struct {
      DB *sql.DB
  }
  
  type OverviewStats struct {
      TotalUsers         int `json:"total_users"`
      TotalOrganizations int `json:"total_organizations"`
      TotalChatbots      int `json:"total_chatbots"`
      TotalConversations int `json:"total_conversations"`
      
      UsersToday         int `json:"users_today"`
      ConversationsToday int `json:"conversations_today"`
      
      ActivePlans map[string]int `json:"active_plans"` // plan_id -> count
  }
  
  func (h *AdminStatsHandlers) GetOverviewStats(w http.ResponseWriter, r *http.Request) {
      // Query DB for platform-wide stats
      // Return OverviewStats as JSON
  }
  ```

- [ ] **Create `internal/db/admin_stats.go`**
  
  Queries for:
  - `GetPlatformUserCount(ctx, db, filter)`
  - `GetPlatformOrgCount(ctx, db)`
  - `GetPlatformChatbotCount(ctx, db)`
  - `GetPlatformConversationCount(ctx, db, since)`
  - `GetPlanDistribution(ctx, db)`

---

## Step 1.7: Admin Users Handler

Create user management endpoints.

### Tasks

- [ ] **Create `internal/api/handlers/admin_users.go`**
  
  Implement:
  - `ListUsers(w, r)` - paginated list with search/filter
  - `GetUser(w, r)` - detailed user info with related data
  - `UpdateUser(w, r)` - suspend/activate user
  
  ```go
  type UserListParams struct {
      Search  string `query:"search"`
      Status  string `query:"status"` // active, suspended, all
      Plan    string `query:"plan"`
      Page    int    `query:"page"`
      PerPage int    `query:"per_page"`
  }
  
  type UserDetail struct {
      User            models.User           `json:"user"`
      Organizations   []OrgMembership       `json:"organizations"`
      ChatbotsOwned   int                   `json:"chatbots_owned"`
      TotalTokensUsed int64                 `json:"total_tokens_used"`
      LastActive      *time.Time            `json:"last_active"`
  }
  ```

- [ ] **Create `internal/db/admin_users.go`**
  
  Queries for:
  - `ListUsersAdmin(ctx, db, params)` - with pagination
  - `GetUserDetailAdmin(ctx, db, userID)`
  - `SuspendUser(ctx, db, userID)`
  - `ActivateUser(ctx, db, userID)`

---

## Step 1.8: Admin Organizations Handler

Create organization management endpoints.

### Tasks

- [ ] **Create `internal/api/handlers/admin_orgs.go`**
  
  Implement:
  - `ListOrganizations(w, r)` - paginated list with stats
  - `GetOrganization(w, r)` - detailed org info
  
  ```go
  type OrgListItem struct {
      ID           string `json:"id"`
      Name         string `json:"name"`
      PlanID       string `json:"plan_id"`
      MemberCount  int    `json:"member_count"`
      ChatbotCount int    `json:"chatbot_count"`
      CreatedAt    time.Time `json:"created_at"`
  }
  
  type OrgDetail struct {
      Organization  models.Organization `json:"organization"`
      Members       []OrgMember         `json:"members"`
      Workspaces    []Workspace         `json:"workspaces"`
      Chatbots      []ChatbotSummary    `json:"chatbots"`
      UsageStats    UsageStats          `json:"usage_stats"`
  }
  ```

- [ ] **Create `internal/db/admin_orgs.go`**
  
  Queries for:
  - `ListOrganizationsAdmin(ctx, db, params)`
  - `GetOrganizationDetailAdmin(ctx, db, orgID)`

---

## Step 1.9: Admin Bootstrap CLI

Create CLI command to make a user admin.

### Tasks

- [ ] **Create `cmd/cli/main.go`** (if doesn't exist)
  
  ```go
  package main
  
  import (
      "flag"
      "fmt"
      "os"
  )
  
  func main() {
      if len(os.Args) < 2 {
          printUsage()
          os.Exit(1)
      }
      
      switch os.Args[1] {
      case "make-admin":
          makeAdminCmd(os.Args[2:])
      case "remove-admin":
          removeAdminCmd(os.Args[2:])
      default:
          printUsage()
          os.Exit(1)
      }
  }
  
  func makeAdminCmd(args []string) {
      fs := flag.NewFlagSet("make-admin", flag.ExitOnError)
      email := fs.String("email", "", "User email to make admin")
      fs.Parse(args)
      
      if *email == "" {
          fmt.Println("Error: --email is required")
          os.Exit(1)
      }
      
      // Connect to DB and update user
  }
  ```

- [ ] **Add Makefile target**
  
  ```makefile
  cli-make-admin:
  	go run cmd/cli/main.go make-admin --email=$(EMAIL)
  
  cli-remove-admin:
  	go run cmd/cli/main.go remove-admin --email=$(EMAIL)
  ```

---

## Verification

### Unit Tests
```bash
# Run admin middleware tests
go test ./internal/api/middleware/... -v -run Admin

# Run admin service tests
go test ./internal/services/... -v -run Admin
```

### Integration Tests
```bash
# After setting up test fixtures
make test-all
```

### Manual Verification

1. **Apply migration:**
   ```bash
   make migrate-up
   ```

2. **Create admin user:**
   ```bash
   make cli-make-admin EMAIL=your-email@example.com
   ```

3. **Verify admin column:**
   ```bash
   make psql
   SELECT email, is_platform_admin FROM users WHERE email = 'your-email@example.com';
   ```

4. **Test API access:**
   ```bash
   # Login as admin user and get token
   # Then:
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/admin/stats/overview
   ```

---

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `db/migrations/000041_admin_platform.up.sql` | CREATE | Migration for admin tables |
| `db/migrations/000041_admin_platform.down.sql` | CREATE | Down migration |
| `internal/models/user.go` | MODIFY | Add IsPlatformAdmin field |
| `internal/db/user.go` | MODIFY | Update queries |
| `internal/auth/jwt.go` | MODIFY | Add admin flag to claims |
| `internal/api/middleware/admin.go` | CREATE | Admin auth middleware |
| `internal/api/middleware/admin_test.go` | CREATE | Middleware tests |
| `internal/db/admin_audit.go` | CREATE | Audit log queries |
| `internal/db/admin_stats.go` | CREATE | Stats queries |
| `internal/db/admin_users.go` | CREATE | User admin queries |
| `internal/db/admin_orgs.go` | CREATE | Org admin queries |
| `internal/services/admin_service.go` | CREATE | Admin business logic |
| `internal/api/handlers/admin_stats.go` | CREATE | Stats handlers |
| `internal/api/handlers/admin_users.go` | CREATE | User handlers |
| `internal/api/handlers/admin_orgs.go` | CREATE | Org handlers |
| `internal/api/routes/admin.go` | CREATE | Route registration |
| `cmd/cli/main.go` | CREATE | CLI for admin management |
