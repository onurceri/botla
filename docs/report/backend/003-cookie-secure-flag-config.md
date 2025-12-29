# Backend Task 003: Make Cookie Secure Flag Configurable

## Background
In `internal/api/handlers/auth.go`, the `Secure` flag for cookies is hardcoded to `false`. This is acceptable for local development (HTTP) but insecure for production (HTTPS). It must be dynamically configured based on the environment or specific configuration.

**File:** `internal/api/handlers/auth.go`
**Location:** Lines 337-348

## Integration Plan
1.  **Update Config**
    - Check `pkg/config/config.go` for a `COOKIE_SECURE` or `ENVIRONMENT` variable (e.g., "production").
    - If missing, add `COOKIE_SECURE` boolean to config (default false).

2.  **Update Handlers**
    - In `auth.go` (`generateAndSendTokens` and `LogoutHandler`), read the secure constraint from configuration.
    - Set `Secure: cfg.CookieSecure` (or logic based on `cfg.Env == "production"`).

3.  **Verify**
    - Test in local environment (Secure=false) -> Login works.
    - Test conceptually for production (Secure=true) -> Cookies require HTTPS.

## Checklist
- [x] Add `CookieSecure` field to `Config` struct if needed, or use existing Env check
- [x] Update `generateAndSendTokens` in `internal/api/handlers/auth.go`
- [x] Update `LogoutHandler` cookie clearing logic to match
- [x] Verify `SameSite` attribute is also appropriate (StrictMode is good)
