# Plan 1.5: URL Checkbox Seçimi UI

## Özet

Crawl veya sitemap sonrası keşfedilen URL'leri checkbox listesi ile kullanıcıya sunarak seçim yapmasını sağlama.

---

## Mevcut Durum

### Mevcut Akış

```
URL Ekleme → Otomatik Crawl → Tüm alt sayfalar eklenir
                                      ↑
                        Kullanıcı kontrolü YOK
```

### İlgili Dosyalar

| Dosya | Mevcut Durum |
|-------|--------------|
| `internal/processing/url_processor.go` | `discoverSubPages` otomatik ekler |
| Frontend source listesi | Sadece eklenmiş kaynakları gösterir |

---

## Hedef Mimari

```
┌────────────────────────────────────────┐
│ Yeni Akış:                             │
│                                        │
│ URL Ekle → "Keşfet" modu seç           │
│         ↓                              │
│ Backend: Crawl + Link extraction       │
│         ↓                              │
│ Frontend: Pending URLs listesi göster  │
│         ↓                              │
│ Kullanıcı: Checkbox ile seç            │
│         ↓                              │
│ Seçilenler kaynak olarak eklenir       │
└────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Şeması - Pending URLs

**Dosya:** `db/migrations/000011_pending_discovered_urls.up.sql`

```sql
CREATE TABLE IF NOT EXISTS pending_discovered_urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    source_id UUID REFERENCES data_sources(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    discovered_at TIMESTAMP DEFAULT NOW(),
    status TEXT DEFAULT 'pending', -- pending, selected, rejected
    
    UNIQUE(chatbot_id, url)
);

CREATE INDEX idx_pending_urls_chatbot ON pending_discovered_urls(chatbot_id, status);
```

**Down migration:**
```sql
DROP TABLE IF EXISTS pending_discovered_urls;
```

### Adım 2: SQLC Queries

**Dosya:** `db/queries/chatbots/pending_urls.sql` (YENİ)

```sql
-- name: InsertPendingURL :exec
INSERT INTO pending_discovered_urls (chatbot_id, source_id, url)
VALUES ($1, $2, $3)
ON CONFLICT (chatbot_id, url) DO NOTHING;

-- name: ListPendingURLs :many
SELECT * FROM pending_discovered_urls
WHERE chatbot_id = $1 AND status = 'pending'
ORDER BY discovered_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPendingURLs :one
SELECT COUNT(*) FROM pending_discovered_urls
WHERE chatbot_id = $1 AND status = 'pending';

-- name: UpdatePendingURLStatus :exec
UPDATE pending_discovered_urls
SET status = $3
WHERE chatbot_id = $1 AND id = ANY($2::uuid[]);

-- name: DeletePendingURLsByChatbot :exec
DELETE FROM pending_discovered_urls
WHERE chatbot_id = $1;
```

### Adım 3: URLProcessor Değişikliği - Auto vs Discovery Mode

**Dosya:** `internal/processing/url_processor.go`

**Mevcut `discoverSubPages` değişikliği:**

```go
type DiscoveryMode string

const (
    DiscoveryModeAuto     DiscoveryMode = "auto"     // Mevcut davranış
    DiscoveryModePending  DiscoveryMode = "pending"  // Pending tablosuna ekle
    DiscoveryModeDisabled DiscoveryMode = "disabled" // Hiç crawl yapma
)

func (p *URLProcessor) discoverSubPages(ctx context.Context, s *models.DataSource, bot *models.Chatbot, plan *models.Plan, content string, mode DiscoveryMode) {
    // ... link extraction
    
    switch mode {
    case DiscoveryModeAuto:
        // Mevcut davranış: doğrudan source oluştur
    case DiscoveryModePending:
        // YENİ: pending_discovered_urls tablosuna ekle
        for _, link := range links {
            db.InsertPendingURL(ctx, p.DB, bot.ID, s.ID, link)
        }
    case DiscoveryModeDisabled:
        // Hiçbir şey yapma
    }
}
```

### Adım 4: API - Pending URLs Endpoint'leri

**Dosya:** `internal/api/handlers/source.go`

**Yeni Endpoint'ler:**

```
GET /api/chatbots/:id/pending-urls
Response:
{
    "urls": [
        {"id": "uuid", "url": "https://...", "discovered_at": "..."},
        ...
    ],
    "total": 50,
    "page": 1,
    "per_page": 20
}

POST /api/chatbots/:id/pending-urls/approve
Request:
{
    "url_ids": ["uuid1", "uuid2", ...]
}
Response:
{
    "approved_count": 5,
    "sources_created": 5
}

POST /api/chatbots/:id/pending-urls/reject
Request:
{
    "url_ids": ["uuid1", "uuid2", ...]
}
Response:
{
    "rejected_count": 3
}

