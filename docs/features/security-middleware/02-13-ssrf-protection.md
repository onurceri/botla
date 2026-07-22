> Extracted from: security-middleware.md — Section: 2.13 SSRF Protection (`pkg/urlutil/ssrf.go`)

## 2.13 SSRF Protection (`pkg/urlutil/ssrf.go`)

- [ ] Blocks requests to private IP ranges: `10.0.0.0/8`
- [ ] Blocks requests to private IP ranges: `172.16.0.0/12`
- [ ] Blocks requests to private IP ranges: `192.168.0.0/16`
- [ ] Blocks requests to `127.0.0.1` and `::1`
- [ ] Blocks requests to `169.254.0.0/16` (link-local)
- [ ] Blocks requests to `100.64.0.0/10` (CGNAT)
- [ ] Blocks metadata endpoints: AWS (`169.254.169.254`), GCP, Azure
- [ ] Blocks requests to `0.0.0.0`
- [ ] Supports allowlist for internal services (configured via env)
- [ ] URL is resolved before IP check (handles DNS rebinding)
- [ ] IPv6 private ranges are also blocked
- [ ] Redirects are followed but re-checked against blocklist
- [ ] Non-HTTP protocols (file://, ftp://, gopher://) are rejected
