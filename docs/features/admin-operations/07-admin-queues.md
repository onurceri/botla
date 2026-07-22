> Extracted from: admin-operations.md — Section: 1.7 Admin Queues

## 1.7 Admin Queues

- [ ] `GET /api/v1/admin/queues` — Queue stats (`AdminQueueHandlers.GetQueues` → `QueueRepo.GetQueueStats`)
- [ ] `GET /api/v1/admin/queues/stuck?threshold=30m` — Get stuck jobs (`AdminQueueHandlers.GetStuckJobs` → `QueueRepo.GetStuckJobs`)
  - Default threshold: 30 minutes
  - Configurable via `threshold` query param (duration string)
- [ ] `POST /api/v1/admin/queues/{id}/retry` — Retry stuck/failed job (`AdminQueueHandlers.RetryJob` → `SourceRepo.UpdateForRefresh`)
  - Logs action via `AdminService.LogAction`
- [ ] `DELETE /api/v1/admin/queues/{id}` — Delete job (`AdminQueueHandlers.DeleteJob` → `SourceRepo.Delete`)
  - Logs action via `AdminService.LogAction`

