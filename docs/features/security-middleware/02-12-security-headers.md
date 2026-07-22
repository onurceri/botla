> Extracted from: security-middleware.md — Section: 2.12 Security Headers (`security.go`)

## 2.12 Security Headers (`security.go`)

- [ ] `Content-Security-Policy` header is set (configurable)
- [ ] `Strict-Transport-Security` header is set for HTTPS enforcement
- [ ] `X-Content-Type-Options: nosniff` is set
- [ ] `X-Frame-Options: DENY` is set (prevents clickjacking)
- [ ] `Referrer-Policy: strict-origin-when-cross-origin` is set
- [ ] `Permissions-Policy` header restricts feature access (camera, microphone, etc.)
- [ ] `Cache-Control: no-store` for authenticated responses
- [ ] Custom security headers can be added via configuration
