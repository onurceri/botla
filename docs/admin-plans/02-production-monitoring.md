# Phase 2: Production Monitoring Backend

> **Estimated Time:** 2-3 days  
> **Priority:** Critical (Week 1)  
> **Depends On:** Phase 1 (Steps 1.1, 1.3)  

This phase adds production monitoring capabilities: health checks, queue monitoring, and error tracking.

---

## Step 2.1: Enhanced Health Check Endpoints

Extend the existing health handler to provide detailed status.

### Tasks

- [ ] **Create `internal/api/handlers/admin_health.go`**
  
  ```go
  package handlers
  
  type AdminHealthHandlers struct {
      DB  *sql.DB
      Cfg *config.Config
  }
  
  type DependencyStatus struct {
      Name      string  `json:"name"`
      Status    string  `json:"status"` // "ok", "degraded", "down"
      Latency   float64 `json:"latency_ms"`
      Message   string  `json:"message,omitempty"`
      CheckedAt time.Time `json:"checked_at"`
  }
  
  type DetailedHealth struct {
      Status       string             `json:"status"` // "healthy", "degraded", "unhealthy"
      Version      string             `json:"version"`
      Uptime       string             `json:"uptime"`
      Dependencies []DependencyStatus `json:"dependencies"`
      Environment  string             `json:"environment"`
  }
  
  func (h *AdminHealthHandlers) GetDetailedHealth(w http.ResponseWriter, r *http.Request) {
      // Check each dependency in parallel
      // Aggregate results
  }
  ```

- [ ] **Implement dependency checks**
  
  Create helper functions:
  ```go
  func checkPostgres(ctx context.Context, db *sql.DB) DependencyStatus
  func checkRedis(ctx context.Context, cfg *config.Config) DependencyStatus
  func checkQdrant(ctx context.Context, cfg *config.Config) DependencyStatus
  func checkOpenAI(ctx context.Context, cfg *config.Config) DependencyStatus
  func checkStorage(ctx context.Context, cfg *config.Config) DependencyStatus  // S3/R2
  ```

- [ ] **Add Redis health check**
  
  ```go
  func checkRedis(ctx context.Context, cfg *config.Config) DependencyStatus {
      start := time.Now()
      client := redis.NewClient(&redis.Options{
          Addr: cfg.REDIS_URL,
      })
      defer client.Close()
      
      err := client.Ping(ctx).Err()
      latency := time.Since(start).Milliseconds()
      
      if err != nil {
          return DependencyStatus{
              Name:    "redis",
              Status:  "down",
              Latency: float64(latency),
              Message: err.Error(),
          }
      }
      
      return DependencyStatus{
          Name:    "redis",
          Status:  "ok",
          Latency: float64(latency),
      }
  }
  ```

- [ ] **Add OpenAI health check** (ping models endpoint)
  
  ```go
  func checkOpenAI(ctx context.Context, cfg *config.Config) DependencyStatus {
      start := time.Now()
      client := &http.Client{Timeout: 5 * time.Second}
      
      req, _ := http.NewRequestWithContext(ctx, "GET", 
          "https://api.openai.com/v1/models", nil)
      req.Header.Set("Authorization", "Bearer "+cfg.OPENAI_API_KEY)
      
      resp, err := client.Do(req)
      latency := time.Since(start).Milliseconds()
      
      if err != nil || resp.StatusCode != 200 {
          return DependencyStatus{
              Name:    "openai",
              Status:  "down",
              Latency: float64(latency),
              Message: "API unreachable",
          }
      }
      
      return DependencyStatus{
          Name:    "openai",
          Status:  "ok",
          Latency: float64(latency),
      }
  }
  ```

- [ ] **Add storage health check** (S3/R2)
  
  ```go
  func checkStorage(ctx context.Context, cfg *config.Config) DependencyStatus {
      // HEAD request to bucket or list with limit 1
  }
  ```

---

