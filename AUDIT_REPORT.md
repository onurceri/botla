# Botla Backend - Comprehensive Codebase Audit Report

This report documents the findings, potential bugs, architectural observations, and performance optimization opportunities discovered during an exhaustive, multi-layered audit of the Botla backend codebase.

---

## 1. Executive Summary

### 1.1 Overview
Botla is an AI-powered chatbot platform featuring a robust RAG (Retrieval-Augmented Generation) pipeline, multi-tenant workspace isolation, and a real-time embeddable chat widget. This audit covers the backend Go services, API routers, database repositories, vector db interactions, background queues, and utility packages.

### 1.2 Core Strengths of the Codebase
- **Clean Architectural Separation:** The codebase strictly segregates concerns (HTTP API Handlers $\rightarrow$ Business Logic Services $\rightarrow$ Repository/Data Access layers).
- **Strong Safety Mechanisms:** Blocking SSRF (Server-Side Request Forgery) using custom IP and CIDR validators during scraping is an excellent security measure.
- **Robust Error Contexts:** Wrapping of Go errors using the `pkg/errors` package provides highly readable error stack traces.
- **No-ORM Philosophy:** Relying on `sqlc` or raw/parameterized queries prevents common ORM performance pitfalls and guarantees compile-time check benefits.

### 1.3 Key Areas for Improvement
While the code is well-structured, several high-impact concurrency vulnerabilities, resource leaks, database constraints, and performance bottlenecks have been identified. Addressing these issues will significantly improve production stability, eliminate rare background crashes, and optimize resource utilization.

---

## 2. High Severity / Critical Vulnerabilities & Potential Bugs

### 2.1 Unique Email Constraint Collision on Soft-Deleted Users
- **Location:** `db/migrations/000001_init.up.sql` (line 515)
- **Vulnerability Type:** Database Constraint / Logic Defect
- **Impact:** Medium-High
- **Description:**
  The `users` table has a standard unique constraint:
  ```sql
  ALTER TABLE ONLY users ADD CONSTRAINT users_email_key UNIQUE (email);
  ```
  However, the platform supports GDPR-compliant soft-deletion (`deleted_at` field). If a user soft-deletes their account, `deleted_at` is set to `NOW()`. If another user (or the same user) later tries to sign up with the exact same email address, the registration will **fail** because the database-level unique constraint does not ignore rows where `deleted_at IS NOT NULL`.
- **Recommendation:**
  Replace the standard unique constraint on `email` with a partial unique index that only enforces uniqueness on active (non-deleted) users:
  ```sql
  ALTER TABLE users DROP CONSTRAINT users_email_key;
  CREATE UNIQUE INDEX ux_users_email_active ON users (email) WHERE (deleted_at IS NULL);
  ```

---

### 2.2 Thread-Safety and Concurrency Panic in `CircuitBreakerManager`
- **Location:** `internal/rag/circuit_breaker.go` (line 126)
- **Vulnerability Type:** Concurrency Bug (Race Condition)
- **Impact:** High
- **Description:**
  `CircuitBreakerManager` manages provider circuit breakers inside a standard Go map:
  ```go
  type CircuitBreakerManager struct {
      breakers map[string]*CircuitBreakerClient
      metrics  *LLMMetrics
      settings CircuitBreakerSettings
  }
  ```
  The methods `RegisterClient`, `GetClient`, `GetAllStates`, and `GetAllProviderStatus` read and write to this `breakers` map. Standard Go maps are **not thread-safe**. When concurrent HTTP chat requests invoke the LLM pipeline, they will access/register clients on this map concurrently. This will inevitably result in a **concurrent map read-write panic**, crashing the entire application process instantly.
- **Recommendation:**
  Protect the `breakers` map with a `sync.RWMutex` (Read-Write Mutex) to ensure concurrent safe operations:
  ```go
  type CircuitBreakerManager struct {
      breakers map[string]*CircuitBreakerClient
      mu       sync.RWMutex
      metrics  *LLMMetrics
      settings CircuitBreakerSettings
  }
  ```
  Ensure all reads lock `mu.RLock()` and writes lock `mu.Lock()`.

---

### 2.3 Unhandled Async Goroutine Panic in `trackAnalyticsAsync`
- **Location:** `internal/services/chat_helpers.go` (line 241)
- **Vulnerability Type:** Concurrency Bug (Process Crash)
- **Impact:** High
- **Description:**
  In `ChatService.trackAnalyticsAsync`, if analytics are configured to run asynchronously (`s.SyncAnalytics` is false), the service spawns a raw goroutine:
  ```go
  if s.SyncAnalytics {
      fn()
  } else {
      go fn()
  }
  ```
  If any panic occurs inside `fn()` (e.g., if a pointer like `s.AnalyticsRepo`, `s.Log`, or database session gets de-referenced as `nil`), the panic will propagate to the top of the newly spawned goroutine. Since Go does not recover goroutines spawned with `go` at the web-framework middleware layer, **the entire Go process will immediately terminate (crash)**.
