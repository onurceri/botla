> Extracted from: security-middleware.md — Section: 2.2 Auth Middleware (`auth.go`)

## 2.2 Auth Middleware (`auth.go`)

### 2.2.1 AuthMiddleware (required)

- [ ] Reads `botla_token` cookie from request
- [ ] Validates JWT signature and expiry
- [ ] Extracts user ID and org ID from JWT claims
- [ ] Sets user context on the request
- [ ] Returns 401 if cookie is missing
- [ ] Returns 401 if JWT is expired
- [ ] Returns 401 if JWT signature is invalid
- [ ] Returns 401 if JWT payload is malformed
- [ ] Returns 401 if user is not found in database
- [ ] Sets `X-User-Id` response header for debugging

### 2.2.2 OptionalAuthMiddleware

- [ ] Reads `botla_token` cookie if present
- [ ] If valid, sets user context (same as AuthMiddleware)
- [ ] If absent or invalid, proceeds without user context
- [ ] Does NOT return 401 on missing token