## Step 2.2: Queue Monitoring

Monitor scraping and refresh queues.

### Tasks

- [ ] **Create `internal/db/admin_queue.go`**
  
  ```go
  package db
  
  type QueueStats struct {
      QueueName       string    `json:"queue_name"`
      PendingCount    int       `json:"pending_count"`
      ProcessingCount int       `json:"processing_count"`
      FailedCount     int       `json:"failed_count"`
      OldestPending   *time.Time `json:"oldest_pending"`
  }
  
  type StuckJob struct {
      ID            string    `json:"id"`
      QueueName     string    `json:"queue_name"`
      SourceID      string    `json:"source_id,omitempty"`
      ChatbotID     string    `json:"chatbot_id,omitempty"`
      Status        string    `json:"status"`
      StartedAt     time.Time `json:"started_at"`
      StuckDuration string    `json:"stuck_duration"`
      ErrorMessage  string    `json:"error_message,omitempty"`
  }
  
  // GetQueueStats returns stats for source-related queues
  func GetQueueStats(ctx context.Context, pool *sql.DB) ([]QueueStats, error) {
      // Query sources table grouped by status
      // pending_discovered_urls counts
      // sources with status='processing' for > 10 minutes
  }
  
  // GetStuckJobs returns jobs that have been processing too long
  func GetStuckJobs(ctx context.Context, pool *sql.DB, threshold time.Duration) ([]StuckJob, error) {
      // sources WHERE status='processing' AND updated_at < NOW() - threshold
      // pending_discovered_urls stuck
  }
  ```

- [ ] **Create `internal/api/handlers/admin_queues.go`**
  
  ```go
  package handlers
  
  type AdminQueueHandlers struct {
      DB *sql.DB
  }
  
  // GetQueues returns queue statistics
  func (h *AdminQueueHandlers) GetQueues(w http.ResponseWriter, r *http.Request) {
      stats, err := db.GetQueueStats(r.Context(), h.DB)
      // Return stats
  }
  
  // GetStuckJobs returns jobs that appear stuck
  func (h *AdminQueueHandlers) GetStuckJobs(w http.ResponseWriter, r *http.Request) {
      threshold := 30 * time.Minute // Configurable via query param
      if t := r.URL.Query().Get("threshold"); t != "" {
          if d, err := time.ParseDuration(t); err == nil {
              threshold = d
          }
      }
      
      jobs, err := db.GetStuckJobs(r.Context(), h.DB, threshold)
      // Return jobs
  }
  
  // RetryJob resets a stuck job for retry
  func (h *AdminQueueHandlers) RetryJob(w http.ResponseWriter, r *http.Request) {
      jobID := chi.URLParam(r, "id")
      // Reset job status to 'pending'
      // Log admin action
  }
  
  // DeleteJob removes a stuck job
  func (h *AdminQueueHandlers) DeleteJob(w http.ResponseWriter, r *http.Request) {
      jobID := chi.URLParam(r, "id")
      // Delete or mark as failed
      // Log admin action
  }
  ```

- [ ] **Add queue routes** (update `internal/api/routes/admin.go`)
  
  ```go
  // Queues
  r.Get("/queues", h.GetQueues)
  r.Get("/queues/stuck", h.GetStuckJobs)
  r.Post("/queues/{id}/retry", h.RetryJob)
  r.Delete("/queues/{id}", h.DeleteJob)
  ```

---

## Step 2.3: Error Tracking Service

Persistent error logging for production issues.

### Tasks

- [ ] **Create `internal/services/error_logger.go`**
  
  ```go
  package services
  
  type ErrorLogger struct {
      DB  *sql.DB
      Log *logger.Logger
  }
  
  type ErrorEntry struct {
      ErrorType      string
      Message        string
      StackTrace     string
      RequestPath    string
      RequestMethod  string
      UserID         *string
      ChatbotID      *string
      OrganizationID *string
      Severity       string  // "info", "warning", "error", "critical"
      Context        map[string]any
  }
  
  func (e *ErrorLogger) LogError(ctx context.Context, entry ErrorEntry) error {
      // Insert into error_logs table
  }
  
  func (e *ErrorLogger) LogCritical(ctx context.Context, entry ErrorEntry) error {
      entry.Severity = "critical"
      return e.LogError(ctx, entry)
  }
  ```

