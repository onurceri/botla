> Extracted from: security-middleware.md — Section: 1. Feature Overview

# 1. Feature Overview

A layered HTTP middleware chain providing authentication (JWT HttpOnly cookies), CORS enforcement, rate limiting (Redis sliding window), SSRF protection, request body size limits (52 MB), security headers (CSP, HSTS, X-Frame-Options), panic recovery, structured request logging, deleted-account blocking, plan-based access control, and multi-tenant context extraction.
