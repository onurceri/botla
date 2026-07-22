> Extracted from: auth.md — Section 1

# Authentication & Authorization — Feature Testing

> **Document Version**: 1.0
> **Last Updated**: July 2026
> **Scope**: Authentication & Authorization
> **Backend**: Go 1.25+ API
> **Test Frameworks**: Go `testing` (backend unit/integration), Playwright (frontend E2E)

---

# 1. Authentication & Authorization

## 1.1 Feature Overview

The authentication system uses JWT-based auth with HttpOnly cookie delivery. Access tokens expire in 1 hour, refresh tokens in 7 days with rotation. Passwords are bcrypt-hashed with strength validation. The frontend treats `GET /api/v1/me` as the auth source of truth.

## 1.2 Detailed Sub-Features

### 1.2.1 Registration (`POST /api/v1/auth/register`)

- [ ] Register with valid email, password meeting all strength requirements, and full name
- [ ] Register with minimum-valid password (8 chars, one upper, one lower, one digit, one special)
- [ ] Register with email containing sub-addressing (user+tag@example.com)
- [ ] Reject registration with existing email (duplicate)
- [ ] Reject registration with password shorter than 8 characters
- [ ] Reject registration with password missing uppercase letter
- [ ] Reject registration with password missing lowercase letter
- [ ] Reject registration with password missing digit
- [ ] Reject registration with password missing special character (@$!%*?&)
- [ ] Reject registration with malformed email (no @, no domain, spaces)
- [ ] Reject registration with empty request body
- [ ] Reject registration with missing required fields
- [ ] Reject registration with password containing only whitespace
- [ ] Reject registration with extremely long email (>254 chars)
- [ ] Reject registration with unsupported character encoding
- [ ] Verify response includes HttpOnly `botla_token` cookie (access)
- [ ] Verify response includes HttpOnly `botla_refresh_token` cookie (refresh)
- [ ] Verify response body includes `user` object with `id`, `email`, `fullName`, `createdAt`
- [ ] Verify default organization is created upon registration
- [ ] Verify default workspace is created upon registration
- [ ] Verify password is bcrypt-hashed in database (not plaintext, not Argon2)
- [ ] Verify rate limiting on registration endpoint (5 req/min per IP)

### 1.2.2 Login (`POST /api/v1/auth/login`)

- [ ] Login with correct email and password
- [ ] Login with email in different case (USER@Example.com vs user@example.com)
- [ ] Login with leading/trailing whitespace in email (trimmed server-side)
- [ ] Reject login with incorrect password
- [ ] Reject login with non-existent email
- [ ] Reject login for soft-deleted user account
- [ ] Reject login with empty credentials
- [ ] Verify response sets `botla_token` (access, 1h expiry, HttpOnly, Secure, SameSite=Lax)
- [ ] Verify response sets `botla_refresh_token` (refresh, 7d expiry, HttpOnly, Secure, SameSite=Lax)
- [ ] Verify refresh token is SHA-256 hashed in DB
- [ ] Verify response body excludes password hash
- [ ] Verify rate limiting on login endpoint (5 req/min per IP, 3 req/min after 3 failures)
- [ ] Verify `Set-Cookie` headers have correct `Path`, `Domain`, `HttpOnly`, `Secure`, `SameSite` attributes
- [ ] Login with remember-me and verify refresh token expiry is set appropriately

### 1.2.3 Token Refresh (`POST /api/v1/auth/refresh`)

- [ ] Refresh with valid refresh token cookie
- [ ] Verify old refresh token is revoked (rotation)
- [ ] Verify new refresh token is different from old (SHA-256 hash differs)
- [ ] Verify new access token is returned
- [ ] Reject refresh with expired refresh token (>7 days)
- [ ] Reject refresh with already-rotated (reused) refresh token
- [ ] Reject refresh with missing cookie
- [ ] Reject refresh with malformed/invalid JWT
- [ ] Reject refresh with access token instead of refresh token
- [ ] Verify concurrent refresh calls — only first succeeds, second gets revoked
- [ ] Verify rate limiting on refresh endpoint

### 1.2.4 Logout (`POST /api/v1/auth/logout`)

- [ ] Logout with valid session
- [ ] Verify refresh token is revoked (removed from DB)
- [ ] Verify `Set-Cookie` clears `botla_token` (max-age=0)
- [ ] Verify `Set-Cookie` clears `botla_refresh_token` (max-age=0)
- [ ] Verify subsequent API calls with cleared cookies return 401
- [ ] Logout without cookies (idempotent — should still succeed)
- [ ] Logout with already-expired token
- [ ] Verify rate limiting on logout endpoint

### 1.2.5 Protected Endpoint (`ANY /api/v1/protected`)

