> Extracted from: security-middleware.md — Section: 2.4 Deleted Account Middleware (`deleted_account.go`)

## 2.4 Deleted Account Middleware (`deleted_account.go`)

- [ ] Checks if authenticated user's account is soft-deleted
- [ ] Returns 410 Gone for deleted accounts
- [ ] Response includes `X-Reason: account_deleted` header
- [ ] Non-authenticated routes are not affected
- [ ] Admin routes bypass the check for platform admins
