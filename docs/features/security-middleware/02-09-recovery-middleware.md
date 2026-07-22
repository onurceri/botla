> Extracted from: security-middleware.md — Section: 2.9 Recovery Middleware (`recovery.go`)

## 2.9 Recovery Middleware (`recovery.go`)

- [ ] Panics are caught and logged with stack trace
- [ ] Returns 500 Internal Server Error to the client
- [ ] Does not leak stack trace to the client (production)
- [ ] Request ID is included in the error response
- [ ] Server continues running after panic (no crash)
