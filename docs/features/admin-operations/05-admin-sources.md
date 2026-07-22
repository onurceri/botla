> Extracted from: admin-operations.md — Section: 1.5 Admin Sources

## 1.5 Admin Sources

- [ ] `GET /api/v1/admin/sources?limit=&offset=&chatbot_id=&source_type=&status=&owner_id=` — Paginated source list (`AdminSourceHandlers.ListSources` → `AdminRepo.AdminListSources`)
  - Default `limit=20`, `offset=0`
- [ ] `GET /api/v1/admin/sources/stats` — Source statistics (`AdminSourceHandlers.GetSourceStats` → `AdminRepo.AdminGetSourceStats`)
- [ ] `GET /api/v1/admin/sources/types` — Available source types + statuses (`AdminSourceHandlers.GetSourceTypes`)
  - Types: `url`, `file`, `pdf`, `text`
  - Statuses: `pending`, `processing`, `ready`, `failed`
- [ ] `GET /api/v1/admin/sources/{id}` — Get source detail (`AdminSourceHandlers.GetSource` → `AdminRepo.AdminGetSourceByID`)
- [ ] `POST /api/v1/admin/sources/{id}/reprocess` — Reprocess a source (`AdminSourceHandlers.ReprocessSource`)
  - Deletes existing Qdrant vectors (`RagService.DeleteSourceVectors`)
  - Resets source to pending (`AdminRepo.AdminReprocessSource`)
  - Queues for processing (`Queue.Enqueue`)
  - Logs action via `AdminService.LogAction`

