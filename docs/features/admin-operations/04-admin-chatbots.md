> Extracted from: admin-operations.md — Section: 1.4 Admin Chatbots

## 1.4 Admin Chatbots

- [ ] `GET /api/v1/admin/chatbots?limit=&offset=&name=&organization_id=&owner_id=` — Paginated chatbot list (`AdminChatbotHandlers.ListChatbots` → `AdminChatbotRepo.ListChatbots`)
  - Default `limit=20`, `offset=0`
- [ ] `GET /api/v1/admin/chatbots/{id}` — Get chatbot detail (`AdminChatbotHandlers.GetChatbot` → `AdminChatbotRepo.GetByID`)
- [ ] `POST /api/v1/admin/chatbots/{id}/force-refresh` — Force re-process all sources (`AdminChatbotHandlers.ForceRefreshChatbot`)
  - Deletes vectors from Qdrant (`RagService.DeleteBotVectors`)
  - Resets chunk counts (`AdminChatbotRepo.DeleteVectors`)
  - Resets all sources to pending (`AdminChatbotRepo.ResetSources`)
  - Queues each source for re-processing (`Queue.Enqueue`)
  - Logs action via `AdminService.LogAction`

