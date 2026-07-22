# Fix 3.1 — Database connection pool tuning

- **Source:** `PROJECT_ISSUES_REPORT.md` §3.1
- **Severity:** Medium-High
- **Validation:** ✅ Confirmed TRUE
- **Type:** Database performance

## Context

With `MaxOpenConns(25)` but `MaxIdleConns(5)`, bursty load opens up to 25 connections and then
closes 20 as soon as traffic dips — TCP connection churn that raises latency and DB CPU. No
`SetConnMaxLifetime`/`SetConnMaxIdleTime` is set anywhere, so connections are never recycled,
risking silent drops by load balancers/firewalls.

## Validated current state

`internal/db/db.go` **lines 26–41** (driver is `github.com/jackc/pgx/v5/stdlib` via `database/sql`,
so these are the correct knobs — not `pgxpool`):
```go
conn, err := sql.Open("pgx", dsn)
...
conn.SetMaxOpenConns(25)   // line 32
conn.SetMaxIdleConns(5)    // line 33
```
`grep` confirms **no** `SetConnMaxLifetime` / `SetConnMaxIdleTime` anywhere in the codebase.

## Plan

In `New()` (after `SetMaxIdleConns`):
```go
conn.SetMaxOpenConns(25)
conn.SetMaxIdleConns(15)
conn.SetConnMaxLifetime(30 * time.Minute)
conn.SetConnMaxIdleTime(5 * time.Minute)
```
- Raising `MaxIdleConns` toward `MaxOpenConns` keeps warm connections across bursts.
- `ConnMaxLifetime` guards against stale/LB-dropped connections; `ConnMaxIdleTime` reclaims unused
  ones.
- Consider sourcing these from `config.Config` (with the above as defaults) so they're tunable per
  environment without a code change.

## Files to modify
- `internal/db/db.go`
- Optionally `pkg/config/config.go` (if making the values configurable)

## Verification
- `go build ./...`.
- Boot the service against a local Postgres; under a burst (e.g. a load script hitting a DB-backed
  endpoint), watch `db.sql` pool stats (`sql.DBStats`) to confirm `Idle` stays near the configured
  floor and `Open` doesn't thrash.
- Confirm connections recycle after the lifetime by checking `pg_stat_activity` churn over time.