- [ ] **Create `internal/db/admin_errors.go`**
  
  ```go
  package db
  
  type ErrorLogEntry struct {
      ID             string         `json:"id"`
      ErrorType      string         `json:"error_type"`
      Message        string         `json:"message"`
      StackTrace     string         `json:"stack_trace,omitempty"`
      RequestPath    string         `json:"request_path,omitempty"`
      RequestMethod  string         `json:"request_method,omitempty"`
      UserID         *string        `json:"user_id,omitempty"`
      ChatbotID      *string        `json:"chatbot_id,omitempty"`
      OrganizationID *string        `json:"organization_id,omitempty"`
      Severity       string         `json:"severity"`
      Context        map[string]any `json:"context,omitempty"`
      CreatedAt      time.Time      `json:"created_at"`
  }
  
  type ErrorFilter struct {
      ErrorType   string
      Severity    string
      Since       time.Time
      Until       time.Time
      ChatbotID   string
      UserID      string
  }
  
  func ListErrorLogs(ctx context.Context, pool *sql.DB, filter ErrorFilter, limit, offset int) ([]ErrorLogEntry, int, error) {
      // Query with filters and pagination
  }
  
  func GetErrorLog(ctx context.Context, pool *sql.DB, id string) (*ErrorLogEntry, error) {
      // Get single error with full details
  }
  
  func GetErrorStats(ctx context.Context, pool *sql.DB, since time.Time) (map[string]int, error) {
      // Count errors grouped by type
  }
  ```

- [ ] **Create `internal/api/handlers/admin_errors.go`**
  
  ```go
  package handlers
  
  type AdminErrorHandlers struct {
      DB *sql.DB
  }
  
  func (h *AdminErrorHandlers) ListErrors(w http.ResponseWriter, r *http.Request) {
      // Parse filter params
      // Query and return paginated results
  }
  
  func (h *AdminErrorHandlers) GetError(w http.ResponseWriter, r *http.Request) {
      id := chi.URLParam(r, "id")
      // Return full error details
  }
  
  func (h *AdminErrorHandlers) GetErrorStats(w http.ResponseWriter, r *http.Request) {
      // Return error counts by type for chart
  }
  ```

- [ ] **Add error routes** (update `internal/api/routes/admin.go`)
  
  ```go
  // Errors
  r.Get("/errors", h.ListErrors)
  r.Get("/errors/stats", h.GetErrorStats)
  r.Get("/errors/{id}", h.GetError)
  ```

---

## Step 2.4: Error Logging Middleware

Automatically log critical errors from handlers.

### Tasks

- [ ] **Create error recovery middleware** `pkg/middleware/error_logger.go`
  
  ```go
  package middleware
  
  import (
      "runtime/debug"
      "github.com/onurceri/botla-co/internal/services"
  )
  
  func ErrorRecovery(errLogger *services.ErrorLogger) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              defer func() {
                  if err := recover(); err != nil {
                      stack := string(debug.Stack())
                      
                      // Extract user context if available
                      var userID, chatbotID *string
                      if u := UserIDFromContext(r.Context()); u != "" {
                          userID = &u
                      }
                      
                      errLogger.LogCritical(r.Context(), services.ErrorEntry{
                          ErrorType:     "panic",
                          Message:       fmt.Sprintf("%v", err),
                          StackTrace:    stack,
                          RequestPath:   r.URL.Path,
                          RequestMethod: r.Method,
                          UserID:        userID,
                          Severity:      "critical",
                      })
                      
                      http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                  }
              }()
              next.ServeHTTP(w, r)
          })
      }
  }
  ```