- **Recommendation:**
  Always wrap asynchronous goroutines in a panic recovery block:
  ```go
  go func() {
      defer func() {
          if r := recover(); r != nil && s.Log != nil {
              s.Log.Error("async_analytics_panic", map[string]any{"panic": r})
          }
      }()
      fn()
  }()
  ```

---

### 2.4 BrowserPool Resource and Scaling Bottleneck after Idle Reaping
- **Location:** `internal/scraper/browser.go` (line 55-90)
- **Vulnerability Type:** Logic Bug / Resource Utilization
- **Impact:** Medium
- **Description:**
  The `BrowserPool` reaps headless browsers if they have been idle longer than `idleTTL` by setting `bp.browsers = nil` and closing the browser connections:
  ```go
  if time.Since(lastUse) > idleTTL {
      bp.mu.Lock()
      browsers := bp.browsers
      bp.browsers = nil
      bp.mu.Unlock()
      for _, b := range browsers {
          _ = b.Close()
      }
  }
  ```
  However, inside `nextBrowser()`, if `len(bp.browsers) == 0`, it only re-creates a **single** browser and puts it in the slice:
  ```go
  if len(bp.browsers) == 0 {
      l := launcher.New().Headless(true)
      // ...
      b := rod.New().ControlURL(u).MustConnect()
      bp.browsers = append(bp.browsers, b)
      bp.idx = 0
      return b
  }
  ```
  If the pool was originally configured with a size of 10, after idle reaping it gets stuck at size 1 because subsequent calls to `nextBrowser` will see `len(bp.browsers) == 1` and continuously reuse/rotate that single browser connection instead of expanding back to the pool size.
- **Recommendation:**
  The pool should either lazily scale up to its maximum size on demand or re-initialize all of its browsers up to the configured pool limit, rather than sticking to a single re-created instance.

---

### 2.5 Inaccurate sliding window rate tracking in `RedisLimiter`
- **Location:** `pkg/ratelimit/redis.go` (line 74)
- **Vulnerability Type:** Logic Bug / Rate Limiting Precision
- **Impact:** Medium
- **Description:**
  The Redis rate limiter uses a sliding window log algorithm using Redis Sorted Sets with a microsecond-precision calculation:
  ```lua
  local score = now + (i / 1000000)
  redis.call('ZADD', key, score, now .. ':' .. i)
  ```
  However, the `now` argument is passed as `now.Unix()`, which has **second-level precision** (e.g. 1761820000). Consecutive requests within the same second from different clients will have the exact same `now .. ':' .. i` value. Since members in sorted sets must be unique, they will collide and overwrite each other, causing the rate limiter to under-count actual requests and bypass configured restrictions.
- **Recommendation:**
  Pass sub-second or nanosecond timestamps as score inputs, such as `now.UnixNano()` or `now.UnixMilli()`, or append a random call suffix/UUID to the member name string to prevent sorted set collisions.

---

## 3. Performance & Resource Optimizations

### 3.1 Database Connection Churn (Low MaxIdleConns relative to MaxOpenConns)
- **Location:** `internal/db/db.go` (line 28-29)
- **Optimization Type:** Database Performance
- **Impact:** Medium-High
- **Description:**
  The Postgres database pool is configured as:
  ```go
  conn.SetMaxOpenConns(25)
  conn.SetMaxIdleConns(5)
  ```
  Having a maximum of 25 open connections but only 5 idle connections means that during traffic spikes (bursty chat or ingestion loads), the application will open up to 25 connections, but immediately close 20 of them as soon as the load subsides. This creates massive **TCP connection churn**, increasing latency and database CPU overhead.
  Additionally, neither `SetConnMaxLifetime` nor `SetConnMaxIdleTime` are configured. This can lead to silent TCP connection drop failures by load balancers or firewalls.
- **Recommendation:**
  Increase `MaxIdleConns` to be closer to `MaxOpenConns` (e.g., 25 or at least 15) and define lifetimes:
  ```go
  conn.SetMaxOpenConns(25)
  conn.SetMaxIdleConns(15)
  conn.SetConnMaxLifetime(30 * time.Minute)
  conn.SetConnMaxIdleTime(5 * time.Minute)
  ```

---

### 3.2 SQL N-Roundtrip Bottleneck in `SaveMessageSources`
- **Location:** `internal/repository/conversation_repo.go` (line 220-235)
- **Optimization Type:** Database Performance / Roundtrip Reduction
- **Impact:** Medium
- **Description:**
  When persisting RAG sources used for a message, the repository loops over the sources and executes an `INSERT` statement individually:
  ```go
  for _, src := range sources {
      _, err = stmt.ExecContext(ctx, messageID, src.SourceID, src.ChunkIndex, src.SourceType, src.Score)
      // ...
  }
  ```
  If a message references 5-10 chunks, this causes 5-10 synchronous database roundtrips sequentially. Under high chat throughput, this creates massive write amplification and latency.
