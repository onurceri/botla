# Backend Security Assessment Report

**Date:** 2025-12-24
**Scope:** Public API Endpoints
**Status:** Pre-Production Review

---

## Executive Summary

This security assessment reviewed all public-facing API endpoints in the Botla backend. The authentication system is well-implemented with proper token security, rotation, and revocation. However, several security gaps were identified that should be addressed before production deployment.

**Risk Level:** Medium

---

## Public Endpoints Reviewed

| Endpoint | Handler | Auth Required |
|----------|---------|---------------|
| `POST /api/v1/auth/register` | `handlers/auth.go:60` | No |
| `POST /api/v1/auth/login` | `handlers/auth.go:148` | No |
| `POST /api/v1/auth/refresh` | `handlers/auth.go:183` | No (token-based) |
| `POST /api/v1/auth/logout` | `handlers/auth.go:223` | No (token-based) |
| `POST /api/v1/public/chatbots/{id}/chat` | `handlers/public.go:166` | No |
| `POST /api/v1/public/chatbots/{id}/feedback` | `handlers/public.go:338` | No |
| `POST /api/v1/public/chatbots/{id}/handoff` | `handlers/handoff.go:37` | No |
| `POST /api/v1/public/chatbots/{id}/handoff/{requestId}/contact` | `handlers/handoff.go:151` | No |
| `GET /api/v1/public/chatbots/{id}/config` | `handlers/public.go:49` | No |
| `GET /health` | `handlers/health.go:19` | No |

---

## Security Measures - Well Implemented

### 1. JWT Authentication (`internal/auth/jwt.go`)

| Measure | Implementation |
|---------|---------------|
| Signing Algorithm | HS256 |
| Token ID (JTI) | 16-byte random hex |
| Issuer Validation | `"botla-co"` |
| Audience Validation | `"botla-api"` |
| Access Token TTL | 1 hour |
| Refresh Token TTL | 7 days |

### 2. Token Security (`internal/api/handlers/auth.go`)

| Measure | Implementation |
|---------|---------------|
| Refresh Token Storage | SHA-256 hashed (not plaintext) |
| Token Rotation | Old token revoked on refresh |
| Revocation Check | Database lookup on every refresh |
| Access Control | TokenType claim validation (`"access"` vs `"refresh"`) |

### 3. Rate Limiting (`pkg/middleware/ratelimit.go`)

| Tier | Limit | Window |
|------|-------|--------|
| Global (IP-based) | 60 req/min | 60s |
| Authenticated Users | 120 req/min | 60s |
| Chat Endpoint (`/api/v1/chat`) | 20 req/min | 60s |
| Sources Endpoint (`/api/v1/sources`) | 10 req/min | 60s |

**Backend:** Redis (with memory fallback for dev)

### 4. Secure Embed Validation (`handlers/public.go:190-257`)

Domain validation uses proper URL parsing to prevent hostname suffix bypass:

```go
parsed, err := url.Parse(origin)
hostname := parsed.Hostname()
// Correctly rejects: https://example.com.evil.com
if hostname == d || strings.HasSuffix(hostname, "."+d)
```

Embed token validation includes `chatbot_id` claim verification.

### 5. CORS Configuration (`pkg/middleware/cors.go`)

- Whitelist-based origin validation
- Vary: Origin header set
- Credentials support enabled

### 6. Input Validation

| Endpoint | Validation |
|----------|------------|
| Registration | `mail.ParseAddress()` for email |
| Login | Case-insensitive email lookup |
| Chat | UUID validation, message trimming |
| Feedback | Message ID verification |
| Config | UUID validation, cache invalidation |

---

## Security Gaps Identified

### HIGH Severity

#### 1. Missing Security Headers

**Location:** Global middleware chain (`cmd/server/main.go:167`)

**Missing Headers:**
- `X-Frame-Options` - Prevents clickjacking
- `X-Content-Type-Options` - Prevents MIME sniffing
- `Content-Security-Policy` - Prevents XSS
- `Strict-Transport-Security` (HSTS) - Enforces HTTPS
- `Referrer-Policy` - Controls referrer information

**Recommendation:** Add security headers middleware:

```go
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("X-Frame-Options", "DENY")
            w.Header().Set("X-Content-Type-Options", "nosniff")
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
            if r.TLS != nil {
                w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

**Priority:** Critical for production

---

#### 2. No Request Body Size Limits

**Location:** All handlers (`handlers/auth.go`, `handlers/public.go`, `handlers/handoff.go`)

**Risk:** DoS via large request bodies

**Recommendation:** Add `http.MaxBytesReader` to all handlers:

```go
// Add at start of handlers that accept JSON bodies
r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1MB limit
```

**Priority:** Critical for production

---

### MEDIUM Severity

#### 3. Weak Password Policy

**Location:** `handlers/auth.go:80`

**Current Implementation:**
```go
if len(req.Password) < 8 {
    respondError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
    return
}
```

**Issue:** Only enforces minimum length, no complexity requirements.

**Recommendation:**
```go
var passwordRegex = regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$`)

if !passwordRegex.MatchString(req.Password) {
    respondError(w, http.StatusBadRequest, "Password must contain at least 8 characters, including uppercase, lowercase, number, and special character")
    return
}
```

**Priority:** Should implement before production

---

#### 4. Permissive Email Validation in Handoff

**Location:** `handlers/handoff.go:175`

