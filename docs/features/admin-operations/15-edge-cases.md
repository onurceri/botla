> Extracted from: admin-operations.md — Section: 4. Edge Cases

# 4. Edge Cases

### 4.1 Admin — General

- [ ] **Non-existent user ID:** `GET /api/v1/admin/users/{uuid}` returns 404
- [ ] **Malformed UUID in path:** Route won't match `{id}` pattern; returns 404
- [ ] **Pagination edge cases:** `offset > total` returns empty array; `limit = 0` uses default; `offset = 0` returns first page; negative offset treated as 0
- [ ] **Concurrent admin actions:** Two admins updating same resource simultaneously — last write wins
- [ ] **Deleted user referenced in audit:** Audit log retains admin + user IDs even after deletion
- [ ] **Large result sets:** Pagination prevents unbounded response size

### 4.2 Admin — Forced Refresh / Reprocess

- [ ] **Chatbot with zero sources:** Force refresh returns `sources_reset: 0, sources_queued: 0, status: "refreshing"` — graceful
- [ ] **Queue unavailable during reprocess:** Source is still reset to pending but `queued: false` — manual retry possible
- [ ] **Qdrant unavailable during vector deletion:** Error logged but process continues (non-fatal)
- [ ] **Source already pending:** Reprocess while already pending — resets and re-queues

### 4.3 Admin — Privacy

- [ ] **Export URL with special characters:** File download — `Content-Disposition` filename sanitized
- [ ] **Storage service nil:** All download/URL endpoints return 503 Service Unavailable
- [ ] **Export at exactly expiry time:** Use `time.Now().After()` — at exact expiry, still valid (not yet after)
- [ ] **Approving already-completed request:** Should fail or be idempotent
- [ ] **Denying already-denied request:** Should be idempotent
- [ ] **Privacy request for non-existent user:** `POST /api/v1/admin/privacy/export/{userId}` with non-existent ID — service handles gracefully

### 4.4 Admin — Health Check

- [ ] **All dependencies down:** Status `unhealthy` (Postgres critical) or `degraded` (if Postgres still up)
- [ ] **Context timeout (10s):** Long-running checks cancelled; partial results returned
- [ ] **Redis cache with stale health data:** Cached response returned for up to 1 minute
- [ ] **Qdrant health endpoint unreachable:** Returns `down` — checked via HTTP GET to `/healthz`

---

