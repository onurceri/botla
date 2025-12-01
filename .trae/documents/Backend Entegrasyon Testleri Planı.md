## Amaç ve Kapsam
- Tüm backend akışlarını uçtan uca doğrulamak: kayıt → giriş → korumalı erişim → chatbot CRUD → kaynak içe alma (pdf/url/metin) → kuyruk ve vektör oluşturma → status/delete → chat cevaplama.
- Güvenlik ve sahiplik kontrolleri, hata/kenar senaryoları ve dış servis entegrasyonu (Postgres, Qdrant, OpenAI) doğrulaması.
- Şu anda kodda olmayan özellikler (rate limiting, analytics) için uygulanabilir test ve geliştirme planı.

## Maliyet Yönetimi (OpenAI Embeddings)
- Testlerde küçük fixture’lar kullanıyoruz (kısa metinler, küçük PDF’ler); chunk sayısı az olduğundan embedding çağrıları sınırlı.
- Varsayılan model `text-embedding-3-small`; kısa metinlerde token sayısı düşük → maliyet çok küçük.
- İki modlu koşum:
  - Offline (varsayılan): OpenAI çağrıları devre dışı veya stub’a yönlendirilir; sıfır maliyet.
  - Online (isteğe bağlı): Sınırlı senaryoları gerçek API ile çalıştırır; test başına toplam embedding çağrısı < 25 olacak şekilde sınırlandırılır.
- Kuyruk testlerinde içerik uzunluğunu 1–2 chunk üretecek kadar küçük tutuyoruz; batch gerekmiyorsa çağrı sayısı minimal kalır.

## Sistem Envanteri (Hızlı Özet)
- Router: `cmd/server/main.go:36-66` (mux, middleware, queue başlatma)
- Auth: `internal/api/handlers/auth.go:34-123`, middleware `pkg/middleware/auth.go:15-37`
- Chatbot: `internal/api/handlers/chatbot.go:29-83`, `internal/api/handlers/chatbot.go:85-173`
- Sources: `internal/api/handlers/source.go:25-167`, `internal/api/handlers/source.go:169-225`
- Chat: `internal/api/handlers/chat.go:37-137`
- Health: `internal/api/handlers/health.go:19-40`
- DB: `internal/db/*` (chatbot, source, conversation)
- RAG: `internal/rag/*` (openai, qdrant, search, tokens)

## Test Ortamı
- Postgres ve Qdrant: docker-compose ile gerçek servisler veya Testcontainers (Go) ile ephemeral konteynerlar.
- OpenAI iki mod:
  - Offline: Stub server ile yanıt emülasyonu (OpenAI çağrıları atlanır veya sabit yanıt döner).
  - Online: Gerçek API key ile, kısa metin ve az chunk; toplam çağrı sayısı sınırlandırılır.
- Config: `.env` üzerinden `DB_*`, `QDRANT_URL`, `OPENAI_API_KEY`, `JWT_SECRET`, `PORT`. Test spesifik izolasyon.
- Migrasyonlar: eksik olduğundan test setup’ında şema oluşturma SQL’lerini çalıştırma (Users, Chatbots, DataSources, Conversations, Messages). Alternatif: `db/migrations` dizinine migration eklenmesi ve setup’ta çalıştırılması.

## Test Veri ve Yardımcılar
- Yardımcı HTTP client: Bearer token ekleyen, JSON parse eden yardımcı fonksiyonlar.
- Fixture’lar:
  - Küçük geçerli PDF (≤1MB) ve hatalı tip/limit dosyası.
  - Basit URL (ör. `https://example.com`) ve stub içerikleri.
  - Kısa serbest metin örnekleri (1–2 chunk üretir).
- Polling yardımcıları: Kaynak status’u `pending → processing → completed/failed` için bekleme.

## Test Vakaları (Uçtan Uca)
- Auth
  - `POST /api/v1/auth/register`: yeni kullanıcı 201, tekrar deneme 409.
  - `POST /api/v1/auth/login`: doğru şifre 200 + token; yanlış şifre 401.
  - `GET /api/v1/protected`: token ile 200; tokensız 401.
- Chatbot CRUD
  - `POST /api/v1/chatbots`: minimum alanlar ile 201, default’ların set edilmesi.
  - `GET /api/v1/chatbots`: kullanıcıya ait liste 200.
  - `GET /api/v1/chatbots/{id}`: sahiplik 200; farklı kullanıcı 403; yoksa 404.
  - `PUT /api/v1/chatbots/{id}`: alan güncelleme 200; invalid body 400.
  - `DELETE /api/v1/chatbots/{id}`: 204 ve sonrasında GET 404.
