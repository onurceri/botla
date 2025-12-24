# Phase 5: Testing & Documentation

> **Estimated Time:** 2-3 days  
> **Priority:** High (Week 5)  
> **Depends On:** Phases 1-4  

This phase focuses on comprehensive testing, security hardening, and documentation.

---

## Step 5.1: Backend Integration Tests

Create integration tests for admin APIs.

### Tasks

- [ ] **Create `internal/integration/admin_test.go`**
  
  ```go
  package integration
  
  func TestAdminStatsEndpoint(t *testing.T) {
      // Setup: Create test users, orgs, chatbots
      // Login as admin user
      // Call /api/v1/admin/stats/overview
      // Verify response contains correct counts
  }
  
  func TestAdminUsersList(t *testing.T) {
      // Setup: Create multiple test users
      // Login as admin
      // Test pagination
      // Test search filter
      // Test status filter
  }
  
  func TestAdminUserSuspend(t *testing.T) {
      // Create regular user
      // Login as admin
      // Suspend user via API
      // Verify user cannot login
      // Reactivate user
      // Verify user can login again
  }
  
  func TestNonAdminCannotAccessAdminEndpoints(t *testing.T) {
      // Create regular user (non-admin)
      // Login as regular user
      // Try to access /api/v1/admin/users
      // Should return 403 Forbidden
  }
  ```

- [ ] **Create `internal/integration/admin_health_test.go`**
  
  ```go
  func TestAdminHealthEndpoint(t *testing.T) {
      // Login as admin
      // Call /api/v1/admin/health/detailed
      // Verify all dependencies are checked
      // Verify response structure
  }
  ```

- [ ] **Create `internal/integration/admin_queues_test.go`**
  
  ```go
  func TestAdminQueuesEndpoint(t *testing.T) {
      // Create some pending sources
      // Call /api/v1/admin/queues
      // Verify queue counts
  }
  
  func TestAdminRetryStuckJob(t *testing.T) {
      // Create a stuck source (processing for > 30 min)
      // Call /api/v1/admin/queues/stuck
      // Verify job appears
      // Call /api/v1/admin/queues/{id}/retry
      // Verify status reset to pending
  }
  ```

- [ ] **Run integration tests**
  ```bash
  make test-all
  # Or specifically:
  go test ./internal/integration/... -v -run Admin
  ```

---

## Step 5.2: Privacy/KVKK Integration Tests

Test privacy compliance flows.

### Tasks

- [ ] **Create `internal/integration/privacy_test.go`**
  
  ```go
  func TestUserDataExportFlow(t *testing.T) {
      // Create user with data (chatbots, conversations, etc.)
      // User requests data export
      // Verify request created
      // Admin approves request
      // Verify export file generated
      // Verify export contains expected data structure
  }
  
  func TestUserDeletionFlow(t *testing.T) {
      // Create user with data
      // User requests account deletion
      // Verify request created with 'pending' status
      // Admin approves deletion
      // Verify user is soft-deleted
      // Verify user data is anonymized
      // Verify user cannot login
  }
  
  func TestConsentManagement(t *testing.T) {
      // Create user
      // Get default consents
      // Update consents
      // Verify changes persisted
      // Verify audit log created
  }
  ```

---

## Step 5.3: Unit Tests for Admin Middleware

### Tasks

- [ ] **Create/Update `internal/api/middleware/admin_test.go`**
  
  ```go
  func TestRequirePlatformAdmin_NoUser(t *testing.T) {
      // Request without user context
      // Should return 401 Unauthorized
  }
  
  func TestRequirePlatformAdmin_NonAdmin(t *testing.T) {
      // Request with regular user (IsPlatformAdmin = false)
      // Should return 403 Forbidden
  }
  
  func TestRequirePlatformAdmin_Admin(t *testing.T) {
      // Request with admin user (IsPlatformAdmin = true)
      // Should pass through to handler
  }
  ```

---

## Step 5.4: Frontend Tests

Create frontend component tests.

### Tasks

- [ ] **Create `frontend/src/pages/admin/__tests__/AdminDashboardPage.test.tsx`**
  
  ```typescript
  import { render, screen, waitFor } from '@testing-library/react';
  import { AdminDashboardPage } from '../AdminDashboardPage';
  import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
  
  // Mock API
  vi.mock('@/api/admin', () => ({
    adminApi: {
      getOverviewStats: vi.fn().mockResolvedValue({
        data: {
          total_users: 100,
          total_organizations: 25,
          total_chatbots: 50,
          total_conversations: 1000,
          users_today: 5,
          conversations_today: 50,
        }
      }),
    },
  }));
  
  describe('AdminDashboardPage', () => {
    it('displays stats cards with correct values', async () => {
      render(/* ... */);
      
      await waitFor(() => {
        expect(screen.getByText('100')).toBeInTheDocument(); // Total users
        expect(screen.getByText('25')).toBeInTheDocument();  // Orgs
      });
    });
  });
  ```

