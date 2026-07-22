> Extracted from: security-middleware.md — Section: 4. Edge Cases

# 4. Edge Cases

## 4.1 Security & Middleware Edge Cases

| # | Edge Case | Expected Behavior |
|---|---|---|
| 4.1.1 | JWT signed with wrong algorithm (alg:none) | Rejected; 401 returned |
| 4.1.2 | JWT with extremely long payload | Rejected (size validation); 401 returned |
| 4.1.3 | Multiple `botla_token` cookies | Only the first valid one is used or all are rejected |
| 4.1.4 | Cookie with `botla_token` but wrong path | Cookie not sent by browser; treated as missing |
| 4.1.5 | Rate limit — burst followed by idle | Burst is limited; after idle window, limits reset |
| 4.1.6 | Rate limit — distributed attack from many IPs | Global rate limit provides protection |
| 4.1.7 | CORS — Origin header with port mismatch (`https://example.com:3000` vs `https://example.com`) | Port is considered part of origin; match must be exact |
| 4.1.8 | CORS — null origin (file://, data://) | Rejected or handled based on configuration |
| 4.1.9 | SSRF — Domain that resolves to multiple IPs (round-robin DNS) | All resolved IPs are checked; if any is private, blocked |
| 4.1.10 | SSRF — Domain using DNS rebinding protection | URL is resolved once and the IP is used for the connection |
| 4.1.11 | SSRF — IPv6 mapped IPv4 addresses (`::ffff:10.0.0.1`) | Detected as private IPv4; blocked |
| 4.1.12 | MaxBytes — Request body exactly 52 MB | Accepted (boundary test) |
| 4.1.13 | MaxBytes — Request body 52 MB + 1 byte | Rejected with 413 |
| 4.1.14 | Panic recovery — panic in middleware itself | Recovery middleware still catches it (it wraps all downstream code) |
| 4.1.15 | Audit log — concurrent admin actions | Each action gets its own log entry; no race conditions |
| 4.1.16 | Audit log — database unavailable | Action is logged to stderr as a fallback; admin action still proceeds? (depends on configuration) |
| 4.1.17 | Password — extremely long (> 72 chars) | bcrypt truncates; validation may warn (policy decision) |
| 4.1.18 | Plan config — invalid JSON in feature flags | Startup fails; system does not boot with invalid config |
| 4.1.19 | Plan config — missing rate limit values | Default values are used; safe defaults |
| 4.1.20 | Request body with transfer-encoding: chunked | Processed correctly; MaxBytes still applies to total decoded size |
| 4.1.21 | Request ID collision (UUID v4, extremely unlikely) | Collision probability is negligible; no special handling needed |
| 4.1.22 | Middleware panics with `http.CloseNotifier` deprecation | Recovery handles the panic; response is 500 |
