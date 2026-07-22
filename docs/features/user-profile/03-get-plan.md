> Extracted from: user-profile.md — Section 1.2.2

# 1.2.2 Plan Information (`GET /api/v1/me/plan`)

- [ ] Return current plan with ID, name, code, and limits
- [ ] Return plan limits structure with all entitlement fields
- [ ] Verify rate limit values in response match plan config
- [ ] Return 401 when not authenticated
- [ ] Return cached plan data (Redis) without hitting DB on subsequent calls