- [ ] **Create `frontend/src/features/admin/__tests__/AdminRoute.test.tsx`**
  
  ```typescript
  describe('AdminRoute', () => {
    it('redirects non-admin users to dashboard', () => {
      // Mock useAuth to return non-admin user
      // Render AdminRoute with children
      // Should navigate to /dashboard
    });
    
    it('allows admin users to access children', () => {
      // Mock useAuth to return admin user
      // Render AdminRoute with children
      // Children should be rendered
    });
  });
  ```

- [ ] **Run frontend tests**
  ```bash
  cd frontend
  npm test
  ```

---

## Step 5.5: Security Hardening

Implement additional security measures.

### Tasks

- [ ] **Add rate limiting to admin endpoints**
  
  Update `pkg/middleware/rate_limit.go`:
  ```go
  // Stricter limits for admin endpoints
  var AdminRateLimiter = rate.NewLimiter(rate.Every(time.Minute/10), 10)
  
  func AdminRateLimit(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          if !AdminRateLimiter.Allow() {
              http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
              return
          }
          next.ServeHTTP(w, r)
      })
  }
  ```

- [ ] **Apply rate limit to admin routes**
  
  ```go
  r.Route("/admin", func(r chi.Router) {
      r.Use(middleware.RequirePlatformAdmin)
      r.Use(middleware.AdminRateLimit)
      // ... routes
  })
  ```

- [ ] **Add IP allowlist (optional)**
  
  ```go
  // Environment variable: ADMIN_ALLOWED_IPS=1.2.3.4,5.6.7.8
  func AdminIPAllowlist(allowedIPs []string) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              if len(allowedIPs) == 0 {
                  next.ServeHTTP(w, r)
                  return
              }
              
              clientIP := r.RemoteAddr
              // Parse and check against allowed list
          })
      }
  }
  ```

- [ ] **Shorter session timeout for admin**
  
  Update JWT token generation to use shorter expiry for admin users:
  ```go
  expiry := 24 * time.Hour  // Normal users
  if user.IsPlatformAdmin {
      expiry = 30 * time.Minute  // Admin users
  }
  ```

---

## Step 5.6: Audit Log Verification

Ensure all admin actions are logged.

### Tasks

- [ ] **Review all admin handlers** and verify each creates an audit log entry:
  
  | Handler | Action | Should Log |
  |---------|--------|------------|
  | `UpdateUser` | suspend/activate | ✅ |
  | `UpdateOrganization` | suspend/change plan | ✅ |
  | `DeleteChatbot` | delete chatbot | ✅ |
  | `ForceRefreshChatbot` | force refresh | ✅ |
  | `ReprocessSource` | reprocess | ✅ |
  | `DeleteSource` | delete source | ✅ |
  | `RetryJob` | retry stuck job | ✅ |
  | `DeleteJob` | delete stuck job | ✅ |
  | `ProcessPrivacyRequest` | approve/deny | ✅ |
  | `GenerateUserExport` | generate export | ✅ |

- [ ] **Write test to verify audit logging**
  
  ```go
  func TestAuditLogCreated(t *testing.T) {
      // Login as admin
      // Suspend a user
      // Query audit_logs table
      // Verify entry exists with correct action, target, admin ID
  }
  ```

---

## Step 5.7: API Documentation

Document admin API endpoints.

### Tasks

- [ ] **Create `docs/admin-api.md`**
  
  ```markdown
  # Admin API Documentation
  
  ## Authentication
  
  All admin endpoints require a valid JWT token with `is_platform_admin: true`.
  
  Send the token in the Authorization header:
  ```
  Authorization: Bearer <token>
  ```
  
  ## Endpoints
  
  ### Overview Stats
  
  `GET /api/v1/admin/stats/overview`
  
  Returns platform-wide statistics.
  
  **Response:**
  ```json
  {
    "total_users": 100,
    "total_organizations": 25,
    "total_chatbots": 50,
    "total_conversations": 1000,
    "users_today": 5,
    "conversations_today": 50,
    "active_plans": {
      "free": 80,
      "pro": 15,
      "enterprise": 5
    }
  }
  ```
  
  ### Users
  
  `GET /api/v1/admin/users`
  
  List users with pagination and filters.
  
  **Query Parameters:**
  - `page` (int): Page number, default 1
  - `per_page` (int): Items per page, default 20, max 100
  - `search` (string): Search by email
  - `status` (string): Filter by status (active, suspended)
  
  <!-- ... more endpoints ... -->
  ```

- [ ] **Document all endpoints** following the structure above

---

## Step 5.8: KVKK Compliance Documentation

Document KVKK compliance features.

### Tasks