POST /api/chatbots/:id/pending-urls/clear
Response:
{
    "cleared_count": 42
}
```

### Adım 5: Chatbot Settings - Discovery Mode

**Dosya:** `internal/models/chatbot.go`

```go
DiscoveryMode string `json:"discovery_mode"` // auto, pending, disabled
```

**Migration:**
```sql
ALTER TABLE chatbots 
ADD COLUMN IF NOT EXISTS discovery_mode TEXT DEFAULT 'auto';
```

### Adım 6: Frontend - Pending URLs Panel

**Dosya:** `frontend/src/features/sources/PendingURLsPanel.tsx` (YENİ)

**UI Tasarımı:**

```
┌─────────────────────────────────────────────────────────────┐
│ 🔍 Keşfedilen URL'ler (42 adet)                      [↻ Yenile]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ [✓ Tümünü Seç] [✗ Temizle] 5 seçili                        │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ☑️ https://example.com/blog/post-1                     │ │
│ │ ☑️ https://example.com/blog/post-2                     │ │
│ │ ☐ https://example.com/about                            │ │
│ │ ☐ https://example.com/contact                          │ │
│ │ ☑️ https://example.com/docs/intro                      │ │
│ │ ...                                                     │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ◄ 1 2 3 4 5 ►                                              │
│                                                             │
│     [❌ Seçilenleri Reddet]   [✓ Seçilenleri Onayla]        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Dosya:** `frontend/src/features/chatbot/ChatbotSettings.tsx`

**Discovery Mode seçeneği ekle:**

```
┌─────────────────────────────────────────┐
│ 🔗 URL Keşif Modu                       │
├─────────────────────────────────────────┤
│ ○ Otomatik - Bulunan tüm URL'ler        │
│              otomatik eklenir           │
│ ● Onay Bekle - URL'ler size sunulur     │
│                ve onayınızı bekler      │
│ ○ Kapalı - Alt sayfa keşfi yapılmaz     │
└─────────────────────────────────────────┘
```

### Adım 7: Sources Tab'a Entegrasyon

**Dosya:** `frontend/src/pages/ChatbotSources.tsx`

**Tab yapısı:**

```
┌──────────────────────────────────────────────────────────────┐
│ [Kaynaklar] [Keşfedilen URL'ler (42)] [Sitemap İçe Aktar]    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│   Tab içeriği...                                             │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000011_*.sql` | YENİ | pending_discovered_urls tablosu |
| `db/queries/chatbots/pending_urls.sql` | YENİ | SQLC queries |
| `internal/models/chatbot.go` | GÜNCELLE | discovery_mode |
| `internal/processing/url_processor.go` | GÜNCELLE | Discovery mode switch |
| `internal/api/handlers/source.go` | GÜNCELLE | Pending URL endpoints |
| `internal/api/router.go` | GÜNCELLE | Routes |
| `frontend/src/features/sources/PendingURLsPanel.tsx` | YENİ | UI |
| `frontend/src/pages/ChatbotSources.tsx` | GÜNCELLE | Tab entegrasyonu |
| `frontend/src/api/source.ts` | GÜNCELLE | API calls |

---

## Test Planı

### Unit Testler

**Dosya:** `internal/processing/url_processor_test.go`

```go
func TestDiscoverSubPages_PendingMode(t *testing.T) {
    // Setup: chatbot with discovery_mode = "pending"
    // Process URL
    // Verify: links added to pending_discovered_urls, NOT to sources
}

func TestDiscoverSubPages_AutoMode(t *testing.T) {
    // Setup: chatbot with discovery_mode = "auto"
    // Process URL
    // Verify: links added directly to sources (mevcut davranış)
}
```

### API Testleri

**Dosya:** `internal/api/handlers/source_test.go`

```go
func TestListPendingURLs(t *testing.T) {
    // GET /chatbots/:id/pending-urls
}

func TestApprovePendingURLs(t *testing.T) {
    // POST /chatbots/:id/pending-urls/approve
    // Verify sources created
}

func TestRejectPendingURLs(t *testing.T) {
    // POST /chatbots/:id/pending-urls/reject
    // Verify status updated
}
```

### Manuel Test Prosedürü

1. **Discovery Mode Test:**
   - Chatbot → Ayarlar → URL Keşif Modu: "Onay Bekle" seç
   - Yeni URL kaynağı ekle
   - "Keşfedilen URL'ler" tab'ında URL'lerin göründüğünü doğrula
   - Birkaçını seç ve "Onayla"
   - "Kaynaklar" tab'ında göründüğünü doğrula

2. **Auto Mode Test:**
   - Discovery Mode: "Otomatik" yap
   - URL ekle
   - Alt sayfaların otomatik eklendiğini doğrula

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Migration başarılı | `make migrate-up` |
| ✅ Pending mode çalışıyor | Unit test |
| ✅ Auto mode geriye uyumlu | Unit test |
| ✅ API endpoints | API test |
| ✅ Frontend UI | Manuel test |
| ✅ %90 coverage | `make cover-gate` |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration + SQLC | 1-2 saat |
| URLProcessor update | 2-3 saat |
| API endpoints | 3-4 saat |
| Frontend UI | 6-8 saat |
| Settings entegrasyonu | 2-3 saat |
| Testler | 3-4 saat |
| **TOPLAM** | **~1 hafta** |

---

## Bağımlılıklar

**Önceki:** 
- Plan 1.2 (Path Filtering) - Filter uygulama
- Plan 1.4 (Sitemap Parser) - Bulk URL handling

**Sonraki:** Plan 1.6 (Auto-Refresh) bu altyapıyı kullanır
