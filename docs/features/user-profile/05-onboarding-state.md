> Extracted from: user-profile.md — Section 1.2.4

# 1.2.4 Onboarding State (`GET /api/v1/me/onboarding`)

- [ ] Return onboarding state as `{ "completed": false, "currentStep": 0, "skipped": false }` for new users
- [ ] Return current step after partial progress
- [ ] Return `"completed": true` after completion
- [ ] Return `"skipped": true` after skip
- [ ] Return complete `onboardingData` with all step responses