- [ ] Access protected endpoint with valid access token cookie
- [ ] Reject access with no cookie
- [ ] Reject access with expired access token
- [ ] Reject access with malformed JWT
- [ ] Reject access with tampered JWT (modified claims)
- [ ] Reject access with refresh token (not access token)
- [ ] Reject access from soft-deleted user
- [ ] Reject access with JWT signed by wrong secret
- [ ] Verify 401 response includes `WWW-Authenticate` header
- [ ] Verify 401 response has consistent error body format

### 1.2.6 Current User (`GET /api/v1/me`)

- [ ] Fetch current user with valid access token
- [ ] Verify response includes `id`, `email`, `fullName`, `avatarUrl`, `planId`, `languageCode`, `createdAt`
- [ ] Verify response excludes `passwordHash`, `deletedAt`
- [ ] Verify `isPlatformAdmin` field is boolean
- [ ] Verify `onboardingData` is included if onboarding was started
- [ ] Verify `languageCode` defaults to `"en"` if not set
- [ ] Fetch with expired token (should 401)
- [ ] Fetch after soft delete (should 401)
- [ ] Verify CORS headers allow `withCredentials`

### 1.2.7 Frontend Auth Source of Truth

- [ ] Verify frontend `/api/v1/me` response is cached as auth state
- [ ] Verify `/api/v1/me` is called on app mount
- [ ] Verify 401 from `/api/v1/me` triggers redirect to `/login`
- [ ] Verify refresh queue: concurrent 401s trigger single refresh call
- [ ] Verify session expiry modal on refresh failure

## 1.3 API Endpoints Table

| Method | Path | Handler | Expected Status Codes |
|--------|------|---------|----------------------|
| `POST` | `/api/v1/auth/register` | `ah.RegisterHandler` | 201 Created, 400 Bad Request (validation), 409 Conflict (duplicate email), 429 Too Many Requests |
| `POST` | `/api/v1/auth/login` | `ah.LoginHandler` | 200 OK, 401 Unauthorized (bad credentials), 403 Forbidden (deleted), 429 Too Many Requests |
| `POST` | `/api/v1/auth/refresh` | `ah.RefreshHandler` | 200 OK, 401 Unauthorized (expired/revoked), 429 Too Many Requests |
| `POST` | `/api/v1/auth/logout` | `ah.LogoutHandler` | 200 OK, 401 Unauthorized |
| `GET` | `/api/v1/me` | `mh.Me` | 200 OK, 401 Unauthorized |
| `ANY` | `/api/v1/protected` | `handlers.ProtectedHandler` | 200 OK, 401 Unauthorized |

## 1.4 Test Scenarios

### Scenario 1: Complete Registration → Login → Refresh → Logout Cycle

```
Step 1: Register a new user
  - [ ] POST /api/v1/auth/register with body:
        { "email": "test@example.com", "password": "ValidP@ss1", "fullName": "Test User" }
  - [ ] Assert: 201 Created
  - [ ] Assert: Response has Set-Cookie for botla_token and botla_refresh_token
  - [ ] Assert: Response body has user.id, user.email, user.fullName
  - [ ] Assert: User exists in DB with bcrypt password hash
  - [ ] Assert: Default org and workspace created in DB
  - [ ] Assert: Refresh token stored in DB (SHA-256)

Step 2: Verify access token works
  - [ ] GET /api/v1/me (with botla_token cookie)
  - [ ] Assert: 200 OK
  - [ ] Assert: email matches registered email

Step 3: Logout
  - [ ] POST /api/v1/auth/logout
  - [ ] Assert: 200 OK
  - [ ] Assert: Refresh token removed from DB
  - [ ] Assert: Cookies cleared

Step 4: Verify tokens are invalidated
  - [ ] GET /api/v1/me (with old botla_token cookie)
  - [ ] Assert: 401 Unauthorized
```

### Scenario 2: Token Refresh Race Condition (Concurrent Rotation)

```
  - [ ] Register and login to get tokens
  - [ ] Capture botla_refresh_token cookie value
  - [ ] Send 3 concurrent POST /api/v1/auth/refresh requests with the same cookie
  - [ ] Assert: Exactly 1 returns 200 OK with new tokens
  - [ ] Assert: Exactly 2 return 401 Unauthorized (revoked)
  - [ ] Assert: Only 1 refresh token exists in DB for this user
  - [ ] Assert: The winning refresh token hash matches the newest token
```

### Scenario 3: Password Strength Validation Matrix

