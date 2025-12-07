# Plan 1.4: Sitemap İçe Alma ve Parse Etme

## Özet

XML sitemap dosyasını okuyarak URL listesi çıkarma ve kullanıcıya seçim sunma.

---

## Mevcut Durum

Sitemap desteği **yok**. Kullanıcılar sadece:
- Tek URL ekleyebilir
- Crawler ile alt sayfaları keşfedebilir

---

## Hedef Mimari

```
Kullanıcı Akışı:
┌────────────────────────────────────────┐
│ 1. Sitemap URL gir                     │
│    https://example.com/sitemap.xml     │
└────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────┐
│ 2. Backend: XML Parse                  │
│    - Standard sitemap format           │
│    - Sitemap index desteği             │
│    - Lastmod, priority okuma           │
└────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────┐
│ 3. Frontend: URL Listesi Göster        │
│    - Checkbox ile seçim                │
│    - Lastmod'a göre sıralama           │
│    - Toplu seçim butonları             │
└────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────┐
│ 4. Seçilen URL'leri Kaynak Olarak Ekle │
│    POST /sources/bulk                  │
└────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Sitemap Parser Modülü

**Dosya:** `internal/scraper/sitemap_parser.go` (YENİ)

**Struct'lar:**

```go
type SitemapURL struct {
    Loc        string     `xml:"loc" json:"loc"`
    LastMod    *time.Time `xml:"lastmod" json:"lastmod,omitempty"`
    ChangeFreq string     `xml:"changefreq" json:"changefreq,omitempty"`
    Priority   float64    `xml:"priority" json:"priority,omitempty"`
}

type SitemapIndex struct {
    Sitemaps []struct {
        Loc     string `xml:"loc"`
        LastMod string `xml:"lastmod,omitempty"`
    } `xml:"sitemap"`
}

type URLSet struct {
    URLs []SitemapURL `xml:"url"`
}
```

**Fonksiyonlar:**

```go
// ParseSitemap URL'den sitemap'i parse eder
// Sitemap index ise recursive olarak tüm sitemapleri okur
func ParseSitemap(ctx context.Context, sitemapURL string) ([]SitemapURL, error)

// isSitemapIndex sitemap index mi kontrol eder
func isSitemapIndex(xmlContent []byte) bool

// parseSitemapXML tek bir sitemap XML'i parse eder
func parseSitemapXML(xmlContent []byte) ([]SitemapURL, error)
```

### Adım 2: API Endpoint - Sitemap Discover

**Dosya:** `internal/api/handlers/source.go`

**Yeni Endpoint:**

```
POST /api/chatbots/:id/sitemap/discover

Request:
{
    "sitemap_url": "https://example.com/sitemap.xml"
}

Response:
{
    "urls": [
        {
            "loc": "https://example.com/page1",
            "lastmod": "2024-01-15T10:30:00Z",
            "priority": 0.8
        },
        ...
    ],
    "total_count": 150
}
```

**Handler:**

```go
func (h *SourceHandlers) DiscoverSitemap(c *gin.Context) {
    // 1. Sitemap URL'i al
    // 2. ParseSitemap çağır
    // 3. Path filter uygula (varsa)
    // 4. URL listesini döndür
}
```

### Adım 3: Bulk Source Creation

**Dosya:** `internal/api/handlers/source.go`

**Mevcut endpoint'i güncelle veya yeni ekle:**

```
POST /api/chatbots/:id/sources/bulk

Request:
{
    "urls": [
        "https://example.com/page1",
        "https://example.com/page2"
    ]
}

Response:
{
    "created_count": 2,
    "skipped_count": 0,
    "errors": []
}
```

### Adım 4: Frontend - Sitemap Tab

**Dosya:** `frontend/src/features/sources/SitemapImport.tsx` (YENİ)

**UI Tasarımı:**

```
┌─────────────────────────────────────────────────────────────┐
│ 🗺️ Sitemap'ten İçe Aktar                                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Sitemap URL'si:                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ https://example.com/sitemap.xml                     [🔍]│ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ─────────────────────────────────────────────────────────── │
│                                                             │
│ ℹ️ 150 URL bulundu                                          │
│                                                             │
│ [✓ Tümünü Seç] [✗ Hiçbirini Seçme] [⟳ Son 30 Günü Seç]     │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ☑️ /blog/post-1          2024-12-01   0.8            │ │
│ │ ☑️ /blog/post-2          2024-11-28   0.8            │ │
│ │ ☐ /about                 2024-01-15   0.5            │ │
│ │ ☐ /contact               2024-01-15   0.5            │ │
│ │ ... (pagination)                                        │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│                    [24 URL Seçildi - Ekle]                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Bileşenler:**
- `SitemapInput`: URL girişi ve discover tetikleme
- `URLSelectionList`: Checkbox listesi
- `BulkActions`: Toplu seçim butonları

