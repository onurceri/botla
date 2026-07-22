> Extracted from: user-profile.md — Section 1.2.12

# 1.2.12 Privacy Request — Account Deletion (`POST /api/v1/me/privacy/delete`)

- [ ] Create account deletion request
- [ ] Verify request appears in list with status `pending`
- [ ] Verify user is NOT immediately deleted (soft-delete workflow)
- [ ] Verify user cannot authenticate after deletion request is processed
- [ ] Reject deletion request if one is already pending