- [ ] **Add to middleware chain** in main router setup
  
  ```go
  r.Use(middleware.ErrorRecovery(errorLogger))
  ```

---

## Step 2.5: Integrate Error Logging in Critical Paths

Add error logging to key services.

### Tasks

- [ ] **Update `internal/services/chat_service.go`**
  
  Log errors when chat responses fail:
  ```go
  if err != nil {
      s.ErrorLogger.LogError(ctx, services.ErrorEntry{
          ErrorType:   "chat_response_error",
          Message:     err.Error(),
          ChatbotID:   &chatbotID,
          Severity:    "error",
          Context:     map[string]any{"query": query},
      })
  }
  ```

- [ ] **Update `internal/scraper/worker.go`**
  
  Log scraping failures:
  ```go
  if err != nil {
      s.ErrorLogger.LogError(ctx, services.ErrorEntry{
          ErrorType:   "scraping_error",
          Message:     err.Error(),
          ChatbotID:   &job.ChatbotID,
          Severity:    "warning",
          Context:     map[string]any{"url": url},
      })
  }
  ```

- [ ] **Update `internal/rag/embedder.go`**
  
  Log embedding failures:
  ```go
  if err != nil {
      s.ErrorLogger.LogError(ctx, services.ErrorEntry{
          ErrorType:   "embedding_error",
          Message:     err.Error(),
          Severity:    "error",
      })
  }
  ```

---

## Step 2.6: Chatbot Admin Handler

Admin endpoints for chatbot management.

### Tasks

- [ ] **Create `internal/api/handlers/admin_chatbots.go`**
  
  ```go
  package handlers
  
  type AdminChatbotHandlers struct {
      DB *sql.DB
  }
  
  type ChatbotListItem struct {
      ID               string    `json:"id"`
      Name             string    `json:"name"`
      OrganizationID   string    `json:"organization_id"`
      OrganizationName string    `json:"organization_name"`
      Status           string    `json:"status"` // active, suspended, error
      SourceCount      int       `json:"source_count"`
      ConversationCount int      `json:"conversation_count"`
      CreatedAt        time.Time `json:"created_at"`
      LastMessageAt    *time.Time `json:"last_message_at"`
  }
  
  func (h *AdminChatbotHandlers) ListChatbots(w http.ResponseWriter, r *http.Request) {
      // Paginated list with search, filter by status/org
  }
  
  func (h *AdminChatbotHandlers) GetChatbot(w http.ResponseWriter, r *http.Request) {
      // Full chatbot details including config, sources, recent errors
  }
  
  func (h *AdminChatbotHandlers) GetChatbotConversations(w http.ResponseWriter, r *http.Request) {
      // List conversations for a chatbot
  }
  
  func (h *AdminChatbotHandlers) SuspendChatbot(w http.ResponseWriter, r *http.Request) {
      // Suspend chatbot (prevent responses)
  }
  
  func (h *AdminChatbotHandlers) ForceRefreshChatbot(w http.ResponseWriter, r *http.Request) {
      // Queue refresh for all sources
  }
  ```

- [ ] **Create `internal/db/admin_chatbots.go`**
  
  ```go
  func ListChatbotsAdmin(ctx context.Context, pool *sql.DB, params ChatbotListParams) ([]ChatbotListItem, int, error)
  func GetChatbotDetailAdmin(ctx context.Context, pool *sql.DB, id string) (*ChatbotDetail, error)
  func SuspendChatbot(ctx context.Context, pool *sql.DB, id string) error
  func UnsuspendChatbot(ctx context.Context, pool *sql.DB, id string) error
  ```

