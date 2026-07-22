> Extracted from: security-middleware.md — Section: 2.10 Request ID Middleware (`request_id.go`)

## 2.10 Request ID Middleware (`request_id.go`)

- [ ] Generates a unique request ID for every request (UUID v4)
- [ ] Sets `X-Request-ID` response header
- [ ] Propagates existing `X-Request-ID` from incoming request (if present)
- [ ] Request ID is available in request context for logging
- [ ] All log entries include the request ID
