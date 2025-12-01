## Hedefler
- Tüm özelliklerin eksiksiz çalıştığını, hatasız olduğunu ve geri dönüşlerde korunmuş olduğunu ispatlayan test paketi.
- Kritik akışlarda (auth, kaynak ingest, chat, feedback, analytics) yüksek güvence ve geriye dönük bozulma koruması.
- CI’de otomatik çalışan kalite kapıları (coverage eşikleri, lint/type-check, entegrasyon testleri ve smoke).

## Test Stratejisi: Backend
- Birim testleri (mevcutların yanına ek):
  - JWT üretim/doğrulama sınır durumları: süresi dolmuş token, yanlış `TokenType` (internal/auth/jwt.go:15-45).
  - AuthHandlers: refresh rotasyonu ve revoke akışı; invalid/expired `refresh_token` senaryoları (internal/api/handlers/auth.go:118-155).
  - Middleware Auth: eksik/bozuk `Authorization` header, yanlış `Bearer` formatı (pkg/middleware/auth.go:15-37).
  - RAG yardımcıları: chunker/tokens sınırları (mevcut testleri genişletme).
- Entegrasyon testleri (mevcutları güçlendirme):
  - Feedback endpoint koruma: yetkisiz → 401, yetkili → 200 (cmd/server/main.go:80-85; internal/api/handlers/chat.go:160-194).
  - Sources ingest: PDF/URL/Text için başarı ve hata durumları; storage unavailable → 503 doğrulaması (internal/api/handlers/source.go:83-147,102-117,118-147).
  - Chat akışı: Qdrant yoksa fallback metni (“Yeterli bilgi bulamadım.”), Qdrant var iken kaynak meta dönmesi (internal/api/handlers/chat.go:109-115,118-129).
  - Analytics: 7 günlük seride COALESCE ile doğru toplama; boş veri (internal/api/handlers/analytics.go:21-77).
- Ortam/mocking:
  - OpenAI HTTP client stub’u: yanıt ve hata senaryoları (internal/rag/openai.go).
  - Qdrant client stub’u: koleksiyon yok → oluştur, arama hataları (internal/rag/qdrant.go).
  - R2 storage fake: Upload/Download/Delete in-memory taklit; dev/test env’de 503’lerin kalkması.
- Test DB ve veri yaşam döngüsü:
  - Ephemeral Postgres (Docker) veya `testcontainers-go` ile otomatik DB; migrasyonların test öncesi uygulanması.
  - Her testte izole kullanıcı/Chatbot/Source kaydı ve temizleme.

## Test Stratejisi: Frontend
- Birim/etkileşim testleri:
  - Route guard: `PrivateRoute` botla_token yokken login’e yönlendirme, varken dashboard (frontend/src/App.tsx:11-18,30-39).
  - Auth akışı: LoginPage form gönderimi, axios post stub, token/refresh set; refresh interceptor 401 → refresh → retry (frontend/src/api/client.ts:8-46; pages/LoginPage.tsx:24-33).
  - SourcesUploader: PDF/URL/Text modları; başarılı ve hatalı yüklemelerde toast mesajları (frontend/src/components/chatbot/SourceUploader.tsx:21-35,37-56).
  - Analytics UI: loading/empty/error durumları ve gerçek veri ile grafik render (frontend/src/pages/AnalyticsPage.tsx).
  - Chat UI: mesaj gönderme, thumbs feedback butonları; backend stub ile durumlar.
- Görsel/durum testleri:
  - Skeleton/empty states: Analytics ve listeler.
  - Hata banner/toast tutarlılığı.
- E2E (Cypress önerisi):
  - Login → Chatbot oluştur → Source ekle → Chat → Feedback → Analytics görüntüle smoke akışları.

## Eksiklerin/Sorunların Tespiti ve Giderim Planı
- Token anahtar tutarlılığı:
  - Tüm frontend auth kontrollerinin `botla_token` ile çalıştığını doğrulayan tarama ve testler.
  - `useAuth` ve route guard uyumu; logout’ta temizleme.
- Feedback endpoint güvenliği:
  - Koruma eklendi; entegrasyon testleri ile doğrulama; UI’de yetkisiz durum handling.
- Sources hata yönetimi:
  - Storage unavailable, PDF boyut/simge uyumsuzluğu durumlarının standart mesajlaştırması ve frontend toast.
- Analytics veri akışı:
  - API entegrasyonu, cache ve yeniden deneme politikası; tarihe göre filtre doğruluk testi.
- Chat akışı esneklikleri:
  - OpenAI hatalarında kullanıcı dostu mesaj; Qdrant yokken fallback doğrulama; token/timeouts.
- Ortam bağımlılıkları:
  - R2/Qdrant/OpenAI için test doubles; env’ler yokken entegrasyon testlerinde deterministik sonuçlar.

## Ortam ve Konfigürasyon
- `.env.test` önerisi (frontend/backend) ile base URL ve anahtarların stub değerleri.
- Backend test başlatıcısı: in-memory/fake hizmetlerle handler’ların bağlı çalıştırılması.
- Frontend `vitest` yapılandırması: jsdom, globals, `setupTests` içinde localStorage ve ToastProvider yardımcıları.

## Coverage ve Kalite Kapıları
- Frontend coverage hedefi: statements/branches ≥ 80% (kritik akışlar ≥ 90%).
- Backend coverage hedefi: handlers+middleware ≥ 80%, auth/rag yardımcıları ≥ 90%.
- CI adımları: `lint`, `type-check`, `test`, `build`; `go test ./... -count=1` ve `npm test -s`.

## Risk ve Edge Cases
- Token rotation yarış durumları (eşzamanlı refresh) ve `_retry` bayrağı.
- Büyük PDF (50MB+) ve PDF dışı içerik.
- Chat timeout’ları ve yeniden deneme politikaları.
- Analytics generate_series ile timezone etkileri.

## Teslimatlar
- Test doubles ve yardımcılar (OpenAI, Qdrant, R2).
- Genişletilmiş backend entegrasyon testleri (feedback, sources, chat, analytics).
- Frontend birim/E2E testleri (auth, sources, analytics, chat).
- CI/coverage raporları ve pre-prod doğrulama kontrol listesi.

## Uygulama Adımları
1. Test doubles altyapısını oluşturma (backend clients ve storage için stub’lar).
2. Backend entegrasyon testlerini env bağımlılıklarından arındırıp yeşile çekme.
3. Frontend test paketlerini genişletme ve coverage hedeflerine ulaşma.
4. Cypress ile e2e smoke akışlarını kurma.
5. CI pipeline ve kalite kapılarını devreye alma; pre-prod checklist ile son doğrulama.

Onayla birlikte bu plana göre test altyapısını ve eksik giderimlerini uygulamaya başlayacağım.