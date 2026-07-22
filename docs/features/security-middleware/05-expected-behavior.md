> Extracted from: security-middleware.md — Section: 5. Expected Behavior

# 5. Expected Behavior — Success/Failure Conditions

## 5.1 Security & Middleware

| Condition | Success Indicator | Failure Indicator |
|---|---|---|
| Valid JWT | 200, user context set | 401, access denied |
| Expired JWT | 401, "token expired" | 500, panic |
| Invalid JWT | 401, "invalid token" | 200, unauthorized access |
| Rate limit OK | 200, headers show remaining | 429 not triggered when it should be |
| Rate limit exceeded | 429, Retry-After header | 200, abuse possible |
| Allowed CORS origin | ACAO header matches origin | Missing ACAO, browser blocks |
| Disallowed CORS origin | 403 or no ACAO header | ACAO: *, credentials leak |
| Body under limit | Processed normally | 413 when should be allowed |
| Body over limit | 413 Payload Too Large | Processed, memory exhaustion |
| SSRF blocked | Error, "private IP blocked" | Request to internal service succeeds |
| SSRF allowed (allowlist) | Request proceeds | Request blocked incorrectly |
| Deleted account | 410, X-Reason header | 200, deleted user access |
| Panic in handler | 500, request ID, server stays up | Stack trace leaked, server crashes |
| Security headers sent | Headers present on response | Headers missing, security risk |