**Current Implementation:**
```go
if req.Email == "" || !strings.Contains(req.Email, "@") {
    api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "valid email is required"})
    return
}
```

**Issue:** Only checks for `@` character. Accepts invalid emails like `@@@` or `a@`.

**Recommendation:** Use `net/mail` package like registration handler:
```go
if _, err := mail.ParseAddress(req.Email); err != nil {
    api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "valid email is required"})
    return
}
```

**Priority:** Should implement before production

---

#### 5. No CAPTCHA on Auth Endpoints

**Location:** `routes_auth.go`

**Risk:** Brute force attacks on `/auth/login` and `/auth/register`

**Current Mitigation:**
- Rate limiting (60 req/min global)
- Generic error messages ("Invalid credentials")

**Recommendation Options:**
1. Add CAPTCHA (e.g., hCaptcha, reCAPTCHA) for register/login
2. Implement progressive delays after failed attempts
3. Stricter rate limits for auth endpoints (5 req/min with backoff)

**Priority:** Recommended for production

---

#### 6. Public Config Endpoint Data Exposure

**Location:** `handlers/public.go:49-134`

**Endpoint:** `GET /api/v1/public/chatbots/{id}/config`

**Exposed Data:**
- Theme colors, welcome message
- Bot display name, icon
- Chat appearance settings
- Custom branding (if hide_branding is true)
- Suggested questions
- Handoff enabled status

**Assessment:** This is **intentional** - the widget needs this configuration. However, it should be documented that chatbot appearance configurations are publicly accessible.

**Priority:** Informational - Document as expected behavior

---

### LOW Severity

#### 7. Insecure CORS Function Exists

**Location:** `pkg/middleware/cors.go:8`

**Function:** `CORSMiddleware(origin string)` allows any origin if `origin` is `*` or empty.

**Current Usage:** `CORSMiddlewareAllowOrigins` is used in production (`cmd/server/main.go:164`)

**Risk:** If `CORSMiddleware` is accidentally used with `*`, it could allow cross-origin attacks.

**Recommendation:** Remove `CORSMiddleware` function or add comment warning against production use.

---

#### 8. No Not-Before (nbf) Claim Validation

**Location:** `internal/auth/jwt.go`

**Current:** Only `IssuedAt` and `ExpiresAt` are set.

**Note:** This is a minor concern. JWT `nbf` claim is rarely used and not required for standard authentication flows.

---

## Middleware Chain Review

**Location:** `cmd/server/main.go:167`

```go
handler := middleware.RecoveryMiddleware(app.log)(
    middleware.RequestLogger(app.log)(
        planLoader(
            middleware.RateLimitMiddleware(app.rateLimiter)(mux)
        )
    )
)
```

**Order Assessment:**
1. ✅ Recovery first (catches panics)
2. ✅ Request logging (auditing)
3. ✅ Plan loader (business logic)
4. ✅ Rate limiting (DoS protection)
5. Router

**Note:** Security headers should be added at the very beginning (before recovery) or end (after all handlers) of this chain.

---

## Authentication Security Checklist

| Check | Status |
|-------|--------|
| Password hashed with bcrypt | ✅ |
| Refresh tokens hashed before storage | ✅ |
| Token rotation implemented | ✅ |
| Token revocation check | ✅ |
| JWT signature verification | ✅ |
| Issuer/audience validation | ✅ |
| Access/refresh token type separation | ✅ |
| Short-lived access tokens (1hr) | ✅ |
| Generic error messages | ✅ |
| Rate limiting on auth endpoints | ✅ (but could be stricter) |
| Account lockout on repeated failures | ❌ Not implemented |

---

## Public Chat Security Checklist

| Check | Status |
|-------|--------|
| UUID validation for chatbot ID | ✅ |
| Secure embed domain validation | ✅ |
| Embed token validation | ✅ |
| Chatbot ownership check | ✅ |
| Plan-based limits enforcement | ✅ |
| Monthly token limits | ✅ |
| Message trimming | ✅ |
| Session ID validation | ✅ |
| Request timeout (context) | ✅ |
| Rate limiting | ✅ |

---

## Recommendations Summary

### Before Production Deployment (Critical)

1. **Add Security Headers Middleware**
2. **Add Request Body Size Limits** (1MB recommended)
3. **Strengthen Password Policy**

### Before or After Production (Recommended)

4. **Fix Email Validation** in handoff endpoint
5. **Add CAPTCHA or Stricter Rate Limits** for auth endpoints
6. **Remove or Document** insecure CORS function
7. **Document** public config endpoint exposure

---

## Testing Recommendations

1. **Authentication Tests:**
   - Brute force login attempts
   - Token replay attacks
   - JWT forgery attempts

2. **Rate Limiting Tests:**
   - Bypass via X-Forwarded-For header
   - Concurrent requests at limit

3. **Input Validation Tests:**
   - SQL injection attempts
   - Malformed JSON bodies
   - Oversized request bodies
   - Invalid UUID formats

4. **CORS Tests:**
   - Origin header bypass attempts
   - Credentials with wildcard origin

---

## Conclusion

The backend has a solid foundation with proper JWT implementation, token rotation, and rate limiting. The secure embed feature demonstrates good security awareness with proper domain validation.

**Key actions for production:**
1. Add security headers
2. Implement request body limits
3. Strengthen password requirements

These are straightforward fixes that significantly improve the security posture without major architectural changes.

---

*Report generated for internal use. Review again before major code changes.*
