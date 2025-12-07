# Plan 1.3: CSS Selector ile Bölge Seçimi

## Özet

Web sayfası taranırken sadece belirli HTML elementlerinden (`.content`, `#article`, `main`) metin çıkarmak için CSS selector desteği.

---

## Mevcut Durum

### İlgili Dosyalar

| Dosya | Mevcut Durum |
|-------|--------------|
| `internal/scraper/worker.go` | `visibleText` tüm body'yi alıyor |
| `internal/models/chatbot.go` | Selector alanı yok |

### Mevcut visibleText Fonksiyonu

```go
// internal/scraper/worker.go:28-44
func visibleText(sel *goquery.Selection) string {
    sel.Find("script,style,noscript").Remove()
    // ... tüm body'den metin çıkarır
}
```

**Sorun:** Menü, footer, sidebar gibi gereksiz alanlar da taranıyor.

---

## Hedef Mimari

```
Kullanıcı Tanımlar:
┌────────────────────────────────────────┐
│ selector_whitelist: [".content",       │
│                      "#article-body",  │
│                      "main article"]   │
└────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────┐
│         Scraper Engine                 │
│  ─────────────────────────────────────│
│  1. Eğer selector varsa:               │
│     → Sadece o elementlerden al        │
│  2. Selector yoksa:                    │
│     → Mevcut davranış (tüm body)       │
└────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000010_selector_whitelist.up.sql`

```sql
ALTER TABLE chatbots 
ADD COLUMN IF NOT EXISTS selector_whitelist TEXT[] DEFAULT '{}';

COMMENT ON COLUMN chatbots.selector_whitelist IS 'CSS selectors for content extraction (empty = full body)';
```

**Dosya:** `db/migrations/000010_selector_whitelist.down.sql`

```sql
ALTER TABLE chatbots DROP COLUMN IF EXISTS selector_whitelist;
```

### Adım 2: Model Güncelleme

**Dosya:** `internal/models/chatbot.go`

**Eklenecek:**
```go
SelectorWhitelist []string `json:"selector_whitelist,omitempty"`
```

### Adım 3: Selector Extraction Fonksiyonu

**Dosya:** `internal/scraper/selector_extractor.go` (YENİ)

**Fonksiyonlar:**

```
func ExtractBySelectors(doc *goquery.Document, selectors []string) string
  - Her selector için doc.Find(selector) çağır
  - Bulunan elementlerden visibleText çıkar
  - Tümünü birleştir (duplicate önleme)
  - Boş ise fallback: tüm body

func ValidateSelector(selector string) error
  - CSS selector syntax kontrolü
  - Tehlikeli karakterleri engelle
```

### Adım 4: ScrapeURL Fonksiyonunu Güncelle

**Dosya:** `internal/scraper/worker.go`

**Değişiklik:** `ScrapeURL` ve `ScrapeURLWithFallback` signature güncelle

```go
type ScrapeConfig struct {
    Selectors []string
    // Gelecekte: PathFilter *PathFilter
}

func ScrapeURL(task ScrapingTask, cfg CollectorConfig, scrapeConfig *ScrapeConfig) (string, error)
```

**İç mantık:**
```go
c.OnHTML("body", func(e *colly.HTMLElement) {
    if scrapeConfig != nil && len(scrapeConfig.Selectors) > 0 {
        content = ExtractBySelectors(e.DOM, scrapeConfig.Selectors)
    } else {
        content = visibleText(e.DOM)
    }
})
```

### Adım 5: URLProcessor'ı Güncelle

**Dosya:** `internal/processing/url_processor.go`

**Değişiklik:** `Process` fonksiyonunda ScrapeConfig kullan

```go
scrapeConfig := &scraper.ScrapeConfig{
    Selectors: bot.SelectorWhitelist,
}
content, err := scraper.ScrapeURLWithFallback(task, cfg, allowDynamic, scrapeConfig)
```

### Adım 6: API Güncellemesi

**Dosya:** `internal/api/handlers/chatbot.go`

