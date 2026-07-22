# Fix 2.2 — Thread-safety of `CircuitBreakerManager`

- **Source:** `PROJECT_ISSUES_REPORT.md` §2.2
- **Severity:** High
- **Validation:** ✅ Confirmed TRUE
- **Type:** Concurrency bug (data race → potential `concurrent map read/write` panic)

## Context

`CircuitBreakerManager` stores provider circuit breakers in a plain Go map with no synchronization.
Concurrent access from request paths and the health endpoint is a data race that can crash the
process with a concurrent map read/write panic. Guarding the map with a `sync.RWMutex` removes the
race.

## Validated current state

`internal/rag/circuit_breaker.go`
- Struct at **lines 135–139**, no mutex, no `sync` import:
  ```go
  type CircuitBreakerManager struct {
      breakers map[string]*CircuitBreakerClient
      metrics  *LLMMetrics
      settings CircuitBreakerSettings
  }
  ```
- Unsynchronized accessors:
  - `RegisterClient` (**160–164**) — writes `m.breakers[name] = cbClient`
  - `GetClient` (**167–170**) — reads `m.breakers[name]`
  - `GetAllStates` (**178–184**) — ranges over map
  - `GetAllProviderStatus` (**193–202**) — ranges over map

Concurrency exposure (nuance): `RegisterClient` is only called from `InitCircuitBreakers()`, which
runs eagerly at startup (`internal/api/router/router.go:54`), so steady-state writes are rare. The
real races are:
- `GetAllProviderStatus` (called from the health endpoint, `internal/api/handlers/health.go:52`)
  ranging the map concurrently with any register.
- A **lazy-init race** in `internal/rag/client_factory.go:73–78` (`GetCircuitBreakerClient`): if
  `cbManager` is nil, two concurrent callers can both run `InitCircuitBreakers()`, one writing while
  the other iterates.

The code is correctly flagged as unsafe regardless; it needs synchronization.

## Plan

1. Add a `mu sync.RWMutex` field to `CircuitBreakerManager` and import `sync`.
2. Lock in every accessor:
   - `RegisterClient`: `m.mu.Lock()` / `defer m.mu.Unlock()` around the map write
     (double-checked locking if it also reads-then-writes).
   - `GetClient`: `m.mu.RLock()` / `defer m.mu.RUnlock()`.
   - `GetAllStates`, `GetAllProviderStatus`: `m.mu.RLock()` for the range. Take a snapshot of the
     entries under the lock, then release before calling into each client (avoid holding the lock
     across potentially slow `cb.GetState()`/status calls).
3. Harden the lazy-init path in `client_factory.go` so `InitCircuitBreakers()` cannot run twice
   concurrently — `sync.Once` is the cleanest fix.

## Files to modify
- `internal/rag/circuit_breaker.go`
- `internal/rag/client_factory.go` (guard lazy init with `sync.Once`)

## Verification
- `go build ./...` and `go vet ./...`.
- Run the race detector over any RAG/health tests: `go test -race ./internal/rag/... ./internal/api/handlers/...`.
- Add a targeted test: spawn N goroutines calling `RegisterClient` + `GetClient` +
  `GetAllProviderStatus` concurrently; pass under `-race`.
