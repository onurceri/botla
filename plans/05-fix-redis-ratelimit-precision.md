# Fix 2.5 — Redis sliding-window rate limiter under-counts

- **Source:** `PROJECT_ISSUES_REPORT.md` §2.5
- **Severity:** Medium
- **Validation:** ✅ Confirmed TRUE
- **Type:** Logic bug / rate-limiting precision

## Context

The sliding-window-log limiter stores each request as a sorted-set member `now .. ':' .. i`. `now`
is passed with **second** precision, and `i` is a per-call loop index starting at 1. Two separate
requests in the same second (with the same `n`) generate an identical member (e.g. `1721000000:1`),
so `ZADD` overwrites instead of adding — the limiter under-counts and lets traffic bypass limits.

## Validated current state

`pkg/ratelimit/redis.go`
- Arguments passed to the script (**lines 85–90**): `now.Unix()` (seconds) and
  `windowStart.Unix()` (seconds).
- Lua member creation (**lines 56–60**):
  ```lua
  for i = 1, n do
      -- Use microsecond precision to handle concurrent requests
      local score = now + (i / 1000000)
      redis.call('ZADD', key, score, now .. ':' .. i)
  end
  ```
  The "microsecond precision" comment is misleading — `now` only has second precision, and the
  score's fractional `i/1000000` does not make the *member* unique.

## Plan

Make members globally unique. Two robust options (pick one):

- **Option A (preferred): nanosecond timestamp + per-request nonce.** Pass `now.UnixNano()` for the
  timestamp used in the member, and append a unique suffix so concurrent same-nanosecond requests
  still differ. Simplest: have the Go side pass a unique request id (e.g. a counter/UUID) and build
  the member as `now .. ':' .. reqid .. ':' .. i`.
- **Option B:** Keep seconds but add a per-call unique token to the member (`now .. ':' .. token .. ':' .. i`),
  where `token` is generated in Go per `Allow` call.

Regardless of option:
- Fix the misleading comment to describe the actual uniqueness mechanism.
- Confirm the score still represents the request time for window eviction (the `ZREMRANGEBYSCORE`
  cutoff uses `windowStart`); keep score in a consistent unit (if switching score to nanos, update
  the eviction comparison accordingly).

## Files to modify
- `pkg/ratelimit/redis.go`

## Verification
- `go build ./...`.
- Unit test (miniredis or real Redis): fire `limit+1` requests within the same second from the same
  key and assert the `(limit+1)`-th is rejected — this fails today.
- Test two distinct keys in the same second don't interfere, and window expiry still evicts old
  entries.
