# Fix 3.2 — Bulk-insert in `SaveMessageSources` (N-roundtrip bottleneck)

- **Source:** `PROJECT_ISSUES_REPORT.md` §3.2
- **Severity:** Medium
- **Validation:** ✅ Confirmed TRUE (+ bonus schema-drift finding)
- **Type:** Database performance / roundtrip reduction

## Context

`SaveMessageSources` executes one `INSERT` per source inside a loop. Even within a transaction, each
`ExecContext` is a separate network roundtrip, so a message referencing 5–10 chunks costs 5–10
sequential roundtrips. Converting to a single multi-value `INSERT` removes the per-row roundtrips.

## Validated current state

`internal/repository/conversation_repo.go` **lines 205–240**. Inside a tx (`BeginTx` +
`PrepareContext`), it loops:
```go
for _, src := range sources {                 // line 226
    if src.SourceID == "" { continue }
    _, err = stmt.ExecContext(ctx, messageID, src.SourceID, src.ChunkIndex, src.SourceType, src.Score)
    ...
}
```
Target statement inserts into `message_sources (message_id, source_id, chunk_index, source_type,
relevance_score)` with `ON CONFLICT (message_id, source_id, chunk_index) DO NOTHING`.

Important nuances:
- Driver is `database/sql` (pgx stdlib), so **`pgx.CopyFrom` is not available** on this conn type.
  Use a **multi-value `INSERT`** (optionally built with the squirrel builder already used across the
  repo) rather than `CopyFrom`.
- No existing bulk-insert pattern exists in the repo to copy from.

### ⚠️ Bonus discovery — possible schema drift
The INSERT writes a `source_type` column, but the `message_sources` table in the migration
(`000001_init.up.sql` lines 232–239) **does not define `source_type`**. Either the live DB schema
has diverged from migrations, or this INSERT fails at runtime. **Resolve this before/while doing the
bulk conversion** — verify the real column set with `\d message_sources` and add the missing column
via migration if needed.

## Plan

1. **Verify/repair schema:** confirm `message_sources` columns in the actual DB; if `source_type`
   is missing, add a migration to create it (or remove it from the INSERT if unused).
2. **Rewrite the insert as a single multi-value statement.** Using squirrel (repo convention):
   ```go
   insert := sq.Insert("message_sources").
       Columns("message_id", "source_id", "chunk_index", "source_type", "relevance_score").
       Suffix("ON CONFLICT (message_id, source_id, chunk_index) DO NOTHING")
   for _, src := range sources {
       if src.SourceID == "" { continue }
       insert = insert.Values(messageID, src.SourceID, src.ChunkIndex, src.SourceType, src.Score)
   }
   // run once against the tx
   ```
   Keep the existing tx wrapper (atomicity is still desirable) or, since it's one statement, run it
   directly — either is fine.
3. Keep the `len(sources) == 0` early return.

## Files to modify
- `internal/repository/conversation_repo.go`
- Possibly a new migration under `db/migrations/` if `source_type` must be added

## Verification
- `go build ./...`.
- Unit/integration test: save a message with N sources, assert all N rows land and that duplicate
  `(message_id, source_id, chunk_index)` are ignored (ON CONFLICT DO NOTHING).
- Benchmark/compare roundtrips before vs. after (query log or `pg_stat_statements`) to confirm a
  single statement.
