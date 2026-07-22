> Extracted from: security-middleware.md — Section: 2.1 Middleware Chain (processing order)

## 2.1 Middleware Chain (processing order)

- [ ] RequestID middleware runs first (outermost)
- [ ] SecurityHeaders middleware runs second
- [ ] Recovery middleware runs third
- [ ] RequestLogger middleware runs fourth
- [ ] MaxBytes middleware runs fifth (52 MB limit)
- [ ] PlanLoader middleware runs sixth
- [ ] DeletedAccount middleware runs seventh
- [ ] RateLimit middleware runs eighth
- [ ] Router/mux runs last (innermost)
