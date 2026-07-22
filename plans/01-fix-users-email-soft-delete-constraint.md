# Fix 2.1 — Email unique constraint blocks re-signup of soft-deleted users

- **Source:** `PROJECT_ISSUES_REPORT.md` §2.1
- **Severity:** Medium-High
- **Validation:** ✅ Confirmed TRUE
- **Type:** Database constraint / logic defect

## Context

The platform supports GDPR soft-deletion (`users.deleted_at`). Because the email uniqueness is
enforced with a plain `UNIQUE(email)` constraint, a soft-deleted row still occupies its email, so a
later signup with the same email fails. Fixing this unblocks legitimate re-registration while
preserving uniqueness among *active* users.

## Validated current state

`db/migrations/000001_init.up.sql`
- `users` table defined at **lines 486–506**; `deleted_at timestamp` at **line 500**.
- Plain unique constraint at **lines 601–602**:
  ```sql
  ALTER TABLE ONLY users
      ADD CONSTRAINT users_email_key UNIQUE (email);
  ```
- The only other migration (`000002_add_privacy_requests_deleted_at.up.sql`) touches
  `privacy_requests`, not `users`. No partial index exists today.

Soft-delete is actively used: `internal/repository/user_repo.go` filters `deleted_at IS NULL` in 8
places (lines 74, 110, 173, 197, 232, 252, 292, 310).

**Bonus discovery (must fix together):** the registration pre-check in
`internal/api/handlers/auth.go` **lines 116–121** also ignores soft-deleted rows:
```go
err := h.DB.QueryRowContext(r.Context(), "SELECT id FROM users WHERE email=$1", req.Email).Scan(&existing)
if err == nil && existing != "" {
    api.WriteErrorCode(w, http.StatusConflict, api.ErrEmailExists)
    return
}
```
Today this pre-check masks the DB constraint (it returns 409 before the INSERT). Both the constraint
and this query must change together, or the bug persists.

## Plan

1. **New migration** `db/migrations/000003_users_email_partial_unique.up.sql`:
   ```sql
   ALTER TABLE users DROP CONSTRAINT users_email_key;
   CREATE UNIQUE INDEX users_email_key ON users (email) WHERE deleted_at IS NULL;
   ```
   And a matching `.down.sql` that drops the index and restores the plain constraint.
   > Keep the index named `users_email_key` so existing conflict-handling code paths that may
   > reference the constraint name are unaffected.

2. **`internal/api/handlers/auth.go` (lines 116–121):** add `AND deleted_at IS NULL` to the
   pre-check query so soft-deleted emails are treated as available:
   ```go
   "SELECT id FROM users WHERE email=$1 AND deleted_at IS NULL"
   ```

3. Confirm the INSERT error path (~line 136) still maps `pgerrcode 23505` to `ErrEmailExists`
   (now only reachable for genuinely-active duplicates).

## Files to modify
- `db/migrations/000003_users_email_partial_unique.up.sql` (new)
- `db/migrations/000003_users_email_partial_unique.down.sql` (new)
- `internal/api/handlers/auth.go`

## Verification
- Run migrations up/down cleanly: `migrate -path db/migrations -database "$DSN" up` then `down 1`.
- Insert a user, soft-delete it (`UPDATE users SET deleted_at=NOW() WHERE email=...`), then sign up
  again with the same email → should succeed.
- With an *active* user present, signing up with the same email → 409 `ErrEmailExists`.
- Add/adjust an auth handler test covering soft-deleted re-signup.
