> Extracted from: security-middleware.md — Section: 2.3 CORS Middleware (`cors.go`)

## 2.3 CORS Middleware (`cors.go`)

- [ ] `Access-Control-Allow-Origin` is set to configured origins
- [ ] Wildcard (`*`) is not used in production
- [ ] `Access-Control-Allow-Methods` includes GET, POST, PUT, DELETE, PATCH, OPTIONS
- [ ] `Access-Control-Allow-Headers` includes Authorization, Content-Type, X-Embed-Token
- [ ] `Access-Control-Allow-Credentials: true` for cookie-based auth
- [ ] `Access-Control-Max-Age` is set (caching preflight)
- [ ] Preflight (OPTIONS) requests are handled correctly
- [ ] Requests from disallowed origins are rejected with 403
- [ ] Requests with no Origin header from server-side clients are allowed