- [ ] **Add chatbot routes**
  
  ```go
  // Chatbots
  r.Get("/chatbots", h.ListChatbots)
  r.Get("/chatbots/{id}", h.GetChatbot)
  r.Get("/chatbots/{id}/conversations", h.GetChatbotConversations)
  r.Patch("/chatbots/{id}", h.UpdateChatbot)  // suspend/unsuspend
  r.Post("/chatbots/{id}/force-refresh", h.ForceRefreshChatbot)
  ```

---

## Step 2.7: Sources Admin Handler

Admin endpoints for data source management.

### Tasks

- [ ] **Create `internal/api/handlers/admin_sources.go`**
  
  ```go
  package handlers
  
  type AdminSourceHandlers struct {
      DB *sql.DB
  }
  
  type SourceListItem struct {
      ID            string    `json:"id"`
      ChatbotID     string    `json:"chatbot_id"`
      ChatbotName   string    `json:"chatbot_name"`
      SourceType    string    `json:"source_type"`
      SourceURL     string    `json:"source_url,omitempty"`
      Status        string    `json:"status"`
      ErrorMessage  string    `json:"error_message,omitempty"`
      LastRefreshed *time.Time `json:"last_refreshed"`
      CreatedAt     time.Time `json:"created_at"`
  }
  
  func (h *AdminSourceHandlers) ListSources(w http.ResponseWriter, r *http.Request) {
      // Paginated list with filters
  }
  
  func (h *AdminSourceHandlers) ListFailedSources(w http.ResponseWriter, r *http.Request) {
      // Only sources with status='error'
  }
  
  func (h *AdminSourceHandlers) ReprocessSource(w http.ResponseWriter, r *http.Request) {
      // Reset status and queue for processing
  }
  
  func (h *AdminSourceHandlers) DeleteSource(w http.ResponseWriter, r *http.Request) {
      // Cascade delete source and chunks
  }
  ```

- [ ] **Add source routes**
  
  ```go
  // Sources
  r.Get("/sources", h.ListSources)
  r.Get("/sources/failed", h.ListFailedSources)
  r.Get("/sources/{id}", h.GetSource)
  r.Post("/sources/{id}/reprocess", h.ReprocessSource)
  r.Delete("/sources/{id}", h.DeleteSource)
  ```

---

## Verification

### Unit Tests
```bash
# Health checks
go test ./internal/api/handlers/... -v -run AdminHealth

# Queue monitoring
go test ./internal/db/... -v -run Queue

# Error logging
go test ./internal/services/... -v -run ErrorLogger
```

### Manual Verification

1. **Health endpoint:**
   ```bash
   curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/api/v1/admin/health/detailed | jq
   ```
   
   Expected: All dependencies with status/latency

2. **Queue stats:**
   ```bash
   curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/api/v1/admin/queues | jq
   ```

3. **Trigger test error:**
   - Cause a panic or error in a handler
   - Verify it appears in:
   ```bash
   curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/api/v1/admin/errors | jq
   ```

4. **Stuck job retry:**
   - Create a source that gets stuck
   - List stuck jobs
   - Retry via API

---

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `internal/api/handlers/admin_health.go` | CREATE | Detailed health checks |
| `internal/db/admin_queue.go` | CREATE | Queue monitoring queries |
| `internal/api/handlers/admin_queues.go` | CREATE | Queue management |
| `internal/services/error_logger.go` | CREATE | Error persistence |
| `internal/db/admin_errors.go` | CREATE | Error log queries |
| `internal/api/handlers/admin_errors.go` | CREATE | Error viewing |
| `pkg/middleware/error_logger.go` | CREATE | Panic recovery + logging |
| `internal/api/handlers/admin_chatbots.go` | CREATE | Chatbot management |
| `internal/db/admin_chatbots.go` | CREATE | Chatbot queries |
| `internal/api/handlers/admin_sources.go` | CREATE | Source management |
| `internal/api/routes/admin.go` | MODIFY | Add new routes |
| `internal/services/chat_service.go` | MODIFY | Add error logging |
| `internal/scraper/worker.go` | MODIFY | Add error logging |