```
Test each password variation:
  - [ ] "short" → 400 (min 8 chars)
  - [ ] "nouppercase1@" → 400 (no uppercase)
  - [ ] "NOLOWERCASE1@" → 400 (no lowercase)
  - [ ] "NoSpecialChar1" → 400 (no special char)
  - [ ] "NoDigitHere@" → 400 (no digit)
  - [ ] "Valid@123" → 201 (meets all: 9 chars, upper, lower, digit, special)
  - [ ] "A1@" + "x"×5 + "b" → 201 (exactly 8 chars, all rules met)
  - [ ] "  LeadingSpace1@" → 400 (leading whitespace in password is invalid)
```

### Scenario 4: Email Case-Insensitivity

```
  - [ ] Register with "TestUser@Example.COM"
  - [ ] Login with "testuser@example.com" — succeeds
  - [ ] Login with "TESTUSER@EXAMPLE.COM" — succeeds
  - [ ] Register with "TestUser@Example.COM" again — 409 Conflict
  - [ ] Verify DB stores lowercase email
```

### Scenario 5: Rate Limiting on Auth Endpoints

```
  - [ ] Send 6 registration requests in rapid succession from same IP
  - [ ] Assert: 6th request returns 429 Too Many Requests
  - [ ] Send 6 login requests with wrong password from same IP
  - [ ] Assert: 6th request returns 429
  - [ ] Wait for rate limit window to expire
  - [ ] Assert: Next request succeeds
```

### Scenario 6: Soft-Deleted User Cannot Authenticate

```
  - [ ] Register user, get tokens
  - [ ] Delete user account (privacy deletion)
  - [ ] Login with deleted user credentials → 403 Forbidden
  - [ ] Refresh with deleted user's refresh token → 401 Unauthorized
  - [ ] GET /api/v1/me with deleted user's access token → 401 Unauthorized
```

### Scenario 7: Frontend Auth Persistence & Expiry

```
  - [ ] Login via frontend UI
  - [ ] Reload page — verify user stays logged in (cookie-based)
  - [ ] Clear cookies via dev tools
  - [ ] Reload page — verify redirect to /login
  - [ ] Login again
  - [ ] Open second tab — verify both tabs share session
  - [ ] Logout from one tab
  - [ ] Verify other tab detects session termination (BroadcastChannel)
  - [ ] Verify other tab redirects to /login
```

## 1.5 Edge Cases

| Edge Case | Expected Behavior |
|-----------|------------------|
| SQL injection in email field | Parameterized query prevents injection, 400 Bad Request or 401 |
| XSS in fullName during registration | Field is sanitized/output-encoded; stored as-is but rendered safely |
| JWT token with `alg: none` | Rejected — server enforces `alg` verification |
| JWT token with algorithm confusion (RS256 vs HS256) | Rejected — server uses fixed signing method |
| Cookie with `__Host-` prefix requirements | Cookies are set with correct prefixes and `Secure` flag |
| CSRF attack on auth endpoints | SameSite=Lax prevents cross-site POST, no CSRF token needed |
| Refresh token stolen and used by attacker | Rotation detects reuse; attacker gets 401, legitimate user's token is also revoked |
| Multiple rapid registration of same email | Unique constraint on email (DB-level) returns 409 |
| Registration with null bytes in email | Sanitized to 400 Bad Request |
| Extremely long password (>72 chars bcrypt limit) | bcrypt truncates at 72 chars; password comparison still works |
| Login from blocked IP (geolocation/allowlist) | 403 Forbidden (if IP restriction enabled) |
| Browser without cookie support | Auth fails; frontend shows error message |

## 1.6 Expected Behavior

| Condition | HTTP Status | Response Body | Cookies |
|-----------|-------------|---------------|---------|
| Registration success | 201 | `{ "user": { "id", "email", "fullName", "createdAt" } }` | Set `botla_token` (1h), `botla_refresh_token` (7d) |
| Registration validation failure | 400 | `{ "error": { "code": "VALIDATION_ERROR", "message": "...", "fields": {...} } }` | None |
| Duplicate email | 409 | `{ "error": { "code": "EMAIL_EXISTS", "message": "Email already registered" } }` | None |
| Login success | 200 | `{ "user": { "id", "email", "fullName" } }` | Set `botla_token` (1h), `botla_refresh_token` (7d) |
| Login wrong password | 401 | `{ "error": { "code": "INVALID_CREDENTIALS", "message": "..."] } }` | None |
| Login deleted account | 403 | `{ "error": { "code": "ACCOUNT_DELETED", "message": "..." } }` | None |
| Refresh success | 200 | `{ "accessToken": "..." }` | Rotated `botla_refresh_token` (7d) |
| Refresh revoked/expired | 401 | `{ "error": { "code": "INVALID_TOKEN", "message": "..." } }` | Clear both cookies |
| Logout success | 200 | `{ "message": "Logged out" }` | Clear both cookies |
| Unauthenticated request | 401 | `{ "error": { "code": "UNAUTHORIZED", "message": "..." } }` | None |
