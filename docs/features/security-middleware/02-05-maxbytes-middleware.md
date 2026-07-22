> Extracted from: security-middleware.md — Section: 2.5 MaxBytes Middleware (`maxbytes.go`)

## 2.5 MaxBytes Middleware (`maxbytes.go`)

- [ ] Request body is limited to 52 MB
- [ ] Requests exceeding the limit return 413 Payload Too Large
- [ ] Limit is configurable via environment variable
- [ ] File upload endpoints have separate (larger) limits if configured