### Adım 5: Path Filter Entegrasyonu

**Dosya:** `internal/api/handlers/source.go`

**Değişiklik:**

```go
func (h *SourceHandlers) DiscoverSitemap(c *gin.Context) {
    // ... sitemap parse
    
    // Chatbot'un path filterlarını uygula
    filter, _ := scraper.NewPathFilter(
        chatbot.IncludePaths, 
        chatbot.ExcludePaths,
    )
    
    var filtered []SitemapURL
    for _, url := range urls {
        parsed, _ := url.Parse(url.Loc)
        if filter.Match(parsed.Path) {
            filtered = append(filtered, url)
        }
    }
    
    // filtered döndür
}
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `internal/scraper/sitemap_parser.go` | YENİ | Parser modülü |
| `internal/scraper/sitemap_parser_test.go` | YENİ | Testler |
| `internal/api/handlers/source.go` | GÜNCELLE | Yeni endpoint'ler |
| `internal/api/router.go` | GÜNCELLE | Route ekleme |
| `frontend/src/features/sources/SitemapImport.tsx` | YENİ | UI bileşeni |
| `frontend/src/api/source.ts` | GÜNCELLE | API calls |

---

## Test Planı

### Unit Testler

**Dosya:** `internal/scraper/sitemap_parser_test.go` (YENİ)

```go
func TestParseSitemap_StandardFormat(t *testing.T) {
    xml := `<?xml version="1.0"?>
    <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
        <url>
            <loc>https://example.com/page1</loc>
            <lastmod>2024-01-15</lastmod>
        </url>
    </urlset>`
    
    // Test standard sitemap
}

func TestParseSitemap_IndexFormat(t *testing.T) {
    // Test sitemap index
}

func TestParseSitemap_InvalidXML(t *testing.T) {
    // Test error handling
}
```

**Komut:** `go test ./internal/scraper/... -v -run TestParseSitemap`

### API Testleri

**Dosya:** `internal/api/handlers/source_test.go`

```go
func TestDiscoverSitemap(t *testing.T) {
    // Mock sitemap server
    // POST /chatbots/:id/sitemap/discover
    // Verify response
}

func TestBulkCreateSources(t *testing.T) {
    // POST /chatbots/:id/sources/bulk
    // Verify sources created
}
```

### Manuel Test Prosedürü

1. **Gerçek Sitemap Test:**
   - Chatbot oluştur
   - Kaynaklar → Sitemap'ten İçe Aktar
   - URL gir: `https://www.example.com/sitemap.xml`
   - URL listesinin görüntülendiğini doğrula
   - Birkaç URL seç ve ekle
   - Kaynaklar listesinde göründüğünü doğrula

2. **Large Sitemap Test:**
   - 1000+ URL içeren sitemap ile test
   - Pagination çalışıyor mu?
   - Performans kabul edilebilir mi?

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Standard sitemap parse | Unit test |
| ✅ Sitemap index parse | Unit test |
| ✅ API endpoint çalışıyor | API test |
| ✅ Path filter entegrasyonu | API test |
| ✅ Frontend UI | Manuel test |
| ✅ %90 coverage | `make cover-gate` |

---

## Edge Cases

| Durum | Beklenen Davranış |
|-------|-------------------|
| Geçersiz XML | Hata mesajı göster |
| Erişilemeyen URL | Hata mesajı göster |
| Boş sitemap | "Hiç URL bulunamadı" mesajı |
| Çok büyük sitemap | Pagination + server-side filtering |
| Sitemap index | Recursive parse, tüm URL'leri birleştir |
| Duplicate URL | Tekrarları filtrele |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Sitemap parser | 2-3 saat |
| API endpoints | 2-3 saat |
| Path filter entegrasyonu | 1-2 saat |
| Frontend UI | 4-6 saat |
| Testler | 2-3 saat |
| **TOPLAM** | **~3-4 gün** |

---

## Bağımlılıklar

**Önceki:** Plan 1.2 (Path Filtering) tamamlanmış olmalı  
**Sonraki:** Plan 1.5 (URL Checkbox UI) bu altyapıyı kullanacak
