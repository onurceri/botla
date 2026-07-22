> Extracted from: user-profile.md — Section 1.2.1

# 1.2.1 Profile Retrieval (`GET /api/v1/me`)

- [ ] Return user profile with all expected fields
- [ ] Return correct plan information (plan ID, not plan code)
- [ ] Return avatar URL when set (nullable)
- [ ] Return avatar URL as null when not set
- [ ] Return correct language code
- [ ] Return onboarding state data when available
- [ ] Exclude sensitive fields (password hash, deleted at)
- [ ] Return 401 when not authenticated
- [ ] Return 401 when user is soft-deleted