- [ ] **Create `docs/kvkk-compliance.md`**
  
  ```markdown
  # KVKK Compliance Documentation
  
  ## Overview
  
  Botla implements the following KVKK (Kişisel Verilerin Korunması Kanunu) 
  compliance features as required by Turkish data protection law.
  
  ## Data Subject Rights (Article 11)
  
  ### Right to Access (Erişim Hakkı)
  
  Users can request a copy of all their personal data.
  
  **Implementation:**
  - Users access Settings > Privacy > Export My Data
  - Request is created and optionally reviewed by admin
  - JSON export file generated and delivered via secure download link
  - Download link expires after 24 hours
  
  ### Right to Erasure (Silme Hakkı)
  
  Users can request deletion of their account and data.
  
  **Implementation:**
  - Users request deletion via Settings > Privacy > Delete Account
  - Request enters admin review queue
  - Admin approves or denies with reason
  - Upon approval, account is soft-deleted
  - Conversation data is anonymized (PII removed)
  - 30-day grace period before permanent deletion
  
  ## Consent Management
  
  ### Consent Types
  
  | Type | Description | Default |
  |------|-------------|---------|
  | Marketing | Promotional emails | Opt-out |
  | Analytics | Usage analytics | Opt-in |
  | Personalization | AI recommendations | Opt-in |
  | Third Party | Data sharing | Opt-out |
  
  ### Consent Recording
  
  All consent changes are logged with:
  - Timestamp
  - IP address
  - User agent
  - Previous and new state
  
  ## Data Retention
  
  | Data Type | Retention Period |
  |-----------|-----------------|
  | Conversations | 90 days |
  | Error logs | 30 days |
  | Audit logs | 1 year |
  | Account data | Until deletion requested |
  
  ## Admin Audit Trail
  
  All admin actions affecting user data are logged:
  - Who performed the action
  - What action was taken
  - When it occurred
  - IP address and user agent
  ```

---

## Step 5.9: Runbook for Production Issues

Create operational documentation.

### Tasks

- [ ] **Create `docs/admin-runbook.md`**
  
  ```markdown
  # Admin Dashboard Runbook
  
  ## Common Scenarios
  
  ### Stuck Scraping Jobs
  
  **Symptoms:** Sources showing "processing" for > 30 minutes
  
  **Resolution:**
  1. Navigate to Admin > Queues
  2. View stuck jobs list
  3. For each stuck job:
     - Check error message
     - If transient error (timeout, rate limit): Click Retry
     - If permanent error (404, blocked): Click Delete
  
  ### High Error Rate
  
  **Symptoms:** Error panel on dashboard shows spike
  
  **Resolution:**
  1. Navigate to Admin > Errors
  2. Filter by severity: Critical first
  3. Identify error pattern (same type, same chatbot, etc.)
  4. Check stack traces for root cause
  5. Common causes:
     - OpenAI rate limit: Wait and retry
     - Invalid API key: Check environment
     - Database connection: Check DB health
  
  ### User Reporting Issues
  
  **Resolution:**
  1. Navigate to Admin > Users
  2. Search for user by email
  3. View user details
  4. Check:
     - Account status (not suspended?)
     - Plan limits
     - Recent activity
  5. If needed: Use Impersonate feature to reproduce issue
  
  ### Processing KVKK Data Export Request
  
  1. Navigate to Admin > Privacy
  2. View pending requests
  3. For export requests:
     - Verify user identity if needed
     - Click Approve
     - System generates export
     - User receives email with download link
  
  ### Processing Account Deletion Request
  
  1. Navigate to Admin > Privacy
  2. View pending deletion requests
  3. Review user's reason
  4. Verify no outstanding issues (unpaid invoices, etc.)
  5. Click Approve (or Deny with reason)
  6. System will:
     - Soft-delete account
     - Anonymize conversation data
     - Notify user via email
  ```

---

## Verification Checklist

### Tests
```bash
# Backend unit tests
go test ./internal/... -v

# Backend integration tests
make test-all

# Frontend tests
cd frontend && npm test
```

### Security Review

- [ ] Admin middleware blocks non-admin users
- [ ] Rate limiting active on admin endpoints
- [ ] All admin actions create audit logs
- [ ] Admin session timeout is 30 minutes
- [ ] Data exports have expiring URLs
- [ ] Deleted users cannot login

### Documentation Review

- [ ] API documentation covers all endpoints
- [ ] KVKK documentation is complete
- [ ] Runbook covers common scenarios
- [ ] README updated with admin setup instructions

---

## Final Files

| File | Description |
|------|-------------|
| `internal/integration/admin_test.go` | Admin API integration tests |
| `internal/integration/admin_health_test.go` | Health endpoint tests |
| `internal/integration/admin_queues_test.go` | Queue management tests |
| `internal/integration/privacy_test.go` | Privacy flow tests |
| `internal/api/middleware/admin_test.go` | Middleware unit tests |
| `frontend/src/pages/admin/__tests__/*.test.tsx` | Frontend tests |
| `pkg/middleware/rate_limit.go` | Updated with admin rate limit |
| `docs/admin-api.md` | API documentation |
| `docs/kvkk-compliance.md` | KVKK compliance docs |
| `docs/admin-runbook.md` | Operational runbook |
