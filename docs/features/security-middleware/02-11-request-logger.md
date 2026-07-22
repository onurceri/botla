> Extracted from: security-middleware.md — Section: 2.11 Request Logger (`requestlog.go`)

## 2.11 Request Logger (`requestlog.go`)

- [ ] Every request is logged with: method, path, status, duration, request ID
- [ ] Response status code is captured
- [ ] Request duration is measured in milliseconds
- [ ] Slow requests (> 1s) are logged at WARN level
- [ ] Error responses (4xx, 5xx) include additional context in logs
- [ ] Request body is NOT logged (privacy)
- [ ] Sensitive headers (Authorization, Cookie) are redacted