**Değişiklik:**
- `CreateChatbot` ve `UpdateChatbot` handler'larında `selector_whitelist` kabul et
- Selector validasyonu ekle

### Adım 7: Frontend Değişiklikleri

**Dosya:** `frontend/src/features/chatbot/ChatbotSettings.tsx`

**Eklenecek UI:**

```
┌─────────────────────────────────────────┐
│ 🎯 İçerik Seçici (CSS Selector)         │
├─────────────────────────────────────────┤
│ Sadece belirtilen HTML elementlerinden  │
│ metin çıkarılır.                        │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ .content                        [x] │ │
│ │ #article-body                   [x] │ │
│ │ main article                    [x] │ │
│ │ + Yeni selector ekle...             │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ 💡 İpucu: Tarayıcıda sağ tık → Öğeyi    │
│    Denetle ile selector bulabilirsiniz  │
│                                         │
│ ⚠️ Boş bırakılırsa tüm sayfa taranır    │
└─────────────────────────────────────────┘
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000010_*.sql` | YENİ | Migration |
| `internal/models/chatbot.go` | GÜNCELLE | selector_whitelist |
| `internal/scraper/selector_extractor.go` | YENİ | Extraction logic |
| `internal/scraper/selector_extractor_test.go` | YENİ | Testler |
| `internal/scraper/worker.go` | GÜNCELLE | ScrapeConfig |
| `internal/processing/url_processor.go` | GÜNCELLE | Config kullan |
| `internal/api/handlers/chatbot.go` | GÜNCELLE | API |
| `frontend/src/features/chatbot/*` | GÜNCELLE | UI |

---

## Test Planı

### Unit Testler

**Dosya:** `internal/scraper/selector_extractor_test.go` (YENİ)

```go
func TestExtractBySelectors(t *testing.T) {
    html := `
    <html>
    <body>
        <nav>Menu items</nav>
        <main>
            <article class="content">
                <p>Important content here</p>
            </article>
        </main>
        <footer>Footer info</footer>
    </body>
    </html>`
    
    testCases := []struct {
        name      string
        selectors []string
        contains  []string
        notContains []string
    }{
        {
            name:      "single selector",
            selectors: []string{".content"},
            contains:  []string{"Important content here"},
            notContains: []string{"Menu items", "Footer info"},
        },
        {
            name:      "multiple selectors",
            selectors: []string{"main", "footer"},
            contains:  []string{"Important content here", "Footer info"},
            notContains: []string{"Menu items"},
        },
    }
}
```

**Komut:** `go test ./internal/scraper/... -v -run TestExtractBySelectors`

### Integration Test

**Dosya:** `internal/scraper/worker_test.go`

```go
func TestScrapeURL_WithSelectors(t *testing.T) {
    // Test server kur
    // Selector ile scrape
    // Sonuçları doğrula
}
```

### Manuel Test Prosedürü

1. **Setup:**
   - https://example.com gibi basit bir site için test

2. **Selector Test:**
   - Chatbot oluştur
   - Selector: `main`, `.content` ekle
   - URL kaynağı ekle
   - Taranan içeriğin sadece selector'lara ait olduğunu doğrula

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Migration başarılı | `make migrate-up` |
| ✅ Selector çıkarımı çalışıyor | Unit test |
| ✅ Fallback (boş selector) | Unit test |
| ✅ API kabul ediyor | API test |
| ✅ Frontend UI | Manuel test |
| ✅ %90 coverage | `make cover-gate` |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration | 30 dk |
| Selector extractor | 2-3 saat |
| Worker update | 1-2 saat |
| Processor update | 1 saat |
| API + Frontend | 3-4 saat |
| Testler | 2-3 saat |
| **TOPLAM** | **~3-4 gün** |

---

## Bağımlılıklar

**Önceki:** Plan 1.2 (Path Filtering) tamamlanmış olmalı
**Sonraki:** Bu plan bağımsız, paralel çalışılabilir
