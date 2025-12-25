# Phase 4: KVKK Compliance Features

> **Estimated Time:** 3-4 days  
> **Priority:** Medium (Week 4)  
> **Depends On:** Phase 1, Phase 3  

This phase implements KVKK (Turkey's GDPR equivalent) compliance features including data export, deletion requests, and consent management.

---

## Step 4.1: Privacy Database Tables

Add tables for privacy requests and consent tracking.

### Tasks

- [x] **Create migration** `db/migrations/000042_kvkk_compliance.up.sql`
  
  ```sql
  -- Privacy/KVKK data requests
  CREATE TABLE privacy_requests (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id UUID REFERENCES users(id) ON DELETE SET NULL,
      user_email TEXT NOT NULL,  -- Store email in case user is deleted
      request_type TEXT NOT NULL CHECK (request_type IN ('export', 'deletion', 'correction')),
      status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'denied')),
      reason TEXT,  -- User's reason for request
      denial_reason TEXT,  -- Admin's reason for denial
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
  CREATE INDEX idx_privacy_requests_created ON privacy_requests(created_at DESC);
  
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
  
  -- Data exports
  CREATE TABLE data_exports (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id UUID REFERENCES users(id) ON DELETE SET NULL,
      requested_by UUID REFERENCES users(id),  -- Admin or user themselves
      format TEXT NOT NULL CHECK (format IN ('json', 'csv')),
      status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
      download_url TEXT,
      file_size_bytes BIGINT,
      expires_at TIMESTAMPTZ,
      error_message TEXT,
      created_at TIMESTAMPTZ DEFAULT NOW(),
      completed_at TIMESTAMPTZ
  );
  
  CREATE INDEX idx_data_exports_user ON data_exports(user_id);
  CREATE INDEX idx_data_exports_status ON data_exports(status);
  ```

- [x] **Create down migration** `db/migrations/000042_kvkk_compliance.down.sql`
  
  ```sql
  DROP TABLE IF EXISTS data_exports;
  DROP TABLE IF EXISTS user_consents;
  DROP TABLE IF EXISTS privacy_requests;
  ```

- [x] **Run migration**
  ```bash
  make migrate-up
  ```

---

## Step 4.2: Privacy Service

Create service to handle data export and deletion.

### Tasks

- [x] **Create `internal/services/privacy_service.go`**
  
  ```go
  package services
  
  type PrivacyService struct {
      DB       *sql.DB
      Log      *logger.Logger
      Storage  StorageClient  // S3/R2 client
  }
  
  // ExportUserData generates a JSON export of all user data
  func (s *PrivacyService) ExportUserData(ctx context.Context, userID string) (*DataExport, error) {
      // 1. Collect all user data:
      //    - User profile
      //    - Organizations & memberships
      //    - Chatbots owned
      //    - Conversations
      //    - Consents
      //    - Activity logs
      
      // 2. Generate JSON file
      
      // 3. Upload to storage with expiring URL
      
      // 4. Create data_exports record
      
      // 5. Return export info
  }
  
  type UserDataExport struct {
      User          UserExportData         `json:"user"`
      Organizations []OrgExportData        `json:"organizations"`
      Chatbots      []ChatbotExportData    `json:"chatbots"`
      Conversations []ConversationExport   `json:"conversations"`
      Consents      []ConsentExportData    `json:"consents"`
      ExportedAt    time.Time              `json:"exported_at"`
  }
  
  // RequestDeletion initiates user data deletion process
  func (s *PrivacyService) RequestDeletion(ctx context.Context, userID, reason string) (*PrivacyRequest, error) {
      // Create pending deletion request
      // Admin will need to approve
  }
  
  // ProcessDeletion performs the actual deletion (admin-initiated)
  func (s *PrivacyService) ProcessDeletion(ctx context.Context, requestID, adminID string) error {
      // 1. Get request and user ID
      // 2. Soft-delete user (set deleted_at)
      // 3. Anonymize conversation data (keep for analytics, remove PII)
      // 4. Delete personal files from storage
      // 5. Update request status to completed
      // 6. Log admin action
  }
  ```

- [x] **Create `internal/db/privacy.go`**
  
  ```go
  package db
  
  type PrivacyRequest struct {
      ID            string     `json:"id"`
      UserID        *string    `json:"user_id"`
      UserEmail     string     `json:"user_email"`
      RequestType   string     `json:"request_type"`
      Status        string     `json:"status"`
      Reason        string     `json:"reason,omitempty"`
      DenialReason  string     `json:"denial_reason,omitempty"`
      ProcessedBy   *string    `json:"processed_by,omitempty"`
      ProcessedAt   *time.Time `json:"processed_at,omitempty"`
      CompletedAt   *time.Time `json:"completed_at,omitempty"`
      CreatedAt     time.Time  `json:"created_at"`
  }
  
  func CreatePrivacyRequest(ctx context.Context, pool *sql.DB, req PrivacyRequest) (*PrivacyRequest, error)
  func GetPrivacyRequest(ctx context.Context, pool *sql.DB, id string) (*PrivacyRequest, error)
  func ListPrivacyRequests(ctx context.Context, pool *sql.DB, status string, limit, offset int) ([]PrivacyRequest, int, error)
  func UpdatePrivacyRequestStatus(ctx context.Context, pool *sql.DB, id, status, adminID string, denialReason *string) error
  
  func GetUserDataForExport(ctx context.Context, pool *sql.DB, userID string) (*UserDataExport, error)
  func AnonymizeUserData(ctx context.Context, pool *sql.DB, userID string) error
  ```

---

## Step 4.3: Privacy Admin Handlers

Admin endpoints for processing privacy requests.

### Tasks

- [x] **Create `internal/api/handlers/privacy.go`**
  
  ```go
  package handlers
  
  type PrivacyHandlers struct {
      DB             *sql.DB
      PrivacyService *services.PrivacyService
      AdminService   *services.AdminService
  }
  
  // ListPrivacyRequests returns pending/processed KVKK requests
  func (h *PrivacyHandlers) ListPrivacyRequests(w http.ResponseWriter, r *http.Request) {
      status := r.URL.Query().Get("status")  // pending, processing, completed, denied
      // Paginated list of requests
  }
  
  // GetPrivacyRequest returns details of a specific request
  func (h *PrivacyHandlers) GetPrivacyRequest(w http.ResponseWriter, r *http.Request) {
      id := chi.URLParam(r, "id")
      // Return request with user details
  }
  
  // ProcessPrivacyRequest approves or denies a request
  func (h *PrivacyHandlers) ProcessPrivacyRequest(w http.ResponseWriter, r *http.Request) {
      var req struct {
          Action       string `json:"action"` // "approve" or "deny"
          DenialReason string `json:"denial_reason,omitempty"`
      }
      
      // Validate action
      // Process based on request type (export, deletion)
      // Log admin action
  }
  
  // GenerateUserExport creates a data export for a user (admin-initiated)
  func (h *PrivacyHandlers) GenerateUserExport(w http.ResponseWriter, r *http.Request) {
      userID := chi.URLParam(r, "userId")
      // Generate export
      // Return download URL
  }
  ```

- [x] **Add privacy routes** (update `internal/api/routes/admin.go`)
  
  ```go
  // KVKK/Privacy
  r.Get("/privacy/requests", h.ListPrivacyRequests)
  r.Get("/privacy/requests/{id}", h.GetPrivacyRequest)
  r.Patch("/privacy/requests/{id}", h.ProcessPrivacyRequest)
  r.Post("/privacy/export/{userId}", h.GenerateUserExport)
  ```

---

## Step 4.4: User-Facing Privacy Endpoints

Allow users to request their own data export/deletion.

### Tasks

- [x] **Create `internal/api/handlers/user_privacy.go`**
  
  ```go
  package handlers
  
  type UserPrivacyHandlers struct {
      DB             *sql.DB
      PrivacyService *services.PrivacyService
  }
  
  // GetMyConsents returns user's current consent settings
  func (h *UserPrivacyHandlers) GetMyConsents(w http.ResponseWriter, r *http.Request) {
      userID := middleware.UserIDFromContext(r.Context())
      // Return consent settings
  }
  
  // UpdateMyConsents updates user's consent settings
  func (h *UserPrivacyHandlers) UpdateMyConsents(w http.ResponseWriter, r *http.Request) {
      var req struct {
          Marketing       *bool `json:"marketing"`
          Analytics       *bool `json:"analytics"`
          Personalization *bool `json:"personalization"`
          ThirdParty      *bool `json:"third_party"`
      }
      // Update each changed consent
  }
  
  // RequestMyDataExport creates a data export request
  func (h *UserPrivacyHandlers) RequestMyDataExport(w http.ResponseWriter, r *http.Request) {
      userID := middleware.UserIDFromContext(r.Context())
      // Create export request
      // Optionally process immediately for small accounts
  }
  
  // RequestAccountDeletion creates a deletion request
  func (h *UserPrivacyHandlers) RequestAccountDeletion(w http.ResponseWriter, r *http.Request) {
      var req struct {
          Reason string `json:"reason"`
      }
      // Create pending deletion request
      // Notify admins
  }
  ```

- [x] **Add user privacy routes** (update regular user routes)
  
  ```go
  // User privacy settings
  r.Get("/me/privacy/consents", h.GetMyConsents)
  r.Patch("/me/privacy/consents", h.UpdateMyConsents)
  r.Post("/me/privacy/export", h.RequestMyDataExport)
  r.Post("/me/privacy/delete-account", h.RequestAccountDeletion)
  ```

---

## Step 4.5: Consent Tracking Database Functions

### Tasks

- [x] **Create `internal/db/consent.go`**
  
  ```go
  package db
  
  type UserConsent struct {
      ID          string     `json:"id"`
      UserID      string     `json:"user_id"`
      ConsentType string     `json:"consent_type"`
      Granted     bool       `json:"granted"`
      GrantedAt   time.Time  `json:"granted_at"`
      RevokedAt   *time.Time `json:"revoked_at,omitempty"`
  }
  
  func GetUserConsents(ctx context.Context, pool *sql.DB, userID string) ([]UserConsent, error)
  
  func UpsertConsent(ctx context.Context, pool *sql.DB, userID, consentType string, granted bool, ip, userAgent string) error {
      // INSERT ON CONFLICT UPDATE
  }
  ```

---

## Step 4.6: Frontend Privacy Page (User)

User-facing privacy settings page.

### Tasks

- [x] **Create `frontend/src/pages/PrivacySettingsPage.tsx`**
  
  ```typescript
  import { useQuery, useMutation } from '@tanstack/react-query';
  import { useState } from 'react';
  
  export function PrivacySettingsPage() {
    const { data: consents } = useQuery({
      queryKey: ['privacy', 'consents'],
      queryFn: () => api.get('/me/privacy/consents'),
    });
    
    const updateMutation = useMutation({
      mutationFn: (data: any) => api.patch('/me/privacy/consents', data),
    });
    
    const exportMutation = useMutation({
      mutationFn: () => api.post('/me/privacy/export'),
    });
    
    const deleteMutation = useMutation({
      mutationFn: (reason: string) => api.post('/me/privacy/delete-account', { reason }),
    });
    
    return (
      <div className="max-w-2xl mx-auto space-y-8">
        <h1 className="text-2xl font-bold">Privacy Settings</h1>
        
        {/* Consent Toggles */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6">
          <h2 className="font-semibold mb-4">Data Processing Consent</h2>
          <div className="space-y-4">
            {/* Toggle for each consent type */}
          </div>
        </section>
        
        {/* Export Data */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6">
          <h2 className="font-semibold mb-2">Export My Data</h2>
          <p className="text-sm text-gray-500 mb-4">
            Download all your personal data in JSON format.
          </p>
          <button
            onClick={() => exportMutation.mutate()}
            disabled={exportMutation.isPending}
            className="px-4 py-2 bg-primary text-white rounded-lg"
          >
            {exportMutation.isPending ? 'Processing...' : 'Request Data Export'}
          </button>
        </section>
        
        {/* Delete Account */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6 border-2 border-red-200">
          <h2 className="font-semibold mb-2 text-red-600">Delete My Account</h2>
          <p className="text-sm text-gray-500 mb-4">
            Request permanent deletion of your account and all data.
            This action cannot be undone.
          </p>
          <button
            onClick={() => {/* Show confirmation modal */}}
            className="px-4 py-2 bg-red-600 text-white rounded-lg"
          >
            Request Account Deletion
          </button>
        </section>
      </div>
    );
  }
  ```

- [x] **Add route** in App.tsx
  
  ```typescript
  <Route path="/settings/privacy" element={<PrivacySettingsPage />} />
  ```

---

## Step 4.7: Admin Privacy Requests Page

Admin page to manage KVKK requests.

### Tasks

- [x] **Create `frontend/src/pages/admin/AdminPrivacyPage.tsx`**

  ```typescript
  import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
  import { adminApi } from '@/api/admin';
  
  export function AdminPrivacyPage() {
    const [status, setStatus] = useState('pending');
    const queryClient = useQueryClient();
    
    const { data } = useQuery({
      queryKey: ['admin', 'privacy', { status }],
      queryFn: () => adminApi.listPrivacyRequests({ status }),
    });
    
    const processMutation = useMutation({
      mutationFn: ({ id, action, reason }: { id: string; action: string; reason?: string }) =>
        adminApi.processPrivacyRequest(id, action, reason),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'privacy'] }),
    });
    
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">KVKK Privacy Requests</h1>
        
        {/* Status Tabs */}
        <div className="flex gap-2">
          {['pending', 'processing', 'completed', 'denied'].map(s => (
            <button
              key={s}
              onClick={() => setStatus(s)}
              className={`px-4 py-2 rounded-lg ${
                status === s ? 'bg-primary text-white' : 'bg-gray-100'
              }`}
            >
              {s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>
        
        {/* Requests Table */}
        <div className="bg-white dark:bg-gray-800 rounded-lg">
          <table className="w-full">
            <thead>
              <tr className="text-left border-b">
                <th className="p-4">User</th>
                <th className="p-4">Type</th>
                <th className="p-4">Reason</th>
                <th className="p-4">Date</th>
                <th className="p-4">Actions</th>
              </tr>
            </thead>
            <tbody>
              {data?.data.data.map(req => (
                <tr key={req.id} className="border-b">
                  <td className="p-4">{req.user_email}</td>
                  <td className="p-4">
                    <span className="px-2 py-1 rounded bg-blue-100 text-blue-800 text-xs">
                      {req.request_type}
                    </span>
                  </td>
                  <td className="p-4 text-sm">{req.reason || '-'}</td>
                  <td className="p-4 text-sm">
                    {new Date(req.created_at).toLocaleDateString()}
                  </td>
                  <td className="p-4">
                    {req.status === 'pending' && (
                      <div className="flex gap-2">
                        <button
                          onClick={() => processMutation.mutate({ id: req.id, action: 'approve' })}
                          className="px-3 py-1 bg-green-500 text-white rounded text-sm"
                        >
                          Approve
                        </button>
                        <button
                          onClick={() => {/* Show denial modal */}}
                          className="px-3 py-1 bg-red-500 text-white rounded text-sm"
                        >
                          Deny
                        </button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }
  ```

---

## Step 4.8: Data Retention Job

Background job for automatic data cleanup.

### Tasks

- [x] **Create `internal/services/retention_job.go`**

  ```go
  package services
  
  type RetentionConfig struct {
      ConversationRetentionDays int  // Default: 90
      ErrorLogRetentionDays     int  // Default: 30
      AuditLogRetentionDays     int  // Default: 365
  }
  
  type RetentionJob struct {
      DB     *sql.DB
      Log    *logger.Logger
      Config RetentionConfig
  }
  
  // Run executes the retention cleanup
  func (j *RetentionJob) Run(ctx context.Context) error {
      // 1. Delete old conversations
      count, err := j.cleanConversations(ctx)
      j.Log.Info("cleaned conversations", "count", count)
      
      // 2. Delete old error logs
      count, err = j.cleanErrorLogs(ctx)
      j.Log.Info("cleaned error logs", "count", count)
      
      // 3. Delete expired data exports
      count, err = j.cleanExpiredExports(ctx)
      j.Log.Info("cleaned expired exports", "count", count)
      
      return nil
  }
  
  func (j *RetentionJob) cleanConversations(ctx context.Context) (int, error) {
      cutoff := time.Now().AddDate(0, 0, -j.Config.ConversationRetentionDays)
      result, err := j.DB.ExecContext(ctx, `
          DELETE FROM messages 
          WHERE conversation_id IN (
              SELECT id FROM conversations 
              WHERE created_at < $1
          )
      `, cutoff)
      // Also delete conversations
  }
  ```

- [x] **Register job** in main.go or scheduler
  
  ```go
  // Run daily at midnight
  scheduler.Every(1).Day().At("00:00").Do(retentionJob.Run)
  ```

---

## Verification

### Unit Tests
```bash
# Privacy service tests
go test ./internal/services/... -v -run Privacy

# Consent database tests
go test ./internal/db/... -v -run Consent
```

### Manual Testing

1. **User data export flow:**
   - Login as regular user
   - Navigate to `/settings/privacy`
   - Click "Request Data Export"
   - Verify request appears in admin panel
   - Admin approves request
   - User receives download link

2. **Account deletion flow:**
   - User requests deletion with reason
   - Admin sees pending request
   - Admin approves deletion
   - User account is soft-deleted

3. **Consent management:**
   - Toggle consent settings
   - Verify changes persist
   - Verify new users have default consents

---

## Files to Create

| File | Description |
|------|-------------|
| `db/migrations/000042_kvkk_compliance.up.sql` | Privacy tables |
| `db/migrations/000042_kvkk_compliance.down.sql` | Down migration |
| `internal/services/privacy_service.go` | Data export/deletion |
| `internal/db/privacy.go` | Privacy request queries |
| `internal/db/consent.go` | Consent queries |
| `internal/api/handlers/privacy.go` | Admin privacy handlers |
| `internal/api/handlers/user_privacy.go` | User privacy handlers |
| `internal/services/retention_job.go` | Data cleanup job |
| `frontend/src/pages/PrivacySettingsPage.tsx` | User privacy page |
| `frontend/src/pages/admin/AdminPrivacyPage.tsx` | Admin privacy page |

---

## Step 4.9: Integration Testing

Verify KVKK features end-to-end.

### Tasks

- [x] **Create `internal/integration/privacy_test.go`**
  - [x] Test User Consent Management
  - [x] Test Data Export Request Flow
  - [x] Test Deletion Request Flow
  - [x] Test Admin Request Processing
  - [x] Test Admin Data Export Generation

- [x] **Verify Retention Job**
  - [x] Create test for retention job in `internal/integration/retention_test.go`
  - [x] Verify expired data is actually deleted
  - [x] Verify non-expired data is preserved

---

## Step 4.10: Documentation

- [x] **Create `docs/kvkk_compliance.md`**
  - [x] Explanation of data collection
  - [x] User rights and how to exercise them
  - [x] Data retention policy details
- [ ] **Update API Documentation**
  - Document new privacy endpoints

---

## Checklist

- [x] Database tables created
- [x] Privacy service implemented
- [x] User-facing API endpoints created
- [x] Admin API endpoints created
- [x] Admin UI pages created
- [x] User settings UI created
- [x] Data retention job implemented
- [x] Integration tests passed
- [x] Documentation updated (KVKK Guide created)
