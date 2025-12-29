# Backend Task 005: Propagate Worker Pool Configuration

## Background
In `cmd/server/main.go`, the analytics worker pool is initialized with a hardcoded size of 10. The application already has a `WORKER_COUNT` configuration (used for source queue), which should ostensibly apply here as well, or a separate `ANALYTICS_WORKER_COUNT` should be introduced.

**File:** `cmd/server/main.go`
**Location:** Line 146

## Integration Plan
1.  **Review Config**
    - Check if `pkg/config` has a suitable field. If reusing `WORKER_COUNT` is appropriate (shared resource limit), use that.
    - If specific control is needed, add `ANALYTICS_WORKER_COUNT` to config.

2.  **Update Main**
    - Replace `workers.NewWorkerPool(log, 10)` with `workers.NewWorkerPool(log, cfg.WORKER_COUNT)` (or the new config field).

3.  **Verify**
    - Verify application starts.
    - Check logs for worker pool initialization if available.

## Checklist
- [ ] Choose appropriate config variable (`WORKER_COUNT` or new one)
- [ ] Update `cmd/server/main.go` initialization
- [ ] Verify build and run
