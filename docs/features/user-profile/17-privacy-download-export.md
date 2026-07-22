> Extracted from: user-profile.md — Section 1.2.16

# 1.2.16 Privacy Request — Download Export (`GET /api/v1/me/privacy/requests/{id}/download` or `GET /api/v1/me/privacy/exports/{id}/download`)

- [ ] Download completed export as JSON/ZIP file
- [ ] Return 404 when export not yet completed
- [ ] Return 404 for non-existent export
- [ ] Return 404 for another user's export
- [ ] Verify correct Content-Type and Content-Disposition headers
