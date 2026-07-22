# Fix 2.4 — `BrowserPool` stuck at size 1 after idle reaping

- **Source:** `PROJECT_ISSUES_REPORT.md` §2.4
- **Severity:** Medium
- **Validation:** ✅ Confirmed TRUE
- **Type:** Logic bug / resource utilization

## Context

After the idle reaper clears the pool, `nextBrowser()` re-creates only a single browser and returns,
so a pool configured for N browsers is permanently stuck at 1 until restart. All N concurrent
callers (the semaphore still admits N) then share one browser, adding contention. The pool should
regrow to its configured size.

## Validated current state

`internal/scraper/browser.go`
- Struct at **lines 16–26** — note there is **no stored `size` field**; the configured size is only
  used to size the `sem` channel (**line 33**) at construction.
- Reaper (**53–81**): on idle timeout sets `bp.browsers = nil` and closes old browsers; keeps running.
- `nextBrowser()` (**101–119**): when `len(bp.browsers) == 0`, launches **one** browser,
  `bp.browsers = append(bp.browsers, b)`, resets `bp.idx = 0`, returns. Subsequent calls round-robin
  over that single browser (`idx % 1 == 0`).

Pool size config: `NewBrowserPool` defaults `size` to 2 if `<= 0` (**lines 29–31**); production uses
`cfg.SCRAPER_BROWSER_POOL_SIZE` (`main.go:201`, default 2 in `pkg/config/config.go:180`);
integration tests use 10 (`internal/integration/realservice.go:150`). So the bug degrades 2→1 in
prod, 10→1 in tests.

## Plan

1. Store the configured size on the struct: add `size int` to `BrowserPool`, set it in
   `NewBrowserPool`.
2. Make browser creation reusable: extract a `launchBrowser() *rod.Browser` helper from the current
   inline launcher logic in `nextBrowser()`.
3. In `nextBrowser()`, when `len(bp.browsers) == 0`, regrow up to `bp.size` (call `launchBrowser()`
   in a loop and append), instead of appending just one. Keep this under `bp.mu` as today.
   - Alternatively, lazily grow on demand (append one per call until `len == size`); either is fine,
     but re-initializing to full size best matches the original intent.
4. Ensure failed launches don't leave the pool wedged — if a launch panics/errors, surface it rather
   than returning a nil browser.

## Files to modify
- `internal/scraper/browser.go`

## Verification
- `go build ./...`.
- Unit test: create a pool with `size=3`, force a reap (set `lastUse` far in the past / trigger the
  reaper), then call `nextBrowser()` and assert `len(bp.browsers) == 3`.
- Integration: run the scraper integration path (pool size 10) and confirm multiple browsers are
  created after an idle cycle.
