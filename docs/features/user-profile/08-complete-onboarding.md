> Extracted from: user-profile.md — Section 1.2.7

# 1.2.7 Complete Onboarding (`POST /api/v1/me/onboarding/complete`)

- [ ] Mark onboarding as completed
- [ ] Verify `completed: true` and `currentStep` = max
- [ ] Complete when already skipped (idempotent)
- [ ] Complete when already completed (idempotent)
