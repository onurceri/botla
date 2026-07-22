# Fix 3.4 — Scope `FOR UPDATE` to the message row only

- **Source:** `PROJECT_ISSUES_REPORT.md` §3.4
- **Severity:** Medium
- **Validation:** ✅ Confirmed TRUE
- **Type:** Database locking & concurrency

## Context

The feedback transaction uses a bare `FOR UPDATE` on a `JOIN` between `messages` and `conversations`.
In PostgreSQL, a bare `FOR UPDATE` locks rows from **every** table in the query, so it unnecessarily
locks the parent conversation row and creates contention with concurrent updates to that
conversation (e.g. `IncrementMessageCount` touching `conversations.updated_at`). Only the message is
updated, so the lock should target `messages` alone.

## Validated current state

`internal/repository/analytics_repo.go` **lines 461–499**; the lock is at **line 475**:
```sql
SELECT m.thumbs_up, c.chatbot_id
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE m.id = $1 FOR UPDATE
```
Followed by `UPDATE messages SET thumbs_up = $2 WHERE id = $1` (**lines 481–484**). This is the
**only** `FOR UPDATE` usage in the repository layer.

## Plan

Change the locking clause to lock only the message row:
```sql
SELECT m.thumbs_up, c.chatbot_id
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE m.id = $1 FOR UPDATE OF m
```
No other changes needed.

## Files to modify
- `internal/repository/analytics_repo.go`

## Verification
- `go build ./...`.
- Unit/integration test: two concurrent feedback updates on different messages of the *same*
  conversation should no longer block each other (previously both locked the conversation row).
- Regression test: feedback update still returns correct prior `thumbs_up` and `chatbot_id`, and the
  `messages.thumbs_up` update persists.