- Sources (İçe Aktarım)
  - `GET /api/v1/chatbots/{id}/sources`: boş liste 200.
  - `POST /api/v1/chatbots/{id}/sources`:
    - PDF: doğru dosya 201 → queue çalışır → status `completed` ve `chunk_count>0` (kısa PDF ile 1–2 chunk).
    - PDF: tip geçersiz veya >50MB → 400/413.
    - URL: geçerli `source_url` 201; boş URL 400.
    - Text: `text` alanı dolu 201; boş 400.
  - `GET /api/v1/sources/{source_id}`: 200 + status, sahiplik kontrolü 403.
  - `DELETE /api/v1/sources/{source_id}`: 204; Qdrant payload’larının silindiği doğrulama (best-effort).
- Kuyruk ve Vektörler
  - `internal/processing/sources_queue.go:worker`: `UpdateSourceProcessing` çağrıları doğru sıralama ve alan güncellemeleri.
  - Qdrant koleksiyonu hazır: `EnsureEmbeddingsCollection` en az bir kez çağrılmış.
  - Upsert payload doğruluğu: `chatbot_id`, `source_id`, `chunk_index`, `source_type`, `original_text` sahaları.
- Chat Akışı
  - `POST /api/v1/chatbots/{id}/chat`: geçerli `message` ve `session_id` ile 200; DB’de iki mesaj (user+assistant) ve `message_count` artışı.
  - Kaynak yoksa cevap fallback: “Yeterli bilgi bulamadım.”, `tokens_used=0`, `sources_used=[]`.
  - Kaynak varsa `sources_used` içeriği: `ChunkIndex` ve `SourceType` dolu, context formatı `SearchContext` ile uyumlu.
  - Sahiplik ve auth kontrolleri: farklı kullanıcı 403; tokensız 401.
- Health
  - `GET /health`: DB ve Qdrant durumu; biri down ise 503.

## Negatif ve Kenar Senaryoları
- Hatalı JSON ve method: `405/400`.
- Token süresi ve doğrulama: `pkg/middleware/auth.go` üzerinden 401.
- Büyük dosyalar ve bozuk PDF (MuPDF build tag yoksa `pdf: extractor unavailable` → status `failed`).
- Qdrant/OpenAI hataları: `chat.go` içinde graceful fallback; queue’da `failed` ve hata mesajı set edilmesi.

## Dış Servis Mocking Stratejisi
- Qdrant: `QDRANT_URL` test stub server’a işaret eder; `SearchSimilar`, `UpsertEmbedding`, `DeleteBySourceID` endpoint’lerini emüle eden `httptest.Server`.
- OpenAI:
  - Offline suite: çağrıları atlama veya sabit embedding/cevap üretim.
  - Online suite: kısa metinler ve az chunk ile sınırlı çağrı sayısı; bütçe güvenli.

## Rate Limiting ve Kotalar (Boşluk ve Plan)
- Mevcutta kullanıcı tipine göre limit/kota yok. Plan:
  - `users` veya ayrı `subscriptions` tablosunda `plan_type`, `monthly_message_limit`, `tokens_limit` alanları.
  - `pkg/middleware` içine `RateLimitMiddleware` (JWT userID bazlı sliding window veya token bucket; 429 döndürür).
  - Chat ve kaynak içe alma uç noktalarında middleware uygulaması.
  - Testler: Free plan için limit dolunca 429; üst planlarda daha yüksek eşikler.

## Teslimatlar
- `internal/integration/` altında Go test suite (httptest + gerçek Postgres/Qdrant ya da Testcontainers).
- Test yardımcıları ve fixture’lar.
- CI job: `docker-compose` ile servisleri kaldırıp suite’i çalıştıran script.
- (Opsiyonel) OpenAI offline/online ayrımı için küçük konfigürasyon ve rate limit middleware + testleri.

## Başlangıç Adımları
- Test şeması/migrasyonlarının hazırlanması.
- Postgres/Qdrant test ortamının ayağa kaldırılması.
- Auth ve Chatbot CRUD entegrasyon testlerinin yazılması.
- Sources ingest ve queue doğrulamaları (kısa içerik + az chunk).
- Chat akış testleri (kaynaklı/kaynaksız).
- Sağlık ve negatif senaryolar.
- Ardından rate limiting ve analytics ekleme + testleri.
