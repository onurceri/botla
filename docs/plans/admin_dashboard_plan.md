# Admin Dashboard Implementation Plan

## Overview

Build a comprehensive platform-wide admin dashboard for Botla to manage all system resources, monitor production health, tackle operational issues, and ensure KVKK (Turkey's data protection law) compliance.

## Requirements Analysis

### Core Entities to Manage

| Entity | Key Management Actions |
|--------|----------------------|
| **Users** | List, view details, suspend/activate, delete (KVKK), export data, impersonate (debug) |
| **Organizations** | List, view members/chatbots, suspend, delete, change plan, view usage |
| **Chatbots** | List, view config, delete, force-refresh sources, view conversations |
| **Data Sources** | List, view status, reprocess, delete, view errors |
| **Conversations** | View chat logs, export transcripts, view confidence scores |
| **Handoff Requests** | View, assign, resolve, analytics |
| **Actions/Tools** | View configurations, execution logs, error rates |
| **Plans** | Manage plan tiers and limits |

### Production Monitoring Requirements

| Feature | Description |
|---------|-------------|
| **Real-time Dashboards** | Live metrics for conversations, API calls, errors |
| **System Health** | Database, Redis, Qdrant, External APIs status |
| **Queue Monitoring** | Scraping queue depth, processing times, stuck jobs |
| **Error Tracking** | Error rates, stack traces, affected users |
| **Usage Analytics** | Token consumption, API calls, rate limit hits |
| **Alerting** | Email/webhook notifications for critical issues |

### KVKK Compliance Requirements

| Requirement | Description |
|-------------|-------------|
| **Data Export (Article 11)** | Users can request all their personal data in machine-readable format |
| **Right to Deletion (Article 11)** | Users can request deletion of their personal data |
| **Consent Management** | Track consent for data processing, marketing communications |
| **Audit Log** | Log all data access and deletion requests |
| **Data Retention Policies** | Configurable retention periods for conversations, logs |

---

## Implementation Plan

### Phase 1: Backend Foundation

#### 1.1 Admin Authentication
- [ ] Create admin role in database (`platform_admin` role)
- [ ] Add `is_platform_admin` flag to users table
- [ ] Create admin middleware (`internal/api/middleware/admin.go`)
- [ ] Add admin routes to router

#### 1.2 Creating the First Admin User

The first platform admin is bootstrapped via environment variable for security:

**.env configuration:**
```bash
# Platform admin email - user with this email becomes super admin
PLATFORM_ADMIN_EMAIL=admin@botla.co
```

**Bootstrap process:**
1. On application startup, check if any user has `is_platform_admin=TRUE`
2. If no admin exists, look up user by `PLATFORM_ADMIN_EMAIL`
3. If user exists, set `is_platform_admin=TRUE` for that user
4. If user doesn't exist yet, log warning to create this user first

**Alternative: CLI command for safety:**
```bash
# Make a user a platform admin
go run cmd/cli/main.go make-admin --email=user@example.com

# Remove admin privileges
go run cmd/cli/main.go remove-admin --email=user@example.com
```

**Security notes:**
- `PLATFORM_ADMIN_EMAIL` is only checked at startup, not on every request
- The CLI command requires existing admin authentication or direct DB access
- Store admin emails in `.env` which should be gitignored and not committed

#### 1.3 Admin API Endpoints

**Users Management**
```
GET    /api/v1/admin/users                    - List users with pagination/search
GET    /api/v1/admin/users/{id}              - Get user details with all related data
PATCH  /api/v1/admin/users/{id}              - Update user (suspend, activate, change plan)
DELETE /api/v1/admin/users/{id}              - Soft-delete (KVKK)
POST   /api/v1/admin/users/{id}/export       - Export user data
GET    /api/v1/admin/users/{id}/activity     - Activity log
POST   /api/v1/admin/users/{id}/impersonate  - Generate impersonation token (for debugging)
```

**Organizations Management**
```
GET    /api/v1/admin/organizations           - List organizations with stats
GET    /api/v1/admin/organizations/{id}      - Details with members, chatbots, usage
PATCH  /api/v1/admin/organizations/{id}      - Suspend/unsuspend, change plan
DELETE /api/v1/admin/organizations/{id}      - Delete org (cascade)
GET    /api/v1/admin/organizations/{id}/usage - Detailed usage breakdown
```

**Chatbots Management**
```
GET    /api/v1/admin/chatbots                - List all chatbots across platform
GET    /api/v1/admin/chatbots/{id}           - Full chatbot details
GET    /api/v1/admin/chatbots/{id}/conversations - View conversations
PATCH  /api/v1/admin/chatbots/{id}           - Suspend/unsuspend
DELETE /api/v1/admin/chatbots/{id}           - Delete chatbot
POST   /api/v1/admin/chatbots/{id}/force-refresh - Force refresh all sources
```

**Data Sources Management**
```
GET    /api/v1/admin/sources                 - List all sources with status
GET    /api/v1/admin/sources/{id}            - Source details with error logs
POST   /api/v1/admin/sources/{id}/reprocess  - Reprocess source
DELETE /api/v1/admin/sources/{id}            - Delete source
GET    /api/v1/admin/sources/failed          - List failed sources
```

**Platform Analytics**
```
GET    /api/v1/admin/stats/overview          - Dashboard stats (totals, today, growth)
GET    /api/v1/admin/stats/users             - User growth and activity
GET    /api/v1/admin/stats/organizations     - Org growth and plan distribution
GET    /api/v1/admin/stats/usage             - Token/usage metrics
GET    /api/v1/admin/stats/chatbots          - Chatbot creation and activity
GET    /api/v1/admin/stats/conversations     - Conversation volume and trends
GET    /api/v1/admin/stats/errors            - Error rates and types
```

**System Health & Monitoring**
```
GET    /api/v1/admin/health/detailed         - Comprehensive health check
GET    /api/v1/admin/health/dependencies     - All external dependencies status
GET    /api/v1/admin/queues                  - Queue depths and processing stats
GET    /api/v1/admin/queues/stuck            - Stuck/failed jobs
POST   /api/v1/admin/queues/{id}/retry       - Retry stuck job
DELETE /api/v1/admin/queues/{id}             - Delete stuck job
GET    /api/v1/admin/errors                  - Recent error log (paginated)
GET    /api/v1/admin/errors/{id}             - Error details with stack trace
```

**KVKK/Privacy Compliance**
```
GET    /api/v1/admin/privacy/requests        - List KVKK requests
GET    /api/v1/admin/privacy/requests/{id}   - Request details
PATCH  /api/v1/admin/privacy/requests/{id}   - Process request (approve/deny)
POST   /api/v1/admin/privacy/export/{userId} - Generate data export
GET    /api/v1/admin/audit-logs              - Admin action audit trail
```

**Actions/Tools Management**
```
GET    /api/v1/admin/actions                 - List all custom actions
GET    /api/v1/admin/actions/{id}/logs       - Execution logs for action
GET    /api/v1/admin/actions/failed          - Failed action executions
```

#### 1.4 Privacy Service (KVKK)
- [ ] Create `internal/services/privacy_service.go`
- [ ] Implement data export (JSON format with all user data)
- [ ] Implement cascading deletion (respecting FK constraints)
- [ ] Create audit logging for data access

#### 1.5 Admin Service
- [ ] Create `internal/services/admin_service.go`
- [ ] Implement platform-wide statistics queries
- [ ] Implement user impersonation token generation
- [ ] Create queue management functions

---

### Phase 2: Production Monitoring Backend

#### 2.1 Health Check Enhancements
- [ ] Extend existing `health.go` handler for detailed checks
- [ ] Add Redis health check
- [ ] Add OpenAI API health check (ping endpoint)
- [ ] Add S3/R2 storage health check
- [ ] Create aggregated status endpoint

#### 2.2 Queue Monitoring
- [ ] Create `internal/db/admin_queue.go` for queue queries
- [ ] Track pending source refreshes
- [ ] Track stuck scraping jobs (not completed in X minutes)
- [ ] Add retry mechanism for stuck jobs

#### 2.3 Error Tracking
- [ ] Create `error_logs` table for persistent error storage
- [ ] Add error logging middleware for critical failures
- [ ] Store stack traces, user context, request details
- [ ] Add error aggregation queries (group by type, time)

#### 2.4 Usage Metrics
- [ ] Create `platform_metrics` table for time-series data
- [ ] Track: API calls, tokens used, rate limit hits, errors
- [ ] Add background job to aggregate metrics hourly

---

### Phase 3: Frontend Admin Dashboard

#### 3.1 Admin Authentication on Frontend

**Route Protection:**
```typescript
// frontend/src/features/admin/AdminRoute.tsx
function AdminRoute({ children }: { children: React.ReactNode }) {
  const { user } = useAuth()

  if (!user?.is_platform_admin) {
    return <Navigate to="/dashboard" replace />
  }

  return <>{children}</>
}
```

**Protected Admin Routes:**
```typescript
// frontend/src/App.tsx
<Route path="/admin" element={
  <AdminRoute>
    <AdminLayout />
  </AdminRoute>
}>
  <Route index element={<AdminDashboardPage />} />
  <Route path="users" element={<AdminUsersPage />} />
  <Route path="users/:id" element={<AdminUserDetailPage />} />
  <Route path="organizations" element={<AdminOrganizationsPage />} />
  <Route path="organizations/:id" element={<AdminOrgDetailPage />} />
  <Route path="chatbots" element={<AdminChatbotsPage />} />
  <Route path="chatbots/:id" element={<AdminChatbotDetailPage />} />
  <Route path="sources" element={<AdminSourcesPage />} />
  <Route path="privacy" element={<AdminPrivacyPage />} />
  <Route path="system" element={<AdminSystemPage />} />
  <Route path="errors" element={<AdminErrorsPage />} />
  <Route path="queues" element={<AdminQueuesPage />} />
  <Route path="audit" element={<AdminAuditPage />} />
</Route>
```

**API Client Updates:**
```typescript
// frontend/src/api/admin.ts
export const adminClient = {
  users: {
    list: (params: UserFilter) => get<PaginatedResponse<User>>('/admin/users', { params }),
    get: (id: string) => get<UserDetail>(`/admin/users/${id}`),
    suspend: (id: string) => patch(`/admin/users/${id}`, { status: 'suspended' }),
    activate: (id: string) => patch(`/admin/users/${id}`, { status: 'active' }),
    export: (id: string) => post(`/admin/users/${id}/export`),
    impersonate: (id: string) => post<{ token: string }>(`/admin/users/${id}/impersonate`),
  },
  organizations: {
    list: (params: OrgFilter) => get<PaginatedResponse<Org>>('/admin/organizations', { params }),
    get: (id: string) => get<OrgDetail>(`/admin/organizations/${id}`),
    // ... etc
  },
  stats: {
    overview: () => get<PlatformStats>('/admin/stats/overview'),
    users: (params: DateRange) => get('/admin/stats/users', { params }),
    errors: (params: DateRange) => get('/admin/stats/errors', { params }),
  },
  health: {
    detailed: () => get<DetailedHealth>('/admin/health/detailed'),
    dependencies: () => get<DependencyStatus[]>('/admin/health/dependencies'),
  },
  queues: {
    list: () => get<QueueStatus[]>('/admin/queues'),
    stuck: () => get<StuckJob[]>('/admin/queues/stuck'),
    retry: (id: string) => post(`/admin/queues/${id}/retry`),
    delete: (id: string) => del(`/admin/queues/${id}`),
  },
  // ... etc
}
```

#### 3.2 Admin Layout
- [ ] Create `frontend/src/pages/admin/AdminLayout.tsx`
- [ ] Create `frontend/src/features/admin/` directory
- [ ] Add admin navigation sidebar with sections
- [ ] Create admin-only route wrapper
- [ ] Add persistent refresh interval selector

#### 3.3 Dashboard Pages

**Overview Dashboard (Priority: High)**
- [ ] Stats cards (users, orgs, chatbots, conversations - today/total/growth)
- [ ] Usage charts (tokens, messages over time)
- [ ] Active plans distribution pie chart
- [ ] System health status panel (green/yellow/red)
- [ ] Recent errors panel
- [ ] Queue depths panel

**Users Page (Priority: High)**
- [ ] User table with search, filters (active, suspended, plan type)
- [ ] Sortable columns (created_at, last_active, plan, usage)
- [ ] Bulk actions (suspend, export)
- [ ] Quick actions dropdown (view, suspend, export, impersonate)

**User Detail Page**
- [ ] User profile section (email, name, plan, created)
- [ ] Organizations membership tab
- [ ] Chatbots owned tab
- [ ] Activity log tab (recent actions)
- [ ] Usage stats tab (tokens, messages)
- [ ] Action buttons (suspend, export data, delete)
- [ ] Impersonate button (opens new tab with user session)

**Organizations Page**
- [ ] Org table with search, plan filter, member count
- [ ] Sortable by usage, member count, chatbot count
- [ ] Quick actions (view, suspend, change plan)

**Organization Detail Page**
- [ ] Org info section
- [ ] Members table with roles
- [ ] Workspaces list
- [ ] Chatbots list with status
- [ ] Usage breakdown chart
- [ ] Action buttons (suspend, delete)

**Chatbots Page**
- [ ] All chatbots view across all orgs
- [ ] Status indicators (active, paused, error, suspended)
- [ ] Message/usage stats columns
- [ ] Filter by status, org, plan
- [ ] Quick actions (view, suspend, force-refresh, delete)

**Chatbot Detail Page**
- [ ] Configuration overview
- [ ] Data sources list with status
- [ ] Recent conversations (with confidence scores)
- [ ] Analytics overview
- [ ] Error log (failed responses)
- [ ] Actions configuration
- [ ] Force refresh button

**Data Sources Page**
- [ ] All sources across platform
- [ ] Status filter (processing, ready, failed, stale)
- [ ] Last refresh time
- [ ] Quick actions (reprocess, delete)
- [ ] Bulk reprocess for failed

**Privacy Requests Page (KVKK)**
- [ ] Request queue table (pending, processing, completed, denied)
- [ ] Request detail modal
- [ ] Process action (approve/deny with reason)
- [ ] Export download link
- [ ] Export status tracking

**System Health Page (Priority: High)**
- [ ] Database connection status and latency
- [ ] Redis connection status and latency
- [ ] Qdrant connection status and latency
- [ ] OpenAI API status and latency
- [ ] S3/R2 storage status
- [ ] Environment info (version, deploy time)
- [ ] Auto-refresh toggle (5s/30s/1m/off)

**Queues Page (Priority: High)**
- [ ] Queue overview (source refresh, scraping)
- [ ] Current depth and processing rate
- [ ] Stuck jobs list (not completed in X minutes)
- [ ] Retry/delete actions for stuck jobs
- [ ] Historical chart (queue depth over time)

**Errors Page (Priority: High)**
- [ ] Recent errors list (paginated)
- [ ] Filter by type, severity, date range
- [ ] Error detail view (stack trace, context)
- [ ] Group by error type chart
- [ ] Affected user/chatbot links

**Audit Log Page**
- [ ] All admin actions (user suspended, export generated, etc.)
- [ ] Filter by admin, action type, date range
- [ ] Export audit log

---

### Phase 4: KVKK Compliance Features

#### 4.1 User-Facing KVKK Portal
- [ ] Add `/account/privacy` page in frontend
- [ ] Export my data button (generates download)
- [ ] Delete my account button (with confirmation)
- [ ] Consent management toggles

#### 4.2 Consent Tracking
- [ ] Add `consents` table (user_id, type, granted_at, revoked_at)
- [ ] Consent types: marketing, analytics, third_party
- [ ] Backend consent endpoints
- [ ] Consent banner integration

#### 4.3 Data Retention
- [ ] Add retention policy config to plans
- [ ] Background job for auto-deletion (`internal/services/retention_job.go`)
- [ ] Configurable: conversations (default 90 days), logs (30 days)
- [ ] Notify before deletion (optional)

---

### Phase 5: Testing & Documentation

#### 5.1 Security Implementation

**Middleware Implementation:**
```go
// internal/api/middleware/admin.go
func RequirePlatformAdmin(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    user := auth.UserFromContext(r.Context())
    if user == nil || !user.IsPlatformAdmin {
      http.Error(w, "Forbidden", http.StatusForbidden)
      return
    }
    next.ServeHTTP(w, r)
  })
}
```

**Rate Limiting:**
```go
// Admin endpoints should use stricter rate limits
// e.g., 10 requests/minute instead of 100
```

**Audit Logging:**
```go
// Every admin action logged
adminSvc.LogAction(r.Context(), adminID, "suspend_user", "user", userID, map[string]any{
  "reason": reason,
  "ip":     r.RemoteAddr,
})
```

#### 5.2 Security Checklist
- [ ] Admin endpoint integration tests
- [ ] KVKK flow tests (export + delete)
- [ ] Admin API documentation
- [ ] KVKK compliance documentation
- [ ] Rate limiting on admin endpoints
- [ ] Audit logging for all admin actions
- [ ] Admin session timeout (shorter than normal: 30 min)
- [ ] IP allowlist for admin routes (optional, env configurable)
- [ ] 2FA requirement for admin accounts (future enhancement)

---

## File Structure

```
internal/
├── api/
│   ├── handlers/
│   │   ├── admin.go              # Admin platform handlers
│   │   ├── admin_users.go        # User management
│   │   ├── admin_orgs.go         # Organization management
│   │   ├── admin_chatbots.go     # Chatbot management
│   │   ├── admin_sources.go      # Source management
│   │   ├── admin_stats.go        # Platform statistics
│   │   ├── admin_health.go       # Detailed health checks
│   │   ├── admin_queues.go       # Queue management
│   │   ├── admin_errors.go       # Error log viewing
│   │   └── privacy.go            # KVKK/Privacy handler
│   ├── middleware/
│   │   └── admin.go              # Admin auth middleware
│   └── routes/
│       └── admin.go              # Admin route registration
├── services/
│   ├── admin_service.go          # Business logic for admin ops
│   ├── privacy_service.go        # KVKK data export/deletion
│   ├── retention_job.go          # Auto-cleanup job
│   └── error_logger.go           # Error persistence service
├── db/
│   ├── admin_users.go            # Admin user queries
│   ├── admin_stats.go            # Platform statistics queries
│   ├── admin_queue.go            # Queue monitoring queries
│   └── admin_audit.go            # Audit log queries

frontend/src/
├── pages/
│   └── admin/
│       ├── AdminLayout.tsx       # Admin shell layout
│       ├── AdminDashboardPage.tsx
│       ├── AdminUsersPage.tsx
│       ├── AdminUserDetailPage.tsx
│       ├── AdminOrganizationsPage.tsx
│       ├── AdminOrgDetailPage.tsx
│       ├── AdminChatbotsPage.tsx
│       ├── AdminChatbotDetailPage.tsx
│       ├── AdminSourcesPage.tsx
│       ├── AdminPrivacyPage.tsx
│       ├── AdminSystemPage.tsx
│       ├── AdminQueuesPage.tsx
│       ├── AdminErrorsPage.tsx
│       └── AdminAuditPage.tsx
├── features/admin/
│   ├── components/
│   │   ├── AdminSidebar.tsx
│   │   ├── StatsCard.tsx
│   │   ├── HealthPanel.tsx
│   │   ├── ErrorsPanel.tsx
│   │   ├── QueuesPanel.tsx
│   │   ├── UserTable.tsx
│   │   ├── OrgTable.tsx
│   │   ├── ChatbotTable.tsx
│   │   └── AuditLogTable.tsx
│   ├── hooks/
│   │   ├── useAdminStats.ts
│   │   ├── useAdminUsers.ts
│   │   ├── useAdminHealth.ts
│   │   └── useAdminQueues.ts
│   └── types.ts
└── api/
    └── admin.ts                  # Admin API client
```

---

## Database Schema Additions

```sql
-- Migration: 000041_admin_platform.up.sql

-- Admin role flag
ALTER TABLE users ADD COLUMN is_platform_admin BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_users_is_platform_admin ON users(is_platform_admin) WHERE is_platform_admin = TRUE;

-- Privacy/KVKK requests
CREATE TABLE privacy_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    request_type TEXT NOT NULL CHECK (request_type IN ('export', 'deletion', 'correction')),
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'denied')),
    reason TEXT,
    denial_reason TEXT,
    processed_by UUID REFERENCES users(id),
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    export_url TEXT,
    export_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_privacy_requests_user ON privacy_requests(user_id);
CREATE INDEX idx_privacy_requests_status ON privacy_requests(status);

-- Consent tracking
CREATE TABLE user_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    consent_type TEXT NOT NULL CHECK (consent_type IN ('marketing', 'analytics', 'personalization', 'third_party')),
    granted BOOLEAN DEFAULT TRUE,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    ip_address INET,
    user_agent TEXT
);

CREATE UNIQUE INDEX idx_user_consents_unique ON user_consents(user_id, consent_type);

-- Data export logs
CREATE TABLE data_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    requested_by UUID REFERENCES users(id),
    format TEXT NOT NULL CHECK (format IN ('json', 'csv')),
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    download_url TEXT,
    file_size_bytes BIGINT,
    expires_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

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

---

## Security Considerations

1. **Admin Access**: Only users with `is_platform_admin=TRUE` can access admin endpoints
2. **Audit Trail**: All admin actions logged to `admin_audit_logs` with IP, user agent
3. **Data Export**: Generate secure, expiring download URLs (S3 presigned or temporary files)
4. **Deletion**: Implement soft-delete with 30-day grace period for KVKK
5. **Rate Limiting**: Strict rate limits on admin endpoints (10 req/min)
6. **Session Timeout**: Admin sessions expire after 30 minutes of inactivity
7. **Impersonation**: All impersonation sessions are logged, limited to 1 hour

---

## KVKK-Specific Features

| Feature | Implementation |
|---------|----------------|
| Right to Access (Article 11) | Data export in JSON format (all user data) |
| Right to Rectification | User profile edit + admin correction endpoint |
| Right to Erasure | Cascading soft-delete with 30-day grace period |
| Right to Restrict Processing | Suspend user/account functionality |
| Right to Data Portability | Standard JSON export format |
| Consent Management | Granular consent toggles per type |
| Privacy Policy | Turkish localized policy page |
| Data Protection Officer | Admin notification workflow |

---

## Priority Order for Implementation

### Week 1: Foundation (Critical for Production Issues)
1. Admin authentication (middleware, database column)
2. System health page (all dependencies)
3. Queue monitoring (stuck jobs, retry)
4. Error logging and viewing
5. Basic overview dashboard

### Week 2: User & Org Management
1. Users list and detail pages
2. Organizations list and detail pages
3. Suspend/activate functionality
4. User impersonation for debugging

### Week 3: Chatbot & Source Management
1. Chatbots list and detail pages
2. Sources list with status
3. Force refresh functionality
4. Conversation viewing

### Week 4: KVKK & Analytics
1. Privacy request handling
2. Data export generation
3. Audit logging
4. Platform analytics charts

### Week 5: Polish & Testing
1. Integration tests
2. Error handling improvements
3. Documentation
4. Security hardening

---

## Production Issue Tackling Features

The admin dashboard specifically addresses these common production issues:

| Issue | Solution |
|-------|----------|
| **Stuck scraping jobs** | Queue monitoring with retry/delete actions |
| **Failed source refreshes** | Failed sources list with reprocess button |
| **High error rates** | Error log with grouping, stack traces |
| **Dependency outages** | Real-time health panel with status |
| **User complaints** | User impersonation to reproduce issues |
| **Usage spikes** | Usage analytics with time-series graphs |
| **KVKK data requests** | Automated data export workflow |
| **Suspicious activity** | Audit log for all admin actions |
| **Chatbot misbehavior** | Conversation viewer with confidence scores |
| **Plan limit issues** | Usage breakdown per org/user |

---

## Next Steps

1. ✅ Review and approve this plan
2. Begin Phase 1: Backend foundation
   - Create database migration for admin tables
   - Implement admin middleware
   - Create admin API handlers
3. Proceed to frontend implementation
4. Add tests and documentation