- **Recommendation:**
  Use PostgreSQL bulk-inserting syntax (inserting multiple values in a single statement) or use the pgx-native copy protocol (`tx.CopyFrom`) to bulk-insert all sources in a single network roundtrip.

---

### 3.3 HTTP Socket & Connection Exhaustion in Admin Health Checks
- **Location:** `internal/api/handlers/admin_health.go` (line 190-250)
- **Optimization Type:** Network Performance / System Resource Preservation
- **Impact:** Medium
- **Description:**
  In `checkQdrant`, `checkOpenAI`, and `checkOpenRouter` health check functions, a new HTTP client is initialized *on every request*:
  ```go
  client := &http.Client{Timeout: 5 * time.Second}
  ```
  Since these clients do not share connection pools, running these health checks frequently will exhaust ephemeral ports under high platform admin usage, leading to TCP socket leaks and connection failures.
- **Recommendation:**
  Define a shared, pre-configured `http.Client` on the `AdminHealthHandlers` struct during instantiation and reuse it across all health checks.

---

### 3.4 Lock Contention on Joined Tables in `UpdateMessageFeedback`
- **Location:** `internal/repository/analytics_repo.go` (line 470)
- **Optimization Type:** Database Locking & Concurrency
- **Impact:** Medium
- **Description:**
  The message feedback transaction performs a joined lock:
  ```sql
  SELECT m.thumbs_up, c.chatbot_id
  FROM messages m
  JOIN conversations c ON m.conversation_id = c.id
  WHERE m.id = $1 FOR UPDATE
  ```
  In PostgreSQL, a bare `FOR UPDATE` clause locks rows across **all** tables in the query (both `messages` and `conversations`). This means updating message feedback will unnecessarily lock the parent conversation, causing lock contention with concurrent user messages or other actions on that same conversation.
- **Recommendation:**
  Specify `FOR UPDATE OF m` to only lock the targeted message row:
  ```sql
  SELECT m.thumbs_up, c.chatbot_id
  FROM messages m
  JOIN conversations c ON m.conversation_id = c.id
  WHERE m.id = $1 FOR UPDATE OF m
  ```

---

## 4. Code Quality, Safety, & Maintainability Observations

### 4.1 Unchecked Error Values on `rows.Close()`
- **Location:** Multiple repository files under `internal/repository/`
- **Issue:** Resource Leak Risk
- **Impact:** Low-Medium
- **Description:**
  There are several instances of `defer func() { _ = rows.Close() }()` or `defer rows.Close()` where the return value of `rows.Close()` is ignored. While `rows.Close()` rarely fails, ignoring it completely goes against strict Go static analysis and linters (like `errcheck`).
- **Recommendation:**
  Ensure that errors returned by `rows.Close()` are logged or verified, particularly in high-volume query loops.

### 4.2 Exposed Database Ports in Development/Integration Configs
- **Location:** `docker-compose.yml` & `docker-compose.dev.yml`
- **Issue:** Infrastructure Security
- **Impact:** Low (Dev environment) / High (If reused in Prod)
- **Description:**
  Postgres port `5432` and Qdrant port `6333` are exposed to the host machine. While convenient for local development, these should be locked down to private container networks or strictly restricted to internal subnets in staging/production setups to prevent external port scanning and attacks.

---

## 5. Concrete Action Plan & Implementation Roadmap

To systematically resolve these issues, the following roadmap is highly recommended:

```
+-----------------------------------------------------------------------------------+
| 1. CONCURRENCY & CRASH PREVENTION (Immediate)                                     |
|    - Implement sync.RWMutex in CircuitBreakerManager.                             |
|    - Wrap trackAnalyticsAsync background goroutine with panic recovery defer.     |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| 2. INFRASTRUCTURE & SCALING FIXES (Short-Term)                                    |
|    - Fix BrowserPool stuck at size 1 by scale-up logic on re-initialization.       |
|    - Use nano precision/uniqueness suffix in RedisLimiter sliding window logs.    |
|    - Replace standard users email unique constraint with partial active index.    |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| 3. PERFORMANCE & DATABASE TUNING (Medium-Term)                                    |
|    - Tune MaxIdleConns to 15-20 and set Connection Max Lifetime/Idle Time.        |
|    - Convert looping individual inserts in SaveMessageSources to multi-value copy. |
|    - Refactor http.Client creation in health checks to reuse a shared client.     |
|    - Restrict FOR UPDATE locks in feedback queries to only targeted message rows. |
+-----------------------------------------------------------------------------------+
```
