## Kapsamlı Keşif Özeti
- Backend: Go `net/http` ile çok katmanlı mimari; JWT ile korunan REST uçları, RAG (OpenAI + Qdrant), dosya depolama (R2).
- Frontend: React + TypeScript + Vite; `react-router-dom` ile yönlendirme; `@tanstack/react-query` ile sunucu durumu; Tailwind tabanlı UI kit.

## Üreteceğimiz Dokümantasyon Yapısı (English)
- Klasör: `docs/features/`
- Dosyalar:
  - `authentication.md`
  - `chatbot-management.md`
  - `sources-ingestion.md` (PDF/URL/Text)
  - `chat-and-feedback.md`
  - `analytics.md`
  - `settings.md`
  - `dashboard.md`
  - `architecture-overview.md` (üst düzey özet)
- Her dosyada bölümler:
  - Purpose & Scope
  - User Flow (step-by-step)
  - Backend Interfaces
    - Endpoints & Handler mapping (file_path:line_number referansları)
    - DB models & queries
    - Middleware & auth etkisi
    - External integrations (OpenAI/Qdrant/R2)
  - Frontend Interfaces
    - Routes/pages/components (file_path:line_number referansları)
    - API client calls & React Query kullanımı
    - State, forms, UI elementleri
  - Error handling & edge cases
  - Configuration & environment
  - Testing strategy & coverage hedefleri

## Feature Envanteri ve Kaynaklar
- Authentication
  - Backend: `cmd/server/main.go:44-49`, `internal/api/handlers/auth.go:39-200`, `pkg/middleware/auth.go:15-37`, `internal/auth/jwt.go:15-45`
  - Frontend: `frontend/src/App.tsx:11-18`, `frontend/src/hooks/useAuth.ts:4-45`, `frontend/src/api/client.ts:8-46`, `frontend/src/pages/LoginPage.tsx`, `frontend/src/pages/RegisterPage.tsx`
- Chatbot Management
  - Backend: `cmd/server/main.go:50-78`, `internal/api/handlers/chatbot.go:39-227`, `internal/db/chatbot.go`
  - Frontend: `frontend/src/pages/ChatbotsPage.tsx`, `frontend/src/pages/ChatbotDetailPage.tsx`, `frontend/src/components/chatbot/*`, `frontend/src/api/chatbot.ts`
- Sources Ingestion (PDF/URL/Text)
  - Backend: `cmd/server/main.go:62-85`, `internal/api/handlers/source.go:23-163,165-227`, `internal/processing/*`, `pkg/storage/*`
  - Frontend: `frontend/src/components/chatbot/SourceUploader.tsx`, `frontend/src/api/source.ts`
- Chat & Feedback
  - Backend: `cmd/server/main.go:65-82`, `internal/api/handlers/chat.go:37-154,160-194`, `internal/rag/*`
  - Frontend: `frontend/src/components/chatbot/*` (chat UI), `frontend/src/api/chatbot.ts`
- Analytics
  - Backend: `cmd/server/main.go:86-88`, `internal/api/handlers/analytics.go:11-77`, `internal/db/analytics.go`
  - Frontend: `frontend/src/pages/AnalyticsPage.tsx`, `frontend/src/api/analytics.ts`, `recharts`
- Settings & Dashboard
  - Frontend: `frontend/src/pages/SettingsPage.tsx`, `frontend/src/pages/DashboardPage.tsx`, `frontend/src/components/layout/*`

## Dokümantasyon İçerik Örnek Şablonları
- `authentication.md`
  - Flow: Register → Login → Token storage → Protected ping → Refresh → Logout
  - Backend: endpoint tabloları ve handler akışları; DB `users`, `refresh_tokens` ilişkisi
  - Frontend: `useAuth`, axios interceptors, route guard
  - Edge cases: 401, token rotation, revoke, UI feedback
  - Tests: unit (JWT/password), integration (auth flow), frontend (Login form)
- `chatbot-management.md`
  - Flow: List/Create → Detail → Update → Soft-delete
  - Backend: `ListOrCreate`, `ByID` ayrıntıları; yetki kontrolleri
  - Frontend: list grid, detail form, API çağrıları
  - Edge cases: bad input, forbidden, not found
  - Tests: integration + frontend form validation
- `sources-ingestion.md`
  - Flow: Sources list → Add (PDF/URL/Text) → Queue → Status → Delete
  - Backend: multipart parsing, storage upload, queue enqueue
  - Frontend: uploader bileşenleri, progress & hata gösterimi
  - Limits: 50MB PDF; content-type doğrulama
  - Tests: PDF/URL/Text ingestion ve silme
- `chat-and-feedback.md`
  - Flow: Session → User msg → Embedding → Vector search → Completion → Analytics → Feedback
  - Backend: timeoutlar, fallback (Qdrant yoksa), background analytics
  - Frontend: chat UI, thumbs up/down
  - Security: feedback endpoint auth durumu
- `analytics.md`
  - Flow: 7-day series → aggregate by user chatbots
  - Backend: SQL `generate_series` + COALESCE
  - Frontend: charts, loading states

## Öncelikli Eksiklik ve İyileştirmeler
- Frontend `PrivateRoute` anahtarı `token` ile kontrol ediyor; sistem `botla_token` kullanıyor (uyumsuzluk).
- `POST /api/v1/messages/{id}/feedback` şu an korumasız; auth sarmalına alınmalı.
- Sources işlemlerinde hata geri bildirimlerini UI’da standart hale getirme.
- Analytics sayfasında loading/empty durumu ve hata handling konsolidasyonu.

## Test Stratejisi
- Backend: mevcut integration testleri çalıştırma ve genişletme (auth edges, sources, chat, rate-limit).
- Frontend: `vitest` + Testing Library ile kapsamlı UI ve API etkileşim testleri; route guard ve token refresh senaryoları.
- E2E: Cypress önerisi (login, chatbot create, source upload, chat, analytics smoke).

## Uygulama Adımları
1. Feature dokümanlarını `docs/features/*` altında İngilizce oluşturma (şablonlara göre).
2. Belgeleri referans alarak frontend uyumsuzluk/eksikleri düzeltme (token anahtarı, korumalı feedback, UI durumları).
3. Backend’de gerekli güvenlik iyileştirmeleri (feedback endpoint).
4. Test yazımı ve çalıştırma: backend (Go) + frontend (Vitest), opsiyonel E2E.
5. Son doğrulama: lint, type-check, dev/preview smoke test, edge-case senaryolar.

Onayınızla birlikte dokümanları üretmeye ve akabinde düzeltme/test aşamalarına başlayacağım.