> Extracted from: security-middleware.md — Section: 3. Test Scenarios

# 3. Test Scenarios

## 3.1 Security & Middleware Scenarios

### 3.1.1 Auth — Valid JWT Cookie

1. Login successfully (receives `botla_token` and `botla_refresh_token` cookies)
2. Make a request to a protected endpoint
3. **Expected:** Auth middleware validates the JWT
4. **Expected:** User context is set on the request
5. **Expected:** Request succeeds (2xx)

### 3.1.2 Auth — Missing JWT Cookie

1. Make a request to a protected endpoint without any cookies
2. **Expected:** Auth middleware returns 401 Unauthorized
3. **Expected:** Response body indicates missing authentication
4. **Expected:** No user context is set

### 3.1.3 Auth — Expired JWT

1. Wait for JWT to expire (or use an expired token)
2. Make a request to a protected endpoint
3. **Expected:** Auth middleware returns 401 Unauthorized
4. **Expected:** Response body indicates token expired

### 3.1.4 Auth — Invalid JWT Signature

1. Tamper with the `botla_token` cookie value
2. Make a request to a protected endpoint
3. **Expected:** Auth middleware returns 401 Unauthorized
4. **Expected:** Response body indicates invalid token

### 3.1.5 Auth — Malformed JWT

1. Set `botla_token` to a random string (not a JWT)
2. Make a request
3. **Expected:** Auth middleware returns 401
4. **Expected:** No panic or internal error

### 3.1.6 Auth — Deleted User

1. Delete a user (soft delete)
2. Attempt to use their valid JWT cookie
3. **Expected:** Auth middleware returns 401
4. **Expected:** User is prompted to re-authenticate

### 3.1.7 Rate Limit — Per-User Exceeded

1. As an authenticated user, rapidly send requests to a rate-limited endpoint
2. Exceed the per-user limit
3. **Expected:** Response is 429 Too Many Requests
4. **Expected:** `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers are set
5. **Expected:** `Retry-After` header indicates cooldown period

### 3.1.8 Rate Limit — Per-IP Exceeded (Unauthenticated)

1. As an unauthenticated user, rapidly send requests
2. Exceed the per-IP limit
3. **Expected:** Same as 3.1.7

### 3.1.9 Rate Limit — Global Exceeded

1. Exceed the global rate limit (all users combined)
2. **Expected:** All users receive 429 responses
3. **Expected:** This acts as a safety net to prevent total system overload

### 3.1.10 Rate Limit — Redis Down

1. Stop Redis server
2. Make a request
3. **Expected:** Based on configuration, rate limiter either fails open (allows request) or fails closed (denies)
4. **Expected:** Error is logged
5. **Expected:** System does not crash

### 3.1.11 CORS — Allowed Origin

1. Make a request from a configured allowed origin
2. **Expected:** `Access-Control-Allow-Origin` header matches the request origin
3. **Expected:** `Access-Control-Allow-Credentials: true` is set
4. **Expected:** Preflight (OPTIONS) succeeds

### 3.1.12 CORS — Disallowed Origin

1. Make a request from `https://evil.com`
2. **Expected:** Response does NOT include `Access-Control-Allow-Origin` (or is blocked)
3. **Expected:** Browser blocks the cross-origin request
4. **Expected:** Server may return 403

### 3.1.13 CORS — Preflight with Custom Headers

1. OPTIONS request with `Access-Control-Request-Method: POST` and `Access-Control-Request-Headers: X-Embed-Token`
2. **Expected:** Responds with 204 No Content
3. **Expected:** `Access-Control-Allow-Headers` includes `X-Embed-Token`

### 3.1.14 Security Headers

1. Make any request to the API
2. **Expected:** Response includes: `Strict-Transport-Security`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Content-Security-Policy`, `Referrer-Policy`

### 3.1.15 MaxBytes — Request Too Large

1. POST a request with body > 52 MB
2. **Expected:** Response is 413 Payload Too Large
3. **Expected:** Request body is not processed

### 3.1.16 MaxBytes — File Upload

1. POST a file upload within the size limit
2. **Expected:** File is accepted (for endpoints that accept uploads)

### 3.1.17 SSRF — Private IP Blocked

1. Configure ingestion with a URL pointing to `http://10.0.0.1/admin`
2. **Expected:** URL is rejected with SSRF error
3. **Expected:** Ingestion does not proceed

### 3.1.18 SSRF — Localhost Blocked

1. Configure ingestion with `http://127.0.0.1:3000/`
2. **Expected:** Rejected with SSRF error

### 3.1.19 SSRF — Metadata Endpoint Blocked

1. Configure ingestion with `http://169.254.169.254/latest/meta-data/`
2. **Expected:** Rejected with SSRF error

### 3.1.20 SSRF — Allowlist Override

1. Configure internal service allowlist
2. Ingest URL from an allowlisted internal service
3. **Expected:** URL is allowed through SSRF protection

### 3.1.21 SSRF — Non-HTTP Protocol Blocked

1. Configure ingestion with `file:///etc/passwd`
2. **Expected:** Rejected with SSRF error (invalid protocol)

### 3.1.22 SSRF — Redirect to Private IP

1. Configure ingestion with an external URL that redirects to `http://192.168.1.1`
2. **Expected:** Redirect is followed, IP is re-checked
3. **Expected:** Request is blocked when private IP is detected

### 3.1.23 Deleted Account — Access Blocked

1. Soft-delete a user account
2. User makes a request with valid JWT
3. **Expected:** DeletedAccount middleware returns 410 Gone
4. **Expected:** `X-Reason: account_deleted` header is set

### 3.1.24 Panic Recovery

1. Trigger a panic in a handler (e.g., nil pointer dereference)
2. **Expected:** Recovery middleware catches the panic
3. **Expected:** Response is 500 with request ID
4. **Expected:** Stack trace is logged (not returned to client)
5. **Expected:** Server continues running

### 3.1.25 Request Logging — Normal Request

1. Make a successful request
2. **Expected:** Log entry includes: method, path, status (2xx), duration, request ID
3. **Expected:** Log level is INFO

### 3.1.26 Request Logging — Error Request

1. Make a request that returns 5xx
2. **Expected:** Log entry includes additional error context
3. **Expected:** Log level is ERROR

### 3.1.27 Request Logging — Slow Request

1. Make a request that takes > 1s
2. **Expected:** Log entry is WARN level with slow-request marker

### 3.1.28 GDPR — Soft Delete Data Preservation

1. Delete a user account
2. **Expected:** User data is preserved in database (deleted_at set, not actually removed)
3. **Expected:** User cannot authenticate
4. **Expected:** Admin can restore the account within the 30-day window

### 3.1.29 GDPR — Hard Delete After Window

1. After soft-delete window expires (30 days)
2. Hard delete runs (cron job or manual trigger)
3. **Expected:** User data is permanently removed
4. **Expected:** Anonymized audit trail remains (no PII)

### 3.1.30 Audit Logging — Admin Action

1. Perform an admin action (e.g., promote user, change plan)
2. **Expected:** Audit log entry is created
3. **Expected:** Entry includes: admin ID, action type, target user ID, timestamp, source IP
4. **Expected:** Log is append-only (cannot be modified or deleted)
