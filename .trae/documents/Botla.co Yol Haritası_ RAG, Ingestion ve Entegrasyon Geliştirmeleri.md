## Faz 1: Temel Ürün Uyumları (Must‑Have)

1. Sitemap İçe Alma ve Seçim UI
- Backend: `POST /chatbots/:id/sitemap/discover`; XML parse → URL listesi
- Frontend: Checkbox listesi ile seçim ve `POST /sources/bulk`
- Referanslar: `internal/processing/url_processor.go:107`, `internal/scraper/worker.go:89`

2. URL Checkbox Seçimi (Crawl Sonrası)
- Backend: “pending discovery” koleksiyonu; path filtreleri yoksa ham liste döndür
- Frontend: “Keşfedilen sayfalar” paneli + toplu ekleme

3. Path Tabanlı Include/Exclude
- Model: `chatbots` → `include_paths[]`, `exclude_paths[]`
- Scraper: `ExtractLinks` sonrası path filtresi

4. CSS Selector ile Bölge Seçimi
- Model: `selector_whitelist[]`
- Scraper: yalnız seçicilerden metin çıkar (`.content`, `#article`)

5. Auto‑Refresh ve Scheduler
- Worker: per‑chatbot refresh policy
- Opsiyonel cron (daily/weekly) → `sources.refresh`
- Kullanım sayacı: mevcut `refresh_count`’a ekle

6. Temperature ve Max Tokens UI
- `Chatbot` formunda slider/input
- API: mevcut alanların güncellenmesi
- Referans: `internal/models/chatbot.go:13`, `internal/rag/openai.go:172`

7. Branding Kaldırma (White‑Label)
- Model: `hide_branding` (plan bazlı)
- Widget: koşullu render (`ChatDrawer.tsx:122`)

## Faz 2: Entegrasyonlar ve Guardrails

1. OpenAI Function Calling + Actions Çerçevesi
- Chat akışına `tools/tool_calls` desteği
- “Custom Actions” registry (HTTP, Zapier webhook)

2. Zapier ve Calendly Entegrasyonları
- Zapier: outbound webhook; token yönetimi
- Calendly: randevu oluşturma endpoint’i

3. Operatör Handoff
- Crisp (SDK/API) üzerinden devralma
- Alternatif: basit yerleşik panel + e‑mail handoff

4. Model Esnekliği
- `Claude` ve `Gemini` istemcileri
- Plan kısıtları (`allowed_models`) ile UI’da seçim

5. Guardrails ve Güven Skoru UI
- `RAG_SCORE_THRESHOLD` ayarı yüzeylenmesi
- “Bilmiyorum” eşiği ve fallback cevap metinleri

## Faz 3: Ajans ve White‑Label Genişlemeleri (Nice‑to‑Have)

1. Çok Kiracılı Organizasyon Yapısı
- Tablolar: `organizations`, `memberships`, `workspaces`
- Faturalama: organizasyon bazlı plan

2. Custom Domain Routing
- `domains` tablosu ve doğrulama (CNAME/TXT)
- Reverse proxy konfigürasyonu (Caddy/Nginx) + host‑bazlı tenant yönlendirme

3. Gelişmiş Analytics
- Kaynak başına kalite/performans; CTR, memnuniyet, yanıtlanamayan oranlar

## Teknik Notlar
- Mevcut RAG akışı: `internal/services/chat_service.go:57` → embedding & Qdrant search → completion
- Kümeleme/embedding: `internal/rag/chunker.go:18`, `internal/rag/openai.go:64`, `internal/rag/qdrant.go:147`
- İnce ayarlar: `internal/rag/search.go:40` (score threshold), `:70` (TR token)

## Başarı Kriterleri
- Kullanıcı, crawl sonrası sayfa seçiminde tam kontrol
- Otomatik güncelleme ve yenileme politikaları görünür/denetlenebilir
- En az bir entegrasyon (Zapier/Calendly) ve bir handoff yolu aktif
- Brand kaldırma ve temel white‑label seçenekleri planlara bağlı
