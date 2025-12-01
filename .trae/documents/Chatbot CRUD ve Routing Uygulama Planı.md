## Genel Yaklaşım
- Mevcut mimariyi izleyerek `net/http` + `ServeMux` ile routing ve `database/sql` ile raw SQL kullanılır.
- Auth desenine uygun olarak tüm chatbot endpoint’leri `AuthMiddleware` ile korunur ve `user_id` context’ten okunur.
- Kod stili, mevcut handler’ların basit JSON parse/encode ve HTTP status kullanımını takip eder.

## Dosya/Dizin Yapısı
1. `internal/models/chatbot.go` — Chatbot veri modeli
2. `internal/db/chatbot.go` — Chatbot CRUD için DB yardımcıları
3. `internal/api/handlers/chatbot.go` — HTTP handler’lar (Create/List/Get/Update/Delete)
4. `cmd/server/main.go` — ServeMux üzerinde yeni route kayıtları

## Model
- `Chatbot` struct, şema ile uyumlu alanlar ve JSON tag’leri içerir:
  - `ID string`, `UserID string`, `Name string`, `Description *string`
  - `SystemPrompt string`, `Model string`, `Temperature float32`, `MaxTokens int`
  - `ThemeColor string`, `WelcomeMessage string`
  - `CreatedAt time.Time`, `UpdatedAt time.Time`, `DeletedAt *time.Time`
- Şema referansı: `db/migrations/0001_init.up.sql:25-39` (tablo sütunları, varsayılanlar ve `deleted_at`).

## DB Katmanı (raw SQL yardımcıları)
- `CreateChatbot(ctx, db, bot) (string, error)`
  - `INSERT INTO chatbots (...) VALUES (...) RETURNING id`
  - `deleted_at IS NULL` gerektirmez; oluşturma sırasında varsayılanlar kullanılır.
- `GetChatbotsByUserID(ctx, db, userID) ([]Chatbot, error)`
  - `SELECT ... FROM chatbots WHERE user_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`
- `GetChatbotByID(ctx, db, id) (*Chatbot, error)`
  - `SELECT ... WHERE id=$1 AND deleted_at IS NULL`
- `UpdateChatbot(ctx, db, bot) error`
  - Sadece sahiplik doğrulaması sonrası alanları günceller; `updated_at=NOW()`.
- `SoftDeleteChatbot(ctx, db, id, userID) error`
  - `UPDATE chatbots SET deleted_at=NOW() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`.
- Not: Yardımcılar `*sql.DB` parametresi alır; mevcut projede repository katmanı yok, ancak bu dosya handler’ları sadeleştirmek için kullanılır.

## HTTP Handler’lar
- `type ChatbotHandlers struct { DB *sql.DB }`
- `CreateChatbot(w, r)`
  - `POST /api/v1/chatbots`
  - `UserIDFromContext(r.Context())` ile `userID` alınır.
  - JSON body doğrulaması (name zorunlu; opsiyonel alanlar trim edilir).
  - DB’de oluşturur, `201` ve yeni `id` + chatbot JSON döndürür.
- `ListChatbots(w, r)`
  - `GET /api/v1/chatbots`
  - Context’ten `userID`; ilgili kullanıcının chatbot’larını `200` ile döndürür.
- `GetChatbot(w, r)`
  - `GET /api/v1/chatbots/:id`
  - URL’den `id` çıkarılır.
  - Kayıt bulunursa sahiplik doğrulanır (`user_id == context userID`), aksi halde `404`/`403`.
- `UpdateChatbot(w, r)`
  - `PUT /api/v1/chatbots/:id`
  - Sahiplik kontrolü sonrası güncelleme; `200` ile güncellenen kaydı döndürür.
- `DeleteChatbot(w, r)`
  - `DELETE /api/v1/chatbots/:id`
  - Sahiplik kontrolü ve soft delete; `204`.
- JSON cevaplarda `Content-Type: application/json` başlığı ve mevcut handler’larla tutarlı status kodları kullanılır.

## Routing
- `cmd/server/main.go` içinde `ChatbotHandlers` örneklenir ve route’lar eklenir.
- Tüm chatbot route’ları `AuthMiddleware(cfg.JWT_SECRET)` ile wrap edilir:
  - `POST   /api/v1/chatbots`
  - `GET    /api/v1/chatbots`
  - `GET    /api/v1/chatbots/:id`
  - `PUT    /api/v1/chatbots/:id`
  - `DELETE /api/v1/chatbots/:id`
- Referans: mevcut kurulum ve middleware uygulaması `cmd/server/main.go:34-43`, `pkg/middleware/auth.go:15-37`.

## Validasyon ve Güvenlik
- Sahiplik: `GetChatbotByID` sonucu `user_id != context userID` ise `403`.
- Soft delete: tüm `SELECT`’ler `deleted_at IS NULL` filtresi içerir.
- Input sanitizasyon: `strings.TrimSpace`, tip doğrulama ve minimal alan seti (örn. sıcaklık aralığı `0.0-1.0`).
- Hata durumlarında `500/400/401/403/404` tutarlı kullanım; body’de basit hata mesajı opsiyonel.

## Test Planı (Postman/cURL)
1. Auth: `/api/v1/auth/register` ve `/api/v1/auth/login` ile token al.
2. Create: `POST /api/v1/chatbots` — zorunlu alan `name`; opsiyonellerle dene.
3. List: `GET /api/v1/chatbots` — kullanıcıya ait kayıtlar gelmeli.
4. Get: `GET /api/v1/chatbots/:id` — sahip olmayan id için `403/404`.
5. Update: `PUT /api/v1/chatbots/:id` — alanlar güncellenmeli, `updated_at` değişmeli.
6. Delete: `DELETE /api/v1/chatbots/:id` — `204`; sonrasında list/get `404`/liste dışı.

## Kod Referansları
- Router/middleware akışı: `cmd/server/main.go:34-43`
- Auth middleware context: `pkg/middleware/auth.go:15-43`
- Handler stili ve JSON encode: `internal/api/handlers/auth.go:34-83`, `internal/api/handlers/auth.go:85-123`
- Şema: `db/migrations/0001_init.up.sql:25-39` (chatbots tablosu)

## Sonraki Adım
- Onayınızla birlikte dosyaları ekleyip handler ve routing’i uygulayacağım; ardından basit cURL komutlarıyla doğrulayıp çıktılarını paylaşacağım.